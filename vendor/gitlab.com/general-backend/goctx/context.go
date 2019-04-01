package goctx

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
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
	LogFields() (ret logrus.Fields)
	SetTimeout(duration time.Duration) (cancel context.CancelFunc)
	SetDeadline(d time.Time) (cancel context.CancelFunc)
	Cancel() (cancel context.CancelFunc)
	LogKeyMap() (ret map[string]string)
	LogKeyFields() (ret logrus.Fields)
	LogKeySet(headerField, headerValue string)
	SetHTTPHeaders(s Setter)
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
	c.Context = context.WithValue(c.Context, mapContextKey(key), value)
	if c.keys == nil {
		c.keys = make(map[string]struct{})
	}
	c.mu.Lock()
	c.keys[key] = nullStruct
	c.mu.Unlock()
}

func (c *MapContext) Get(key string) interface{} {
	return c.Value(mapContextKey(key))
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
	ret = make(map[string]interface{})
	for k := range c.keys {
		ret[k] = c.Value(mapContextKey(k))
	}
	return
}

func (c *MapContext) LogFields() (ret logrus.Fields) {
	ret = make(logrus.Fields)
	for k := range c.keys {
		ret[k] = c.Value(mapContextKey(k))
	}
	delete(ret, LogKeyTrace)
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
