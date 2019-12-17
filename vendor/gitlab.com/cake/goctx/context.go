package goctx

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var nullStruct = struct{}{}

// follow the suggestion from "context" package to avoid collision
type mapContextKey string

// Context defines interface of common usage, combing log, value, go context methods.
type Context interface {
	Set(key string, value interface{})
	Get(key string) interface{}
	GetCID() (string, error)
	GetString(key string) (value string, ok bool)
	GetInt(key string) (value int, ok bool)
	GetInt64(key string) (value int64, ok bool)

	Map() (ret map[string]interface{})
	MapString() (ret map[string]string)
	HeaderKeyMap() (ret map[string]string)

	// header key transformation
	SetShortenKey(key string, value interface{})
	InjectHTTPHeader(s Setter)

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
	context.Context

	mu   sync.Mutex          // protects following fields
	keys map[string]struct{} // for record keys
}

func Background() Context {
	return &MapContext{Context: context.Background()}
}
func TODO() Context {
	return &MapContext{Context: context.TODO()}
}

func (c *MapContext) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Context = context.WithValue(c.Context, mapContextKey(key), value)
	if c.keys == nil {
		c.keys = make(map[string]struct{})
	}
	c.keys[key] = nullStruct
}

func (c *MapContext) Get(key string) interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Value(mapContextKey(key))
}

func (c *MapContext) Done() <-chan struct{} {
	c.mu.Lock()
	defer c.mu.Unlock()
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

func (c *MapContext) Map() (ret map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ret = make(map[string]interface{})
	for k := range c.keys {
		ret[k] = c.Value(mapContextKey(k))
	}
	return
}

func (c *MapContext) MapString() (ret map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ret = make(map[string]string)
	for k := range c.keys {
		if v, ok := c.Value(mapContextKey(k)).(string); ok {
			ret[k] = v
		}
	}
	if _, ok := c.keys[LogKeyTrace]; ok {
		ret[LogKeyTrace] = fmt.Sprintf("%s", c.Value(mapContextKey(LogKeyTrace)))
	}
	return
}

func (c *MapContext) SetTimeout(duration time.Duration) (cancel context.CancelFunc) {
	c.Context, cancel = context.WithTimeout(c.Context, duration)
	return
}

func (c *MapContext) SetDeadline(d time.Time) (cancel context.CancelFunc) {
	c.Context, cancel = context.WithDeadline(c.Context, d)
	return
}

func (c *MapContext) Cancel() (cancel context.CancelFunc) {
	c.Context, cancel = context.WithCancel(c.Context)
	return
}

// Note: if parent is not a cancelCtx, then there is no way to cancel children ctx from the parent
func (parent *MapContext) WithCancel() (ctx *MapContext, cancel context.CancelFunc) {
	c, cancel := context.WithCancel(parent.Context)
	ctx = &MapContext{Context: c, keys: parent.keys}
	return
}

// Note: if parent is not a cancelCtx, then there is no way to cancel children ctx from the parent
func (parent *MapContext) WithDeadline(d time.Time) (ctx *MapContext, cancel context.CancelFunc) {
	c, cancel := context.WithDeadline(parent.Context, d)
	ctx = &MapContext{Context: c, keys: parent.keys}
	return
}
