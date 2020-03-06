package command

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/cake/gotrace"
	"gitlab.com/cake/m800log"
	"gitlab.com/cake/mgopool"
	"gitlab.com/cake/redispool"

	"gitlab.com/cake/go-project-template/apiserver"
	"gitlab.com/cake/go-project-template/gpt"
)

var (
	quit = make(chan os.Signal, 5)
)

func NewServerCmd() *cobra.Command {
	var serverConfigFile string
	var cmdAPI = &cobra.Command{
		Use:   "server",
		Short: "Start the go-project-template server",
		Long:  `Run the http server go-project-template`,
		Run: func(cmd *cobra.Command, args []string) {
			defer log.Println("server main thread exiting")
			log.Println("config path:", serverConfigFile)

			var err error
			err = initInfra(serverConfigFile)
			if err != nil {
				panic("init infra error:" + err.Error())
			}
			m800log.Infof(systemCtx, "[go-project-template] init config: %+v", viper.AllSettings())

			if viper.GetBool("app.prof") {
				ActivateProfile()
			}

			// Init HTTP server, to provide readiness information at the very beginning
			httpServer, err := apiserver.InitGinServer(systemCtx)
			if err != nil {
				panic("api server init error:" + err.Error())
			}

			// Init Kafka producer
			// pConf := newKafkaProducerConfig()
			// pCtx := goctx.Background()
			// for k, v := range systemCtx.Map() {
			// 	pCtx.Set(k, v)
			// }
			// pCtx.Set("name", "kafkaProducer")
			// if err := kafkautil.InitKafkaProducer(pCtx, pConf, nil); err != nil {
			// 	panic(err)
			// }
			// apiserver.SetKafkaFlag()

			// Init redis
			// conf := NewRedisConfig()
			// conf.RedisDB = 3
			// redisPool, err := redispool.NewPool(conf)
			// if err != nil {
			// 	panic("init redis rate limit db error:" + err.Error())
			// }

			// Init local mongo
			err = mgopool.Initialize(getLocalMongoDBInfo())
			if err != nil {
				m800log.Errorf(systemCtx, "local mongo connect error: %v, config: %+v", err, getLocalMongoDBInfo())
				panic(err)
			}
			defer mgopool.Close()

			defer func(httpServer *http.Server) {
				log.Println("shutdown api server ...")
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := httpServer.Shutdown(ctx); err != nil {
					log.Printf("http server shutdown error:%v\n", err)
				}
			}(httpServer)

			// gracefully shutdown
			signal.Notify(quit, syscall.SIGINT, syscall.SIGKILL)
			<-quit
		},
	}

	cmdAPI.Flags().StringVarP(&serverConfigFile, "config", "c", "./local.toml", "Path to Config File")
	return cmdAPI
}

func initInfra(config string) error {
	viper.AutomaticEnv()
	viper.SetConfigFile(config)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	// Init log
	log.SetFlags(log.LstdFlags)
	err = m800log.Initialize(viper.GetString("log.output"), viper.GetString("log.level"))
	if err != nil {
		return err
	}

	m800log.SetM800JSONFormatter(viper.GetString("log.timestamp_format"), gpt.GetAppName(), gpt.GetVersion().Version, gpt.GetPhaseEnv(), gpt.GetNamespace())
	_ = m800log.SetAccessLevel(viper.GetString("log.access_level"))
	// Init tracer
	err = initTracer()
	if err != nil {
		return err
	}
	return nil
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
		LocalAgentHostPort:  viper.GetString("jaeger.host"),
		LogSpans:            viper.GetBool("jaeger.log_spans"),
	}
	log.Printf("Sampler Config:%+v\nReporterConfig:%+v\n", sConf, rConf)
	if err := gotrace.InitJaeger(gpt.GetAppName(), sConf, rConf); err != nil {
		return fmt.Errorf("init tracer error:%s", err.Error())
	}
	return nil
}

func newKafkaProducerConfig() *kafka.ConfigMap {
	return &kafka.ConfigMap{
		"bootstrap.servers": viper.GetString("kafka.bootstrap_servers"),
		"security.protocol": viper.GetString("kafka.security_protocol"),
		"ssl.ca.location":   viper.GetString("kafka.ssl_ca_location"),
		"sasl.mechanism":    viper.GetString("kafka.sasl_mechanism"),
		"sasl.username":     viper.GetString("kafka.sasl_username"),
		"sasl.password":     viper.GetString("kafka.sasl_password"),

		// producer config
		"go.batch.producer":       viper.GetBool("kafka.go_batch_producer"),
		"go.events.channel.size":  viper.GetInt("kafka.events_channel_size"),
		"go.produce.channel.size": viper.GetInt("kafka.produce_channel_size"),
		// idempotence reduce duplicate message
		"enable.idempotence": viper.GetBool("kafka.enable_idempotence"),
		"acks":               viper.GetInt("kafka.acks"),
	}
}

func getLocalMongoDBInfo() *mgopool.DBInfo {
	name := viper.GetString("database.mgo.name")
	mgoUser := viper.GetString("database.mgo.user")
	mgoPassword := viper.GetString("database.mgo.password")
	mgoAuthDatabase := viper.GetString("database.mgo.authdatabase")
	mgoMaxConn := viper.GetInt("database.mgo.max_conn")
	mgoTimeout := viper.GetDuration("database.mgo.timeout")
	mgoDirect := viper.GetBool("database.mgo.direct")
	mgoSecondary := viper.GetBool("database.mgo.secondary")
	mgoMongos := viper.GetBool("database.mgo.mongos")
	mgoAddrs := strings.Split(viper.GetString("database.mgo.hosts"), ";")
	return mgopool.NewDBInfo(name, mgoAddrs, mgoUser, mgoPassword,
		mgoAuthDatabase, mgoTimeout, mgoMaxConn, mgoDirect, mgoSecondary, mgoMongos)
}

func newRedisConfig() *redispool.Config {
	return &redispool.Config{
		Hosts:            strings.Split(viper.GetString("database.redis.host"), ";"),
		MasterName:       viper.GetString("database.redis.master"),
		SentinelPassword: viper.GetString("database.redis.sentinel_password"),
		Password:         viper.GetString("database.redis.password"),
		Timeout:          viper.GetDuration("database.redis.connect_timeout"),
		IdleTimeout:      viper.GetDuration("database.redis.idle_timeout"),
		MaxIdle:          viper.GetInt("database.redis.max_idle"),
		MaxActive:        viper.GetInt("database.redis.max_active"),
	}
}
