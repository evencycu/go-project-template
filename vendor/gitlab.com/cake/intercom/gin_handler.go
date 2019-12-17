package intercom

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/cake/goctx"
)

const (
	pageNotFound     = "page not found"
	methodNotAllowed = "method not allowed"
)

func NoRouteHandler(code int) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := Response{
			Code:    code,
			Message: pageNotFound,
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
			Message: methodNotAllowed,
			CID:     c.GetHeader(goctx.HTTPHeaderCID),
		}
		c.Set(goctx.LogKeyErrorCode, code)
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, response)
	}
}
