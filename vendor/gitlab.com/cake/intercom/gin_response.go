package intercom

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"gitlab.com/general-backend/goctx"
	"gitlab.com/general-backend/gopkg"
)

// Response defines the JSON RESTful response
type Response struct {
	Code    int         `json:"code"`
	Result  interface{} `json:"result,omitempty"`
	Message string      `json:"message,omitempty"`
	CID     string      `json:"cid,omitempty"`
}

type ListResponse struct {
	Code    int         `json:"code"`
	Result  interface{} `json:"result,omitempty"`
	Message string      `json:"message,omitempty"`
	Total   int         `json:"total"`
	Offset  int         `json:"offset"`
	Count   int         `json:"count"`
	CID     string      `json:"cid,omitempty"`
}

type JsonResponse struct {
	Code    int             `json:"code"`
	Result  json.RawMessage `json:"result,omitempty"`
	Message string          `json:"message,omitempty"`
	Total   int             `json:"total"`
	Offset  int             `json:"offset"`
	Count   int             `json:"count"`
	CID     string          `json:"cid,omitempty"`
}

type ImmutableMap struct {
	data   map[int]int
	rwLock sync.RWMutex
}

func (im *ImmutableMap) Set(key, value int) error {
	im.rwLock.Lock()
	defer im.rwLock.Unlock()
	if _, ok := im.data[key]; ok {
		return fmt.Errorf("existed key")
	}
	im.data[key] = value
	return nil
}

func (im *ImmutableMap) Get(key int) (int, bool) {
	v, ok := im.data[key]
	return v, ok
}

var ErrorHttpStatusMapping = &ImmutableMap{
	data:   make(map[int]int),
	rwLock: sync.RWMutex{},
}

// GinOKResponse defines the interface of success response
func GinOKResponse(c *gin.Context, result interface{}) {
	response := Response{}
	response.Result = result
	c.JSON(http.StatusOK, response)
}

// GinOKListResponse defines the interface of success response
func GinOKListResponse(c *gin.Context, result interface{}, total, offset, count int) {
	response := ListResponse{}
	response.Result = result
	response.Total = total
	response.Offset = offset
	response.Count = count
	c.JSON(http.StatusOK, response)
}

// GinAllResponse defines the interface of error response with result
func GinAllResponse(c *gin.Context, result interface{}, err gopkg.CodeError) {
	response := Response{}
	response.Result = result
	response.Code = err.ErrorCode()
	response.Message = err.Error()
	response.CID = c.GetHeader(goctx.HTTPHeaderCID)
	c.JSON(http.StatusOK, response)
}

// GinOKError defines the interface of error response
func GinOKError(c *gin.Context, err gopkg.CodeError) {
	response := Response{}
	response.Code = err.ErrorCode()
	response.Message = err.Error()
	response.CID = c.GetHeader(goctx.HTTPHeaderCID)
	ctx := GetContextFromGin(c)
	ctx.Set(goctx.LogKeyErrorCode, err.ErrorCode())
	dumpRequest(ctx, ErrorTraceLevel, c.Request)
	c.Set(TraceTagGinError, err.FullError())
	c.AbortWithStatusJSON(http.StatusOK, response)
}

// GinError defines the interface of error response
func GinError(c *gin.Context, err gopkg.CodeError) {
	status, ok := ErrorHttpStatusMapping.Get(err.ErrorCode())
	if !ok {
		status = http.StatusInternalServerError
	}

	response := Response{
		Code:    err.ErrorCode(),
		Message: err.Error(),
		CID:     c.GetHeader(goctx.HTTPHeaderCID),
	}
	ctx := GetContextFromGin(c)
	ctx.Set(goctx.LogKeyErrorCode, err.ErrorCode())
	dumpRequest(ctx, ErrorTraceLevel, c.Request)
	c.Set(TraceTagGinError, err.FullError())
	c.AbortWithStatusJSON(status, response)
}

// GinErrorCodeMsg defines the interface of error response
func GinErrorCodeMsg(c *gin.Context, code int, msg string) {
	status, ok := ErrorHttpStatusMapping.Get(code)
	if !ok {
		status = http.StatusInternalServerError
	}

	response := Response{
		Code:    code,
		Message: msg,
		CID:     c.GetHeader(goctx.HTTPHeaderCID),
	}
	ctx := GetContextFromGin(c)
	ctx.Set(goctx.LogKeyErrorCode, code)
	dumpRequest(ctx, ErrorTraceLevel, c.Request)
	c.Set(TraceTagGinError, msg)
	c.AbortWithStatusJSON(status, response)
}
