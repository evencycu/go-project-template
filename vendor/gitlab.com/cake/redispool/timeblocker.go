package redispool

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"golang.org/x/sync/singleflight"

	"gitlab.com/cake/goctx"
	"gitlab.com/cake/m800log"
)

const (
	// prefix key + component name + custom key(from input)
	prefixKeyPatternTimeBlocker       = "timeBlocker:%s:%s"
	prefixKeySignalPatternTimeBlocker = ":"
)

type TimeBlocker struct {
	RedisPool *Pool

	DefaultRedisCleanTimeout time.Duration
	// map[weights type] weights timeout limit
	WeightsTimeout             map[string]time.Duration
	singleFlightGroup          singleflight.Group
	SingleFlightExpireDuration time.Duration
}

type TimeBlockerCheckPassResult struct {
	Key              string        `json:"key,omitempty"`
	ComponentName    string        `json:"componentName,omitempty"`
	CustomKey        string        `json:"customKey,omitempty"`
	Passed           bool          `json:"passed,omitempty"`
	RemainingTimeout time.Duration `json:"remainingTimeoutNanoSec,omitempty"`
	Error            error         `json:"error,omitempty"`
}

func NewTimeBlocker(weightsTimeout map[string]time.Duration, defaultRedisCleanTimeout time.Duration, redisPool *Pool) (blocker *TimeBlocker, err error) {
	blocker = &TimeBlocker{
		RedisPool:                  redisPool,
		DefaultRedisCleanTimeout:   defaultRedisCleanTimeout,
		WeightsTimeout:             weightsTimeout,
		singleFlightGroup:          singleflight.Group{},
		SingleFlightExpireDuration: time.Millisecond * 100,
	}
	return
}

func GetTimeBlockerCacheKeyName(componentName, customKey string) string {
	return fmt.Sprintf(prefixKeyPatternTimeBlocker, componentName, customKey)
}

func GetComponentNameFromTimeBlockerKey(key string) string {
	res := strings.Split(key, prefixKeySignalPatternTimeBlocker)
	if len(res) < 3 {
		return ""
	}
	return res[1]
}

func GetCustomKeyFromTimeBlockerKey(key string) string {
	res := strings.Split(key, prefixKeySignalPatternTimeBlocker)
	if len(res) < 3 {
		return ""
	}
	return res[2]
}

func (tb *TimeBlocker) GetInt(ctx goctx.Context, key string) (currentScore int, remainingTimeout time.Duration, err error) {
	funcName := "TimeBlocker.GetInt"

	forgetSingleFlightGroupKey := func(sfKey string) {
		time.Sleep(tb.SingleFlightExpireDuration)
		tb.singleFlightGroup.Forget(sfKey)
	}
	type recordResult struct {
		Value            int
		RemainingTimeout time.Duration
	}
	// single flight: first one to do the job, others will be blocking and wait.
	// when the first has been completed, others will got same result too
	sfResult, err, _ := tb.singleFlightGroup.Do(key, func() (output interface{}, errSF error) {
		go forgetSingleFlightGroupKey(key)
		score, errSF := tb.RedisPool.GetInt(key)
		if errSF != nil {
			m800log.Errorf(ctx, "[%s] failed to get score from redis key: %s, error: %v", funcName, key, errSF)
			return
		}
		resRemainingTimeout, errSF := tb.GetExpiredTime(key)
		output = recordResult{
			Value:            score,
			RemainingTimeout: resRemainingTimeout,
		}
		m800log.Debugf(ctx, "[%s] got result by key: %s, score: %d", funcName, key, score)
		return
	})
	if err != nil {
		// because redis record is not exists, so unnecessary return nil error
		if err == redis.ErrNil {
			err = nil
		}
		return
	}
	result := sfResult.(recordResult)
	currentScore = result.Value
	remainingTimeout = result.RemainingTimeout
	return
}

func (tb *TimeBlocker) IsPassed(ctx goctx.Context, componentName, customKey string) (pass bool, remainingTimeout time.Duration, err error) {
	funcName := "TimeBlocker.IsPassed"
	key := GetTimeBlockerCacheKeyName(componentName, customKey)
	pass, remainingTimeout, err = tb.isPassedByKey(ctx, key)
	if err != nil {
		m800log.Debugf(ctx, "[%s] get score from redis key: %s, error: %s", funcName, key, err)
	}
	return
}

func (tb *TimeBlocker) isPassedByKey(ctx goctx.Context, key string) (pass bool, remainingTimeout time.Duration, err error) {
	score, remainingTimeout, err := tb.GetInt(ctx, key)
	if err != nil {
		return
	}
	pass = score == 0
	return
}

func (tb *TimeBlocker) IsKeysPassed(ctx goctx.Context, componentName string, customKeys []string) (results map[string]TimeBlockerCheckPassResult, err error) {
	funcName := "TimeBlocker.IsKeysPassed"
	results = map[string]TimeBlockerCheckPassResult{}
	// for validate
	if len(customKeys) == 0 {
		m800log.Errorf(ctx, "[%s] empty custom keys", funcName)
		err = fmt.Errorf("empty custom keys")
		return
	}
	if componentName == "" {
		m800log.Errorf(ctx, "[%s] empty component name", funcName)
		err = fmt.Errorf("empty component name")
		return
	}

	tbKeys := []string{}
	for i := range customKeys {
		tbKeys = append(tbKeys, GetTimeBlockerCacheKeyName(componentName, customKeys[i]))
	}
	resp, err := tb.RedisPool.MGet(tbKeys)
	if err != nil {
		m800log.Errorf(ctx, "[%s] failed to mget block score redis keys: %v, error: %s", funcName, tbKeys, err)
		return
	}
	ttls, err := tb.RedisPool.TTLs(tbKeys)
	if err != nil {
		m800log.Errorf(ctx, "[%s] failed to ttls redis keys: %v, error: %s", funcName, tbKeys, err)
		return
	}
	for key := range resp {
		score := 0
		var errParse error
		if resp[key] != nil {
			num, errConv := strconv.Atoi(fmt.Sprintf("%s", resp[key]))
			if errConv != nil {
				m800log.Debugf(ctx, "[%s] failed to parse redis data, key: %s, error: %s", funcName, key, errConv)
				errParse = errConv
			}
			score = num
		}

		pass := score == 0
		remainingTimeout := time.Second * time.Duration(ttls[key])
		results[key] = TimeBlockerCheckPassResult{
			Key:              key,
			ComponentName:    GetComponentNameFromTimeBlockerKey(key),
			CustomKey:        GetCustomKeyFromTimeBlockerKey(key),
			Passed:           pass,
			RemainingTimeout: remainingTimeout,
			Error:            errParse,
		}

	}
	return
}

func (tb *TimeBlocker) Incr(ctx goctx.Context, componentName, weightsType, customKey string) (currentRecord int, err error) {
	funcName := "TimeBlocker.Incr"
	key := GetTimeBlockerCacheKeyName(componentName, customKey)

	score := 1
	currentRecord, err = tb.RedisPool.IncrBy(key, score)
	if err != nil {
		m800log.Errorf(ctx, "[%s] failed to increment score to redis, key: %s, error: %+v", funcName, key, err)
		return
	}
	currentTimeout, err := tb.GetExpiredTime(key)
	if err != nil {
		m800log.Errorf(ctx, "[%s] failed to increment score to redis, key: %s, error: %+v", funcName, key, err)
		return
	}
	newTimeout := tb.getExpiredTimeByWeightsType(weightsType)
	// when blocker redis record timeout small than new timeout, redis timeout will be refresh
	if currentTimeout < newTimeout {
		timeoutSec := int(newTimeout.Seconds())
		if err = tb.RedisPool.Expire(key, strconv.Itoa(timeoutSec)); err != nil {
			m800log.Errorf(ctx, "[%s] failed to set expire timeout to redis, key: %s, error: %+v", funcName, key, err)
			return
		}
		currentTimeout = newTimeout
	}

	m800log.Debugf(ctx, "[%s] key: %s, current score: %d, remaining timeout: %v", funcName, key, currentRecord, currentTimeout)
	return
}

func (tb *TimeBlocker) getExpiredTimeByWeightsType(weightsType string) (resTimeout time.Duration) {
	resTimeout = tb.DefaultRedisCleanTimeout
	if tb.WeightsTimeout == nil {
		return
	}
	if timeout, exists := tb.WeightsTimeout[weightsType]; exists {
		resTimeout = timeout
		return
	}
	return
}

func (tb *TimeBlocker) GetExpiredTime(key string) (expireTime time.Duration, err error) {
	redisExpireTime, errTTL := tb.RedisPool.TTL(key)
	if errTTL != nil {
		err = fmt.Errorf("[GetExpiredTime] failed to redis pool get ttl by key: %s, error: %v", key, errTTL)
		return
	}
	expireTime = time.Second * time.Duration(redisExpireTime)
	return
}
