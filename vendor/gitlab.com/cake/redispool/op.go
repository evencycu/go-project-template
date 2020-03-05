package redispool

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/gomodule/redigo/redis"
)

const (
	RedisSet     = "SET"
	RedisSetex   = "SETEX"
	RedisSadd    = "SADD"
	RedisExpire  = "EXPIRE"
	RedisDel     = "DEL"
	RedisGet     = "GET"
	RedisWatch   = "WATCH"
	RedisMulti   = "MULTI"
	RedisExec    = "EXEC"
	RedisPing    = "PING"
	RedisPong    = "PONG"
	RedisNx      = "NX"
	RedisHSet    = "HSET"
	RedisHMSet   = "HMSET"
	RedisHGet    = "HGET"
	RedisHKeys   = "HKEYS"
	RedisHMGet   = "HMGET"
	RedisHGetAll = "HGETALL"
	RedisHLen    = "HLEN"
	RedisHDel    = "HDEL"
	RedisIncr    = "INCR"
	RedisIncrBy  = "INCRBY"
	RedisKeys    = "KEYS"
	RedisEx      = "EX"
	RedisTTL     = "TTL"
	RedisExists  = "EXISTS"

	RedisOk     = "OK"
	RedisQueued = "QUEUED"
)

func (p *Pool) FlushDB() (err error) {
	c := p.pool.Get()
	defer c.Close()

	str, err := redis.String(c.Do("FLUSHDB"))
	if err != nil {
		return
	}
	if str != RedisOk {
		return fmt.Errorf("unknown resp:%s", str)
	}
	return
}

// Set key, value only
func (p *Pool) Set(key string, value interface{}) error {
	c := p.pool.Get()
	defer c.Close()
	str, err := redis.String(c.Do(RedisSet, key, value))
	if err != nil {
		return err
	}
	if str != RedisOk {
		return fmt.Errorf("unknown resp:%s", str)
	}
	return nil
}

// Setex set key, value with TTL
func (p *Pool) Setex(key, ttlSec string, value interface{}) error {
	c := p.pool.Get()
	defer c.Close()
	str, err := redis.String(c.Do(RedisSetex, key, ttlSec, value))
	if err != nil {
		return err
	}
	if str != RedisOk {
		return fmt.Errorf("unknown resp:%s", str)
	}
	return nil
}

// Setnx set key value and TTL if not exist
func (p *Pool) Setnx(key, ttlSec string, value interface{}) error {
	c := p.pool.Get()
	defer c.Close()
	str, err := redis.String(c.Do(RedisSet, key, value, RedisEx, ttlSec, RedisNx))
	if err != nil {
		return err
	}
	if str != RedisOk {
		return fmt.Errorf("unknown resp:%s", str)
	}
	return nil
}

func (p *Pool) Exists(key string) (value bool, err error) {
	c := p.pool.Get()
	defer c.Close()

	return redis.Bool(c.Do(RedisExists, key))
}

func (p *Pool) TTL(key string) (value int, err error) {
	c := p.pool.Get()
	defer c.Close()

	return redis.Int(c.Do(RedisTTL, key))
}

func (p *Pool) GetBytes(key string) (value []byte, err error) {
	c := p.pool.Get()
	defer c.Close()

	return redis.Bytes(c.Do(RedisGet, key))
}

func (p *Pool) GetString(key string) (value string, err error) {
	c := p.pool.Get()
	defer c.Close()

	return redis.String(c.Do(RedisGet, key))
}

func (p *Pool) GetBool(key string) (value bool, err error) {
	c := p.pool.Get()
	defer c.Close()

	return redis.Bool(c.Do(RedisGet, key))
}

func (p *Pool) GetInt(key string) (value int, err error) {
	c := p.pool.Get()
	defer c.Close()

	return redis.Int(c.Do(RedisGet, key))
}

func (p *Pool) Expire(key, ttlSec string) error {
	c := p.pool.Get()
	defer c.Close()
	resp, err := redis.Int(c.Do(RedisExpire, key, ttlSec))
	if err != nil {
		return err
	}
	if resp != 1 {
		return fmt.Errorf("unknown resp:%d", resp)
	}
	return nil
}

func (p *Pool) Delete(key string) error {
	c := p.pool.Get()
	defer c.Close()
	resp, err := redis.Int(c.Do(RedisDel, key))
	if err != nil {
		return err
	}
	if resp != 1 {
		return fmt.Errorf("unknown resp:%d", resp)
	}
	return nil
}

func (p *Pool) getAndDelete(key string) (value interface{}, err error) {
	c := p.pool.Get()
	defer c.Close()
	if err = c.Send(RedisMulti); err != nil {
		return
	}
	if err = c.Send(RedisGet, key); err != nil {
		return
	}
	if err = c.Send(RedisDel, key); err != nil {
		return
	}
	var resps []interface{}
	resps, err = redis.Values(c.Do(RedisExec))
	if err != nil {
		return
	}
	value = resps[0]
	return
}

func (p *Pool) GetStringAndDelete(key string) (value string, err error) {
	return redis.String(p.getAndDelete(key))
}

func (p *Pool) GetIntAndDelete(key string) (value int, err error) {
	return redis.Int(p.getAndDelete(key))
}

func (p *Pool) GetBytesAndDelete(key string) (value []byte, err error) {
	return redis.Bytes(p.getAndDelete(key))
}

func (p *Pool) HSetJSON(key, field string, value interface{}) (int, error) {
	c := p.pool.Get()
	defer c.Close()

	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return 0, err
	}

	return redis.Int(c.Do(RedisHSet, key, field, jsonBytes))
}

//if the field does not exist, then reutrn redis.ErrNil
func (p *Pool) HGetJSON(key, field string, result interface{}) error {
	c := p.pool.Get()
	defer c.Close()

	bytes, err := redis.Bytes(c.Do(RedisHGet, key, field))
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, result)
}

func (p *Pool) HKeys(key string) ([]string, error) {
	c := p.pool.Get()
	defer c.Close()

	return redis.Strings(c.Do(RedisHKeys, key))
}

func (p *Pool) HMGetJSON(key string, values map[string]interface{}) error {
	c := p.pool.Get()
	defer c.Close()

	fs := make([]interface{}, 0)
	fs = append(fs, key)
	fields := make([]string, 0)
	for k, _ := range values {
		fs = append(fs, k)
		fields = append(fields, k)
	}

	vs, err := redis.Values(c.Do(RedisHMGet, fs...))
	if err != nil {
		return err
	}

	for i, v := range vs {
		if v == nil {
			values[fields[i]] = nil
			continue
		}

		bytes, ok := v.([]byte)
		if !ok {
			errMsg := fmt.Sprintf("v %T is not of type bytes", v)
			return errors.New(errMsg)
		}

		err = json.Unmarshal(bytes, values[fields[i]])
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Pool) HLen(key string) (int, error) {
	c := p.pool.Get()
	defer c.Close()

	return redis.Int(c.Do(RedisHLen, key))
}

func (p *Pool) HMSetJSON(key string, values map[string]interface{}) error {
	c := p.pool.Get()
	defer c.Close()

	fs := make([]interface{}, 0)
	fs = append(fs, key)
	for k, v := range values {
		bytes, err := json.Marshal(v)
		if err != nil {
			return err
		}
		fs = append(fs, k)
		fs = append(fs, bytes)
	}

	resp, err := redis.String(c.Do(RedisHMSet, fs...))
	if err != nil {
		return err
	}

	if resp != RedisOk {
		return errors.New(resp)
	}

	return nil
}

func (p *Pool) HDel(key, field string) (int, error) {
	c := p.pool.Get()
	defer c.Close()

	return redis.Int(c.Do(RedisHDel, key, field))
}

func (p *Pool) HGetAllJSON(key string, result map[string]interface{}, toType reflect.Type) error {
	c := p.pool.Get()
	defer c.Close()

	vs, err := redis.Values(c.Do(RedisHGetAll, key))
	if err != nil {
		return err
	}

	k := ""
	for i, v := range vs {
		if i%2 == 0 {
			//key
			if keyValue, ok := v.([]byte); !ok {
				return fmt.Errorf("unknown key type %+v", v)
			} else {
				k = string(keyValue)
			}
		}

		if i%2 == 1 {
			//value
			if v == nil {
				result[k] = nil
				continue
			}

			bytes, ok := v.([]byte)
			if !ok {
				errMsg := fmt.Sprintf("v %T is not of type bytes", v)
				return errors.New(errMsg)
			}

			reflectValue := reflect.New(toType)

			newObj := reflectValue.Interface()
			err = json.Unmarshal(bytes, newObj)
			if err != nil {
				return err
			}

			result[k] = newObj
		}

	}

	return nil
}

func (p *Pool) MultiExec(commands []string, args [][]interface{}) ([]interface{}, error) {
	c := p.pool.Get()
	defer c.Close()
	if len(commands) == 0 {
		return nil, nil
	}

	if len(commands) != len(args) {
		return nil, errors.New("unmatch number of commands and args")
	}
	if err := c.Send(RedisMulti); err != nil {
		return nil, err
	}

	for i := 0; i < len(commands); i++ {
		if err := c.Send(commands[i], args[i]...); err != nil {
			return nil, err
		}
	}

	var resps []interface{}
	resps, err := redis.Values(c.Do(RedisExec))
	if err != nil {
		return nil, err
	}

	return resps, nil
}

func (p *Pool) WatchMultiExec(wme *WatchMultiExecutor) ([]interface{}, error) {
	c := p.pool.Get()
	defer c.Close()

	return wme.Exec(c)
}

func (p *Pool) Keys(key string) ([]string, error) {
	c := p.pool.Get()
	defer c.Close()

	return redis.Strings(c.Do(RedisKeys, key))
}

func (p *Pool) Incr(key string) (int, error) {
	c := p.pool.Get()
	defer c.Close()

	return redis.Int(c.Do(RedisIncr, key))
}

func (p *Pool) IncrBy(key string, amount int) (int, error) {
	c := p.pool.Get()
	defer c.Close()

	return redis.Int(c.Do(RedisIncrBy, key, amount))
}

func (p *Pool) GetJSON(key string, result interface{}) error {
	bytes, err := p.GetBytes(key)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, result)
}

func (p *Pool) SetJSON(key string, content interface{}) error {
	bytes, err := json.Marshal(content)
	if err != nil {
		return err
	}

	return p.Set(key, bytes)
}

type WatchMultiExecutor struct {
	watchKeys []string
	commands  []Command
}

type Command struct {
	name string
	args []interface{}
}

func (w *WatchMultiExecutor) AddWatches(keys ...string) *WatchMultiExecutor {
	w.watchKeys = append(w.watchKeys, keys...)
	return w
}

func (w *WatchMultiExecutor) AddCommand(name string, args ...interface{}) *WatchMultiExecutor {
	w.commands = append(w.commands, Command{name: name, args: args})
	return w
}

func (w *WatchMultiExecutor) Exec(c redis.Conn) ([]interface{}, error) {
	if len(w.commands) == 0 {
		return nil, nil
	}

	if len(w.watchKeys) != 0 {
		unique := make(map[string]bool)
		watchKeys := make([]interface{}, 0)
		for _, watchKey := range w.watchKeys {
			if _, ok := unique[watchKey]; !ok {
				watchKeys = append(watchKeys, watchKey)
			}
		}

		resp, err := redis.String(c.Do(RedisWatch, watchKeys...))
		if err != nil {
			return nil, err
		}

		if resp != RedisOk {
			return nil, fmt.Errorf("cannot WATCH keys %+v: %s", watchKeys, resp)
		}
	}

	if err := c.Send(RedisMulti); err != nil {
		return nil, err
	}

	for _, command := range w.commands {
		if err := c.Send(command.name, command.args...); err != nil {
			return nil, err
		}
	}

	resps, err := redis.Values(c.Do(RedisExec))
	if err != nil {
		return nil, err
	}

	return resps, nil
}
