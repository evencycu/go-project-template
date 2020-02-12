package err

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/cake/go-project-template/gpt"
	"gitlab.com/cake/goctx"
	"gitlab.com/cake/gopkg"
	"gitlab.com/cake/intercom"
	"gitlab.com/cake/m800log"
)

func randomErr(c *gin.Context) {
	ctx := intercom.GetContextFromGin(c)
	handlerName := c.HandlerName()

	err := genErr()
	m800log.Errorf(ctx, "[%s] error occurs: %s", handlerName, err)

	if ok := errors.As(err, &numberError{}); ok && (errors.Is(err, ErrTooLarge) || errors.Is(err, ErrTooSmall)) {
		intercom.GinError(c, gopkg.NewWrappedCarrierCodeError(gpt.CodeBadRequest, "request error", err))
		return
	} else if ok := errors.As(err, &sessionError{}); ok && errors.Is(err, ErrPermission) {
		intercom.GinError(c, gopkg.NewWrappedCarrierCodeError(gpt.CodeForbidden, "permission error", err))
		return
	} else if ok := errors.As(err, &sessionError{}); ok && errors.Is(err, ErrClosed) {
		intercom.GinError(c, gopkg.NewWrappedCarrierCodeError(gpt.CodeInternalServerError, "network error", err))
		return
	}

	intercom.GinError(c, gopkg.NewWrappedCarrierCodeError(gpt.CodeInternalServerError, "database is burning", err))
	return
}

func upstreamErr(c *gin.Context) {
	ctx := intercom.GetContextFromGin(c)
	handlerName := c.HandlerName()

	resp, err := upstreamToRandomError(ctx)
	if err != nil {
		m800log.Errorf(ctx, "[%s] error occurs: %s", handlerName, err)
		intercom.GinError(c, gopkg.NewWrappedCarrierCodeError(gpt.CodeUpstreamSpecific, "upstream error", err))
		return
	}

	intercom.GinOKResponse(c, resp)
}

func upstreamToRandomError(ctx goctx.Context) (result *intercom.JsonResponse, err gopkg.CodeError) {
	var req *http.Request

	uri := gpt.APIErrorPath
	req, err = intercom.HTTPNewRequest(
		ctx, http.MethodGet,
		fmt.Sprintf("http://localhost:8999%s", uri),
		nil)
	if err != nil {
		return
	}

	result, err = intercom.M800Do(ctx, req)

	return
}

func genErr() error {
	getErrFunc := getErrFuncList[rand.Intn(len(getErrFuncList))]
	return getErrFunc()
}
