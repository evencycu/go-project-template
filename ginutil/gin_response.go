package ginutil

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gitlab.com/general-backend/gopkg"
	"gitlab.com/general-backend/m800log"
)

// Response defines the JSON RESTful response
type Response struct {
	Code    int         `json:"code"`
	Result  interface{} `json:"result,omitempty"`
	Message string      `json:"message,omitempty"`
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

// GinAllResponse defines the interface of error response with result
func GinAllResponse(c *gin.Context, result interface{}, err gopkg.CodeError) {
	response := Response{}
	response.Result = result
	response.Code = err.ErrorCode()
	response.Message = err.Error()
	c.JSON(http.StatusOK, response)
}

// GinOKError defines the interface of error response with HTTP200
func GinOKError(c *gin.Context, err gopkg.CodeError) {
	response := Response{}
	response.Code = err.ErrorCode()
	response.Message = err.Error()

	// here, we check level first, because we have to do requestDump
	if m800log.GetLogger().Level >= logrus.DebugLevel {
		requestDump, _ := httputil.DumpRequest(c.Request, true)
		m800log.Debug(GetContextFromGin(c), "Gin Request:", string(requestDump), "Error:", err.ErrorCode())
	}

	c.AbortWithStatusJSON(http.StatusOK, response)
}

// GinError defines the interface of error response with HTTP code by setting ErrorHttpStatusMapping
func GinError(c *gin.Context, err gopkg.CodeError) {
	status, ok := ErrorHttpStatusMapping.Get(err.ErrorCode())
	if !ok {
		status = http.StatusInternalServerError
	}

	response := Response{
		Code:    err.ErrorCode(),
		Message: err.Error(),
	}
	// here, we check level first, because we have to do requestDump
	if m800log.GetLogger().Level >= logrus.DebugLevel {
		requestDump, _ := httputil.DumpRequest(c.Request, true)
		m800log.Debug(GetContextFromGin(c), "Gin Request:", string(requestDump), "Erorr:", err.ErrorCode())
	}

	c.AbortWithStatusJSON(status, response)
}
