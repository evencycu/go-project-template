package redispool

import (
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

type RedisPubSub struct {
	channels *redis.Args
	psc      *redis.PubSubConn
}

func NewRedisPubSub(psc *redis.PubSubConn, channels *redis.Args) *RedisPubSub {
	return &RedisPubSub{
		psc:      psc,
		channels: channels,
	}
}

func (r *RedisPubSub) UnsubAndClose() error {
	err := r.psc.PUnsubscribe(redis.Args{}.AddFlat(r.channels)...)
	if err != nil {
		fmt.Printf("[PUnsubscribe] %+v failed, err: %+v", r.channels, err)
	}

	return r.psc.Close()
}

func (r *RedisPubSub) Subscribe() error {
	return r.psc.PSubscribe(*r.channels...)
}

func (r *RedisPubSub) Receive() (*redis.Message, error) {
	switch n := r.psc.Receive().(type) {
	case error:
		return nil, n
	case redis.Message:
		return &n, nil
	}
	return nil, nil
}

//Ref: https://redis.io/topics/notifications
func (p *Pool) GetRedisPubSub(channels []string, eventType string) (*RedisPubSub, error) {
	c := p.pool.Get()

	reply, err := c.Do("CONFIG", "SET", "notify-keyspace-events", eventType)
	if err != nil {
		errClose := c.Close()
		if errClose != nil {
			return nil, fmt.Errorf("[GetRedisPubSub] set config err %+v, close connection err: %+v", err, errClose)
		}
		return nil, err
	}

	if reply.(string) != "OK" {
		errClose := c.Close()
		if errClose != nil {
			return nil, fmt.Errorf("[GetRedisPubSub] set config reply not OK, close connection err: %+v", errClose)
		}
		return nil, errors.New("not ok")
	}

	args := redis.Args{}.AddFlat(channels)
	psc := redis.PubSubConn{Conn: c}

	return NewRedisPubSub(&psc, &args), nil
}
