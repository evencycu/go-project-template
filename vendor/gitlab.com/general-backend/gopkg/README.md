# Introduction

The common errors, shared utilities of Golang based microservices.

## Requirement

* `Golang v1.8.0+`

## TODO

* utils/slack

## Project Structure

```sh
├── CHANGELOG.md
├── README.md
├── errors.go
├── errors_test.go
└── util
    └── slack
```

## The concept about error

### CodeError

The basic error type of production usage (interface). We need one more `Error Code` in most use case. Therefore, we define the `CodeError` interface implements not only the `Error()` but also `ErrorCode()`.

* Methods
  * Error() - error message as standard `error`
  * ErrorCode() - error code in **7** digits

### AsCodeError

* According to the name, you will know it could be handle as Code Error
* But it was created to support more actually.
* The Implemented struct type is TraceCodeError.
* Interface Detail
* AsEqual - It should contain a CodeError, AsEqual should be used in error handling.