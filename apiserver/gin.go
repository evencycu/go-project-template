package apiserver

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	ginprometheus "gitlab.com/cake/gin-prometheus"
	new_err "gitlab.com/cake/go-project-template/err"
	"gitlab.com/cake/go-project-template/gpt"
	"gitlab.com/cake/go-project-template/metric_api"
	"gitlab.com/cake/goctx"
	"gitlab.com/cake/intercom"
	"gitlab.com/cake/m800log"
	"gitlab.com/cake/mgopool"
)

var (
	metricSystem = "gin"
)

func InitGinServer(ctx goctx.Context) *http.Server {
	// Create gin http server.
	gin.SetMode(viper.GetString("http.mode"))
	router := gin.New()
	router.Use(intercom.M800Recovery(gpt.CodeInternalServerError))

	// Add gin prometheus metrics
	p, err := ginprometheus.NewPrometheus(metricSystem,
		ginprometheus.HistogramMetrics(metricSystem, ginprometheus.DefaultDurationBucket, ginprometheus.DefaultSizeBucket),
		ginprometheus.HistogramHandleFunc())
	if err != nil {
		panic("gin prometheus init error:" + err.Error())
	}
	p.Use(router)

	// general service for debugging
	router.GET("/health", health)
	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gpt.GetVersion())
	})
	router.GET("/config", appConfig)
	router.GET("/mongo", mongoInfo)
	router.GET("/ready", ready)

	rootGroup := router.Group("")

	// Add application API
	new_err.AddErrorEndpoint(rootGroup)
	metric_api.AddMetricEndpoint(rootGroup)

	port := viper.GetString("http.port")
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  viper.GetDuration("http.read_timeout"),
		WriteTimeout: viper.GetDuration("http.write_timeout"),
	}

	go func() {
		m800log.Infof(ctx, "Server is running and listening port: %s", port)
		// service connections
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	return httpServer
}

func mongoInfo(c *gin.Context) {
	if mgopool.IsNil() {
		c.JSON(http.StatusOK, nil)
		return
	}
	status := map[string]interface{}{}
	status["Len"] = mgopool.Len()
	status["IsAvailable"] = mgopool.IsAvailable()
	status["Cap"] = mgopool.Cap()
	status["Mode"] = mgopool.Mode()
	status["Config"] = mgopool.ShowConfig()
	status["LiveServers"] = mgopool.LiveServers()
	c.JSON(http.StatusOK, status)
}

func appConfig(c *gin.Context) {
	settings := viper.AllSettings()
	delete(settings, "database")
	c.JSON(http.StatusOK, settings)
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func ready(c *gin.Context) {
	ctx := intercom.GetContextFromGin(c)
	result := gin.H{}
	result["local mongo"] = true
	code := http.StatusOK
	mongoNil := mgopool.IsNil()
	if mongoNil {
		code = http.StatusServiceUnavailable
		result["local mongo"] = false
	} else {
		if errMongo := mgopool.Ping(ctx); errMongo != nil {
			m800log.Errorf(ctx, "[ready] local mongo unhealthy: %v", errMongo)
			code = http.StatusServiceUnavailable
			result["local mongo"] = false
		}
	}

	response := gin.H{}
	response["code"] = 0
	response["result"] = result
	c.JSON(code, response)
}
