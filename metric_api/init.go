package metric_api

import (
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gitlab.com/cake/intercom"
	"gitlab.com/cake/go-project-template/gpt"
	"gitlab.com/cake/go-project-template/metric"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func preRegisterLabels() {
	metric.RegisterMetrics()
	for _, svc := range ServiceList {
		for _, t := range TypeList {
			metric.LabeledCounter.WithLabelValues(svc, t)
			metric.LabeledGauge.WithLabelValues(svc, t)
			metric.LabeledHistogram.WithLabelValues(svc, t)
			metric.LabeledSummary.WithLabelValues(svc, t)
		}
	}
}

func AddMetricEndpoint(rootGroup *gin.RouterGroup) {
	// settings
	apiTimeout := viper.GetDuration("http.api_timeout")

	MetricGroup := rootGroup.Group(gpt.APIMetricPath,
		intercom.AccessMiddleware(apiTimeout, gpt.GetNamespace()),
	)
	{
		MetricGroup.POST("/counter", counter)
		MetricGroup.POST("/gauge", gauge)
		MetricGroup.POST("/histogram", histogram)
		MetricGroup.POST("/summary", summary)
	}

	MetricLabelGroup := rootGroup.Group(gpt.APILabeledMetricPath,
		intercom.AccessMiddleware(apiTimeout, gpt.GetNamespace()),
	)
	{
		MetricLabelGroup.POST("/counter", labeledCounter)
		MetricLabelGroup.POST("/gauge", labeledGauge)
		MetricLabelGroup.POST("/histogram", labeledHistogram)
		MetricLabelGroup.POST("/summary", labeledSummary)
	}

	preRegisterLabels()
}
