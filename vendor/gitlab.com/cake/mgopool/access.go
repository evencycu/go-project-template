package mgopool

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	opentracing "github.com/opentracing/opentracing-go"
	"gitlab.com/cake/goctx"
	"gitlab.com/cake/gopkg"
	"gitlab.com/cake/gotrace/v2"
)

const (
	TraceTagCriteria = "mongo.criteria"
	TraceTagError    = "mongo.error"
)

const (
	FuncPoolWaiting = "PoolWaiting"
)

const (
	METHOD_CREATE_DOC  = "CreateDocument"
	METHOD_UPDATE_DOC  = "UpdateDocument"
	METHOD_UPDATE_DOCS = "UpdateDocuments"
	METHOD_READ_DOC    = "ReadDocument"
	METHOD_DELETE_DOC  = "DeleteDocument"
	METHOD_DELETE_DOCS = "DeleteDocuments"
	METHOD_CREATE_COL  = "CreateCollection"
	METHOD_DROP_COL    = "DropCollection"
	METHOD_RENAME_COL  = "RenameCollection"
	METHOD_CREATE_BULK = "CreateBulker"
	METHOD_BULK_INSERT = "BulkInsert"
	METHOD_BULK_UPSERT = "BulkUpsert"
	METHOD_BULK_UPDATE = "BulkUpdate"
	METHOD_BULK_DELETE = "BulkDelete"

	FuncBulk               = "Bulk"
	FuncBulkDelete         = "BulkDelete"
	FuncBulkInsert         = "BulkInsert"
	FuncBulkUpdate         = "BulkUpdate"
	FuncBulkUpsert         = "BulkUpsert"
	FuncCollectionCount    = "CollectionCount"
	FuncCreateCollection   = "CreateCollection"
	FuncCreateIndex        = "CreateIndex"
	FuncDBRun              = "DBRun"
	FuncDropCollection     = "DropCollection"
	FuncDropIndex          = "DropIndex"
	FuncDropIndexName      = "DropIndexName"
	FuncDistinct           = "Distinct"
	FuncEnsureIndex        = "EnsureIndex"
	FuncFindAndModify      = "FindAndModify"
	FuncFindAndRemove      = "FindAndRemove"
	FuncGetCollectionNames = "GetCollectionNames"
	FuncIndexes            = "Indexes"
	FuncInsert             = "Insert"
	FuncPing               = "Ping"
	FuncPipe               = "Pipe"
	FuncQueryAll           = "QueryAll"
	FuncQueryCount         = "QueryCount"
	FuncQueryOne           = "QueryOne"
	FuncQueryApply         = "QueryApply"
	FuncQueryDistinct      = "QueryDistinct"
	FuncQueryExplain       = "QueryExplain"
	FuncQueryMapReduce     = "QueryMapReduce"
	FuncRemove             = "Remove"
	FuncRemoveAll          = "RemoveAll"
	FuncRenameCollection   = "RenameCollection"
	FuncRun                = "Run"
	FuncUpdate             = "Update"
	FuncUpdateAll          = "UpdateAll"
	FuncUpdateId           = "UpdateId"
	FuncUpsert             = "Upsert"
)

type MongoPool interface {
	Init(dbi *DBInfo) error
	IsAvailable() bool
	Len() int
	LiveServers() []string
	Cap() int
	Mode() mgo.Mode
	Config() *DBInfo
	Close()
	ShowConfig() map[string]interface{}
	Recover() error

	Ping(ctx goctx.Context) (err gopkg.CodeError)
	GetCollectionNames(ctx goctx.Context, dbName string) (names []string, err gopkg.CodeError)
	CollectionCount(ctx goctx.Context, dbName, collection string) (n int, err gopkg.CodeError)
	Run(ctx goctx.Context, cmd interface{}, result interface{}) gopkg.CodeError
	DBRun(ctx goctx.Context, dbName string, cmd, result interface{}) gopkg.CodeError
	Insert(ctx goctx.Context, dbName, collection string, doc interface{}) gopkg.CodeError
	Remove(ctx goctx.Context, dbName, collection string, selector interface{}) (err gopkg.CodeError)
	RemoveAll(ctx goctx.Context, dbName, collection string, selector interface{}) (info *mgo.ChangeInfo, err gopkg.CodeError)
	Update(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (err gopkg.CodeError)
	UpdateAll(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (info *mgo.ChangeInfo, err gopkg.CodeError)
	UpdateId(ctx goctx.Context, dbName, collection string, id interface{}, update interface{}) (err gopkg.CodeError)
	Upsert(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (info *mgo.ChangeInfo, err gopkg.CodeError)
	BulkInsert(ctx goctx.Context, dbName, collection string, documents []bson.M) (err gopkg.CodeError)
	BulkUpsert(ctx goctx.Context, dbName, collection string, selectors, documents []bson.M) (result *mgo.BulkResult, err gopkg.CodeError)
	BulkUpdate(ctx goctx.Context, dbName, collection string, selectors, documents []bson.M) (result *mgo.BulkResult, err gopkg.CodeError)
	BulkDelete(ctx goctx.Context, dbName, collection string, documents []bson.M) (result *mgo.BulkResult, err gopkg.CodeError)
	GetBulk(ctx goctx.Context, dbName, collection string) (MongoBulk, gopkg.CodeError)
	GetQuery(ctx goctx.Context, dbName, collection string, selector interface{}) (MongoQuery, gopkg.CodeError)
	QueryCount(ctx goctx.Context, dbName, collection string, selector interface{}) (n int, err gopkg.CodeError)
	QueryAll(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, skip, limit int, sort ...string) (err gopkg.CodeError)
	QueryAllWithCollation(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, collation *mgo.Collation, skip, limit int, sort ...string) (err gopkg.CodeError)
	QueryOne(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, skip, limit int, sort ...string) (err gopkg.CodeError)
	FindAndModify(ctx goctx.Context, dbName, collection string, result, selector, update, fields interface{}, skip, limit int, upsert, returnNew bool, sort ...string) (info *mgo.ChangeInfo, err gopkg.CodeError)
	FindAndRemove(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, skip, limit int, sort ...string) (info *mgo.ChangeInfo, err gopkg.CodeError)
	Indexes(ctx goctx.Context, dbName, collection string) (result []mgo.Index, err gopkg.CodeError)
	CreateIndex(ctx goctx.Context, dbName, collection string, key []string, sparse, unique bool, name string) (err gopkg.CodeError)
	EnsureIndex(ctx goctx.Context, dbName, collection string, index mgo.Index) (err gopkg.CodeError)
	DropIndex(ctx goctx.Context, dbName, collection string, keys []string) (err gopkg.CodeError)
	DropIndexName(ctx goctx.Context, dbName, collection, name string) (err gopkg.CodeError)
	CreateCollection(ctx goctx.Context, dbName, collection string, info *mgo.CollectionInfo) (err gopkg.CodeError)
	DropCollection(ctx goctx.Context, dbName, collection string) (err gopkg.CodeError)
	RenameCollection(ctx goctx.Context, dbName, oldName, newName string) (err gopkg.CodeError)
	Pipe(ctx goctx.Context, dbName, collection string, pipeline, result interface{}) (err gopkg.CodeError)
}

type MongoBulk interface {
	Insert(docs ...interface{})
	Update(pairs ...interface{})
	Upsert(pairs ...interface{})
	Remove(selectors ...interface{})
	RemoveAll(selectors ...interface{})
	Run() (result *mgo.BulkResult, err gopkg.CodeError)
	RunBulkError() (*mgo.BulkResult, *mgo.BulkError, gopkg.CodeError)
}

type MongoQuery interface {
	All(result interface{}) gopkg.CodeError
	Apply(change mgo.Change, result interface{}) (info *mgo.ChangeInfo, err gopkg.CodeError)
	Batch(n int) MongoQuery
	Collation(collation *mgo.Collation) MongoQuery
	Comment(comment string) MongoQuery
	Count() (n int, err gopkg.CodeError)
	Distinct(key string, result interface{}) gopkg.CodeError
	Explain(result interface{}) gopkg.CodeError

	Hint(indexKey ...string) MongoQuery
	Limit(n int) MongoQuery
	LogReplay() MongoQuery
	MapReduce(job *mgo.MapReduce, result interface{}) (info *mgo.MapReduceInfo, err gopkg.CodeError)
	One(result interface{}) (err gopkg.CodeError)
	Prefetch(p float64) MongoQuery
	Select(selector interface{}) MongoQuery
	SetMaxScan(n int) MongoQuery
	SetMaxTime(d time.Duration) MongoQuery
	Skip(n int) MongoQuery
	Snapshot() MongoQuery
	Sort(fields ...string) MongoQuery

	// add Iter when need
	// Iter() *Iter
	// Tail(timeout time.Duration) *Iter
	// For(result interface{}, f func() error) gopkg.CodeError
}

func CreateMongoSpan(ctx goctx.Context, funcName string) opentracing.Span {
	return gotrace.CreateChildOfSpan(ctx, funcName)
}

func needReconnect(s string) bool {
	switch {
	// dc, cluster not healthy
	case strings.Contains(s, MongoMsgNotMaster), strings.Contains(s, MongoMsgNoReachableServers), strings.Contains(s, MongoMsgEOF), strings.Contains(s, MongoMsgKernelEOF), strings.Contains(s, MongoMsgClose):
		return true
		// no host case
	case strings.HasPrefix(s, MongoMsgNoHost), strings.HasPrefix(s, MongoMsgNoHost2), strings.HasPrefix(s, MongoMsgWriteUnavailable):
		return true
		// io timeout case
	case strings.HasSuffix(s, MongoMsgTimeout), strings.HasPrefix(s, MongoMsgReadTCP), strings.HasPrefix(s, MongoMsgWriteTCP):
		return true
	}

	return false
}

func badUpdateOperator(errorString string) bool {
	if strings.HasPrefix(errorString, MongoMsgEmptySet) || strings.HasPrefix(errorString, MongoMsgEmptyUnset) || strings.HasPrefix(errorString, MongoMsgBadModifier) || strings.HasPrefix(errorString, MongoMsgEmptyInc) || strings.HasPrefix(errorString, MongoMsgEmptyRename) {
		return true
	}
	return false
}

func getMongoCollection(s *Session, dbName, colName string) *mgo.Collection {
	return s.Session().DB(dbName).C(colName)
}

// resultHandling
// 1. put back session to pool
// 2. check the error and do error handling
func (p *Pool) resultHandling(err error, ctx goctx.Context, s *Session) gopkg.CodeError {
	if err == nil {
		p.put(s)
		return nil
	}
	errorString := err.Error()
	code := UnknownError

	if needReconnect(errorString) {
		// do reconnect
		errLog(ctx, s.Addr(), fmt.Sprintf("API Lost Connection, error:%s", errorString))
		go p.backgroundReconnect(s)
		code = APIConnectDatabase
		// special case, we won't put session right now
		ctx.Set(goctx.LogKeyErrorCode, code)
		return gopkg.NewCarrierCodeError(code, errorString)
	}

	// return pool first
	p.put(s)
	switch {
	case err == mgo.ErrNotFound || strings.HasPrefix(errorString, MongoMsgUnknown):
		code = NotFound
	case errorString == MongoMsgNsNotFound || errorString == MongoMsgCollectionNotFound || errorString == MongoMsgNamespaceNotFound:
		code = CollectionNotFound
	case strings.HasPrefix(errorString, MongoMsgE11000) || strings.HasPrefix(errorString, MongoMsgBulk):
		code = DocumentConflict
		errorString = "document already exists:" + strings.Replace(errorString, MongoMsgE11000, "", 1)
	case strings.HasSuffix(errorString, MongoMsgCollectionConflict):
		code = CollectionConflict
	case strings.HasSuffix(errorString, MongoMsgArray):
		code = QueryInputArray
		errorString = "query format error:" + errorString
	case strings.HasPrefix(errorString, MongoMsgEachArray) || strings.HasPrefix(errorString, MongoMsgPullAllArray):
		code = UpdateInputArray
		errorString = "Add/AddUnique/Remove requires an array argument"
	case badUpdateOperator(errorString):
		code = BadUpdateOperatorUsage
		errorString = "update requires an object and not empty"
	case strings.HasPrefix(errorString, MongoMsgIncrement):
		code = IncrementNumeric
	case errorString == MongoMsgRegexString:
		code = RegexString
	case strings.HasPrefix(errorString, MongoMsgDotField):
		code = DotField
	case strings.HasPrefix(errorString, MongoMsgwiredTigerIndex):
		code = StringIndexTooLong
	}
	ctx.Set(goctx.LogKeyErrorCode, code)
	return gopkg.NewCarrierCodeError(code, errorString)
}

func (p *Pool) Ping(ctx goctx.Context) (err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}
	err = p.resultHandling(s.Session().Ping(), ctx, s)
	return
}

// GetCollectionNames return all collection names in the db.
func (p *Pool) GetCollectionNames(ctx goctx.Context, dbName string) (names []string, err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}
	names, errDB := s.Session().DB(dbName).CollectionNames()
	err = p.resultHandling(errDB, ctx, s)
	return
}

func (p *Pool) CollectionCount(ctx goctx.Context, dbName, collection string) (n int, err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	// start := time.Now()
	col := getMongoCollection(s, dbName, collection)

	sp := CreateMongoSpan(ctx, FuncCollectionCount)
	defer sp.Finish()
	n, errDB := col.Count()
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	sp.SetTag(TraceTagCriteria, collection)
	// accessLog(ctx, s.Addr(), METHOD_CREATE_DOC, fmt.Sprintf("Collection:%s,Stuff:%+v", collection, doc), start)
	return
}

func (p *Pool) Run(ctx goctx.Context, cmd interface{}, result interface{}) gopkg.CodeError {
	s, err := p.get(ctx)
	if err != nil {
		return err
	}

	sp := CreateMongoSpan(ctx, FuncRun)
	defer sp.Finish()
	errDB := s.Session().Run(cmd, result)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("DB Operations:%+v ,Result:%+v", cmd, result)
	sp.SetTag(TraceTagCriteria, logMsg)
	infoLog(ctx, s.Addr(), logMsg)
	return err
}

func (p *Pool) DBRun(ctx goctx.Context, dbName string, cmd, result interface{}) gopkg.CodeError {
	s, err := p.get(ctx)
	if err != nil {
		return err
	}

	sp := CreateMongoSpan(ctx, FuncDBRun)
	defer sp.Finish()
	errDB := s.Session().DB(dbName).Run(cmd, result)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("DB Operations:%+v ,Result:%+v", cmd, result)
	sp.SetTag(TraceTagCriteria, logMsg)
	infoLog(ctx, s.Addr(), logMsg)
	return err
}

func (p *Pool) Insert(ctx goctx.Context, dbName, collection string, doc interface{}) gopkg.CodeError {
	s, err := p.get(ctx)
	if err != nil {
		return err
	}

	start := time.Now()
	col := getMongoCollection(s, dbName, collection)

	// Insert document to collection
	sp := CreateMongoSpan(ctx, FuncInsert)
	defer sp.Finish()
	errDB := col.Insert(doc)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Collection:%s,Stuff:%+v", collection, doc)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, s.Addr(), METHOD_CREATE_DOC, logMsg, start)
	return err
}

func (p *Pool) Distinct(ctx goctx.Context, dbName, collection string, selector bson.M, field string, result interface{}) (err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	start := time.Now()
	col := getMongoCollection(s, dbName, collection)
	sp := CreateMongoSpan(ctx, FuncDistinct)
	defer sp.Finish()
	errDB := col.Find(selector).Distinct(field, result)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	logMsg := fmt.Sprintf("Collection:%s,Field:%+v", collection, field)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, s.Addr(), METHOD_DELETE_DOC, logMsg, start)
	return
}

func (p *Pool) Remove(ctx goctx.Context, dbName, collection string, selector interface{}) (err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	start := time.Now()
	col := getMongoCollection(s, dbName, collection)
	sp := CreateMongoSpan(ctx, FuncRemove)
	defer sp.Finish()
	errDB := col.Remove(selector)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v", collection, selector)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, s.Addr(), METHOD_DELETE_DOC, logMsg, start)
	return
}
func (p *Pool) RemoveAll(ctx goctx.Context, dbName, collection string, selector interface{}) (info *mgo.ChangeInfo, err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	start := time.Now()
	col := getMongoCollection(s, dbName, collection)

	sp := CreateMongoSpan(ctx, FuncRemoveAll)
	defer sp.Finish()
	info, errDB := col.RemoveAll(selector)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v", collection, selector)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, s.Addr(), METHOD_DELETE_DOCS, logMsg, start)
	return
}

func (p *Pool) Update(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	start := time.Now()
	col := getMongoCollection(s, dbName, collection)

	sp := CreateMongoSpan(ctx, FuncUpdate)
	defer sp.Finish()
	errDB := col.Update(selector, update)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v,Update:%+v", collection, selector, update)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, s.Addr(), METHOD_UPDATE_DOC, logMsg, start)
	return
}

func (p *Pool) UpdateAll(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (info *mgo.ChangeInfo, err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	start := time.Now()
	col := getMongoCollection(s, dbName, collection)

	sp := CreateMongoSpan(ctx, FuncUpdateAll)
	defer sp.Finish()
	info, errDB := col.UpdateAll(selector, update)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v,Update:%+v", collection, selector, update)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, s.Addr(), METHOD_UPDATE_DOCS, logMsg, start)
	return
}

func (p *Pool) UpdateId(ctx goctx.Context, dbName, collection string, id interface{}, update interface{}) (err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	start := time.Now()
	col := getMongoCollection(s, dbName, collection)

	sp := CreateMongoSpan(ctx, FuncUpdateId)
	defer sp.Finish()
	errDB := col.UpdateId(id, update)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v,Update:%+v", collection, id, update)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, s.Addr(), METHOD_UPDATE_DOC, logMsg, start)
	return
}

func (p *Pool) Upsert(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (info *mgo.ChangeInfo, err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	start := time.Now()
	col := getMongoCollection(s, dbName, collection)

	sp := CreateMongoSpan(ctx, FuncUpsert)
	defer sp.Finish()
	var errDB error
	info, errDB = col.Upsert(selector, update)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
		return
	}
	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v,Upsert:%+v", collection, selector, update)
	if info.UpsertedId != nil {
		accessLog(ctx, s.Addr(), METHOD_CREATE_DOC, logMsg, start)
	} else {
		accessLog(ctx, s.Addr(), METHOD_UPDATE_DOC, logMsg, start)
	}
	sp.SetTag(TraceTagCriteria, logMsg)
	return
}

func (p *Pool) BulkInsert(ctx goctx.Context, dbName, collection string, documents []bson.M) (err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	start := time.Now()
	col := getMongoCollection(s, dbName, collection)
	bulk := col.Bulk()
	bulk.Unordered()
	b := len(documents)
	for i := 0; i < b; i++ {
		// NOTE: bson.NewObjectId would fail with goroutine, even there is only one goroutine worker.
		bulk.Insert(documents[i])
	}
	// Set document _id if not set
	// Insert document to collection
	sp := CreateMongoSpan(ctx, FuncBulkInsert)
	defer sp.Finish()
	_, errDB := bulk.Run()
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	sp.SetTag(TraceTagCriteria, collection)
	accessLog(ctx, s.Addr(), METHOD_BULK_INSERT, collection, start)
	return
}

func (p *Pool) BulkUpsert(ctx goctx.Context, dbName, collection string, selectors, documents []bson.M) (result *mgo.BulkResult, err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	start := time.Now()
	col := getMongoCollection(s, dbName, collection)
	bulk := col.Bulk()
	bulk.Unordered()
	b := len(documents)
	for i := 0; i < b; i++ {
		// NOTE: bson.NewObjectId would fail with goroutine, even there is only one goroutine worker.
		bulk.Upsert(selectors[i], documents[i])
	}
	// Set document _id if not set
	// Insert document to collection
	sp := CreateMongoSpan(ctx, FuncBulkUpsert)
	defer sp.Finish()
	var errDB error
	result, errDB = bulk.Run()
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	sp.SetTag(TraceTagCriteria, collection)
	accessLog(ctx, s.Addr(), METHOD_BULK_UPSERT, collection, start)
	return
}

func (p *Pool) BulkUpdate(ctx goctx.Context, dbName, collection string, selectors, documents []bson.M) (result *mgo.BulkResult, err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	start := time.Now()
	col := getMongoCollection(s, dbName, collection)
	bulk := col.Bulk()
	bulk.Unordered()
	b := len(documents)
	for i := 0; i < b; i++ {
		// NOTE: bson.NewObjectId would fail with goroutine, even there is only one goroutine worker.
		bulk.Update(selectors[i], documents[i])
	}
	// Set document _id if not set
	// Insert document to collection
	sp := CreateMongoSpan(ctx, FuncBulkUpdate)
	defer sp.Finish()
	var errDB error
	result, errDB = bulk.Run()
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	sp.SetTag(TraceTagCriteria, collection)
	accessLog(ctx, s.Addr(), METHOD_BULK_UPDATE, collection, start)
	return
}

func (p *Pool) BulkDelete(ctx goctx.Context, dbName, collection string, documents []bson.M) (result *mgo.BulkResult, err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	start := time.Now()
	col := getMongoCollection(s, dbName, collection)
	bulk := col.Bulk()
	bulk.Unordered()
	b := len(documents)
	for i := 0; i < b; i++ {
		// NOTE: bson.NewObjectId would fail with goroutine, even there is only one goroutine worker.
		bulk.Remove(documents[i])
	}
	// Set document _id if not set
	// Insert document to collection
	sp := CreateMongoSpan(ctx, FuncBulkDelete)
	defer sp.Finish()
	var errDB error
	result, errDB = bulk.Run()
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	sp.SetTag(TraceTagCriteria, collection)
	accessLog(ctx, s.Addr(), METHOD_BULK_DELETE, collection, start)
	return
}

func (p *Pool) GetBulk(ctx goctx.Context, dbName, collection string) (MongoBulk, gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return nil, err
	}

	col := getMongoCollection(s, dbName, collection)
	bulk := col.Bulk()
	bulk.Unordered()
	return &Bulk{
		bulk:       bulk,
		collection: collection,
		ctx:        ctx,
		session:    s,
		pool:       p,
	}, nil
}

type Query struct {
	query      *mgo.Query
	pool       *Pool
	ctx        goctx.Context
	session    *Session
	collection string
	selector   interface{}
	skip       int
	limit      int
}

func (p *Pool) GetQuery(ctx goctx.Context, dbName, collection string, selector interface{}) (MongoQuery, gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return nil, err
	}

	q := getMongoCollection(s, dbName, collection).Find(selector)
	return &Query{
		query:      q,
		collection: collection,
		ctx:        ctx,
		session:    s,
		pool:       p,
		selector:   selector,
	}, nil
}

func (q *Query) All(result interface{}) gopkg.CodeError {
	start := time.Now()
	sp := CreateMongoSpan(q.ctx, FuncQueryAll)
	defer sp.Finish()
	errDB := q.query.All(result)
	err := q.pool.resultHandling(errDB, q.ctx, q.session)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Collection:%s,Query:%+v,Skip:%d,Limit:%d",
		q.collection,
		q.selector,
		q.skip,
		q.limit,
	)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(q.ctx, q.session.Addr(), METHOD_READ_DOC, logMsg, start)
	return err
}
func (q *Query) Apply(change mgo.Change, result interface{}) (info *mgo.ChangeInfo, err gopkg.CodeError) {
	start := time.Now()
	sp := CreateMongoSpan(q.ctx, FuncQueryApply)
	defer sp.Finish()
	var errDB error
	info, errDB = q.query.Apply(change, result)
	err = q.pool.resultHandling(errDB, q.ctx, q.session)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("[Apply]Collection:%s,Query:%+v,Skip:%d,Limit:%d",
		q.collection,
		q.selector,
		q.skip,
		q.limit,
	)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(q.ctx, q.session.Addr(), METHOD_READ_DOC, logMsg, start)
	return
}
func (q *Query) Batch(n int) MongoQuery {
	q.query = q.query.Batch(n)
	return q
}
func (q *Query) Collation(collation *mgo.Collation) MongoQuery {
	q.query = q.query.Collation(collation)
	return q
}
func (q *Query) Comment(comment string) MongoQuery {
	q.query = q.query.Comment(comment)
	return q
}
func (q *Query) Count() (n int, err gopkg.CodeError) {
	start := time.Now()
	sp := CreateMongoSpan(q.ctx, FuncQueryCount)
	defer sp.Finish()
	var errDB error
	n, errDB = q.query.Count()
	err = q.pool.resultHandling(errDB, q.ctx, q.session)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("[Count]Collection:%s,Query:%+v",
		q.collection,
		q.selector,
	)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(q.ctx, q.session.Addr(), METHOD_READ_DOC, logMsg, start)
	return
}
func (q *Query) Distinct(key string, result interface{}) gopkg.CodeError {
	start := time.Now()
	sp := CreateMongoSpan(q.ctx, FuncQueryCount)
	defer sp.Finish()

	errDB := q.query.Distinct(key, result)
	err := q.pool.resultHandling(errDB, q.ctx, q.session)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("[Distinct]Collection:%s,Query:%+v,Skip:%d,Limit:%d",
		q.collection,
		q.selector,
		q.skip,
		q.limit,
	)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(q.ctx, q.session.Addr(), METHOD_READ_DOC, logMsg, start)
	return err
}
func (q *Query) Explain(result interface{}) gopkg.CodeError {
	start := time.Now()
	sp := CreateMongoSpan(q.ctx, FuncQueryExplain)
	defer sp.Finish()

	errDB := q.query.Explain(result)
	err := q.pool.resultHandling(errDB, q.ctx, q.session)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("[Explain]Collection:%s,Query:%+v,Skip:%d,Limit:%d",
		q.collection,
		q.selector,
		q.skip,
		q.limit,
	)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(q.ctx, q.session.Addr(), METHOD_READ_DOC, logMsg, start)
	return err
}

func (q *Query) Hint(indexKey ...string) MongoQuery {
	q.query = q.query.Hint(indexKey...)
	return q
}
func (q *Query) Limit(n int) MongoQuery {
	q.query = q.query.Limit(n)
	return q
}
func (q *Query) LogReplay() MongoQuery {
	q.query = q.query.LogReplay()
	return q
}
func (q *Query) MapReduce(job *mgo.MapReduce, result interface{}) (info *mgo.MapReduceInfo, err gopkg.CodeError) {
	start := time.Now()
	sp := CreateMongoSpan(q.ctx, FuncQueryMapReduce)
	defer sp.Finish()
	var errDB error
	info, errDB = q.query.MapReduce(job, result)
	err = q.pool.resultHandling(errDB, q.ctx, q.session)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Collection:%s,Query:%+v,Skip:%d,Limit:%d,MapReduce:%+v",
		q.collection,
		q.selector,
		q.skip,
		q.limit,
		job,
	)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(q.ctx, q.session.Addr(), METHOD_READ_DOC, logMsg, start)
	return
}
func (q *Query) One(result interface{}) (err gopkg.CodeError) {
	start := time.Now()
	sp := CreateMongoSpan(q.ctx, FuncQueryOne)
	defer sp.Finish()

	errDB := q.query.One(result)
	err = q.pool.resultHandling(errDB, q.ctx, q.session)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("[One]Collection:%s,Query:%+v,Skip:%d",
		q.collection,
		q.selector,
		q.skip,
	)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(q.ctx, q.session.Addr(), METHOD_READ_DOC, logMsg, start)
	return err
}
func (q *Query) Prefetch(p float64) MongoQuery {
	q.query = q.query.Prefetch(p)
	return q
}
func (q *Query) Select(selector interface{}) MongoQuery {
	q.query = q.query.Select(selector)
	q.selector = selector
	return q
}
func (q *Query) SetMaxScan(n int) MongoQuery {
	q.query = q.query.SetMaxScan(n)
	return q
}
func (q *Query) SetMaxTime(d time.Duration) MongoQuery {
	q.query = q.query.SetMaxTime(d)
	return q
}
func (q *Query) Skip(n int) MongoQuery {
	q.query = q.query.Skip(n)
	return q
}
func (q *Query) Snapshot() MongoQuery {
	q.query = q.query.Snapshot()
	return q
}
func (q *Query) Sort(fields ...string) MongoQuery {
	q.query = q.query.Sort(fields...)
	return q
}

type Bulk struct {
	bulk       *mgo.Bulk
	pool       *Pool
	ctx        goctx.Context
	session    *Session
	collection string
}

func (b *Bulk) Insert(docs ...interface{}) {
	b.bulk.Insert(docs...)
}
func (b *Bulk) Update(pairs ...interface{}) {
	b.bulk.Update(pairs...)
}
func (b *Bulk) Upsert(pairs ...interface{}) {
	b.bulk.Upsert(pairs...)
}
func (b *Bulk) Remove(selectors ...interface{}) {
	b.bulk.Remove(selectors...)
}
func (b *Bulk) RemoveAll(selectors ...interface{}) {
	b.bulk.RemoveAll(selectors...)
}
func (b *Bulk) Run() (result *mgo.BulkResult, err gopkg.CodeError) {
	result, _, err = b.RunBulkError()
	return
}

func (b *Bulk) RunBulkError() (*mgo.BulkResult, *mgo.BulkError, gopkg.CodeError) {
	start := time.Now()
	sp := CreateMongoSpan(b.ctx, FuncBulk)
	defer sp.Finish()
	result, errDB := b.bulk.Run()
	err := b.pool.resultHandling(errDB, b.ctx, b.session)
	var errB *mgo.BulkError
	if errDB != nil {
		errB = errDB.(*mgo.BulkError)
	}
	sp.SetTag(TraceTagCriteria, b.collection)
	accessLog(b.ctx, b.session.Addr(), METHOD_CREATE_BULK, b.collection, start)
	return result, errB, err
}

//expect data is the pointer to an array or a slice
//elements at error indexes will be removed from data
func RemoveErrorElements(ctx goctx.Context, cases []mgo.BulkErrorCase, data interface{}) gopkg.CodeError {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Ptr {
		return gopkg.NewCodeError(TypeNotSupported, "data is not a pointer")
	}
	switch v.Elem().Kind() {
	case reflect.Array, reflect.Slice:
		slice := v.Elem()
		if slice.Len() == 0 {
			return nil
		}
		idx := 0
		i := 0
		newLen := slice.Len() - len(cases)
		result := reflect.MakeSlice(reflect.SliceOf(slice.Index(0).Type()), newLen, newLen)
		for _, errCase := range cases {
			subList := slice.Slice(i, errCase.Index)
			for j := 0; j < subList.Len(); j++ {
				result.Index(idx).Set(subList.Index(j))
				idx++
			}
			i = errCase.Index + 1
		}

		subList := slice.Slice(i, slice.Len())
		for j := 0; j < subList.Len(); j++ {
			result.Index(idx).Set(subList.Index(j))
			idx++
		}
		slice.Set(result)
		return nil
	}

	return gopkg.NewCodeError(TypeNotSupported, "unsupported type")
}

func (p *Pool) QueryCount(ctx goctx.Context, dbName, collection string, selector interface{}) (n int, err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	start := time.Now()
	query := getMongoCollection(s, dbName, collection).Find(selector)

	sp := CreateMongoSpan(ctx, FuncQueryCount)
	defer sp.Finish()
	n, errDB := query.Count()
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Collection:%s,Query:%+v",
		collection,
		selector,
	)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, s.Addr(), METHOD_READ_DOC, logMsg, start)
	return
}

func (p *Pool) QueryAll(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, skip, limit int, sort ...string) (err gopkg.CodeError) {
	return p.QueryAllWithCollation(ctx, dbName, collection, result, selector, fields, nil, skip, limit, sort...)
}

func (p *Pool) QueryAllWithCollation(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, collation *mgo.Collation, skip, limit int, sort ...string) (err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	start := time.Now()
	query := getMongoCollection(s, dbName, collection).Find(selector)
	if fields != nil {
		query.Select(fields)
	}
	query.Skip(skip)
	query.Limit(limit)
	if len(sort) > 0 {
		query.Sort(sort...)
	}
	if collation != nil {
		query.Collation(collation)
	}

	sp := CreateMongoSpan(ctx, FuncQueryAll)
	defer sp.Finish()
	errDB := query.All(result)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Collection:%s,Query:%+v,Projection::%+v,Sort:%+v,Skip:%d,Limit:%d",
		collection,
		selector,
		fields,
		sort,
		skip,
		limit,
	)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, s.Addr(), METHOD_READ_DOC, logMsg, start)
	return
}

func (p *Pool) QueryOne(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, skip, limit int, sort ...string) (err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	start := time.Now()
	query := getMongoCollection(s, dbName, collection).Find(selector)
	if fields != nil {
		query.Select(fields)
	}
	query.Skip(skip)
	query.Limit(limit)
	if len(sort) > 0 {
		query.Sort(sort...)
	}

	sp := CreateMongoSpan(ctx, FuncQueryOne)
	defer sp.Finish()
	errDB := query.One(result)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Collection:%s,Filter:%+v,Projection::%+v,Sort:%+v,Skip:%d,Limit:%d",
		collection,
		selector,
		fields,
		sort,
		skip,
		limit,
	)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, s.Addr(), METHOD_READ_DOC, logMsg, start)
	return
}

// FindAndModify can only update one doc (by mongodb description)
func (p *Pool) FindAndModify(ctx goctx.Context, dbName, collection string, result, selector, update, fields interface{},
	skip, limit int, upsert, returnNew bool, sort ...string) (info *mgo.ChangeInfo, err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}
	start := time.Now()
	query := getMongoCollection(s, dbName, collection).Find(selector)
	if fields != nil {
		query.Select(fields)
	}
	query.Skip(skip)
	query.Limit(limit)
	if len(sort) > 0 {
		query.Sort(sort...)
	}

	change := mgo.Change{
		Update:    update,
		ReturnNew: returnNew,
		Upsert:    upsert,
	}

	sp := CreateMongoSpan(ctx, FuncFindAndModify)
	defer sp.Finish()
	var errDB error
	info, errDB = query.Apply(change, result)

	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Collection:%s,Query:%+v,Projection::%+v,Sort:%+v,Skip:%d,Limit:%d,Change:%+v",
		collection,
		selector,
		fields,
		sort,
		skip,
		limit,
		change,
	)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, s.Addr(), METHOD_UPDATE_DOC, logMsg, start)
	return
}

func (p *Pool) FindAndRemove(ctx goctx.Context, dbName, collection string, result, selector, fields interface{},
	skip, limit int, sort ...string) (info *mgo.ChangeInfo, err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	start := time.Now()
	query := getMongoCollection(s, dbName, collection).Find(selector)
	if fields != nil {
		query.Select(fields)
	}
	query.Skip(skip)
	query.Limit(limit)
	if len(sort) > 0 {
		query.Sort(sort...)
	}

	change := mgo.Change{
		Remove: true,
	}

	sp := CreateMongoSpan(ctx, FuncFindAndRemove)
	defer sp.Finish()
	var errDB error
	info, errDB = query.Apply(change, result)

	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Collection:%s,Query:%+v,Projection::%+v,Sort:%+v,Skip:%d,Limit:%d,Change:%+v",
		collection,
		selector,
		fields,
		sort,
		skip,
		limit,
		change,
	)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, s.Addr(), METHOD_READ_DOC, logMsg, start)
	return
}

func (p *Pool) Indexes(ctx goctx.Context, dbName, collection string) (result []mgo.Index, err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	sp := CreateMongoSpan(ctx, FuncIndexes)
	defer sp.Finish()
	result, errDB := getMongoCollection(s, dbName, collection).Indexes()
	err = p.resultHandling(errDB, ctx, s)
	sp.SetTag(TraceTagCriteria, collection)
	return
}

func (p *Pool) CreateIndex(ctx goctx.Context, dbName, collection string, key []string, sparse, unique bool, name string) (err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	index := mgo.Index{
		Key:        key,
		Background: true,
		Sparse:     sparse,
		Unique:     unique,
		Name:       name,
	}
	sp := CreateMongoSpan(ctx, FuncCreateIndex)
	defer sp.Finish()
	errDB := getMongoCollection(s, dbName, collection).EnsureIndex(index)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
		errLog(ctx, s.Addr(), collection+" add index:"+strings.Join(index.Key, ",")+" err:"+err.Error())
	}
	sp.SetTag(TraceTagCriteria, fmt.Sprintf("Collection:%s,Create Index:%+v", collection, index))
	return
}

func (p *Pool) EnsureIndex(ctx goctx.Context, dbName, collection string, index mgo.Index) (err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}
	sp := CreateMongoSpan(ctx, FuncEnsureIndex)
	defer sp.Finish()
	errDB := getMongoCollection(s, dbName, collection).EnsureIndex(index)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
		errLog(ctx, s.Addr(), collection+" add index:"+strings.Join(index.Key, ",")+" err:"+err.Error())
	}

	sp.SetTag(TraceTagCriteria, fmt.Sprintf("Collection:%s,EnsureIndex:%+v", collection, index))
	return
}

func (p *Pool) DropIndex(ctx goctx.Context, dbName, collection string, keys []string) (err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}
	sp := CreateMongoSpan(ctx, FuncDropIndex)
	defer sp.Finish()
	errDB := getMongoCollection(s, dbName, collection).DropIndex(keys...)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	sp.SetTag(TraceTagCriteria, fmt.Sprintf("Collection:%s,DropIndex Key:%+v", collection, keys))
	return
}

func (p *Pool) DropIndexName(ctx goctx.Context, dbName, collection, name string) (err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}
	sp := CreateMongoSpan(ctx, FuncDropIndexName)
	defer sp.Finish()
	errDB := getMongoCollection(s, dbName, collection).DropIndexName(name)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	sp.SetTag(TraceTagCriteria, fmt.Sprintf("Collection:%s,DropIndexName:%s", collection, name))
	return
}

func (p *Pool) CreateCollection(ctx goctx.Context, dbName, collection string, info *mgo.CollectionInfo) (err gopkg.CodeError) {
	start := time.Now()
	s, err := p.get(ctx)
	if err != nil {
		return
	}
	sp := CreateMongoSpan(ctx, FuncCreateCollection)
	defer sp.Finish()
	errDB := getMongoCollection(s, dbName, collection).Create(info)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	sp.SetTag(TraceTagCriteria, collection)
	accessLog(ctx, s.Addr(), METHOD_CREATE_COL, collection, start)
	return
}

func (p *Pool) DropCollection(ctx goctx.Context, dbName, collection string) (err gopkg.CodeError) {
	start := time.Now()
	s, err := p.get(ctx)
	if err != nil {
		return
	}
	sp := CreateMongoSpan(ctx, FuncDropCollection)
	defer sp.Finish()
	errDB := getMongoCollection(s, dbName, collection).DropCollection()
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}

	sp.SetTag(TraceTagCriteria, collection)
	accessLog(ctx, s.Addr(), METHOD_DROP_COL, collection, start)
	return
}

func (p *Pool) RenameCollection(ctx goctx.Context, dbName, oldName, newName string) (err gopkg.CodeError) {
	start := time.Now()
	s, err := p.get(ctx)
	if err != nil {
		return
	}
	// result is useless, error has message, so we just throw
	result := bson.M{}
	from := fmt.Sprintf("%s.%s", dbName, oldName)
	to := fmt.Sprintf("%s.%s", dbName, newName)

	sp := CreateMongoSpan(ctx, FuncRenameCollection)
	defer sp.Finish()
	errDB := s.Session().Run(bson.D{{Name: "renameCollection", Value: from}, {Name: "to", Value: to}}, result)
	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Rename Collection from:%s to:%s", from, to)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, s.Addr(), METHOD_RENAME_COL, logMsg, start)
	return
}

func (p *Pool) Pipe(ctx goctx.Context, dbName, collection string, pipeline, result interface{}) (err gopkg.CodeError) {
	start := time.Now()
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	sp := CreateMongoSpan(ctx, FuncPipe)
	defer sp.Finish()
	pipe := getMongoCollection(s, dbName, collection).Pipe(pipeline)
	errDB := pipe.All(result)

	err = p.resultHandling(errDB, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	logMsg := fmt.Sprintf("Collection:%s,Pipeline:%+v",
		collection,
		pipeline,
	)
	sp.SetTag(TraceTagCriteria, logMsg)
	accessLog(ctx, s.Addr(), METHOD_READ_DOC, logMsg, start)
	return
}
