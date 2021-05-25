# gotrace

utils for integrating jaeger tracing system

## TODO

* unit test

## Usage examples

### Initialization

```go
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

closer, errJaeger := gotrace.InitJaeger(ds.GetAppName(), sConf, rConf)
if errJaeger != nil {
    m800log.Errorf(systemCtx, "tracer init error, sampler config: %+v, reporter onfig: %+v", sConf, rConf)
    panic(errJaeger)
}
if closer != nil {
    defer closer.Close()
}
```

### Entry of component: In Gin Middleware

```go
sp, isNew := gotrace.CreateSpan(ctx,"span name")
if isNew {
    defer sp.Finish()
}
```

### Outgoing point: In mgopool

```go
sp := CreateMongoSpan(ctx, FuncCollectionCount, fmt.Sprintf("Collection:%s", collection))
defer sp.Finish()
n, dberr := col.Count()  // process go out to mongodb
```
