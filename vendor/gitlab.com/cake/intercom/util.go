package intercom

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"gitlab.com/general-backend/gopkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gitlab.com/general-backend/goctx"
	"gitlab.com/general-backend/m800log"
)

// ParseJSONReq read body, and put the body back to http req
func ParseJSONReq(ctx goctx.Context, req *http.Request, v interface{}) gopkg.CodeError {
	if req.Body == nil {
		return gopkg.NewCodeError(CodeParseJSON, "nil body")
	}

	raw, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return gopkg.NewCodeError(CodeParseJSON, err.Error())
	}
	req.Body.Close()
	req.Body = ioutil.NopCloser(bytes.NewReader(raw))
	err = json.Unmarshal(raw, v)
	if err != nil {
		m800log.Errorf(ctx, "err:%v, req body: %s", err.Error(), string(raw))
		return gopkg.NewCodeError(CodeParseJSON, err.Error())
	}
	return nil
}

func readFromReadCloser(readCloser io.ReadCloser) ([]byte, gopkg.CodeError) {
	if readCloser == nil {
		return nil, gopkg.NewCodeError(CodeReadAll, "nil readCloser")
	}

	raw, err := ioutil.ReadAll(readCloser)
	if err != nil {
		return nil, gopkg.NewCodeError(CodeReadAll, err.Error())
	}
	readCloser.Close()
	return raw, nil
}

// ParseJSONReadCloser
func ParseJSONReadCloser(ctx goctx.Context, readCloser io.ReadCloser, v interface{}) gopkg.CodeError {
	raw, err := readFromReadCloser(readCloser)
	if err != nil {
		return err
	}

	errJSON := json.Unmarshal(raw, v)
	if errJSON != nil {
		m800log.Errorf(ctx, "err:%v, req body: %s", errJSON.Error(), raw)
		return gopkg.NewCodeError(CodeParseJSON, errJSON.Error())
	}
	return nil
}

// ParseJSON
func ParseJSON(ctx goctx.Context, data []byte, v interface{}) gopkg.CodeError {
	err := json.Unmarshal(data, v)
	if err != nil {
		m800log.Errorf(ctx, "err:%v, input: %s", err.Error(), string(data))
		return gopkg.NewCodeError(CodeParseJSON, err.Error())
	}
	return nil
}

// GetStringFromIO returns the string from given io.ReadCloser
func GetStringFromIO(readCloser io.ReadCloser) string {
	defer readCloser.Close()
	bytes, _ := ioutil.ReadAll(readCloser)
	return string(bytes)
}

func dumpRequest(ctx goctx.Context, level logrus.Level, req *http.Request) {
	req.Header.Del(HeaderAuthorization)
	requestDump, _ := httputil.DumpRequest(req, true)
	m800log.Log(ctx, level, "DumpRequest:\n", string(requestDump))
}

func dumpRequestAndBody(ctx goctx.Context, level logrus.Level, req *http.Request, body []byte) {
	req.Header.Del(HeaderAuthorization)
	requestDump, _ := httputil.DumpRequest(req, false)
	m800log.Logf(ctx, level, "DumpRequest:\n%s\nBody:%s", requestDump, body)
}

// LogDumpRequest check level first, because we don't want to waste resource on DumpRequest
func LogDumpRequest(ctx goctx.Context, level logrus.Level, req *http.Request) {
	if m800log.GetLogger().Level >= level {
		dumpRequest(ctx, level, req)
	}
}

// LogDumpRequestAndBody
func LogDumpRequestAndBody(ctx goctx.Context, level logrus.Level, req *http.Request, body []byte) {
	if m800log.GetLogger().Level >= level {
		dumpRequestAndBody(ctx, level, req, body)
	}
}

// LogDumpResponse
func LogDumpResponse(ctx goctx.Context, level logrus.Level, resp *http.Response) {
	if m800log.GetLogger().Level >= level {
		respDump, _ := httputil.DumpResponse(resp, true)
		m800log.Log(ctx, level, "DumpResponse:\n", string(respDump))
	}
}

// LogDumpResponseAndBody
func LogDumpResponseAndBody(ctx goctx.Context, level logrus.Level, resp *http.Response, body []byte) {
	if m800log.GetLogger().Level >= level {
		respDump, _ := httputil.DumpResponse(resp, false)
		m800log.Logf(ctx, level, "DumpResponse:\n%s\nBody:%s", respDump, body)
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
