package ginutil

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
	"gitlab.com/general-backend/goctx"
	"gitlab.com/general-backend/m800log"
)

func ParseJSON(ctx goctx.Context, readCloser io.ReadCloser, v interface{}) error {
	if readCloser == nil {
		return fmt.Errorf("nil input")
	}

	defer readCloser.Close()
	raw, err := ioutil.ReadAll(readCloser)
	if err != nil {
		return err
	}
	err = json.Unmarshal(raw, v)
	if err != nil {
		m800log.Error(ctx, err.Error()+" : "+string(raw))
	}
	return err
}

// TBD: should we add debug log here for error handling?
func GetStringFromIO(readCloser io.ReadCloser) string {
	defer readCloser.Close()
	bytes, _ := ioutil.ReadAll(readCloser)
	return string(bytes)
}

// LogDumpRequest check level first, because we don't want to waste resource on DumpRequest
func LogDumpRequest(ctx goctx.Context, req *http.Request, v ...interface{}) {
	if m800log.GetLogger().Level >= logrus.DebugLevel {
		requestDump, _ := httputil.DumpRequest(req, true)
		m800log.Debug(ctx, "Request:", string(requestDump), v)
	}
}

func LogDumpResponse(ctx goctx.Context, resp *http.Response, v ...interface{}) {
	if m800log.GetLogger().Level >= logrus.DebugLevel {
		respDump, _ := httputil.DumpResponse(resp, true)
		m800log.Debug(ctx, "Response:", string(respDump), v)
	}
}

// GetContextFromGin is a util generated the goctx from gin.Context
func GetContextFromGin(c *gin.Context) goctx.Context {
	if ctxI, gok := c.Get(goctx.ContextKey); gok {
		ctx, rok := ctxI.(goctx.Context)
		if rok {
			return ctx
		}
	}
	return goctx.GetContextFromGetHeader(c)
}

func GenUUID(ctx goctx.Context) uuid.UUID {
	uuid, err := uuid.NewV4()
	if err != nil {
		m800log.Error(ctx, fmt.Sprintf("failed to generate UUID: %v", err))
	}
	m800log.Debug(ctx, fmt.Sprintf("generated Version 4 UUID %v", uuid))
	return uuid
}
