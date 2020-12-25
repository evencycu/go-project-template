package goctx

import (
	"net/http"
	"net/textproto"
	"strings"
)

const (
	// ContextKey is the common field if we have to put goctx in other context object
	ContextKey = "X-Ctx"

	// for trace
	HTTPHeaderTrace         = "uber-trace-id"
	HTTPHeaderJaegerDebug   = "jaeger-debug-id"
	HTTPHeaderJaegerBaggage = "jaeger-baggage"

	// for http passing
	HTTPHeaderCID            = "x-correlation-id"
	HTTPHeaderEID            = "x-m800-eid"
	HTTPHeaderService        = "x-m800-svc"
	HTTPHeaderClientIP       = "x-forwarded-for"
	HTTPHeaderClientPort     = "x-forwarded-port"
	HTTPHeaderClientPlatform = "x-m800-platform"
	HTTPHeaderDeviceID       = "x-m800-deviceid"
	HTTPHeaderWebTabID       = "x-m800-tabid"

	HTTPHeaderUserHome       = "x-m800-usr-home"
	HTTPHeaderServiceHome    = "x-m800-svc-home"
	HTTPHeaderServiceType    = "x-m800-svc-type"
	HTTPHeaderUserRole       = "x-m800-usr-role"
	HTTPHeaderUserGroup      = "x-m800-usr-group"
	HTTPHeaderUserAnms       = "x-m800-usr-anms"
	HTTPHeaderCustomRole     = "x-m800-custom-role" // This field is for customer-defined role.
	HTTPHeaderUserDataScope  = "x-m800-usr-data-scope"
	HTTPHeaderUserDepartment = "x-m800-usr-dept"
	HTTPHeaderSuperAdmin     = "x-m800-usr-super-admin"

	HTTPHeaderInternalCaller = "x-m800-internal-caller"

	// for logger
	LogKeyTrace         = "uti"
	LogKeyJaegerDebug   = "jdi"
	LogKeyJaegerBaggage = "jb"

	// FROM m800 log format document: https://issuetracking.maaii.com:9443/pages/viewpage.action?pageId=65128541#Loggingformatdesign(v1.0)-Backendteam(BE)

	// LogKeyCID is the cid field key
	LogKeyCID            = "cid"
	LogKeyEID            = "eid"
	LogKeyService        = "svc"
	LogKeyClientIP       = "clientIP"
	LogKeyClientPort     = "clientPort"
	LogKeyClientPlatform = "platform"
	LogKeyDeviceID       = "deviceID"
	LogKeyWebTabID       = "webTabID"

	LogKeyUserHome                    = "usrHome"
	LogKeyServiceHome                 = "svcHome"
	LogKeyServiceType                 = "svcType"
	LogKeyUserRole                    = "usrRole"
	LogKeyUserGroup                   = "usrGroup"
	LogKeyUserAnms                    = "usrAnms"
	LogKeyCustomRole                  = "customRole"
	LogKeyUserDepartment              = "usrDept"
	LogKeyUserDataScope               = "usrDataScope"
	LogKeyInternalCaller              = "internalCaller"
	LogKeyAuditFeatureName            = "feature"
	LogKeyAuditNeedChangelog          = "changelog"
	LogKeyRolePermissionAPIMappingKey = "permitKey"
	LogKeySuperAdmin                  = "superadmin"
	// LogKeyTimestamp is the time field key
	LogKeyTimestamp = "time"
	// LogKeyLevel is the level field key
	LogKeyLevel = "level"
	// LogKeyMessage is the message field key
	LogKeyMessage          = "message"
	LogKeyEntryType        = "entryType"
	LogKeyURI              = "uri"
	LogKeyMethod           = "method"
	LogKeyCallFunc         = "xfunc"
	LogKeyCallLine         = "xline"
	LogKeyHTTPMethod       = "httpMethod"
	LogKeyInstance         = "instanceID"
	LogKeyVersion          = "vid"
	LogKeyCase             = "caseType"
	LogKeyNamespace        = "ns"
	LogKeyEnv              = "env"
	LogKeyErrorCode        = "eCode"
	LogKeyErrorMessage     = "eMessage"
	LogKeyErrorType        = "eType"
	LogKeyWrapErrorCode    = "ueCode"
	LogKeyWrapErrorMessage = "ueMessage"

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

// Deprecated
// remove for code simplify
// GetHeaderer interface designed for getting goctx from gin.Context
type GetHeaderer interface {
	GetHeader(key string) string
}

var (
	sKMap, hKMap map[string]string
)

func init() {
	// role special handle
	sKMap = map[string]string{
		LogKeyTrace:          HTTPHeaderTrace,
		LogKeyJaegerDebug:    HTTPHeaderJaegerDebug,
		LogKeyJaegerBaggage:  HTTPHeaderJaegerBaggage,
		LogKeyCID:            HTTPHeaderCID,
		LogKeyEID:            HTTPHeaderEID,
		LogKeyService:        HTTPHeaderService,
		LogKeyDeviceID:       HTTPHeaderDeviceID,
		LogKeyWebTabID:       HTTPHeaderWebTabID,
		LogKeyClientPlatform: HTTPHeaderClientPlatform,
		LogKeyClientIP:       HTTPHeaderClientIP,
		LogKeyClientPort:     HTTPHeaderClientPort,
		LogKeyUserHome:       HTTPHeaderUserHome,
		LogKeyServiceHome:    HTTPHeaderServiceHome,
		LogKeyServiceType:    HTTPHeaderServiceType,
		// LogKeyUserRole:       HTTPHeaderUserRole,
		LogKeyUserGroup:      HTTPHeaderUserGroup,
		LogKeyUserAnms:       HTTPHeaderUserAnms,
		LogKeyCustomRole:     HTTPHeaderCustomRole,
		LogKeyInternalCaller: HTTPHeaderInternalCaller,
		LogKeyUserDataScope:  HTTPHeaderUserDataScope,
		LogKeyUserDepartment: HTTPHeaderUserDepartment,
		LogKeySuperAdmin:     HTTPHeaderSuperAdmin,
	}

	hKMap = map[string]string{
		HTTPHeaderTrace:          LogKeyTrace,
		HTTPHeaderJaegerDebug:    LogKeyJaegerDebug,
		HTTPHeaderJaegerBaggage:  LogKeyJaegerBaggage,
		HTTPHeaderCID:            LogKeyCID,
		HTTPHeaderEID:            LogKeyEID,
		HTTPHeaderService:        LogKeyService,
		HTTPHeaderClientIP:       LogKeyClientIP,
		HTTPHeaderClientPort:     LogKeyClientPort,
		HTTPHeaderDeviceID:       LogKeyDeviceID,
		HTTPHeaderWebTabID:       LogKeyWebTabID,
		HTTPHeaderClientPlatform: LogKeyClientPlatform,
		HTTPHeaderUserHome:       LogKeyUserHome,
		HTTPHeaderServiceHome:    LogKeyServiceHome,
		HTTPHeaderServiceType:    LogKeyServiceType,
		// HTTPHeaderUserRole:       LogKeyUserRole,
		HTTPHeaderUserGroup:      LogKeyUserGroup,
		HTTPHeaderUserAnms:       LogKeyUserAnms,
		HTTPHeaderCustomRole:     LogKeyCustomRole,
		HTTPHeaderInternalCaller: LogKeyInternalCaller,
		HTTPHeaderUserDataScope:  LogKeyUserDataScope,
		HTTPHeaderUserDepartment: LogKeyUserDepartment,
		HTTPHeaderSuperAdmin:     LogKeySuperAdmin,
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

func GetContextFromMap(p map[string]interface{}) Context {
	c := Background()
	for k, v := range p {
		c.Set(k, v)
	}
	return c
}

func GetContextFromMapString(p map[string]string) Context {
	c := Background()
	for k, v := range p {
		if k == LogKeyUserRole {
			c.Set(k, strings.Split(v, ","))
		} else {
			c.Set(k, v)
		}
	}
	return c
}

func GetContextFromHeaderKeyMap(p map[string]string) Context {
	c := Background()
	for hk, sk := range hKMap {
		v := p[hk]
		if len(v) > 0 {
			c.Set(sk, v)
		}
	}
	return c
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

	if role := g.Get(HTTPHeaderUserRole); role != "" {
		c.Set(LogKeyUserRole, strings.Split(role, ","))
	}

	return c
}

// Deprecated
// remove for code simplify
func GetContextFromGetHeader(g GetHeaderer) Context {
	c := Background()
	for hk, sk := range hKMap {
		v := g.GetHeader(hk)
		if len(v) > 0 {
			c.Set(sk, v)
		}
	}
	if role := g.GetHeader(HTTPHeaderUserRole); role != "" {
		c.Set(LogKeyUserRole, strings.Split(role, ","))
	}

	return c
}

func GetContextFromHeader(g http.Header) Context {
	c := Background()
	for hk, sk := range hKMap {
		v := g[textproto.CanonicalMIMEHeaderKey(hk)]
		if len(v) > 0 {
			c.Set(sk, strings.Join(v, ","))
		}
	}
	if role := g.Get(HTTPHeaderUserRole); role != "" {
		c.Set(LogKeyUserRole, strings.Split(role, ","))
	}

	return c
}

// CopyContext copy context value and span
func CopyContext(ctx Context) Context {
	newCtx := Background()
	for k, v := range ctx.Map() {
		newCtx.Set(k, v)
	}
	if ctx.GetSpan() != nil {
		newCtx.SetSpan(ctx.GetSpan())
	}
	return newCtx
}
