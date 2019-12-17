# mgopool package

## Dependency

* gitlab.com/cake/goctx
* gitlab.com/cake/gopkg
* gitlab.com/cake/m800log
* github.com/globalsign/mgo

## Initialize

```go
err := mongo.Initialize(mgoConfig)
```

## Log

mongo package 預設需要使用 m800log.Logger。

## Config

* Example:

```toml
[database]
[database.mgo.default]
  name = "Mgo"
  user = "test"
  password = "test"
  authdatabase = "admin"
  max_conn = 5
  host_num = 2
  timeout = "30s"
  direct = false
  secondary = false
  mongos = false

[database.mgo.instance.0]
  host = "10.128.112.181"
  port = 7379
  enabled = true

[database.mgo.instance.1]
  host = "10.128.112.181"
  port = 7380
  enabled = true
```

## Test Result

```bash
go test -cover
PASS
coverage: 96.5% of statements
ok      gitlab.com/cake/mgopool      74.419s
```