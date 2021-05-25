package intercom

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/cake/goctx"
)

const (
	pageNotFound     = "page not found, method: %s uri: %s"
	methodNotAllowed = "method not allowed, method: %s uri: %s"
)

func NoRouteHandler(code int) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := Response{
			Code:    code,
			Message: fmt.Sprintf(pageNotFound, c.Request.Method, c.Request.Host+c.Request.URL.Path),
			CID:     c.GetHeader(goctx.HTTPHeaderCID),
		}
		c.Set(goctx.LogKeyErrorCode, code)
		c.AbortWithStatusJSON(http.StatusNotFound, response)
	}
}

func NoMethodHandler(code int) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := Response{
			Code:    code,
			Message: fmt.Sprintf(methodNotAllowed, c.Request.Method, c.Request.Host+c.Request.URL.Path),
			CID:     c.GetHeader(goctx.HTTPHeaderCID),
		}
		c.Set(goctx.LogKeyErrorCode, code)
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, response)
	}
}
