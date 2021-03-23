package redispool

import (
	"fmt"
	"log"

	"github.com/ory/dockertest/v3"
	"gitlab.com/cake/goctx"
	"gitlab.com/cake/m800log"
)

type RedisDockerTester struct {
	DockerPool            *dockertest.Pool
	RedisResource         *dockertest.Resource
	RedisSentinelResource *dockertest.Resource
	Pool                  *Pool
}

func (r *RedisDockerTester) Teardown() {
	if err := r.DockerPool.Purge(r.RedisResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := r.DockerPool.Purge(r.RedisSentinelResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func SetupRedisDockers(ctx goctx.Context, conf *Config) *RedisDockerTester {
	var err error
	tester := RedisDockerTester{}

	InitRedisDockers(ctx, &tester)

	sentinel := fmt.Sprintf("127.0.0.1:%s", tester.RedisSentinelResource.GetPort("26379/tcp"))
	conf.Hosts = []string{sentinel}

	tester.Pool, err = NewPool(conf)
	if err != nil {
		panic("redispool init error:" + err.Error())
	}
	err = tester.DockerPool.Retry(func() error {
		errI := tester.Pool.Ping()
		return errI
	})
	if err != nil {
		panic("connect docker fail, error:" + err.Error())
	}

	return &tester
}

func InitRedisDockers(ctx goctx.Context, tester *RedisDockerTester) {
	var err error
	tester.DockerPool, err = dockertest.NewPool("")
	if err != nil {
		panic("docker test init fail, error:" + err.Error())
	}

	tester.RedisResource, err = tester.DockerPool.Run("redis", "5.0", nil)
	if err != nil {
		panic("redis docker init fail, error:" + err.Error())
	}
	redisMaster := fmt.Sprintf("127.0.0.1:%s", tester.RedisResource.GetPort("6379/tcp"))
	m800log.Debugf(ctx, "redisMaster: %+v", redisMaster)

	tester.RedisSentinelResource, err = tester.DockerPool.Run("s7anley/redis-sentinel-docker", "4.0", []string{"MASTER=" + redisMaster, "MASTER_NAME=master"})
	if err != nil {
		panic("redis docker init fail, error:" + err.Error())
	}
}
