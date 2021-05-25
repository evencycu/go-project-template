package redispool

import (
	"fmt"
	"time"

	"github.com/FZambia/sentinel"
	"github.com/eaglerayp/go-conntrack"
	"github.com/gomodule/redigo/redis"
	"gitlab.com/cake/goctx"
	"gitlab.com/cake/m800log"
)

type Config struct {
	Hosts            []string
	MasterName       string
	SentinelPassword string
	Password         string
	Timeout          time.Duration
	IdleTimeout      time.Duration
	MaxIdle          int
	MaxActive        int
	RedisDB          int
}

type Pool struct {
	ctx               goctx.Context
	pool              *redis.Pool
	conf              *Config
	needReconnectTime *time.Time
	// for unit test
	reconnectTryCounter *int
}

func NewPool(conf *Config) (*Pool, error) {
	ctx := goctx.Background()

	host := conf.Hosts
	master := conf.MasterName
	pw := conf.Password
	sentinelPw := conf.SentinelPassword
	timeout := conf.Timeout
	db := conf.RedisDB
	maxIdle := conf.MaxIdle
	maxActive := conf.MaxActive
	idleTimeout := conf.IdleTimeout
	now := time.Now()
	tp := &now
	counter := 0

	conntrackDialer := conntrack.NewDialFunc(
		conntrack.DialWithName("redispool"),
	)

	ctx.Set("DBType", "redis")
	ctx.Set("redis.host", host)
	ctx.Set("redis.masterName", master)
	ctx.Set("redis.db", db)
	ctx.Set("redis.connectTimeout", timeout.String())
	ctx.Set("redis.maxIdle", maxIdle)
	ctx.Set("redis.maxActive", maxActive)
	ctx.Set("redis.idleTimeout", idleTimeout.String())
	m800log.Info(ctx, "Init redis with config in ctx")
	sntnl := &sentinel.Sentinel{
		Addrs:      host,
		MasterName: master,
		Dial: func(addr string) (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr,
				redis.DialNetDial(conntrackDialer),
				redis.DialPassword(sentinelPw),
				redis.DialConnectTimeout(timeout),
				redis.DialWriteTimeout(timeout),
				redis.DialReadTimeout(timeout))
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}

	pool := &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: idleTimeout,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			masterAddr, err := sntnl.MasterAddr()
			if err != nil {
				return nil, err
			}

			ctx.Set("redis.masterAddr", masterAddr)
			m800log.Debug(ctx, "redis dialing, master addr: ", masterAddr)

			c, err := redis.Dial("tcp", masterAddr,
				redis.DialNetDial(conntrackDialer),
				redis.DialPassword(pw),
				redis.DialConnectTimeout(timeout),
				redis.DialWriteTimeout(timeout),
				redis.DialReadTimeout(timeout),
				redis.DialDatabase(db))
			if err != nil {
				return nil, err
			}

			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if t.After(*tp) {
				return nil
			}
			counter++
			if !sentinel.TestRole(c, "master") {
				return fmt.Errorf("redis role check failed")
			} else {
				return nil
			}
		},
	}

	return &Pool{
		ctx,
		pool,
		conf,
		tp,
		&counter,
	}, nil
}

func (p *Pool) Ping() error {
	c := p.pool.Get()
	defer c.Close()
	resp, err := c.Do(RedisPing)
	if err != nil {
		return err
	}
	str, ok := resp.(string)
	if !ok {
		m800log.Info(p.ctx, "redis ping unknown response:", resp)
		return fmt.Errorf("unknow response")
	}
	if str != RedisPong {
		m800log.Info(p.ctx, "redis ping bad response:", resp)
		return fmt.Errorf("%s", str)
	}
	return nil
}

func (p *Pool) GetPool() *redis.Pool {
	return p.pool
}
