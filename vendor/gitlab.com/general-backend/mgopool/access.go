package mgopool

import (
	"fmt"
	"strings"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	opentracing "github.com/opentracing/opentracing-go"
	"gitlab.com/general-backend/goctx"
	"gitlab.com/general-backend/gopkg"
	"gitlab.com/general-backend/gotrace"
)

const (
	TraceTagCriteria = "mongo.criteria"
	TraceTagError    = "mongo.error"
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
	METHOD_BULK_DELETE = "BulkDelete"

	FuncBulkDelete         = "BulkDelete"
	FuncBulkInsert         = "BulkInsert"
	FuncBulkUpsert         = "BulkUpsert"
	FuncCollectionCount    = "CollectionCount"
	FuncCreateCollection   = "CreateCollection"
	FuncCreateIndex        = "CreateIndex"
	FuncDBRun              = "DBRun"
	FuncDropCollection     = "DropCollection"
	FuncDropIndex          = "DropIndex"
	FuncDropIndexName      = "DropIndexName"
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
	FuncRemove             = "Remove"
	FuncRemoveAll          = "RemoveAll"
	FuncRenameCollection   = "RenameCollection"
	FuncRun                = "Run"
	FuncUpdate             = "Update"
	FuncUpdateAll          = "UpdateAll"
	FuncUpdateId           = "UpdateId"
	FuncUpsert             = "Upsert"
)

func CreateMongoSpan(ctx goctx.Context, funcName, criteria string) opentracing.Span {
	tags := make(map[string]string)
	tags[TraceTagCriteria] = criteria
	sp := gotrace.CreateSpanByContext(funcName, ctx, gotrace.ReferenceChildOf, &gotrace.TagsMap{
		Others: tags,
	})
	return sp
}

func needReconnect(s string) bool {
	switch s {
	case MongoMsgNoReachableServers, MongoMsgEOF, MongoMsgKernelEOF, MongoMsgClose, MongoMsgNotMaster:
		return true
	}
	if strings.HasPrefix(s, MongoMsgNoHost) || strings.HasPrefix(s, MongoMsgNoHost2) || strings.HasPrefix(s, MongoMsgWriteUnavailable) {
		return true
	}
	return false
}

func getMongoCollection(s *Session, dbName, colName string) *mgo.Collection {
	return s.Session().DB(dbName).C(colName)
}

func (p *Pool) checkDatabaseError(err error, ctx goctx.Context, s *Session) gopkg.CodeError {
	if err == nil {
		p.put(s)
		return nil
	}
	errorString := err.Error()
	code := UnknownError

	switch {
	case needReconnect(errorString):
		// do reconnect
		errLog(ctx, s.Addr(), fmt.Sprintf("API Lost Connection, error:%s", errorString))
		go p.backgroundReconnect(s)
		code = APIConnectDatabase
		// special case, we won't put session right now
		ctx.Set(goctx.LogKeyErrorCode, code)
		return gopkg.NewCarrierCodeError(code, errorString)
	case err == mgo.ErrNotFound || strings.HasPrefix(errorString, MongoMsgUnknown):
		code = NotFound
	case errorString == MongoMsgNsNotFound || errorString == MongoMsgCollectionNotFound || errorString == MongoMsgNamespaceNotFound:
		code = CollectionNotFound
	case strings.HasPrefix(errorString, MongoMsgE11000) || strings.HasPrefix(errorString, MongoMsgBulk):
		code = DocumentConflict
		errorString = "document already exists:" + strings.Replace(errorString, MongoMsgE11000, "", 1)
	case strings.HasSuffix(errorString, MongoMsgTimeout) ||
		strings.HasPrefix(errorString, MongoMsgReadTCP) || strings.HasPrefix(errorString, MongoMsgWriteTCP):
		infoLog(ctx, s.Addr(), errorString)
		code = Timeout
		s.Session().Refresh()
	case strings.HasSuffix(errorString, MongoMsgCollectionConflict):
		code = CollectionConflict
	case strings.HasSuffix(errorString, MongoMsgArray):
		code = QueryInputArray
		errorString = "query format error:" + errorString
	case strings.HasPrefix(errorString, MongoMsgEachArray) || strings.HasPrefix(errorString, MongoMsgPullAllArray):
		code = UpdateInputArray
		errorString = "Add/AddUnique/Remove requires an array argument"
	case strings.HasPrefix(errorString, MongoMsgIncrement):
		code = IncrementNumeric
	case errorString == MongoMsgRegexString:
		code = RegexString
	case strings.HasPrefix(errorString, MongoMsgDotField):
		code = DotField
	case strings.HasPrefix(errorString, MongoMsgwiredTigerIndex):
		code = StringIndexTooLong
	}
	p.put(s)
	ctx.Set(goctx.LogKeyErrorCode, code)
	return gopkg.NewCarrierCodeError(code, errorString)
}

func (p *Pool) Ping(ctx goctx.Context) (err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}
	err = p.checkDatabaseError(s.Session().Ping(), ctx, s)
	return
}

// GetCollectionNames return all collection names in the db.
func (p *Pool) GetCollectionNames(ctx goctx.Context, dbName string) (names []string, err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}
	names, dberr := s.Session().DB(dbName).CollectionNames()
	err = p.checkDatabaseError(dberr, ctx, s)
	return
}

func (p *Pool) CollectionCount(ctx goctx.Context, dbName, collection string) (n int, err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	// start := time.Now()
	col := getMongoCollection(s, dbName, collection)

	sp := CreateMongoSpan(ctx, FuncCollectionCount, fmt.Sprintf("Collection:%s", collection))
	defer sp.Finish()
	n, dberr := col.Count()
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	// accessLog(ctx, s.Addr(), METHOD_CREATE_DOC, fmt.Sprintf("Collection:%s,Stuff:%+v", collection, doc), start)
	return
}

func (p *Pool) Run(ctx goctx.Context, cmd interface{}, result interface{}) gopkg.CodeError {
	s, err := p.get(ctx)
	if err != nil {
		return err
	}
	logMsg := fmt.Sprintf("DB Operations:%+v ,Result:%+v", cmd, result)
	sp := CreateMongoSpan(ctx, FuncRun, logMsg)
	defer sp.Finish()
	dberr := s.Session().Run(cmd, result)
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	infoLog(ctx, s.Addr(), logMsg)
	return err
}

func (p *Pool) DBRun(ctx goctx.Context, dbName string, cmd, result interface{}) gopkg.CodeError {
	s, err := p.get(ctx)
	if err != nil {
		return err
	}
	logMsg := fmt.Sprintf("DB Operations:%+v ,Result:%+v", cmd, result)
	sp := CreateMongoSpan(ctx, FuncDBRun, logMsg)
	defer sp.Finish()
	dberr := s.Session().DB(dbName).Run(cmd, result)
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
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
	logMsg := fmt.Sprintf("Collection:%s,Stuff:%+v", collection, doc)
	sp := CreateMongoSpan(ctx, FuncInsert, logMsg)
	defer sp.Finish()
	dberr := col.Insert(doc)
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	accessLog(ctx, s.Addr(), METHOD_CREATE_DOC, logMsg, start)
	return err
}

func (p *Pool) Remove(ctx goctx.Context, dbName, collection string, selector interface{}) (err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	start := time.Now()
	col := getMongoCollection(s, dbName, collection)
	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v", collection, selector)
	sp := CreateMongoSpan(ctx, FuncRemove, logMsg)
	defer sp.Finish()
	dberr := col.Remove(selector)
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
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
	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v", collection, selector)
	sp := CreateMongoSpan(ctx, FuncRemoveAll, logMsg)
	defer sp.Finish()
	info, dberr := col.RemoveAll(selector)
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
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
	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v,Update:%+v", collection, selector, update)
	sp := CreateMongoSpan(ctx, FuncUpdate, logMsg)
	defer sp.Finish()
	dberr := col.Update(selector, update)
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
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
	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v,Update:%+v", collection, selector, update)
	sp := CreateMongoSpan(ctx, FuncUpdateAll, logMsg)
	defer sp.Finish()
	info, dberr := col.UpdateAll(selector, update)
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
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
	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v,Update:%+v", collection, id, update)
	sp := CreateMongoSpan(ctx, FuncUpdateId, logMsg)
	defer sp.Finish()
	dberr := col.UpdateId(id, update)
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
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
	logMsg := fmt.Sprintf("Collection:%s,Selector:%+v,Upsert:%+v", collection, selector, update)
	sp := CreateMongoSpan(ctx, FuncUpsert, logMsg)
	defer sp.Finish()
	info, dberr := col.Upsert(selector, update)
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	} else if info.UpsertedId != nil {
		accessLog(ctx, s.Addr(), METHOD_CREATE_DOC, fmt.Sprintf("Collection:%s,Selector:%+v,Stuff:%+v", collection, selector, update), start)
	} else {
		accessLog(ctx, s.Addr(), METHOD_UPDATE_DOC, fmt.Sprintf("Collection:%s,Selector:%+v,Stuff:%+v", collection, selector, update), start)
	}
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
	sp := CreateMongoSpan(ctx, FuncBulkInsert, fmt.Sprintf("Collection:%s", collection))
	defer sp.Finish()
	_, dberr := bulk.Run()
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
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
	sp := CreateMongoSpan(ctx, FuncBulkUpsert, fmt.Sprintf("Collection:%s", collection))
	defer sp.Finish()
	result, dberr := bulk.Run()
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	accessLog(ctx, s.Addr(), METHOD_BULK_UPSERT, collection, start)
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
	sp := CreateMongoSpan(ctx, FuncBulkDelete, fmt.Sprintf("Collection:%s", collection))
	defer sp.Finish()
	result, dberr := bulk.Run()
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	accessLog(ctx, s.Addr(), METHOD_BULK_DELETE, collection, start)
	return
}

func (p *Pool) GetBulk(ctx goctx.Context, dbName, collection string) (*Bulk, gopkg.CodeError) {
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
	start := time.Now()
	result, dberr := b.bulk.Run()
	err = b.pool.checkDatabaseError(dberr, b.ctx, b.session)
	accessLog(b.ctx, b.session.Addr(), METHOD_CREATE_BULK, b.collection, start)
	return
}

func (p *Pool) QueryCount(ctx goctx.Context, dbName, collection string, selector interface{}) (n int, err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	start := time.Now()
	query := getMongoCollection(s, dbName, collection).Find(selector)
	logMsg := fmt.Sprintf("Collection:%s,Query:%+v",
		collection,
		selector,
	)
	sp := CreateMongoSpan(ctx, FuncQueryCount, logMsg)
	defer sp.Finish()
	n, dberr := query.Count()
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	accessLog(ctx, s.Addr(), METHOD_READ_DOC, logMsg, start)
	return
}

func (p *Pool) QueryAll(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, skip, limit int, sort ...string) (err gopkg.CodeError) {
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

	logMsg := fmt.Sprintf("Collection:%s,Query:%+v,Projection::%+v,Sort:%+v,Skip:%d,Limit:%d",
		collection,
		selector,
		fields,
		sort,
		skip,
		limit,
	)
	sp := CreateMongoSpan(ctx, FuncQueryAll, logMsg)
	defer sp.Finish()
	dberr := query.All(result)
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
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
	logMsg := fmt.Sprintf("Collection:%s,Filter:%+v,Projection::%+v,Sort:%+v,Skip:%d,Limit:%d",
		collection,
		selector,
		fields,
		sort,
		skip,
		limit,
	)
	sp := CreateMongoSpan(ctx, FuncQueryOne, logMsg)
	defer sp.Finish()
	dberr := query.One(result)
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
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
	logMsg := fmt.Sprintf("Collection:%s,Query:%+v,Projection::%+v,Sort:%+v,Skip:%d,Limit:%d,Change:%+v",
		collection,
		selector,
		fields,
		sort,
		skip,
		limit,
		change,
	)
	sp := CreateMongoSpan(ctx, FuncFindAndModify, logMsg)
	defer sp.Finish()
	var dberr error
	info, dberr = query.Apply(change, result)

	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
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
	logMsg := fmt.Sprintf("Collection:%s,Query:%+v,Projection::%+v,Sort:%+v,Skip:%d,Limit:%d,Change:%+v",
		collection,
		selector,
		fields,
		sort,
		skip,
		limit,
		change,
	)
	sp := CreateMongoSpan(ctx, FuncFindAndRemove, logMsg)
	defer sp.Finish()
	var dberr error
	info, dberr = query.Apply(change, result)

	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	accessLog(ctx, s.Addr(), METHOD_READ_DOC, logMsg, start)
	return
}

func (p *Pool) Indexes(ctx goctx.Context, dbName, collection string) (result []mgo.Index, err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}

	sp := CreateMongoSpan(ctx, FuncIndexes, fmt.Sprintf("Collection:%s,Query Indexes", collection))
	defer sp.Finish()
	result, dberr := getMongoCollection(s, dbName, collection).Indexes()
	err = p.checkDatabaseError(dberr, ctx, s)
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
	sp := CreateMongoSpan(ctx, FuncCreateIndex, fmt.Sprintf("Collection:%s,Create Index:%+v", collection, index))
	defer sp.Finish()
	dberr := getMongoCollection(s, dbName, collection).EnsureIndex(index)
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
		errLog(ctx, s.Addr(), collection+" add index:"+strings.Join(index.Key, ",")+" err:"+err.Error())
	}
	return
}

func (p *Pool) EnsureIndex(ctx goctx.Context, dbName, collection string, index mgo.Index) (err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}
	sp := CreateMongoSpan(ctx, FuncEnsureIndex, fmt.Sprintf("Collection:%s,EnsureIndex:%+v", collection, index))
	defer sp.Finish()
	dberr := getMongoCollection(s, dbName, collection).EnsureIndex(index)
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
		errLog(ctx, s.Addr(), collection+" add index:"+strings.Join(index.Key, ",")+" err:"+err.Error())
	}
	return
}

func (p *Pool) DropIndex(ctx goctx.Context, dbName, collection string, keys []string) (err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}
	sp := CreateMongoSpan(ctx, FuncDropIndex, fmt.Sprintf("Collection:%s,DropIndex Key:%+v", collection, keys))
	defer sp.Finish()
	dberr := getMongoCollection(s, dbName, collection).DropIndex(keys...)
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	return
}

func (p *Pool) DropIndexName(ctx goctx.Context, dbName, collection, name string) (err gopkg.CodeError) {
	s, err := p.get(ctx)
	if err != nil {
		return
	}
	sp := CreateMongoSpan(ctx, FuncDropIndexName, fmt.Sprintf("Collection:%s,DropIndexName:%s", collection, name))
	defer sp.Finish()
	dberr := getMongoCollection(s, dbName, collection).DropIndexName(name)
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	return
}

func (p *Pool) CreateCollection(ctx goctx.Context, dbName, collection string, info *mgo.CollectionInfo) (err gopkg.CodeError) {
	start := time.Now()
	s, err := p.get(ctx)
	if err != nil {
		return
	}
	sp := CreateMongoSpan(ctx, FuncCreateCollection, fmt.Sprintf("Collection:%s", collection))
	defer sp.Finish()
	dberr := getMongoCollection(s, dbName, collection).Create(info)
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	accessLog(ctx, s.Addr(), METHOD_CREATE_COL, collection, start)
	return
}

func (p *Pool) DropCollection(ctx goctx.Context, dbName, collection string) (err gopkg.CodeError) {
	start := time.Now()
	s, err := p.get(ctx)
	if err != nil {
		return
	}
	sp := CreateMongoSpan(ctx, FuncDropCollection, fmt.Sprintf("DropCollection:%s", collection))
	defer sp.Finish()
	dberr := getMongoCollection(s, dbName, collection).DropCollection()
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
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
	logMsg := fmt.Sprintf("Rename Collection from:%s to:%s", from, to)
	sp := CreateMongoSpan(ctx, FuncRenameCollection, logMsg)
	defer sp.Finish()
	dberr := s.Session().Run(bson.D{{Name: "renameCollection", Value: from}, {Name: "to", Value: to}}, result)
	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	accessLog(ctx, s.Addr(), METHOD_RENAME_COL, logMsg, start)
	return
}

func (p *Pool) Pipe(ctx goctx.Context, dbName, collection string, pipeline, result interface{}) (err gopkg.CodeError) {
	start := time.Now()
	s, err := p.get(ctx)
	if err != nil {
		return
	}
	logMsg := fmt.Sprintf("Collection:%s,Pipeline:%+v",
		collection,
		pipeline,
	)
	sp := CreateMongoSpan(ctx, FuncPipe, logMsg)
	defer sp.Finish()
	pipe := getMongoCollection(s, dbName, collection).Pipe(pipeline)
	dberr := pipe.All(result)

	err = p.checkDatabaseError(dberr, ctx, s)
	if err != nil {
		sp.SetTag(TraceTagError, err.FullError())
	}
	accessLog(ctx, s.Addr(), METHOD_READ_DOC, logMsg, start)
	return
}
