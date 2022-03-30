package redispool

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"gitlab.com/cake/m800log"
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
	RedisDecr    = "DECR"
	RedisDecrBy  = "DECRBY"
	RedisKeys    = "KEYS"
	RedisEx      = "EX"
	RedisTTL     = "TTL"
	RedisExists  = "EXISTS"
	RedisMGet    = "MGET"

	RedisOk     = "OK"
	RedisQueued = "QUEUED"
)

const (
	ErrMsgReadOnlyReplica = "READONLY You can't write against a read only replica"
)

func (p *Pool) checkErrReason(err error) {
	if err == nil {
		return
	}
	errMsg := err.Error()
	if strings.Contains(errMsg, ErrMsgReadOnlyReplica) {
		*p.needReconnectTime = time.Now().Add(p.pool.IdleTimeout)
		m800log.Warnf(p.ctx, "need reconnect err:%v needReconnectTime:%v", err, *p.needReconnectTime)
	}
}

func (p *Pool) FlushDB() (err error) {
	c := p.pool.Get()
	defer c.Close()

	str, err := redis.String(c.Do("FLUSHDB"))
	if err != nil {
		p.checkErrReason(err)
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
		p.checkErrReason(err)
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
		p.checkErrReason(err)
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
		p.checkErrReason(err)
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
	value, err = redis.Bool(c.Do(RedisExists, key))
	p.checkErrReason(err)
	return
}

func (p *Pool) TTL(key string) (value int, err error) {
	c := p.pool.Get()
	defer c.Close()
	value, err = redis.Int(c.Do(RedisTTL, key))
	p.checkErrReason(err)
	return
}

// results =  map[key] timeout seconds
// recommend change to use mongodb, redis not suitable for too complex use case.
func (p *Pool) TTLs(keys []string) (results map[string]int, err error) {
	c := p.pool.Get()
	defer c.Close()

	comnands := []string{}
	args := [][]interface{}{}
	for i := range keys {
		comnands = append(comnands, RedisTTL)
		args = append(args, []interface{}{keys[i]})
	}

	resps, err := p.MultiExec(comnands, args)
	if err != nil {
		p.checkErrReason(err)
		return
	}
	results = map[string]int{}
	for i := range resps {
		num, ok := resps[i].(int64)
		if !ok {
			err = fmt.Errorf("failed to parse response interface to int64: %+v, type: %t", resps[i], resps[i])
			p.checkErrReason(err)
			return
		}
		results[keys[i]] = int(num)
	}

	return
}

func (p *Pool) GetBytes(key string) (value []byte, err error) {
	c := p.pool.Get()
	defer c.Close()
	value, err = redis.Bytes(c.Do(RedisGet, key))
	p.checkErrReason(err)
	return
}

func (p *Pool) GetString(key string) (value string, err error) {
	c := p.pool.Get()
	defer c.Close()
	value, err = redis.String(c.Do(RedisGet, key))
	p.checkErrReason(err)
	return
}

func (p *Pool) GetStrings(key string) (values []string, err error) {
	c := p.pool.Get()
	defer c.Close()
	values, err = redis.Strings(c.Do(RedisGet, key))
	p.checkErrReason(err)
	return
}

func (p *Pool) GetBool(key string) (value bool, err error) {
	c := p.pool.Get()
	defer c.Close()
	value, err = redis.Bool(c.Do(RedisGet, key))
	p.checkErrReason(err)
	return
}

func (p *Pool) GetInt(key string) (value int, err error) {
	c := p.pool.Get()
	defer c.Close()
	value, err = redis.Int(c.Do(RedisGet, key))
	p.checkErrReason(err)
	return
}

func (p *Pool) Expire(key, ttlSec string) error {
	c := p.pool.Get()
	defer c.Close()
	resp, err := redis.Int(c.Do(RedisExpire, key, ttlSec))
	if err != nil {
		p.checkErrReason(err)
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
		p.checkErrReason(err)
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
		p.checkErrReason(err)
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
	res, err := redis.Int(c.Do(RedisHSet, key, field, jsonBytes))
	p.checkErrReason(err)

	return res, err
}

// if the field does not exist, then reutrn redis.ErrNil
func (p *Pool) HGetJSON(key, field string, result interface{}) error {
	c := p.pool.Get()
	defer c.Close()

	bytes, err := redis.Bytes(c.Do(RedisHGet, key, field))
	if err != nil {
		p.checkErrReason(err)
		return err
	}

	return json.Unmarshal(bytes, result)
}

func (p *Pool) HKeys(key string) ([]string, error) {
	c := p.pool.Get()
	defer c.Close()
	res, err := redis.Strings(c.Do(RedisHKeys, key))

	p.checkErrReason(err)
	return res, err
}

// values should be the type of *map[string]struct
// if field data not in redis, values will not include that key
func (p *Pool) HMGetJSON(key string, fields []string, values interface{}) error {
	c := p.pool.Get()
	defer c.Close()
	// is a map[string]interface{}
	valuesv := reflect.ValueOf(values)
	if valuesv.Kind() != reflect.Ptr {
		return errors.New("values argument must be a pointer of map")
	}
	if valuesv.Elem().Kind() != reflect.Map {
		return errors.New("values argument must be a pointer of map")
	}
	if valuesv.Elem().Type().Key().Kind() != reflect.String {
		return errors.New("key of the map must be string type")
	}

	fs := make([]interface{}, 0)
	fs = append(fs, key)
	for _, field := range fields {
		fs = append(fs, field)
	}

	vs, err := redis.Values(c.Do(RedisHMGet, fs...))
	if err != nil {
		p.checkErrReason(err)
		return err
	}

	// Key doesn't exist
	if len(vs) == 0 {
		return nil
	}

	counter := 0

	resultBytes := []byte{'{'}
	for i, v := range vs {
		if v == nil {
			continue
		}
		counter++
		bytes, ok := v.([]byte)
		if !ok {
			errMsg := fmt.Sprintf("v %T is not of type bytes", v)
			return errors.New(errMsg)
		}
		resultBytes = append(resultBytes, '"')
		resultBytes = append(resultBytes, []byte(fields[i])...)
		resultBytes = append(resultBytes, '"')
		resultBytes = append(resultBytes, byte(':'))
		resultBytes = append(resultBytes, bytes...)
		resultBytes = append(resultBytes, byte(','))
	}
	resultBytes[len(resultBytes)-1] = '}'

	// No valid fields
	if counter == 0 {
		return nil
	}
	err = json.Unmarshal(resultBytes, values)
	if err != nil {
		return err
	}
	return nil
}

func (p *Pool) HLen(key string) (int, error) {
	c := p.pool.Get()
	defer c.Close()
	res, err := redis.Int(c.Do(RedisHLen, key))

	p.checkErrReason(err)
	return res, err
}

// values should be the type of map[string]struct
func (p *Pool) HMSetJSON(key string, values interface{}) error {
	c := p.pool.Get()
	defer c.Close()
	// is a map[string]interface{}
	valuesv := reflect.ValueOf(values)
	if valuesv.Kind() != reflect.Map {
		return errors.New("values argument must be a map")
	}
	for _, k := range valuesv.MapKeys() {
		if k.Kind() != reflect.String {
			return errors.New("key of the map must be string type")
		}
	}

	fs := make([]interface{}, 0)
	fs = append(fs, key)
	iter := valuesv.MapRange()
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()
		bytes, err := json.Marshal(v.Interface())
		if err != nil {
			return err
		}

		fs = append(fs, k)
		fs = append(fs, bytes)
	}

	resp, err := redis.String(c.Do(RedisHMSet, fs...))
	if err != nil {
		p.checkErrReason(err)
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
	res, err := redis.Int(c.Do(RedisHDel, key, field))

	p.checkErrReason(err)
	return res, err
}

// values should be the type of *map[string]struct
func (p *Pool) HGetAllJSON(key string, result interface{}) error {
	c := p.pool.Get()
	defer c.Close()
	// is a map[string]interface{}
	resultv := reflect.ValueOf(result)
	if resultv.Kind() != reflect.Ptr {
		return errors.New("result argument must be a pointer of map")
	}
	if resultv.Elem().Kind() != reflect.Map {
		return errors.New("result argument must be a pointer of map")
	}
	if resultv.Elem().Type().Key().Kind() != reflect.String {
		return errors.New("key of the map must be string type")
	}

	vs, err := redis.Values(c.Do(RedisHGetAll, key))
	if err != nil {
		p.checkErrReason(err)
		return err
	}

	// Key doesn't exist
	if len(vs) == 0 {
		return nil
	}
	resultBytes := []byte{'{'}
	for i, v := range vs {
		if i%2 == 0 {
			// key
			if keyValue, ok := v.([]byte); !ok {
				return fmt.Errorf("unknown key type %+v", v)
			} else {
				resultBytes = append(resultBytes, '"')
				resultBytes = append(resultBytes, keyValue...)
				resultBytes = append(resultBytes, '"')
			}
		}

		if i%2 == 1 {
			// value
			bytes, ok := v.([]byte)
			if !ok {
				errMsg := fmt.Sprintf("v %T is not of type bytes", v)
				return errors.New(errMsg)
			}
			resultBytes = append(resultBytes, ':')
			resultBytes = append(resultBytes, bytes...)
			resultBytes = append(resultBytes, ',')
		}

	}
	resultBytes[len(resultBytes)-1] = '}'
	err = json.Unmarshal(resultBytes, result)
	if err != nil {
		return err
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
		p.checkErrReason(err)
		return nil, err
	}

	return resps, nil
}

func (p *Pool) WatchMultiExec(wme *WatchMultiExecutor) ([]interface{}, error) {
	c := p.pool.Get()
	defer c.Close()
	res, err := wme.Exec(c)
	p.checkErrReason(err)
	return res, err
}

// FIXME: should not be exposed because of performance issue
// See warning in: https://redis.io/commands/keys
func (p *Pool) Keys(key string) ([]string, error) {
	c := p.pool.Get()
	defer c.Close()
	res, err := redis.Strings(c.Do(RedisKeys, key))
	p.checkErrReason(err)

	return res, err
}

func (p *Pool) Incr(key string) (int, error) {
	c := p.pool.Get()
	defer c.Close()
	res, err := redis.Int(c.Do(RedisIncr, key))
	p.checkErrReason(err)
	return res, err
}

func (p *Pool) IncrBy(key string, amount int) (int, error) {
	c := p.pool.Get()
	defer c.Close()
	res, err := redis.Int(c.Do(RedisIncrBy, key, amount))
	p.checkErrReason(err)

	return res, err
}

func (p *Pool) Decr(key string) (int, error) {
	c := p.pool.Get()
	defer c.Close()
	res, err := redis.Int(c.Do(RedisDecr, key))
	p.checkErrReason(err)
	return res, err
}

func (p *Pool) DecrBy(key string, amount int) (int, error) {
	c := p.pool.Get()
	defer c.Close()
	res, err := redis.Int(c.Do(RedisDecrBy, key, amount))
	p.checkErrReason(err)

	return res, err
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

func (p *Pool) ScanKeys(pattern string) ([]string, error) {
	c := p.pool.Get()
	defer c.Close()

	iter := 0
	keys := []string{}
	for {
		arr, err := redis.Values(c.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return keys, fmt.Errorf("patterm '%s' keys failed", pattern)
		}

		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}

	return keys, nil
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

// MGet map[key][]byte
func (p *Pool) MGet(keys []string) (map[string]interface{}, error) {
	c := p.pool.Get()
	defer c.Close()

	result := map[string]interface{}{}
	args := []interface{}{}
	for i := range keys {
		args = append(args, keys[i])
	}

	// expect: "1", "2", "3"...
	r, err := c.Do(RedisMGet, args...)
	reply, err := redis.Values(r, err)
	if err != nil {
		p.checkErrReason(err)
		return nil, err
	}

	for i := range reply {
		result[keys[i]] = reply[i]
	}
	return result, err
}
