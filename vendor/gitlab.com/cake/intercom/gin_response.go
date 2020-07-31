package intercom

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"gitlab.com/cake/goctx"
	"gitlab.com/cake/gopkg"
)

var defaultHTTPErrorCode = http.StatusInternalServerError

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
	Code       int             `json:"code"`
	HTTPStatus int             `json:"-"`
	Result     json.RawMessage `json:"result,omitempty"`
	Message    string          `json:"message,omitempty"`
	Total      int             `json:"total"`
	Offset     int             `json:"offset"`
	Count      int             `json:"count"`
	CID        string          `json:"cid,omitempty"`
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

func SetDefaultHTTPErrorCode(httpCode int) error {
	if ok := http.StatusText(httpCode); ok == "" {
		return fmt.Errorf("invalid http code")
	}
	defaultHTTPErrorCode = httpCode
	return nil
}

// GinOKResponse defines the interface of success response
func GinOKResponse(c *gin.Context, result interface{}) {
	response := Response{}
	response.Result = result
	response.CID = c.GetHeader(goctx.HTTPHeaderCID)
	c.AbortWithStatusJSON(http.StatusOK, response)
}

// GinOKListResponse defines the interface of success response
func GinOKListResponse(c *gin.Context, result interface{}, total, offset, count int) {
	response := ListResponse{}
	response.Result = result
	response.Total = total
	response.Offset = offset
	response.Count = count
	response.CID = c.GetHeader(goctx.HTTPHeaderCID)
	c.AbortWithStatusJSON(http.StatusOK, response)
}

// GinAllResponse defines the interface of error response with result
func GinAllResponse(c *gin.Context, result interface{}, err gopkg.CodeError) {
	response := Response{}
	response.Result = result
	response.Code = err.ErrorCode()
	response.Message = err.ErrorMsg()
	response.CID = c.GetHeader(goctx.HTTPHeaderCID)
	c.Set(goctx.LogKeyErrorCode, err.ErrorCode())
	setWrappedErrorCode(c, err)
	c.AbortWithStatusJSON(http.StatusOK, response)
}

// GinOKError defines the interface of error response
func GinOKError(c *gin.Context, err gopkg.CodeError) {
	response := Response{}
	response.Code = err.ErrorCode()
	response.Message = err.ErrorMsg()
	response.CID = c.GetHeader(goctx.HTTPHeaderCID)
	c.Set(goctx.LogKeyErrorCode, err.ErrorCode())
	setWrappedErrorCode(c, err)
	c.AbortWithStatusJSON(http.StatusOK, response)
}

// GinError defines the interface of error response
func GinError(c *gin.Context, err gopkg.CodeError) {
	status, ok := ErrorHttpStatusMapping.Get(err.ErrorCode())
	if !ok {
		status = defaultHTTPErrorCode
	}

	response := Response{
		Code:    err.ErrorCode(),
		Message: err.ErrorMsg(),
		CID:     c.GetHeader(goctx.HTTPHeaderCID),
	}
	c.Set(goctx.LogKeyErrorCode, err.ErrorCode())
	setWrappedErrorCode(c, err)
	c.AbortWithStatusJSON(status, response)
}

// GinAllErrorResponse defines the interface of error response with result
func GinAllErrorResponse(c *gin.Context, result interface{}, err gopkg.CodeError) {
	status, ok := ErrorHttpStatusMapping.Get(err.ErrorCode())
	if !ok {
		status = defaultHTTPErrorCode
	}
	response := Response{}
	response.Result = result
	response.Code = err.ErrorCode()
	response.Message = err.ErrorMsg()
	response.CID = c.GetHeader(goctx.HTTPHeaderCID)
	c.Set(goctx.LogKeyErrorCode, err.ErrorCode())
	setWrappedErrorCode(c, err)
	c.AbortWithStatusJSON(status, response)
}

// GinErrorCodeMsg defines the interface of error response
func GinErrorCodeMsg(c *gin.Context, code int, msg string) {
	status, ok := ErrorHttpStatusMapping.Get(code)
	if !ok {
		status = defaultHTTPErrorCode
	}

	response := Response{
		Code:    code,
		Message: msg,
		CID:     c.GetHeader(goctx.HTTPHeaderCID),
	}
	c.Set(goctx.LogKeyErrorCode, code)
	c.AbortWithStatusJSON(status, response)
}

// GinErrorStatus return with the given HTTP status code, and error response
func GinErrorStatus(c *gin.Context, status int, err gopkg.CodeError) {
	response := Response{
		Code:    err.ErrorCode(),
		Message: err.ErrorMsg(),
		CID:     c.GetHeader(goctx.HTTPHeaderCID),
	}
	c.Set(goctx.LogKeyErrorCode, err.ErrorCode())
	setWrappedErrorCode(c, err)
	c.AbortWithStatusJSON(status, response)
}

func setWrappedErrorCode(c *gin.Context, err gopkg.CodeError) {
	var carrierErr gopkg.CarrierCodeError
	if errors.As(err, &carrierErr) {
		if wrapErr := carrierErr.Unwrap(); wrapErr != nil {
			if errors.As(wrapErr, &carrierErr) {
				c.Set(goctx.LogKeyWrapErrorCode, carrierErr.ErrorCode())
			}
		}
	}
}
