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
    LocalAgentHostPort:  fmt.Sprintf("%s:%d", viper.GetString("jaeger.host"), viper.GetInt("jaeger.port")),
    LogSpans:            viper.GetBool("jaeger.log_spans"),
}
log.Printf("Sampler Config:%+v\nReporterConfig:%+v\n", sConf, rConf)
if err := gotrace.InitJaeger(AppName, sConf, rConf); err != nil {
    return fmt.Errorf("init tracer error:%s", err.Error())
}
return nil
```

### Entry of component: In Gin Middleware

```go
sp := gotrace.ExtractSpanFromReq(c.HandlerName(), c.Request)
defer sp.Finish()
ctx.Set(gotrace.ParentSpan, sp)
```

### Outgoing point: In mongodao 

```go
sp := CreateMongoSpan(ctx, FuncCollectionCount, fmt.Sprintf("Collection:%s", collection))
defer sp.Finish()
n, dberr := col.Count()  // process go out to mongodb
```