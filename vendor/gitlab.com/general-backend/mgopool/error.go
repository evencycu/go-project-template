package mgopool

import "gitlab.com/general-backend/gopkg"

var ErrUnknown gopkg.CodeError

// ErrAPIConnectDatabase is constant error for connecting mongodb fail
var ErrAPIConnectDatabase gopkg.CodeError

// ErrAPIFullResource is constant error for not enough mongo connections
var ErrAPIFullResource gopkg.CodeError

// ErrMongoPoolClosed is constant error for pool closed
var ErrMongoPoolClosed gopkg.CodeError

func init() {
	ErrUnknown = gopkg.NewCarrierCodeError(UnknownError, "unknown error")
	ErrAPIConnectDatabase = gopkg.NewCarrierCodeError(APIConnectDatabase, "cannot connect database, please retry later")
	ErrAPIFullResource = gopkg.NewCarrierCodeError(APIFullResource, "mongo resource not enough")
	ErrMongoPoolClosed = gopkg.NewCarrierCodeError(MongoPoolClosed, "mongo connection pool closed")
}
