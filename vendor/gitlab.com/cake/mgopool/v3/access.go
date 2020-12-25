package mgopool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"gitlab.com/cake/goctx"
	"gitlab.com/cake/gopkg"
	"gitlab.com/cake/gotrace/v2"
	"gitlab.com/cake/m800log"
	"gitlab.com/cake/mgopool/v3/compat"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	traceMsgEmptyElements = "empty elements"
)

const (
	TraceTagCriteria = "mongo.criteria"
	TraceTagInfo     = "mongo.info"
	TraceTagError    = "mongo.error"
)

const (
	FuncBulk                   = "Bulk"
	FuncBulkDelete             = "BulkDelete"
	FuncBulkInsert             = "BulkInsert"
	FuncBulkUpdate             = "BulkUpdate"
	FuncBulkUpsert             = "BulkUpsert"
	FuncCollectionCount        = "CollectionCount"
	FuncCreateCollection       = "CreateCollection"
	FuncCreateIndex            = "CreateIndex"
	FuncDBRun                  = "DBRun"
	FuncDropCollection         = "DropCollection"
	FuncDropIndex              = "DropIndex"
	FuncDropIndexName          = "DropIndexName"
	FuncDistinct               = "Distinct"
	FuncEnsureIndex            = "EnsureIndex"
	FuncFindAndModify          = "FindAndModify"
	FuncFindAndReplace         = "FindAndReplace"
	FuncFindAndRemove          = "FindAndRemove"
	FuncGetCollectionNames     = "GetCollectionNames"
	FuncIndexes                = "Indexes"
	FuncInsert                 = "Insert"
	FuncPing                   = "Ping"
	FuncPipe                   = "Pipe"
	FuncQueryAll               = "QueryAll"
	FuncQueryCount             = "QueryCount"
	FuncQueryOne               = "QueryOne"
	FuncQueryApply             = "QueryApply"
	FuncQueryDistinct          = "QueryDistinct"
	FuncQueryExplain           = "QueryExplain"
	FuncQueryMapReduce         = "QueryMapReduce"
	FuncCursorAll              = "CursorAll"
	FuncRemove                 = "Remove"
	FuncRemoveAll              = "RemoveAll"
	FuncRenameCollection       = "RenameCollection"
	FuncRun                    = "Run"
	FuncUpdate                 = "Update"
	FuncUpdateAll              = "UpdateAll"
	FuncUpdateId               = "UpdateId"
	FuncUpsert                 = "Upsert"
	FuncUpdateWithArrayFilters = "UpdateWithArrayFilters"
)

var (
	emptyBulkResult = BulkResult{}
)

type MongoPool interface {
	Init(dbi *DBInfo) error
	IsAvailable() bool
	Len() int
	LiveServers() []string
	Cap() int
	Mode() readpref.Mode
	Config() *DBInfo
	Close()
	ShowConfig() map[string]interface{}
	Recover() error

	Ping(ctx goctx.Context) (err gopkg.CodeError)
	PingPref(ctx goctx.Context, pref *readpref.ReadPref) (err gopkg.CodeError)
	GetCollectionNames(ctx goctx.Context, dbName string) (names []string, err gopkg.CodeError)
	CollectionCount(ctx goctx.Context, dbName, collection string) (n int, err gopkg.CodeError)
	Run(ctx goctx.Context, cmd interface{}, result interface{}) gopkg.CodeError
	DBRun(ctx goctx.Context, dbName string, cmd, result interface{}) gopkg.CodeError
	Insert(ctx goctx.Context, dbName, collection string, doc interface{}) gopkg.CodeError
	Remove(ctx goctx.Context, dbName, collection string, selector interface{}) (err gopkg.CodeError)
	RemoveAll(ctx goctx.Context, dbName, collection string, selector interface{}) (removedCount int, err gopkg.CodeError)
	Update(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (err gopkg.CodeError)
	ReplaceOne(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}, upsert bool) (err gopkg.CodeError)
	UpdateAll(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (result *mongo.UpdateResult, err gopkg.CodeError)
	UpdateId(ctx goctx.Context, dbName, collection string, id interface{}, update interface{}) (err gopkg.CodeError)
	UpdateWithArrayFilters(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}, arrayFilters interface{}, multi bool) (result *mongo.UpdateResult, err gopkg.CodeError)
	Upsert(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (result *mongo.UpdateResult, err gopkg.CodeError)
	BulkInsert(ctx goctx.Context, dbName, collection string, documents []bson.M) (result *BulkResult, err gopkg.CodeError)
	BulkUpsert(ctx goctx.Context, dbName, collection string, selectors []bson.M, documents []bson.M) (result *BulkResult, err gopkg.CodeError)
	BulkUpdate(ctx goctx.Context, dbName, collection string, selectors []bson.M, documents []bson.M) (result *BulkResult, err gopkg.CodeError)
	BulkInsertInterfaces(ctx goctx.Context, dbName, collection string, documents []interface{}) (result *BulkResult, err gopkg.CodeError)
	BulkUpsertInterfaces(ctx goctx.Context, dbName, collection string, selectors []bson.M, documents []interface{}) (result *BulkResult, err gopkg.CodeError)
	BulkUpdateInterfaces(ctx goctx.Context, dbName, collection string, selectors []bson.M, documents []interface{}) (result *BulkResult, err gopkg.CodeError)
	BulkDelete(ctx goctx.Context, dbName, collection string, documents []bson.M) (result *BulkResult, err gopkg.CodeError)
	QueryCount(ctx goctx.Context, dbName, collection string, selector interface{}) (n int, err gopkg.CodeError)
	QueryCountWithOptions(ctx goctx.Context, dbName, collection string, selector interface{}, skip, limit int) (n int, err gopkg.CodeError)
	QueryAll(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, skip, limit int, sort ...string) (err gopkg.CodeError)
	QueryAllWithCollation(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, collation *options.Collation, skip, limit int, sort ...string) (err gopkg.CodeError)
	QueryOne(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, skip int, sort ...string) (err gopkg.CodeError)
	FindAndModify(ctx goctx.Context, dbName, collection string, result, selector, update, fields interface{}, upsert, returnNew bool, sort ...string) (err gopkg.CodeError)
	FindAndReplace(ctx goctx.Context, dbName, collection string, result, selector, replacement, fields interface{}, upsert, returnNew bool, sort ...string) (err gopkg.CodeError)
	FindAndRemove(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, sort ...string) (err gopkg.CodeError)
	Indexes(ctx goctx.Context, dbName, collection string) (result []map[string]interface{}, err gopkg.CodeError)
	CreateIndex(ctx goctx.Context, dbName, collection string, key []string, sparse, unique bool, name string) (err gopkg.CodeError)
	CreateTTLIndex(ctx goctx.Context, dbName, collection string, key string, ttlSec int) (err gopkg.CodeError)
	EnsureIndex(ctx goctx.Context, dbName, collection string, index mongo.IndexModel) (err gopkg.CodeError)
	EnsureIndexCompat(ctx goctx.Context, dbName, collection string, index compat.Index) (err gopkg.CodeError)
	DropIndex(ctx goctx.Context, dbName, collection string, keys []string) (err gopkg.CodeError)
	DropIndexName(ctx goctx.Context, dbName, collection, name string) (err gopkg.CodeError)
	CreateCollection(ctx goctx.Context, dbName, collection string, options *options.CreateCollectionOptions) (err gopkg.CodeError)
	DropCollection(ctx goctx.Context, dbName, collection string) (err gopkg.CodeError)
	RenameCollection(ctx goctx.Context, dbName, oldName, newName string) (err gopkg.CodeError)
	Pipe(ctx goctx.Context, dbName, collection string, pipeline, result interface{}) (err gopkg.CodeError)
	GetCursor(ctx goctx.Context, dbName, collection string, selector interface{}, option ...*options.FindOptions) (MongoCursor, gopkg.CodeError)
	Distinct(ctx goctx.Context, dbName, collection string, selector bson.M, field string, result interface{}) (err gopkg.CodeError)
	GetBulk(ctx goctx.Context, dbName, collection string, batchSize ...int) *Bulk
}

type BulkResult struct {
	mongo.InsertManyResult
	mongo.BulkWriteResult
	mongo.BulkWriteException
}

type MongoCursor interface {
	Err() error
	Decode(val interface{}) (err gopkg.CodeError)
	Close(ctx context.Context) (err gopkg.CodeError)
	All(ctx context.Context, results interface{}) (err gopkg.CodeError)
	ID() int64
	Next(ctx context.Context) bool
	RemainingBatchLength() int
	TryNext(ctx context.Context) bool
}

func CreateMongoSpan(ctx goctx.Context, funcName string) opentracing.Span {
	return gotrace.CreateChildOfSpan(ctx, funcName)
}

func badConnection(s string) bool {
	switch {
	// dc, cluster not healthy
	case strings.Contains(s, MongoMsgNotMaster), strings.Contains(s, MongoMsgNoReachableServers), strings.Contains(s, MongoMsgEOF), strings.Contains(s, MongoMsgKernelEOF), strings.Contains(s, MongoMsgClose), strings.Contains(s, MongoMsgServerSelectionError):
		return true
		// no host case
	case strings.HasPrefix(s, MongoMsgNoHost), strings.HasPrefix(s, MongoMsgNoHost2), strings.HasPrefix(s, MongoMsgWriteUnavailable):
		return true
		// io timeout case
	case strings.HasPrefix(s, MongoMsgReadTCP), strings.HasPrefix(s, MongoMsgWriteTCP):
		return true
	}

	return false
}

func badUpdateOperator(errorString string) bool {
	if strings.Contains(errorString, MongoMsgEmptySet) || strings.Contains(errorString, MongoMsgEmptyUnset) || strings.Contains(errorString, MongoMsgBadModifier) || strings.Contains(errorString, MongoMsgEmptyInc) || strings.Contains(errorString, MongoMsgEmptyRename) {
		return true
	}
	return false
}

func getMongoCollection(c *mongo.Client, dbName, colName string) *mongo.Collection {
	return c.Database(dbName).Collection(colName, collectionOptions)
}

// resultHandling
// 1. put back session to pool
// 2. check the error and do error handling
func (p *Pool) resultHandling(err error, ctx goctx.Context) gopkg.CodeError {
	if err == nil {
		return nil
	}

	errorString := err.Error()
	code := UnknownError

	switch {
	case errorString == MongoMsgCursorNotFound || errorString == MongoMsgDocumentsNotFound || strings.HasPrefix(errorString, MongoMsgUnknown):
		code = NotFound
	case errorString == MongoMsgNsNotFound || errorString == MongoMsgCollectionNotFound || errorString == MongoMsgNamespaceNotFound:
		code = CollectionNotFound
	case strings.HasPrefix(errorString, MongoMsgDuplicateKey) || strings.Contains(errorString, MongoMsgBulk) || IsDup(err):
		code = DocumentConflict
		errorString = "document already exists:" + strings.Replace(errorString, MongoMsgE11000, "", 1)
	case strings.HasSuffix(errorString, MongoMsgCollectionConflict):
		code = CollectionConflict
	case errorString == MongoMsgDocNil:
		code = BadInputDoc
	case strings.Contains(errorString, MongoMsgArray), strings.Contains(errorString, MongoMsgNonEmpty), strings.Contains(errorString, MongoMsgInArray):
		code = QueryInputArray
		errorString = "query format error:" + errorString
	case badUpdateOperator(errorString):
		code = BadUpdateOperatorUsage
		errorString = "update requires an non-empty object"
	case strings.HasPrefix(errorString, MongoMsgIncrement):
		code = IncrementNumeric
	case errorString == MongoMsgRegexString:
		code = RegexString
	case strings.HasPrefix(errorString, MongoMsgDotField):
		code = DotField
	case strings.HasPrefix(errorString, MongoMsgwiredTigerIndex):
		code = StringIndexTooLong
	case errorString == MongoMsgGetPoolTimeout:
		poolTimeoutCounter.WithLabelValues(p.name).Inc()
		code = PoolTimeout
	case strings.Contains(errorString, MongoMsgContextTimeout), strings.Contains(errorString, MongoMsgTimeout):
		code = Timeout
		m800log.Errorf(ctx, "[op timeout] error:%s", errorString)
		errorString = MgopoolOpTimeout
	case badConnection(errorString):
		code = APIConnectDatabase
		m800log.Errorf(ctx, "[badConnection] error:%s", errorString)
		errorString = MgopoolBadConnection
	}

	ctx.Set(goctx.LogKeyErrorCode, code)
	return gopkg.NewCarrierCodeError(code, errorString)
}

// Ping change default behavior to primary
func (p *Pool) Ping(ctx goctx.Context) (err gopkg.CodeError) {
	errPing := p.client.Ping(ctx, readpref.Primary())
	err = p.resultHandling(errPing, ctx)
	return
}

// PingPref ping with given pref, nil would use client pref
func (p *Pool) PingPref(ctx goctx.Context, pref *readpref.ReadPref) (err gopkg.CodeError) {
	errPing := p.client.Ping(ctx, pref)
	err = p.resultHandling(errPing, ctx)
	return
}

// GetCollectionNames return all collection names in the db.
func (p *Pool) GetCollectionNames(ctx goctx.Context, dbName string) (names []string, err gopkg.CodeError) {
	names, errDB := p.client.Database(dbName, databaseOptions).ListCollectionNames(ctx, bson.M{})
	err = p.resultHandling(errDB, ctx)
	return
}

func (p *Pool) CollectionCount(ctx goctx.Context, dbName, collection string) (n int, err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelCollectionCount, start, &err)

	sp := CreateMongoSpan(ctx, FuncCollectionCount)
	defer sp.Finish()

	col := getMongoCollection(p.client, dbName, collection)
	_n, errDB := col.CountDocuments(ctx, bson.M{})
	n = int(_n)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	sp.SetTag(TraceTagCriteria, collection)
	return
}

func (p *Pool) Run(ctx goctx.Context, cmd interface{}, result interface{}) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelRun, start, &err)

	sp := CreateMongoSpan(ctx, FuncRun)
	defer sp.Finish()

	errDB := p.client.Database("admin", databaseOptions).RunCommand(ctx, cmd).Decode(result)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("DB Operations:%+v ,Result:%+v", cmd, result)
	sp.SetTag(TraceTagCriteria, logMsg)
	infoLog(ctx, p.LiveServers(), logMsg)
	return
}

func (p *Pool) DBRun(ctx goctx.Context, dbName string, cmd, result interface{}) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelDBRun, start, &err)

	sp := CreateMongoSpan(ctx, FuncDBRun)
	defer sp.Finish()

	errDB := p.client.Database(dbName, databaseOptions).RunCommand(ctx, cmd).Decode(result)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("DB Operations:%+v ,Result:%+v", cmd, result)
	sp.SetTag(TraceTagCriteria, logMsg)
	infoLog(ctx, p.LiveServers(), logMsg)
	return
}

func (p *Pool) Insert(ctx goctx.Context, dbName, collection string, doc interface{}) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelInsert, start, &err)

	// Insert document to collection
	sp := CreateMongoSpan(ctx, FuncInsert)
	defer sp.Finish()

	col := getMongoCollection(p.client, dbName, collection)
	_, errDB := col.InsertOne(ctx, doc)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Collection:%s,Stuff:%+v", collection, doc)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, p.LiveServers(), FuncInsert, logMsg, start)
	return
}

func (p *Pool) Distinct(ctx goctx.Context, dbName, collection string, selector bson.M, field string, result interface{}) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelDistinct, start, &err)

	sp := CreateMongoSpan(ctx, FuncDistinct)
	defer sp.Finish()

	col := getMongoCollection(p.client, dbName, collection)
	resDB, errDB := col.Distinct(ctx, field, selector)
	defer func() {
		logMsg := fmt.Sprintf("Collection:%s,Field:%+v", collection, field)
		sp.SetTag(TraceTagCriteria, logMsg)
		accessLog(ctx, p.LiveServers(), FuncDistinct, logMsg, start)
	}()
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
		return
	}
	byteData, _ := json.Marshal(resDB)

	errUnmarshal := json.Unmarshal(byteData, result)
	if errUnmarshal != nil {
		err = gopkg.NewCodeError(UpdateInputArray, errUnmarshal.Error())
	}
	return
}

func (p *Pool) Remove(ctx goctx.Context, dbName, collection string, selector interface{}) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelRemove, start, &err)

	sp := CreateMongoSpan(ctx, FuncRemove)
	defer sp.Finish()

	col := getMongoCollection(p.client, dbName, collection)
	deleteResult, errDB := col.DeleteOne(ctx, selector)
	err = p.resultHandling(errDB, ctx)
	if deleteResult != nil && deleteResult.DeletedCount == int64(0) && err == nil { // align to mgo behavior
		err = gopkg.NewCodeError(NotFound, MongoMsgDocumentsNotFound)
	}
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v", collection, selector)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, p.LiveServers(), FuncRemove, logMsg, start)
	return
}

func (p *Pool) RemoveAll(ctx goctx.Context, dbName, collection string, selector interface{}) (removedCount int, err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelRemoveAll, start, &err)

	sp := CreateMongoSpan(ctx, FuncRemoveAll)
	defer sp.Finish()

	col := getMongoCollection(p.client, dbName, collection)
	res, errDB := col.DeleteMany(ctx, selector)
	if res != nil {
		removedCount = int(res.DeletedCount)
	}
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v", collection, selector)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, p.LiveServers(), FuncRemoveAll, logMsg, start)
	return
}

func (p *Pool) ReplaceOne(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}, upsert bool) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelReplaceOne, start, &err)

	sp := CreateMongoSpan(ctx, FuncUpdate)
	defer sp.Finish()

	col := getMongoCollection(p.client, dbName, collection)
	opts := &options.ReplaceOptions{
		Upsert: &upsert,
	}
	var updateResult *mongo.UpdateResult
	var errDB error
	updateResult, errDB = col.ReplaceOne(ctx, selector, update, opts)
	err = p.resultHandling(errDB, ctx)
	if updateResult != nil && updateResult.MatchedCount == 0 && err == nil { // align to mgo behavior
		err = gopkg.NewCodeError(NotFound, MongoMsgDocumentsNotFound)
	}
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v,Replace:%+v Upsert:%v", collection, selector, update, upsert)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, p.LiveServers(), FuncUpdate, logMsg, start)
	return
}

func (p *Pool) Update(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelUpdate, start, &err)

	sp := CreateMongoSpan(ctx, FuncUpdate)
	defer sp.Finish()

	col := getMongoCollection(p.client, dbName, collection)
	setOp, hasOp := EnsureUpdateOp(update)
	var updateResult *mongo.UpdateResult
	var errDB error
	if hasOp {
		updateResult, errDB = col.UpdateOne(ctx, selector, setOp)
	} else {
		updateResult, errDB = col.ReplaceOne(ctx, selector, update)
	}
	err = p.resultHandling(errDB, ctx)
	if updateResult != nil && updateResult.MatchedCount == 0 && err == nil { // align to mgo behavior
		err = gopkg.NewCodeError(NotFound, MongoMsgDocumentsNotFound)
	}
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v,Update:%+v", collection, selector, update)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, p.LiveServers(), FuncUpdate, logMsg, start)
	return
}

func (p *Pool) UpdateAll(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (result *mongo.UpdateResult, err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelUpdateAll, start, &err)

	sp := CreateMongoSpan(ctx, FuncUpdateAll)
	defer sp.Finish()

	col := getMongoCollection(p.client, dbName, collection)
	setOp, _ := EnsureUpdateOp(update)
	result, errDB := col.UpdateMany(ctx, selector, setOp)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v,Update:%+v", collection, selector, update)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, p.LiveServers(), FuncUpdateAll, logMsg, start)
	return
}

func (p *Pool) UpdateId(ctx goctx.Context, dbName, collection string, id interface{}, update interface{}) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelUpdateId, start, &err)

	sp := CreateMongoSpan(ctx, FuncUpdateId)
	defer sp.Finish()

	col := getMongoCollection(p.client, dbName, collection)
	_, errDB := col.UpdateOne(ctx, bson.M{"_id": id}, update)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v,Update:%+v", collection, id, update)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, p.LiveServers(), FuncUpdateId, logMsg, start)
	return
}

func (p *Pool) UpdateWithArrayFilters(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}, arrayFilters interface{}, multi bool) (result *mongo.UpdateResult, err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelUpdateWithArrayFilters, start, &err)

	sp := CreateMongoSpan(ctx, FuncUpdateWithArrayFilters)
	defer sp.Finish()

	if reflect.TypeOf(arrayFilters).Kind() != reflect.Slice {
		return nil, gopkg.NewCarrierCodeError(QueryInputArray, "query format error: arrayFilters is expected to be an array")
	}
	v := reflect.ValueOf(arrayFilters)
	filters := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		filters[i] = v.Index(i).Interface()
	}

	col := getMongoCollection(p.client, dbName, collection)
	var errDB error
	updateOpt := options.Update().SetArrayFilters(options.ArrayFilters{Filters: filters})
	setOp, _ := EnsureUpdateOp(update)
	result, errDB = col.UpdateMany(ctx, selector, setOp, updateOpt)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
		return
	}

	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v,Update:%+v,ArrayFilters:%+v", collection, selector, update, arrayFilters)
	accessLog(ctx, p.LiveServers(), FuncUpdateWithArrayFilters, logMsg, start)
	sp.SetTag(TraceTagCriteria, logMsg)
	return
}

func (p *Pool) Upsert(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (result *mongo.UpdateResult, err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelUpsert, start, &err)

	sp := CreateMongoSpan(ctx, FuncUpsert)
	defer sp.Finish()

	col := getMongoCollection(p.client, dbName, collection)
	updateOpt := options.Update().SetUpsert(true)
	var errDB error
	update = EnsureUpsertOp(update)
	result, errDB = col.UpdateOne(ctx, selector, update, updateOpt)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
		return
	}

	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v,Upsert:%+v", collection, selector, update)
	accessLog(ctx, p.LiveServers(), FuncUpsert, logMsg, start)
	sp.SetTag(TraceTagCriteria, logMsg)
	return
}

func (p *Pool) BulkInsert(ctx goctx.Context, dbName, collection string, documents []bson.M) (result *BulkResult, err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelPoolBulkInsert, start, &err)

	sp := CreateMongoSpan(ctx, FuncBulkInsert)
	defer sp.Finish()

	if len(documents) == 0 {
		sp.SetTag(TraceTagInfo, traceMsgEmptyElements)
		result = &emptyBulkResult
		return
	}

	var interfaces []interface{}
	for _, doc := range documents {
		if doc == nil {
			continue
		}
		interfaces = append(interfaces, doc)
	}

	col := getMongoCollection(p.client, dbName, collection)
	insertManyOpt := options.InsertMany().SetOrdered(false)
	res, errDB := col.InsertMany(ctx, interfaces, insertManyOpt)
	result = MergeBulkResult(res, errDB)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	sp.SetTag(TraceTagCriteria, collection)
	accessLog(ctx, p.LiveServers(), FuncBulkInsert, collection, start)
	return
}

func (p *Pool) BulkUpsert(ctx goctx.Context, dbName, collection string, selectors []bson.M, documents []bson.M) (result *BulkResult, err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelPoolBulkUpsert, start, &err)

	sp := CreateMongoSpan(ctx, FuncBulkUpsert)
	defer sp.Finish()

	if len(selectors) == 0 || len(documents) == 0 {
		sp.SetTag(TraceTagInfo, traceMsgEmptyElements)
		result = &emptyBulkResult
		return
	}

	var writeModels []mongo.WriteModel
	for i := 0; i < len(documents); i++ {
		if documents[i] == nil {
			continue
		}
		setOp, hasOp := EnsureUpdateOp(documents[i])
		if hasOp {
			m := mongo.NewUpdateManyModel()
			m.SetUpsert(true)
			m.SetFilter(selectors[i])
			m.SetUpdate(setOp)
			writeModels = append(writeModels, m)
		} else {
			m := mongo.NewReplaceOneModel()
			m.SetUpsert(true)
			m.SetFilter(selectors[i])
			m.SetReplacement(documents[i])
			writeModels = append(writeModels, m)
		}
	}

	col := getMongoCollection(p.client, dbName, collection)
	bulkWriteOpt := options.BulkWrite().SetOrdered(false)
	res, errDB := col.BulkWrite(ctx, writeModels, bulkWriteOpt)
	result = MergeBulkResult(res, errDB)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	sp.SetTag(TraceTagCriteria, collection)
	accessLog(ctx, p.LiveServers(), FuncBulkUpsert, collection, start)
	return
}

func (p *Pool) BulkUpdate(ctx goctx.Context, dbName, collection string, selectors []bson.M, documents []bson.M) (result *BulkResult, err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelPoolBulkUpdate, start, &err)

	sp := CreateMongoSpan(ctx, FuncBulkUpdate)
	defer sp.Finish()

	if len(selectors) == 0 || len(documents) == 0 {
		sp.SetTag(TraceTagInfo, traceMsgEmptyElements)
		result = &emptyBulkResult
		return
	}

	var writeModels []mongo.WriteModel
	for i := 0; i < len(documents); i++ {
		if documents[i] == nil {
			continue
		}
		setOp, hasOp := EnsureUpdateOp(documents[i])
		if hasOp {
			m := mongo.NewUpdateManyModel()
			m.SetUpsert(false)
			m.SetFilter(selectors[i])
			m.SetUpdate(setOp)
			writeModels = append(writeModels, m)
		} else {
			m := mongo.NewReplaceOneModel()
			m.SetUpsert(false)
			m.SetFilter(selectors[i])
			m.SetReplacement(documents[i])
			writeModels = append(writeModels, m)
		}
	}

	col := getMongoCollection(p.client, dbName, collection)
	bulkWriteOpt := options.BulkWrite().SetOrdered(false)
	res, errDB := col.BulkWrite(ctx, writeModels, bulkWriteOpt)
	result = MergeBulkResult(res, errDB)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	sp.SetTag(TraceTagCriteria, collection)
	accessLog(ctx, p.LiveServers(), FuncBulkUpdate, collection, start)
	return
}

func (p *Pool) BulkInsertInterfaces(ctx goctx.Context, dbName, collection string, documents []interface{}) (result *BulkResult, err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelPoolBulkInsertInterfaces, start, &err)

	sp := CreateMongoSpan(ctx, FuncBulkInsert)
	defer sp.Finish()

	if len(documents) == 0 {
		sp.SetTag(TraceTagInfo, traceMsgEmptyElements)
		result = &emptyBulkResult
		return
	}

	col := getMongoCollection(p.client, dbName, collection)
	insertManyOpt := options.InsertMany().SetOrdered(false)
	res, errDB := col.InsertMany(ctx, documents, insertManyOpt)
	result = MergeBulkResult(res, errDB)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	sp.SetTag(TraceTagCriteria, collection)
	accessLog(ctx, p.LiveServers(), FuncBulkInsert, collection, start)
	return
}

func (p *Pool) BulkUpsertInterfaces(ctx goctx.Context, dbName, collection string, selectors []bson.M, documents []interface{}) (result *BulkResult, err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelPoolBulkUpsertInterfaces, start, &err)

	sp := CreateMongoSpan(ctx, FuncBulkUpsert)
	defer sp.Finish()

	if len(selectors) == 0 || len(documents) == 0 {
		sp.SetTag(TraceTagInfo, traceMsgEmptyElements)
		result = &emptyBulkResult
		return
	}
	var writeModels []mongo.WriteModel
	for i := 0; i < len(documents); i++ {
		if documents[i] == nil {
			continue
		}
		m := mongo.NewUpdateManyModel()
		m.SetUpsert(true)
		m.SetFilter(selectors[i])
		m.SetUpdate(EnsureUpsertOp(documents[i]))
		writeModels = append(writeModels, m)
	}

	col := getMongoCollection(p.client, dbName, collection)
	bulkWriteOpt := options.BulkWrite().SetOrdered(false)
	res, errDB := col.BulkWrite(ctx, writeModels, bulkWriteOpt)
	result = MergeBulkResult(res, errDB)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	sp.SetTag(TraceTagCriteria, collection)
	accessLog(ctx, p.LiveServers(), FuncBulkUpsert, collection, start)
	return
}

func (p *Pool) BulkUpdateInterfaces(ctx goctx.Context, dbName, collection string, selectors []bson.M, documents []interface{}) (result *BulkResult, err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelPoolBulkUpdateInterfaces, start, &err)

	sp := CreateMongoSpan(ctx, FuncBulkUpdate)
	defer sp.Finish()

	if len(selectors) == 0 || len(documents) == 0 {
		sp.SetTag(TraceTagInfo, traceMsgEmptyElements)
		result = &emptyBulkResult
		return
	}

	var writeModels []mongo.WriteModel
	for i := 0; i < len(documents); i++ {
		if documents[i] == nil {
			continue
		}
		setOp, hasOp := EnsureUpdateOp(documents[i])
		if hasOp {
			m := mongo.NewUpdateManyModel()
			m.SetUpsert(false)
			m.SetFilter(selectors[i])
			m.SetUpdate(setOp)
			writeModels = append(writeModels, m)
		} else {
			m := mongo.NewReplaceOneModel()
			m.SetUpsert(false)
			m.SetFilter(selectors[i])
			m.SetReplacement(documents[i])
			writeModels = append(writeModels, m)
		}
	}

	col := getMongoCollection(p.client, dbName, collection)
	bulkWriteOpt := options.BulkWrite().SetOrdered(false)
	res, errDB := col.BulkWrite(ctx, writeModels, bulkWriteOpt)
	result = MergeBulkResult(res, errDB)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	sp.SetTag(TraceTagCriteria, collection)
	accessLog(ctx, p.LiveServers(), FuncBulkUpdate, collection, start)
	return
}

func (p *Pool) BulkDelete(ctx goctx.Context, dbName, collection string, selector []bson.M) (result *BulkResult, err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelPoolBulkDelete, start, &err)

	sp := CreateMongoSpan(ctx, FuncBulkDelete)
	defer sp.Finish()

	if len(selector) == 0 {
		sp.SetTag(TraceTagInfo, traceMsgEmptyElements)
		result = &emptyBulkResult
		return
	}

	var writeModels []mongo.WriteModel
	for i := 0; i < len(selector); i++ {
		if selector[i] == nil {
			continue
		}
		m := mongo.NewDeleteManyModel()
		m.SetFilter(selector[i])
		writeModels = append(writeModels, m)
	}

	col := getMongoCollection(p.client, dbName, collection)
	bulkWriteOpt := options.BulkWrite().SetOrdered(false)
	res, errDB := col.BulkWrite(ctx, writeModels, bulkWriteOpt)
	result = MergeBulkResult(res, errDB)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	sp.SetTag(TraceTagCriteria, collection)
	accessLog(ctx, p.LiveServers(), FuncBulkDelete, collection, start)
	return
}

func (p *Pool) QueryCount(ctx goctx.Context, dbName, collection string, selector interface{}) (n int, err gopkg.CodeError) {
	return p.QueryCountWithOptions(ctx, dbName, collection, selector, 0, 0)
}

func (p *Pool) QueryCountWithOptions(ctx goctx.Context, dbName, collection string, selector interface{}, skip, limit int) (n int, err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelQueryCountWithOptions, start, &err)

	sp := CreateMongoSpan(ctx, FuncQueryCount)
	defer sp.Finish()

	opt := options.Count()
	if skip > 0 {
		opt.SetSkip(int64(skip))
	}
	if limit > 0 {
		opt.SetLimit(int64(limit))
	}

	col := getMongoCollection(p.client, dbName, collection)
	_n, errDB := col.CountDocuments(ctx, selector, opt)
	n = int(_n)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	logMsg := fmt.Sprintf("Collection:%s,Query:%+v", collection, selector)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, p.LiveServers(), FuncQueryCount, logMsg, start)
	return
}

func (p *Pool) QueryAll(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, skip, limit int, sort ...string) (err gopkg.CodeError) {
	return p.QueryAllWithCollation(ctx, dbName, collection, result, selector, fields, nil, skip, limit, sort...)
}

func (p *Pool) QueryAllWithCollation(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, collation *options.Collation, skip, limit int, sort ...string) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelQueryAllWithOptions, start, &err)

	sp := CreateMongoSpan(ctx, FuncQueryAll)
	defer sp.Finish()

	findOpt := options.Find()
	findOpt.SetSkip(int64(skip))
	findOpt.SetLimit(int64(limit))

	if fields != nil {
		findOpt.SetProjection(fields)
	}
	if len(sort) > 0 {
		findOpt.SetSort(ParseSortField(sort...))
	}
	if collation != nil {
		findOpt.SetCollation(collation)
	}
	// compatible with mgo
	if selector == nil {
		selector = bson.M{}
	}

	col := getMongoCollection(p.client, dbName, collection)
	cursor, errDB := col.Find(ctx, selector, findOpt)
	defer func() {
		logMsg := fmt.Sprintf("Collection:%s,Query:%+v,Projection::%+v,Sort:%+v,Skip:%d,Limit:%d",
			collection,
			selector,
			fields,
			sort,
			skip,
			limit,
		)
		sp.SetTag(TraceTagCriteria, logMsg)
		accessLog(ctx, p.LiveServers(), FuncQueryAll, logMsg, start)
	}()
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
		return
	}
	errDB = cursor.All(ctx, result)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	return
}

func (p *Pool) QueryOne(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, skip int, sort ...string) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelQueryOne, start, &err)

	sp := CreateMongoSpan(ctx, FuncQueryOne)
	defer sp.Finish()

	findOpt := options.FindOne()
	findOpt.SetSkip(int64(skip))
	findOpt.SetProjection(fields)
	if len(sort) > 0 {
		findOpt.SetSort(ParseSortField(sort...))
	}

	col := getMongoCollection(p.client, dbName, collection)
	errDB := col.FindOne(ctx, selector, findOpt).Decode(result)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	logMsg := fmt.Sprintf("Collection:%s,Filter:%+v,Projection::%+v,Sort:%+v,Skip:%d", collection, selector, fields, sort, skip)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, p.LiveServers(), FuncQueryOne, logMsg, start)
	return
}

// FindAndModify can only update one doc (by mongodb description)
func (p *Pool) FindAndModify(ctx goctx.Context, dbName, collection string, result, selector, update, fields interface{},
	upsert, returnNew bool, sort ...string) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelFindAndModify, start, &err)

	sp := CreateMongoSpan(ctx, FuncFindAndModify)
	defer sp.Finish()

	if result == nil {
		result = bson.M{}
	}

	col := getMongoCollection(p.client, dbName, collection)
	setOp, hasOp := EnsureUpdateOp(update)
	var errDB error
	if hasOp {
		foamOpt := options.FindOneAndUpdate()
		if returnNew {
			foamOpt.SetReturnDocument(options.After)
		}
		foamOpt.SetUpsert(upsert)
		foamOpt.SetProjection(fields)
		errDB = col.FindOneAndUpdate(ctx, selector, setOp, foamOpt).Decode(result)
	} else {
		foamOpt := options.FindOneAndReplace()
		if returnNew {
			foamOpt.SetReturnDocument(options.After)
		}
		foamOpt.SetUpsert(upsert)
		foamOpt.SetProjection(fields)
		errDB = col.FindOneAndReplace(ctx, selector, update, foamOpt).Decode(result)
	}

	err = p.resultHandling(errDB, ctx)
	if upsert && !returnNew && err != nil && err.ErrorCode() == NotFound { // align to mgo behavior
		err = nil
	}
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	logMsg := fmt.Sprintf("Collection:%s,Query:%+v,Projection::%+v,Sort:%+v", collection, selector, fields, sort)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, p.LiveServers(), FuncFindAndModify, logMsg, start)
	return
}

func (p *Pool) FindAndReplace(ctx goctx.Context, dbName, collection string, result, selector, replacement, fields interface{},
	upsert, returnNew bool, sort ...string) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelFindAndReplace, start, &err)

	sp := CreateMongoSpan(ctx, FuncFindAndReplace)
	defer sp.Finish()

	foarOpt := options.FindOneAndReplace()
	if returnNew {
		foarOpt.SetReturnDocument(options.After)
	}
	foarOpt.SetUpsert(upsert)
	foarOpt.SetProjection(fields)

	if result == nil {
		result = bson.M{}
	}

	col := getMongoCollection(p.client, dbName, collection)
	errDB := col.FindOneAndReplace(ctx, selector, replacement, foarOpt).Decode(result)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	logMsg := fmt.Sprintf("Collection:%s,Query:%+v,Projection::%+v,Sort:%+v", collection, selector, fields, sort)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, p.LiveServers(), FuncFindAndReplace, logMsg, start)
	return
}

func (p *Pool) FindAndRemove(ctx goctx.Context, dbName, collection string, result, selector, fields interface{},
	sort ...string) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelFindAndRemove, start, &err)

	sp := CreateMongoSpan(ctx, FuncFindAndRemove)
	defer sp.Finish()

	foarOpt := options.FindOneAndDelete()
	foarOpt.SetProjection(fields)
	if len(sort) > 0 {
		foarOpt.SetSort(ParseSortField(sort...))
	}

	if result == nil {
		result = bson.M{}
	}

	col := getMongoCollection(p.client, dbName, collection)
	errDB := col.FindOneAndDelete(ctx, selector, foarOpt).Decode(result)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	logMsg := fmt.Sprintf("Collection:%s,Query:%+v,Projection::%+v,Sort:%+v", collection, selector, fields, sort)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, p.LiveServers(), FuncFindAndRemove, logMsg, start)
	return
}

func (p *Pool) Indexes(ctx goctx.Context, dbName, collection string) (result []map[string]interface{}, err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelIndexes, start, &err)

	sp := CreateMongoSpan(ctx, FuncIndexes)
	defer sp.Finish()

	sp.SetTag(TraceTagCriteria, collection)
	indexView := getMongoCollection(p.client, dbName, collection).Indexes()
	cursor, errDB := indexView.List(ctx)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		return
	}
	result = []map[string]interface{}{}
	errDB = cursor.All(ctx, &result)
	err = p.resultHandling(errDB, ctx)
	return
}

func (p *Pool) CreateIndex(ctx goctx.Context, dbName, collection string, key []string, sparse, unique bool, name string) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelCreateIndex, start, &err)

	sp := CreateMongoSpan(ctx, FuncCreateIndex)
	defer sp.Finish()

	indexOpt := options.Index()
	indexOpt.SetBackground(true)
	indexOpt.SetSparse(sparse)
	indexOpt.SetUnique(unique)
	if name != "" {
		indexOpt.SetName(name)
	}
	keys, _ := ParseIndexKey(key)
	index := mongo.IndexModel{
		Keys:    keys,
		Options: indexOpt,
	}

	indexView := getMongoCollection(p.client, dbName, collection).Indexes()
	name, errDB := indexView.CreateOne(ctx, index)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
		errLog(ctx, p.LiveServers(), fmt.Sprintf("%s add index:%s name:%s err:%s", collection, strings.Join(key, ","), name, err.Error()))
	}
	sp.SetTag(TraceTagCriteria, fmt.Sprintf("Collection:%s,Create Index:%+v", collection, index))
	return
}

func (p *Pool) CreateTTLIndex(ctx goctx.Context, dbName, collection string, key string, ttlSec int) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelCreateTTLIndex, start, &err)

	sp := CreateMongoSpan(ctx, FuncCreateIndex)
	defer sp.Finish()

	indexOpt := options.Index()
	indexOpt.SetBackground(true)
	indexOpt.SetExpireAfterSeconds(int32(ttlSec))

	keys, _ := ParseIndexKey([]string{key})
	index := mongo.IndexModel{
		Keys:    keys,
		Options: indexOpt,
	}

	indexView := getMongoCollection(p.client, dbName, collection).Indexes()
	name, errDB := indexView.CreateOne(ctx, index)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
		errLog(ctx, p.LiveServers(), fmt.Sprintf("%s add ttl index:%s name:%s err:%s", collection, key, name, err.Error()))
	}
	sp.SetTag(TraceTagCriteria, fmt.Sprintf("Collection:%s,Create Index:%+v", collection, index))
	return
}

func (p *Pool) EnsureIndex(ctx goctx.Context, dbName, collection string, index mongo.IndexModel) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelEnsureIndex, start, &err)

	sp := CreateMongoSpan(ctx, FuncEnsureIndex)
	defer sp.Finish()

	indexView := getMongoCollection(p.client, dbName, collection).Indexes()
	name, errDB := indexView.CreateOne(ctx, index)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
		errLog(ctx, p.LiveServers(), fmt.Sprintf("%s add index:%+v name:%s err:%s", collection, index.Keys, name, err.Error()))
	}

	sp.SetTag(TraceTagCriteria, fmt.Sprintf("Collection:%s,EnsureIndex:%+v", collection, index))
	return
}

func (p *Pool) EnsureIndexCompat(ctx goctx.Context, dbName, collection string, indexCompat compat.Index) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelEnsureIndexCompat, start, &err)

	indexOptions := options.IndexOptions{
		Collation: indexCompat.Collation,
	}

	if indexCompat.Unique {
		indexOptions.Unique = &indexCompat.Unique
	}

	if indexCompat.Background {
		indexOptions.Background = &indexCompat.Background
	}

	if indexCompat.Sparse {
		indexOptions.Sparse = &indexCompat.Sparse
	}

	if len(indexCompat.PartialFilter) != 0 {
		indexOptions.PartialFilterExpression = &indexCompat.PartialFilter
	}

	if indexCompat.Min != 0 {
		indexOptions.Min = &indexCompat.Min
	}

	if indexCompat.Max != 0 {
		indexOptions.Max = &indexCompat.Max
	}

	if indexCompat.DefaultLanguage != "" {
		indexOptions.DefaultLanguage = &indexCompat.DefaultLanguage
	}

	if indexCompat.LanguageOverride != "" {
		indexOptions.LanguageOverride = &indexCompat.LanguageOverride
	}

	if len(indexCompat.Weights) != 0 {
		indexOptions.Weights = &indexCompat.Weights
	}

	if indexCompat.ExpireAfter != 0 {
		expireAfterSeconds := int32(indexCompat.ExpireAfter.Seconds())
		indexOptions.ExpireAfterSeconds = &expireAfterSeconds
	}

	if indexCompat.BucketSize != 0 {
		bucketSize := int32(indexCompat.BucketSize)
		indexOptions.BucketSize = &bucketSize
	}

	if indexCompat.Bits != 0 {
		bits := int32(indexCompat.Bits)
		indexOptions.Bits = &bits
	}

	if indexCompat.Name != "" {
		indexOptions.Name = &indexCompat.Name
	}

	keys, _ := ParseIndexKey(indexCompat.Key)

	index := mongo.IndexModel{
		Keys:    keys,
		Options: &indexOptions,
	}

	err = p.EnsureIndex(ctx, dbName, collection, index)
	return
}

func (p *Pool) DropIndex(ctx goctx.Context, dbName, collection string, key []string) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelDropIndex, start, &err)

	sp := CreateMongoSpan(ctx, FuncDropIndex)
	defer sp.Finish()

	// get default index name
	_, indexName := ParseIndexKey(key)
	indexView := getMongoCollection(p.client, dbName, collection).Indexes()
	_, errDB := indexView.DropOne(ctx, indexName)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	sp.SetTag(TraceTagCriteria, fmt.Sprintf("Collection:%s,DropIndex Key:%+v", collection, key))
	return
}

func (p *Pool) DropIndexName(ctx goctx.Context, dbName, collection, name string) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelDropIndexName, start, &err)

	sp := CreateMongoSpan(ctx, FuncDropIndexName)
	defer sp.Finish()

	indexView := getMongoCollection(p.client, dbName, collection).Indexes()
	_, errDB := indexView.DropOne(ctx, name)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	sp.SetTag(TraceTagCriteria, fmt.Sprintf("Collection:%s,DropIndexName:%s", collection, name))
	return
}

func (p *Pool) CreateCollection(ctx goctx.Context, dbName, collection string, opt *options.CreateCollectionOptions) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelCreateCollection, start, &err)

	sp := CreateMongoSpan(ctx, FuncCreateCollection)
	defer sp.Finish()

	errDB := p.client.Database(dbName, databaseOptions).CreateCollection(ctx, collection, opt)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	sp.SetTag(TraceTagCriteria, collection)
	accessLog(ctx, p.LiveServers(), FuncCreateCollection, collection, start)
	return
}

func (p *Pool) DropCollection(ctx goctx.Context, dbName, collection string) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelDropCollection, start, &err)

	sp := CreateMongoSpan(ctx, FuncDropCollection)
	defer sp.Finish()

	errDB := getMongoCollection(p.client, dbName, collection).Drop(ctx)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	sp.SetTag(TraceTagCriteria, collection)
	accessLog(ctx, p.LiveServers(), FuncDropCollection, collection, start)
	return
}

func (p *Pool) RenameCollection(ctx goctx.Context, dbName, oldName, newName string) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelRenameCollection, start, &err)

	sp := CreateMongoSpan(ctx, FuncRenameCollection)
	defer sp.Finish()

	from := fmt.Sprintf("%s.%s", dbName, oldName)
	to := fmt.Sprintf("%s.%s", dbName, newName)

	cmd := bson.D{{"renameCollection", from}, {"to", to}}
	errDB := p.client.Database("admin", databaseOptions).RunCommand(ctx, cmd).Err()
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Rename Collection from:%s to:%s", from, to)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, p.LiveServers(), FuncRenameCollection, logMsg, start)
	return
}

func (p *Pool) Pipe(ctx goctx.Context, dbName, collection string, pipeline, result interface{}) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelPipe, start, &err)

	sp := CreateMongoSpan(ctx, FuncPipe)
	defer sp.Finish()

	var queryAll bool
	resultsVal := reflect.ValueOf(result)
	if resultsVal.Kind() != reflect.Ptr {
		return gopkg.NewCodeError(NotPtrInput, fmt.Sprintf("results argument must be a pointer type"))
	}

	sliceVal := resultsVal.Elem()
	if sliceVal.Kind() == reflect.Interface {
		sliceVal = sliceVal.Elem()
	}

	if sliceVal.Kind() == reflect.Slice {
		queryAll = true
	}

	cursor, errDB := getMongoCollection(p.client, dbName, collection).Aggregate(ctx, pipeline)
	defer func() {
		logMsg := fmt.Sprintf("Collection:%s,Pipeline:%+v",
			collection,
			pipeline,
		)
		sp.SetTag(TraceTagCriteria, logMsg)
		accessLog(ctx, p.LiveServers(), FuncPipe, logMsg, start)
	}()
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
		return
	}
	if queryAll {
		errDB = cursor.All(ctx, result)
	} else {
		if cursor.Next(ctx) {
			errDB = cursor.Decode(result)
		} else {
			errDB = errors.New(MongoMsgDocumentsNotFound)
		}
	}

	err = p.resultHandling(errDB, ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	return
}

// GetBulk support two usecase
// 1. not set backgroundBatchSize: must run bulk.Run() to trigger operation
// 2. set backgroundBatchSize: each bulk op may trigger background operation,
// this operation will not return bulk result and error
func (p *Pool) GetBulk(ctx goctx.Context, dbName, collection string, backgroundBatchSize ...int) *Bulk {
	bSize := 0
	if len(backgroundBatchSize) > 0 {
		bSize = backgroundBatchSize[0]
	}

	f := false
	b := &Bulk{
		ctx:        ctx,
		dbName:     dbName,
		collection: collection,
		mtx:        &sync.RWMutex{},
		batchSize:  bSize,
		pool:       p,
		options: &options.BulkWriteOptions{
			Ordered: &f,
		},
	}
	return b
}

// cursor related

type Cursor struct {
	cursor     *mongo.Cursor
	ctx        goctx.Context
	pool       *Pool
	collection string
	selector   interface{}
	opt        *options.FindOptions
}

func (p *Pool) GetCursor(ctx goctx.Context, dbName, collection string, selector interface{}, option ...*options.FindOptions) (cursor MongoCursor, err gopkg.CodeError) {
	c, errDB := getMongoCollection(p.client, dbName, collection).Find(ctx, selector, option...)
	err = p.resultHandling(errDB, ctx)
	if err != nil {
		return
	}

	cursor = &Cursor{
		cursor:     c,
		ctx:        ctx,
		pool:       p,
		collection: collection,
		selector:   selector,
		opt:        options.MergeFindOptions(option...),
	}
	return
}

func (c *Cursor) Err() error {
	return c.cursor.Err()
}

func (c *Cursor) Decode(val interface{}) (err gopkg.CodeError) {
	errDB := c.cursor.Decode(val)
	err = c.pool.resultHandling(errDB, c.ctx)
	return
}

func (c *Cursor) Close(ctx context.Context) (err gopkg.CodeError) {
	errDB := c.cursor.Close(ctx)
	err = c.pool.resultHandling(errDB, c.ctx)
	return
}

func (c *Cursor) All(ctx context.Context, result interface{}) (err gopkg.CodeError) {
	sp := CreateMongoSpan(c.ctx, FuncCursorAll)
	defer sp.Finish()
	errDB := c.cursor.All(ctx, result)
	err = c.pool.resultHandling(errDB, c.ctx)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Collection:%s,Query:%+v,Options:%+v",
		c.collection,
		c.selector,
		c.opt,
	)
	sp.SetTag(TraceTagCriteria, logMsg)
	return err
}

func (c *Cursor) ID() int64 {
	return c.cursor.ID()
}

func (c *Cursor) Next(ctx context.Context) bool {
	return c.cursor.Next(ctx)
}

func (c *Cursor) RemainingBatchLength() int {
	return c.cursor.RemainingBatchLength()
}

func (c *Cursor) TryNext(ctx context.Context) bool {
	return c.cursor.TryNext(ctx)
}
