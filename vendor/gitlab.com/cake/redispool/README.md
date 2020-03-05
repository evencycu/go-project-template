# redispool

* wrapper of redigo
* support redis sentinel

## Initialize

```go
conf := NewRedisConfig()
sessionRedisPool, err := redispool.NewPool(conf)
if err != nil {
    panic(err)
}
```

## Sentinel failover test

* redis connection will d/c and re-connect when 1. master died 2. force failover (change master)
* thus, we only get master address when dialing

## TODO

* read-only from slave