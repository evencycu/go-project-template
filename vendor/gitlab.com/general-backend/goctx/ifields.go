package goctx

import (
	uuid "github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

const (
	// ContextKey is the common field if we have to put goctx in other context object
	ContextKey = "X-Ctx"

	// for trace
	HTTPHeaderTrace         = "uber-trace-id"
	HTTPHeaderJaegerDebug   = "jaeger-debug-id"
	HTTPHeaderJaegerBaggage = "jaeger-baggage"

	// for http passing
	HTTPHeaderCID = "x-correlation-id"
	HTTPHeaderApp = "x-app"

	// for logger
	LogKeyTrace         = "uti"
	LogKeyJaegerDebug   = "jdi"
	LogKeyJaegerBaggage = "jb"

	// FROM m800 log format document: https://issuetracking.maaii.com:9443/pages/viewpage.action?pageId=65128541#Loggingformatdesign(v1.0)-Backendteam(BE)

	// LogKeyCID is the cid field key
	LogKeyCID = "cid"
	// LogKeyTimestamp is the time field key
	LogKeyTimestamp = "time"
	// LogKeyLevel is the level field key
	LogKeyLevel = "level"
	// LogKeyMessage is the message field key
	LogKeyMessage      = "message"
	LogKeyEntryType    = "entryType"
	LogKeyURI          = "uri"
	LogKeyMethod       = "method"
	LogKeyHTTPMethod   = "httpMethod"
	LogKeyInstance     = "instanceID"
	LogKeyVersion      = "vid"
	LogKeyCase         = "caseType"
	LogKeyRegion       = "region"
	LogKeyErrorCode    = "eCode"
	LogKeyErrorMessage = "eMessage"
	LogKeyErrorType    = "eType"

	// golang tp add fields

	// LogKeyAccessTime is the log field of access time
	LogKeyAccessTime = "accessTime"
	// LogKeyLogType is the log type field key
	LogKeyLogType = "type"
	// LogKeyApp is the log field of app
	LogKeyApp = "app"
)

// Peeker designed for getting goctx from fasthttp *RequestHeader
type Peeker interface {
	Peek(key string) []byte
}

// Setter designed for setting fasthttp *RequestHeader, net/http Header with goctx
type Setter interface {
	Set(key, value string)
}

// Getter interface designed for getting goctx from net/http Header
type Getter interface {
	Get(key string) string
}

// GetHeaderer interface designed for getting goctx from gin.Context.GetHeader
type GetHeaderer interface {
	GetHeader(key string) string
}

var (
	sKMap, hKMap map[string]string
)

func init() {
	sKMap = map[string]string{
		LogKeyTrace:         HTTPHeaderTrace,
		LogKeyJaegerDebug:   HTTPHeaderJaegerDebug,
		LogKeyJaegerBaggage: HTTPHeaderJaegerBaggage,
		LogKeyCID:           HTTPHeaderCID,
		LogKeyApp:           HTTPHeaderApp,
	}

	hKMap = map[string]string{
		HTTPHeaderTrace:         LogKeyTrace,
		HTTPHeaderJaegerDebug:   LogKeyJaegerDebug,
		HTTPHeaderJaegerBaggage: LogKeyJaegerBaggage,
		HTTPHeaderCID:           LogKeyCID,
		HTTPHeaderApp:           LogKeyApp,
	}
}

// IFieldHeaderKeyMap returns a map, key is HTTP Header Field, value is LogKey Field
func IFieldHeaderKeyMap() (keyMap map[string]string) {
	return hKMap
}

// IFieldLogKeyKeyMap returns a map, key is LogKey Field, value is HTTP Header Field
func IFieldLogKeyKeyMap() (keyMap map[string]string) {
	return sKMap
}

func GetContextFromPeeker(p Peeker) Context {
	c := Background()
	var v []byte
	for hk, sk := range hKMap {
		v = p.Peek(hk)
		if len(v) > 0 {
			c.Set(sk, string(v))
		}
	}
	return c
}

func GetContextFromGetter(g Getter) Context {
	c := Background()
	var v string
	for hk, sk := range hKMap {
		v = g.Get(hk)
		if len(v) > 0 {
			c.Set(sk, v)
		}
	}
	return c
}

func GetContextFromGetHeader(g GetHeaderer) Context {
	c := Background()
	var v string
	for hk, sk := range hKMap {
		v = g.GetHeader(hk)
		if len(v) > 0 {
			c.Set(sk, v)
		}
	}
	return c
}

func (c *MapContext) SetHTTPHeaders(s Setter) {
	for hk, sk := range c.LogKeyMap() {
		s.Set(hk, sk)
	}
}

// LogKeyMap returns a map, key HTTP Header Field, value is the field value stored in Context
func (c *MapContext) LogKeyMap() (ret map[string]string) {
	ret = make(map[string]string)
	for sk, hk := range sKMap {
		v, _ := c.GetString(sk)
		if len(v) > 0 {
			ret[hk] = v
		}
	}
	return
}

// LogKeyFields returns a logrus.Fields, key HTTP Header Field, value is the field value stored in Context
func (c *MapContext) LogKeyFields() (ret logrus.Fields) {
	ret = make(logrus.Fields)
	for sk, hk := range sKMap {
		v, _ := c.GetString(sk)
		if len(v) > 0 {
			ret[hk] = v
		}
	}
	return
}

// LogKeySet sets Context with Header Key, Store in Context with LogKey Key
func (c *MapContext) LogKeySet(headerField, headerValue string) {
	if sk, ok := hKMap[headerField]; ok {
		c.Set(sk, headerValue)
	}
}

// GetCID return the CID, if not, generate one
func (c *MapContext) GetCID() (string, error) {
	cid, ok := c.GetString(LogKeyCID)
	if !ok {
		uuidv4, err := uuid.NewV4()
		if err != nil {
			return "", err
		}
		cid = uuidv4.String()
		c.Set(LogKeyCID, cid)
	}
	return cid, nil
}
