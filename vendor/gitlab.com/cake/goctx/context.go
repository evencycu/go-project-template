package goctx

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
)

type RWMutexInterface interface {
	Lock()
	RLock()
	RLocker() sync.Locker
	RUnlock()
	Unlock()
}

// Context defines interface of common usage, combing log, value, go context methods.
type Context interface {
	Set(key string, value interface{})
	Get(key string) interface{}
	GetCID() (string, error)
	GetString(key string) (value string, ok bool)
	GetStringSlice(key string) (value []string, ok bool)
	GetBool(key string) (value bool, ok bool)
	GetInt(key string) (value int, ok bool)
	GetInt64(key string) (value int64, ok bool)

	Map() (ret map[string]interface{})
	MapString() (ret map[string]string)
	HeaderKeyMap() (ret map[string]string)

	// tracing method
	GetSpan() opentracing.Span
	StartSpanFromContext(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span
	SetSpan(span opentracing.Span)

	// header key transformation
	SetShortenKey(key string, value interface{})
	InjectHTTPHeader(s http.Header)

	// context related methods
	SetTimeout(duration time.Duration) (cancel context.CancelFunc)
	SetDeadline(d time.Time) (cancel context.CancelFunc)
	Cancel() (cancel context.CancelFunc)

	// create child goctx
	WithCancel() (ctx *MapContext, cancel context.CancelFunc)
	WithDeadline(d time.Time) (ctx *MapContext, cancel context.CancelFunc)

	// golang context() interface
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
	Deadline() (deadline time.Time, ok bool)
}

// MapContext implements the map concept and mixins the golang context
type MapContext struct {
	mu RWMutexInterface // protects following fields
	context.Context

	m *sync.Map
}

func Background() Context {
	return &MapContext{
		Context: context.Background(),
		mu:      &sync.RWMutex{},
		m:       &sync.Map{},
	}
}
func TODO() Context {
	return &MapContext{
		Context: context.TODO(),
		mu:      &sync.RWMutex{},
		m:       &sync.Map{},
	}
}

func (c *MapContext) Set(key string, value interface{}) {
	c.m.Store(key, value)
}

func (c *MapContext) Get(key string) interface{} {
	return c.Value(key)
}

func (c *MapContext) Done() <-chan struct{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Context.Done()
}

func (c *MapContext) GetString(key string) (value string, ok bool) {
	v := c.Get(key)
	if v == nil {
		return
	}
	value, ok = v.(string)
	return
}

func (c *MapContext) GetStringSlice(key string) (value []string, ok bool) {
	v := c.Get(key)
	if v == nil {
		return
	}
	value, ok = v.([]string)
	return
}

func (c *MapContext) GetBool(key string) (value, ok bool) {
	v := c.Get(key)
	if v == nil {
		return
	}
	value, ok = v.(bool)
	return
}

func (c *MapContext) GetInt(key string) (value int, ok bool) {
	v := c.Get(key)
	if v == nil {
		return
	}
	value, ok = v.(int)
	return
}

func (c *MapContext) GetInt64(key string) (value int64, ok bool) {
	v := c.Get(key)
	if v == nil {
		return
	}
	value, ok = v.(int64)
	return
}

// Map needs directly call c.Context.Value prevent duplicate lock
// usecase: copy context, print log get data
func (c *MapContext) Map() (ret map[string]interface{}) {
	ret = make(map[string]interface{})
	c.m.Range(func(k, v interface{}) bool {
		ret[k.(string)] = v
		return true
	})
	return
}

// MapString needs directly call c.Context.Value prevent duplicate lock
func (c *MapContext) MapString() (ret map[string]string) {
	ret = make(map[string]string)
	c.m.Range(func(k, v interface{}) bool {
		switch v.(type) {
		case string:
			ret[k.(string)] = v.(string)
		case []string:
			ret[k.(string)] = strings.Join(v.([]string), ",")
		}

		return true
	})

	if sp := opentracing.SpanFromContext(c.Context); sp != nil {
		uti := fmt.Sprintf("%s", sp)
		if uti != "" && uti != emptyObjectUti {
			ret[LogKeyTrace] = uti
		}
	}
	return
}

func (c *MapContext) SetTimeout(duration time.Duration) (cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Context, cancel = context.WithTimeout(c.Context, duration)
	return
}

func (c *MapContext) SetDeadline(d time.Time) (cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Context, cancel = context.WithDeadline(c.Context, d)
	return
}

func (c *MapContext) Cancel() (cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Context, cancel = context.WithCancel(c.Context)
	return
}

func (c *MapContext) InjectHTTPHeader(rh http.Header) {
	for hk, sk := range c.HeaderKeyMap() {
		if s := rh.Get(hk); len(s) == 0 {
			rh.Set(hk, sk)
		}
	}
}

// HeaderKeyMap returns a map, key is HTTP Header Field, value is the field value stored in Context
func (c *MapContext) HeaderKeyMap() (ret map[string]string) {
	ret = make(map[string]string)
	for sk, hk := range sKMap {
		v, _ := c.GetString(sk)
		if len(v) > 0 {
			ret[hk] = v
		}
	}
	if sp := c.GetSpan(); sp != nil {
		uti := fmt.Sprintf("%s", sp)
		if uti != "" && uti != emptyObjectUti {
			ret[HTTPHeaderTrace] = uti
		}
	}
	switch role := c.Get(LogKeyUserRole).(type) {
	case string:
		ret[HTTPHeaderUserRole] = role
	case []string:
		ret[HTTPHeaderUserRole] = strings.Join(role, ",")
	}
	return
}

// SetShortenKey sets Context with Header Key, Store in Context with LogKey Key
func (c *MapContext) SetShortenKey(headerField string, headerValue interface{}) {
	if sk, ok := hKMap[headerField]; ok {
		c.Set(sk, headerValue)
	}
}

// GetCID return the CID, if not, generate one
func (c *MapContext) GetCID() (string, error) {
	cid, ok := c.GetString(LogKeyCID)
	if !ok {
		cid = strconv.FormatInt(time.Now().UnixNano(), 10)
		c.Set(LogKeyCID, cid)
	}
	return cid, nil
}

func (c *MapContext) Err() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Context.Err()
}

func (c *MapContext) Value(key interface{}) interface{} {
	value, _ := c.m.Load(key)
	return value
}

func (c *MapContext) Deadline() (deadline time.Time, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Context.Deadline()
}

// Note: if parent is not a cancelCtx, then there is no way to cancel children ctx from the parent
func (parent *MapContext) WithCancel() (ctx *MapContext, cancel context.CancelFunc) {
	c, cancel := context.WithCancel(parent.Context)
	ctx = &MapContext{
		Context: c,
		m:       copySyncMap(parent.m),
		mu:      &sync.RWMutex{},
	}
	return
}

// Note: if parent is not a cancelCtx, then there is no way to cancel children ctx from the parent
func (parent *MapContext) WithDeadline(d time.Time) (ctx *MapContext, cancel context.CancelFunc) {
	c, cancel := context.WithDeadline(parent.Context, d)
	ctx = &MapContext{
		Context: c,
		m:       copySyncMap(parent.m),
		mu:      &sync.RWMutex{},
	}
	return
}

func copySyncMap(m *sync.Map) *sync.Map {
	var cp sync.Map

	m.Range(func(k, v interface{}) bool {
		vm, ok := v.(sync.Map)
		if ok {
			cp.Store(k, copySyncMap(&vm))
		} else {
			cp.Store(k, v)
		}

		return true
	})

	return &cp
}
