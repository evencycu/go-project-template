package ginutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gitlab.com/general-backend/gopkg"
)

func TestGinResponse(t *testing.T) {
	router := gin.Default()
	router.GET("/ok", func(c *gin.Context) {
		GinOKResponse(c, "ok")
	})
	router.GET("/error", func(c *gin.Context) {
		GinOKError(c, gopkg.NewCarrierCodeError(1234567, "error"))
	})

	router.GET("/init", func(c *gin.Context) {
		GinAllResponse(c, "abc", gopkg.NewCarrierCodeError(1234567, "error"))
	})

	req, _ := http.NewRequest(http.MethodGet, "/ok", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	// http 200
	assert.Equal(t, http.StatusOK, resp.Code)
	result := &Response{}

	err := json.Unmarshal(resp.Body.Bytes(), result)
	assert.NoError(t, err)
	assert.Equal(t, "ok", result.Result.(string))
	assert.Equal(t, 0, result.Code)

	req2, _ := http.NewRequest(http.MethodGet, "/error", nil)
	resp2 := httptest.NewRecorder()
	router.ServeHTTP(resp2, req2)
	assert.Equal(t, http.StatusOK, resp2.Code)

	result = &Response{}
	err = json.Unmarshal(resp2.Body.Bytes(), result)

	assert.NoError(t, err)
	assert.Equal(t, "error", result.Message)
	assert.Equal(t, 1234567, result.Code)

	req3, _ := http.NewRequest(http.MethodGet, "/init", nil)
	resp3 := httptest.NewRecorder()
	router.ServeHTTP(resp3, req3)
	assert.Equal(t, http.StatusOK, resp3.Code)

	result = &Response{}
	err = json.Unmarshal(resp3.Body.Bytes(), result)

	assert.NoError(t, err)
	assert.Equal(t, "error", result.Message)
	assert.Equal(t, "abc", result.Result)
	assert.Equal(t, 1234567, result.Code)
}
