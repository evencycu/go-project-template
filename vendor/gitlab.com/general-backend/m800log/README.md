# M800Log

## Prerequisite

* `gitlab.com/general-backend/goctx`
* `github.com/sirupsen/logrus`

## Getting Started

`Initialize(output, level string) (err error)` defines the init input.

* `output`: specify the output file name, `Discard`, `Stdout` are special case
* `level`: specify the log level i.e., `debug`, `info`, `warning`, `error`, `fatal`, `panic`

`SetStackTrace(stackEnabled bool)` set the option to add stack info in log

* Example:

```json
{"app":"app","eStack":"goroutine 9 [running]:\ngitlab.com/general-backend/m800log.stackTrace(0xc420090e60, 0x6fa83f)\n\t/home/ray/go/src/gitlab.com/general-backend/m800log/logger.go:167 +0xc5\ngitlab.com/general-backend/m800log.GetGeneralEntry(0x735800, 0xc420102640, 0xc420130480)\n\t/home/ray/go/src/gitlab.com/general-backend/m800log/logger.go:157 +0xc7\ngitlab.com/general-backend/m800log.Error(0x735800, 0xc420102640, 0xc42004bf88, 0x1, 0x1)\n\t/home/ray/go/src/gitlab.com/general-backend/m800log/logger.go:116 +0x3f\ngitlab.com/general-backend/m800log.TestStackTrace(0xc4201264b0)\n\t/home/ray/go/src/gitlab.com/general-backend/m800log/logger_test.go:70 +0x1f6\ntesting.tRunner(0xc4201264b0, 0x70eda8)\n\t/usr/local/go/src/testing/testing.go:777 +0xd0\ncreated by testing.(*T).Run\n\t/usr/local/go/src/testing/testing.go:824 +0x2e0\n","instanceID":"ray-M800","level":"error","message":"error","time":"2018-10-02T07:28:42.567507135Z","type":"General","vid":"v1"}
```

`SetM800JSONFormatter(timestampFormat, app, version string)` defines the required fields in JSON logs of M800 foramt.

* `timestampFormat`: `""` for default setting
* `app`: specify the server name in log `app` field
* `version` specify the version in log `vid` field

## Code Example

```go
err = m800log.Initialize(viper.GetString("log.file_name"), viper.GetString("log.level"))
if err != nil {
	panic(err)
}
m800log.SetM800JSONFormatter(viper.GetString("log.timestamp_format"), AppName, Version)
```

## Config Example

```toml
[log]
  file_name = "Stdout"
  level = "debug"
  timestamp_format = ""
```

## Reference

* [M800 Log format](https://issuetracking.maaii.com:9443/pages/viewpage.action?pageId=65128541)