package mgopool

import (
	"net/http"

	"gitlab.com/cake/gopkg"
	"gitlab.com/cake/intercom"
)

var (
	ErrUnknown            = gopkg.NewCarrierCodeError(UnknownError, "unknown error")
	ErrAPIConnectDatabase = gopkg.NewCarrierCodeError(APIConnectDatabase, "cannot connect database, please retry later")
	ErrMongoPoolClosed    = gopkg.NewCarrierCodeError(MongoPoolClosed, "mongo connection pool closed")
)

func init() {
	_ = intercom.ErrorHttpStatusMapping.Set(ContextTimeout, http.StatusRequestTimeout)

	_ = intercom.ErrorHttpStatusMapping.Set(NotFound, http.StatusBadRequest)
	_ = intercom.ErrorHttpStatusMapping.Set(CollectionNotFound, http.StatusBadRequest)
	_ = intercom.ErrorHttpStatusMapping.Set(CollectionConflict, http.StatusConflict)
	_ = intercom.ErrorHttpStatusMapping.Set(QueryInputArray, http.StatusBadRequest)
	_ = intercom.ErrorHttpStatusMapping.Set(UpdateInputArray, http.StatusBadRequest)
	_ = intercom.ErrorHttpStatusMapping.Set(IncrementNumeric, http.StatusBadRequest)
	_ = intercom.ErrorHttpStatusMapping.Set(RegexString, http.StatusBadRequest)
	_ = intercom.ErrorHttpStatusMapping.Set(DotField, http.StatusBadRequest)
	_ = intercom.ErrorHttpStatusMapping.Set(Timeout, http.StatusRequestTimeout)
	_ = intercom.ErrorHttpStatusMapping.Set(StringIndexTooLong, http.StatusBadRequest)
	_ = intercom.ErrorHttpStatusMapping.Set(BadUpdateOperatorUsage, http.StatusBadRequest)
}
