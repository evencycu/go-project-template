package metric_api

import (
	"math/rand"

	"github.com/gin-gonic/gin"
	"gitlab.com/cake/gopkg"
	"gitlab.com/cake/intercom"
	"gitlab.com/cake/m800log"
	"gitlab.com/cake/go-project-template/gpt"
	"gitlab.com/cake/go-project-template/metric"
)

func gauge(c *gin.Context) {
	ctx := intercom.GetContextFromGin(c)
	handlerName := c.HandlerName()

	var payload struct {
		Value float64
	}
	if err := intercom.ParseJSONGin(ctx, c, &payload); err != nil {
		m800log.Debugf(ctx, "[%s] invalid request: %+v, err: %+v", handlerName, payload, err)
		intercom.GinError(c, gopkg.NewCodeError(gpt.CodeBadRequest, err.Error()))
		return
	}

	metric.Gauge.Set(payload.Value)

	intercom.GinOKResponse(c, nil)
}

func labeledGauge(c *gin.Context) {
	ctx := intercom.GetContextFromGin(c)
	handlerName := c.HandlerName()

	var payload struct {
		Value float64
	}
	if err := intercom.ParseJSONGin(ctx, c, &payload); err != nil {
		m800log.Debugf(ctx, "[%s] invalid request: %+v, err: %+v", handlerName, payload, err)
		intercom.GinError(c, gopkg.NewCodeError(gpt.CodeBadRequest, err.Error()))
		return
	}

	metric.LabeledGauge.WithLabelValues(
		ServiceList[rand.Intn(len(ServiceList))],
		TypeList[rand.Intn(len(TypeList))],
	).Set(payload.Value)

	intercom.GinOKResponse(c, nil)
}
