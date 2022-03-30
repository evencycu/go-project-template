package redispool

import (
	"fmt"
	"strconv"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/patrickmn/go-cache"
	"gitlab.com/cake/goctx"
	"gitlab.com/cake/m800log"
)

const (
	// prefix key + component name + custom key(from input)
	prefixKeyPatternRateLimiter = "rl-%s-%s"
	errMsgEmptyQuotaRedisBucket = "blocking: redis bucket not enough quota by key: %s"
)

type RateLimiter struct {
	LocalPool *cache.Cache
	RedisPool *Pool

	RedisCleanTimeout time.Duration
	// everytime pod will get how many quota from redis MaxQuotaSize
	BatchQuotaSize int64
	// for redis, when redis Incr counter exceeds this value, means not have any quota.
	MaxQuotaSize               int64
	singleFlightGroup          singleflight.Group
	SingleFlightExpireDuration time.Duration
}

func NewRateLimiter(batchQuota, maxQuota int64, redisCleanTimeout time.Duration, redisPool *Pool) (limiter *RateLimiter, err error) {
	localCachePool := cache.New(redisCleanTimeout, redisCleanTimeout)
	limiter = &RateLimiter{
		LocalPool:                  localCachePool,
		RedisPool:                  redisPool,
		RedisCleanTimeout:          redisCleanTimeout,
		BatchQuotaSize:             batchQuota,
		MaxQuotaSize:               maxQuota,
		singleFlightGroup:          singleflight.Group{},
		SingleFlightExpireDuration: time.Millisecond * 100,
	}
	return
}

func GetCacheKeyName(componentName, customKey string) string {
	return fmt.Sprintf(prefixKeyPatternRateLimiter, componentName, customKey)
}

func (r *RateLimiter) getCurrentLocalQuota(key string) (batchQuota int64) {
	value, exists := r.LocalPool.Get(key)
	if exists {
		batchQuota = value.(int64)
	} else {
		batchQuota = 0
	}

	return
}

func (r *RateLimiter) Decrement(ctx goctx.Context, componentName, customKey string, usedCount int64) (err error) {
	funcName := "RateLimiter.Decrement"
	key := GetCacheKeyName(componentName, customKey)

	// case:
	// 1. when local quota is empty, local quota will be got new batch quota from redis
	// 2. when redis bucket is empty, local quota will be -1
	err = r.updateLocalQuota(ctx, key)
	if err != nil {
		m800log.Tracef(ctx, "[%s] failed to sf get batch quota from bucket error: %v", funcName, err)
		return
	}

	// case: avoid when current local quota not enough(including: -1 case) for used count. will be blocking
	if currentQuota := r.getCurrentLocalQuota(key); usedCount > currentQuota {
		err = fmt.Errorf("blocking: local cache not enough quota of this round batch by key: %s", key)
		m800log.Tracef(ctx, "[%s] now is blocking by key: %s, usedCount: %d, batchQuota: %d, error: %v", funcName, key, usedCount, currentQuota, err)
		return
	}

	remainLocalQuota, err := r.LocalPool.DecrementInt64(key, usedCount)
	if err != nil {
		m800log.Errorf(ctx, "[%s] failed to decrement quota from local cache by key: %s, error: %v", funcName, key, err)
		return
	}

	// avoid local quota to overuse (-n)
	if remainLocalQuota < 0 {
		err = fmt.Errorf("blocking: local cache and redis both not enough quota by key: %s", key)
		m800log.Tracef(ctx, "[%s] now is blocking by key: %s, usedCount: %d, batchQuota: %d, error: %v", funcName, key, usedCount, remainLocalQuota, err)
		return
	}

	m800log.Tracef(ctx, "[%s] result remain local quota left: %d, key: %s", funcName, remainLocalQuota, key)
	return
}

func (r *RateLimiter) getBatchQuotaFromRedisBucket(ctx goctx.Context, key string) (nextBatchQuota int64, err error) {
	funcName := "RateLimiter.getBatchQuotaFromRedisBucket"
	batchQuotaSize := int(r.BatchQuotaSize)
	maxQuotaSize := int(r.MaxQuotaSize)

	usedQuota, errIncrBy := r.RedisPool.IncrBy(key, batchQuotaSize)
	if errIncrBy != nil {
		err = fmt.Errorf("failed to redis pool incr value by key: %s, error: %v", key, errIncrBy)
		return
	}
	m800log.Tracef(ctx, "[%s] current used quota: %d", funcName, usedQuota)

	// first time case
	if usedQuota == batchQuotaSize {
		timeoutSec := int(r.RedisCleanTimeout.Seconds())
		if errExpire := r.RedisPool.Expire(key, strconv.Itoa(timeoutSec)); errExpire != nil {
			err = fmt.Errorf("failed to redis pool set expire timeout sec by key: %s, timeout: %d, error: %v", key, timeoutSec, errExpire)
			return
		}
	}

	if usedQuota > maxQuotaSize {
		// latest batch case
		remainQuota := maxQuotaSize - (usedQuota - batchQuotaSize)
		if remainQuota > 0 {
			// handle odd number case, count the last left quota
			nextBatchQuota = int64(remainQuota)
			// record prometheus metric: quota bucket done by key
			return
		}
		// blocking case: redis quota is empty
		nextBatchQuota = -1
		err = fmt.Errorf(errMsgEmptyQuotaRedisBucket, key)
		return
	}

	// enough redis quota and not last quota case
	nextBatchQuota = r.BatchQuotaSize
	return
}

func (r *RateLimiter) updateLocalQuota(ctx goctx.Context, key string) (err error) {
	funcName := "RateLimiter.updateLocalQuota"

	// handle single flight worker stuck case, when timeout will forget sfKey. next one to do job again
	forgetSingleFlightGroupKey := func(sfKey string) {
		time.Sleep(r.SingleFlightExpireDuration)
		r.singleFlightGroup.Forget(sfKey)
	}
	// single flight: first one to do the job, others will be blocking and wait.
	// when the first has been completed, others will got same result too
	_, err, _ = r.singleFlightGroup.Do(key, func() (output interface{}, errSF error) {
		// avoid have others goroutines getting quotas repeatedly
		if currentQuota := r.getCurrentLocalQuota(key); currentQuota != 0 {
			return
		}
		go forgetSingleFlightGroupKey(key)

		// when this round batch quota is done will supplement this pod new batch quota from redis
		newQuota, errGet := r.getBatchQuotaFromRedisBucket(ctx, key)
		if errGet != nil && newQuota != -1 {
			errSF = errGet
			m800log.Errorf(ctx, "[%s] failed to get batch quota from bucket error: %v", funcName, errSF)
			return
		}

		// edge case: avoid next single flight worker got redis quota = -1 will overwrites not exhausted local quota.
		if newQuota == -1 {
			currentQuota := r.getCurrentLocalQuota(key)
			if currentQuota > 0 {
				m800log.Tracef(ctx, "[%s] get batch quota from bucket error: local quota is not empty", funcName)
				errSF = nil
				return
			}
		}

		redisExpireTime, errTTL := r.getRedisKeyExpiredTime(key)
		if errTTL != nil {
			errSF = errTTL
			m800log.Errorf(ctx, "[%s] failed to get redis key expired time error: %v", funcName, errSF)
			return
		}
		output = newQuota
		r.LocalPool.Set(key, newQuota, redisExpireTime)
		m800log.Tracef(ctx, "[%s] stored new quota by key:%s, quota: %d", funcName, key, newQuota)
		return
	})
	return
}

func (r *RateLimiter) getRedisKeyExpiredTime(key string) (expireTime time.Duration, err error) {
	redisExpireTime, errTTL := r.RedisPool.TTL(key)
	if errTTL != nil {
		err = fmt.Errorf("failed to redis pool get ttl by key: %s, error: %v", key, errTTL)
		return
	}
	expireTime = time.Second * time.Duration(redisExpireTime)
	return
}
