package semconv

import (
	"net/textproto"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.com/cake/goctx"
	"go.opentelemetry.io/otel/attribute"
)

var (
	traceKeyMap = map[string]attribute.Key{
		goctx.HTTPHeaderCID:            M800CIDKey,
		goctx.HTTPHeaderEID:            M800EIDKey,
		goctx.HTTPHeaderService:        M800ServiceKey,
		goctx.HTTPHeaderServiceHome:    M800ServiceHomeKey,
		goctx.HTTPHeaderDeviceID:       M800ClientDeviceIDKey,
		goctx.HTTPHeaderWebTabID:       M800ClientWebIDKey,
		goctx.HTTPHeaderClientPlatform: M800ClientPlatformKey,
		goctx.HTTPHeaderUserGroup:      M800UserGroupKey,
		goctx.HTTPHeaderUserAnms:       M800UserAnonymousKey,
		goctx.HTTPHeaderCustomRole:     M800UserCustomRoleKey,
	}
)

func M800AttributesFromHTTPRequest(c *gin.Context) []attribute.KeyValue {
	attrs := []attribute.KeyValue{}

	header := c.Request.Header

	for headerKey, traceKey := range traceKeyMap {
		v := header[textproto.CanonicalMIMEHeaderKey(headerKey)]
		if len(v) > 0 {
			attrs = append(attrs, traceKey.String(strings.Join(v, ",")))
		}
	}
	if role := header.Get(goctx.HTTPHeaderUserRole); role != "" {
		attrs = append(attrs, M800UserRoleKey.StringSlice(strings.Split(role, ",")))
	}

	return attrs
}

func M800ErrorCodeFromResponse(c *gin.Context) []attribute.KeyValue {
	attrs := []attribute.KeyValue{}

	if eCode, ok := c.Get(goctx.LogKeyErrorCode); ok {
		attrs = append(attrs, M800ErrorCodeKey.Int(eCode.(int)))
	}
	if eMsg, ok := c.Get(goctx.LogKeyErrorMessage); ok {
		attrs = append(attrs, M800ErrorMessageKey.String(eMsg.(string)))
	}

	return attrs
}
