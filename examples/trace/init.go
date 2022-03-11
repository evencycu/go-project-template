package trace

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gitlab.com/cake/go-project-template/gpt"
	"gitlab.com/cake/golibs/intercom"
	"gitlab.com/cake/m800log"
	"gitlab.com/cake/mgopool/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func AddMetricEndpoint(rootGroup *gin.RouterGroup) {
	// settings
	apiTimeout := viper.GetDuration("http.api_timeout")

	TraceGroup := rootGroup.Group(gpt.APITracePath,
		intercom.AccessMiddleware(apiTimeout, gpt.GetNamespace()),
	)
	{
		TraceGroup.GET("/:name", nameHandler)
	}
}

func nameHandler(c *gin.Context) {
	ctx := intercom.GetContextFromGin(c)
	test := c.Param("name")

	span := ctx.GetSpan()
	bag := ctx.GetBaggage()

	err := mgopool.Insert(ctx, "testDB", "testCollection", sample{"1", test})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		m800log.Error(ctx, "==================== error:", err.Error())
	}

	span.SetAttributes(attribute.KeyValue{Key: "baggage", Value: attribute.StringValue(bag.String())})
	span.SetAttributes(attribute.KeyValue{Key: "name", Value: attribute.StringValue(test)})

	c.JSON(404, gin.H{
		"code":    2990000,
		"message": test + " not found",
	})
}

type sample struct {
	ID   string `bson:"_id"`
	Name string `bson:"name"`
}
