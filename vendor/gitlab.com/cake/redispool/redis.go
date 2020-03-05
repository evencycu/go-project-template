package redispool

import (
	"fmt"
	"time"

	"github.com/FZambia/sentinel"
	"github.com/gomodule/redigo/redis"
	conntrack "github.com/eaglerayp/go-conntrack"
	"gitlab.com/cake/goctx"
	"gitlab.com/cake/m800log"
)

var systemCtx goctx.Context

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
	pool *redis.Pool
	conf *Config
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

	conntrackDialer := conntrack.NewDialFunc(
		conntrack.DialWithName("redispool"),
	)

	ctx.Set("database", "redis")
	ctx.Set("host", host)
	ctx.Set("masterName", master)
	ctx.Set("db", db)
	ctx.Set("connectTimeout", timeout.String())
	ctx.Set("maxIdle", maxIdle)
	ctx.Set("maxActive", maxActive)
	ctx.Set("idleTimeout", idleTimeout.String())
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
		// TestOnBorrow: sentinelBorrow,
	}

	systemCtx = ctx
	return &Pool{
		pool,
		conf,
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
		m800log.Info(systemCtx, "redis ping unknown response:", resp)
		return fmt.Errorf("unknow response")
	}
	if str != RedisPong {
		m800log.Info(systemCtx, "redis ping bad response:", resp)
		return fmt.Errorf("%s", str)
	}
	return nil
}
