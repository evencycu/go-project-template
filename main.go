package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"gitlab.com/general-backend/goctx"
	"gitlab.com/general-backend/gotrace"
	"gitlab.com/general-backend/m800log"
	"gitlab.com/general-backend/mgopool"
	"gitlab.com/rayshih/template/apiserver"

	"github.com/spf13/viper"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

type options struct {
	version    bool
	prof       bool
	configFile string
}

var opts options
var systemCtx goctx.Context

func init() {
	flag.BoolVar(&opts.version, "build", false, "GoLang build version.")
	flag.BoolVar(&opts.prof, "prof", false, "GoLang profiling function.")
	flag.StringVar(&opts.configFile, "config", "./conf.d/current.toml", "Path to Config File")

	systemCtx = goctx.Background()
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [arguments] <command> \n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if opts.version {
		fmt.Fprintf(os.Stderr, "%s\n", runtime.Version())
	}
	if opts.prof {
		ActivateProfile()
	}
	viper.AutomaticEnv()
	viper.SetConfigFile(opts.configFile)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Printf("[Warn] no config file: %s", err)
	}
	log.SetFlags(log.LstdFlags)

	// Init log
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	err = m800log.Initialize(viper.GetString("log.file_name"), viper.GetString("log.level"))
	if err != nil {
		panic(err)
	}
	m800log.SetM800JSONFormatter(viper.GetString("log.timestamp_format"), appName, version)
	m800log.SetAccessLevel(viper.GetString("log.access_level"))
	// Init tracer
	err = initTracer()
	if err != nil {
		panic(err)
	}

	// Init mongo
	err = mgopool.Initialize(getMongoDBInfo())
	if err != nil {
		m800log.Error(systemCtx, "mongo connect error:", err, ", config:", viper.AllSettings())
		panic(err)
	}

	httpServer := apiserver.InitGinServer(systemCtx, version)

	// graceful shutdown
	quit := make(chan os.Signal, 5)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

func initTracer() error {
	if !viper.GetBool("jaeger.enabled") {
		log.Println("Jaeger disabled")
		return nil
	}
	sConf := &jaegercfg.SamplerConfig{
		Type:  jaeger.SamplerTypeRateLimiting,
		Param: viper.GetFloat64("jaeger.sample_rate"),
	}
	rConf := &jaegercfg.ReporterConfig{
		QueueSize:           viper.GetInt("jaeger.queue_size"),
		BufferFlushInterval: viper.GetDuration("jaeger.flush_interval"),
		LocalAgentHostPort:  fmt.Sprintf("%s:%d", viper.GetString("jaeger.host"), viper.GetInt("jaeger.port")),
		LogSpans:            viper.GetBool("jaeger.log_spans"),
	}
	log.Printf("Sampler Config:%+v\nReporterConfig:%+v\n", sConf, rConf)
	if err := gotrace.InitJaeger(appName, sConf, rConf); err != nil {
		return fmt.Errorf("init tracer error:%s", err.Error())
	}
	return nil
}

func getMongoDBInfo() *mgopool.DBInfo {
	b := viper.GetInt("database.mgo.default.host_num")
	var prefix, key string
	var enabled bool
	name := viper.GetString("database.mgo.default.name")
	mgoMaxConn := viper.GetInt("database.mgo.default.max_conn")
	mgoDefaultUser := viper.GetString("database.mgo.default.user")
	mgoDefaultPassword := viper.GetString("database.mgo.default.password")
	mgoDefaultAuthDatabase := viper.GetString("database.mgo.default.authdatabase")
	mgoDefaultTimeout := viper.GetDuration("database.mgo.default.timeout")
	mgoDirect := viper.GetBool("database.mgo.default.direct")
	mgoSecondary := viper.GetBool("database.mgo.default.secondary")
	mgoMonogs := viper.GetBool("database.mgo.default.mongos")
	mgoAddrs := []string{}
	for i := 0; i < b; i++ {
		prefix = fmt.Sprintf("database.mgo.instance.%d", i)
		key = fmt.Sprintf("%s.enabled", prefix)
		enabled = viper.GetBool(key)
		if enabled {
			key = fmt.Sprintf("%s.host", prefix)
			host := viper.GetString(key)
			key = fmt.Sprintf("%s.port", prefix)
			port := viper.GetString(key)
			mgoAddrs = append(mgoAddrs, host+":"+port)
		}
	}
	return mgopool.NewDBInfo(name, mgoAddrs, mgoDefaultUser, mgoDefaultPassword,
		mgoDefaultAuthDatabase, mgoDefaultTimeout, mgoMaxConn, mgoDirect, mgoSecondary, mgoMonogs)
}
