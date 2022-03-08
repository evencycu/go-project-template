package mgopool

import (
	"gitlab.com/cake/goctx"
	"gitlab.com/cake/gopkg"
	"gitlab.com/cake/mgopool/v3/compat"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	std MongoPool
)

// Initialize init mongo instance
func Initialize(dbi *DBInfo) (err error) {
	std, err = NewSessionPool(dbi)
	return
}

func SetExportedPool(p MongoPool) {
	std = p
}

func Close() {
	std.Close()
}

func IsNil() bool {
	return std == nil
}

func IsAvailable() bool {
	return std.IsAvailable()
}

func Len() int {
	return std.Len()
}

func Cap() int {
	return std.Cap()
}

func Mode() readpref.Mode {
	return std.Mode()
}

func LiveServers() []string {
	return std.LiveServers()
}

func ShowConfig() map[string]interface{} {
	return std.ShowConfig()
}

func Ping(ctx goctx.Context) (err gopkg.CodeError) {
	return std.Ping(ctx)
}

func PingPref(ctx goctx.Context, pref *readpref.ReadPref) (err gopkg.CodeError) {
	return std.PingPref(ctx, pref)
}

// GetCollectionNames return all collection names in the db.
func GetCollectionNames(ctx goctx.Context, dbName string) (names []string, err gopkg.CodeError) {
	return std.GetCollectionNames(ctx, dbName)
}

func CollectionCount(ctx goctx.Context, dbName, collection string) (n int, err gopkg.CodeError) {
	return std.CollectionCount(ctx, dbName, collection)
}

func Run(ctx goctx.Context, cmd interface{}, result interface{}) gopkg.CodeError {
	return std.Run(ctx, cmd, result)
}

func DBRun(ctx goctx.Context, dbName string, cmd, result interface{}) gopkg.CodeError {
	return std.DBRun(ctx, dbName, cmd, result)
}

func Insert(ctx goctx.Context, dbName, collection string, doc interface{}) gopkg.CodeError {
	return std.Insert(ctx, dbName, collection, doc)
}

func Remove(ctx goctx.Context, dbName, collection string, selector interface{}) (err gopkg.CodeError) {
	return std.Remove(ctx, dbName, collection, selector)
}
func RemoveAll(ctx goctx.Context, dbName, collection string, selector interface{}) (removedCount int, err gopkg.CodeError) {
	return std.RemoveAll(ctx, dbName, collection, selector)
}
func Update(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (err gopkg.CodeError) {
	return std.Update(ctx, dbName, collection, selector, update)
}
func UpdateAll(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (result *mongo.UpdateResult, err gopkg.CodeError) {
	return std.UpdateAll(ctx, dbName, collection, selector, update)
}

func UpdateId(ctx goctx.Context, dbName, collection string, id interface{}, update interface{}) (err gopkg.CodeError) {
	return std.UpdateId(ctx, dbName, collection, id, update)
}

func Upsert(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (result *mongo.UpdateResult, err gopkg.CodeError) {
	return std.Upsert(ctx, dbName, collection, selector, update)
}

func BulkInsert(ctx goctx.Context, dbName, collection string, documents []bson.M) (result *BulkResult, err gopkg.CodeError) {
	return std.BulkInsert(ctx, dbName, collection, documents)
}

func BulkUpsert(ctx goctx.Context, dbName, collection string, selectors, documents []bson.M) (result *BulkResult, err gopkg.CodeError) {
	return std.BulkUpsert(ctx, dbName, collection, selectors, documents)
}

func BulkUpdate(ctx goctx.Context, dbName, collection string, selectors, documents []bson.M) (result *BulkResult, err gopkg.CodeError) {
	return std.BulkUpdate(ctx, dbName, collection, selectors, documents)
}

func BulkInsertInterfaces(ctx goctx.Context, dbName, collection string, documents []interface{}) (result *BulkResult, err gopkg.CodeError) {
	return std.BulkInsertInterfaces(ctx, dbName, collection, documents)
}

func BulkUpsertInterfaces(ctx goctx.Context, dbName, collection string, selectors []bson.M, documents []interface{}) (result *BulkResult, err gopkg.CodeError) {
	return std.BulkUpsertInterfaces(ctx, dbName, collection, selectors, documents)
}

func BulkUpdateInterfaces(ctx goctx.Context, dbName, collection string, selectors []bson.M, documents []interface{}) (result *BulkResult, err gopkg.CodeError) {
	return std.BulkUpdateInterfaces(ctx, dbName, collection, selectors, documents)
}

func BulkDelete(ctx goctx.Context, dbName, collection string, documents []bson.M) (result *BulkResult, err gopkg.CodeError) {
	return std.BulkDelete(ctx, dbName, collection, documents)
}

func QueryCount(ctx goctx.Context, dbName, collection string, selector interface{}) (n int, err gopkg.CodeError) {
	return std.QueryCountWithOptions(ctx, dbName, collection, selector, 0, 0)
}

func UpdateWithArrayFilters(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}, arrayFilters interface{}, multi bool) (result *mongo.UpdateResult, err gopkg.CodeError) {
	return std.UpdateWithArrayFilters(ctx, dbName, collection, selector, update, arrayFilters, multi)
}

func QueryCountWithOptions(ctx goctx.Context, dbName, collection string, selector interface{}, skip, limit int) (n int, err gopkg.CodeError) {
	return std.QueryCountWithOptions(ctx, dbName, collection, selector, skip, limit)
}

func QueryAll(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, skip, limit int, sort ...string) (err gopkg.CodeError) {
	return std.QueryAll(ctx, dbName, collection, result, selector, fields, skip, limit, sort...)
}

func QueryAllWithCollation(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, collation *options.Collation, skip, limit int, sort ...string) (err gopkg.CodeError) {
	return std.QueryAllWithCollation(ctx, dbName, collection, result, selector, fields, collation, skip, limit, sort...)
}

func QueryOne(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, skip int, sort ...string) (err gopkg.CodeError) {
	return std.QueryOne(ctx, dbName, collection, result, selector, fields, skip, sort...)
}

func FindAndReplace(ctx goctx.Context, dbName, collection string, result, selector, replacement, fields interface{}, upsert, returnNew bool, sort ...string) (err gopkg.CodeError) {
	return std.FindAndReplace(ctx, dbName, collection, result, selector, replacement, fields, upsert, returnNew, sort...)
}

func FindAndModify(ctx goctx.Context, dbName, collection string, result, selector, update, fields interface{},
	upsert, returnNew bool, sort ...string) (err gopkg.CodeError) {
	return std.FindAndModify(ctx, dbName, collection, result, selector, update, fields, upsert, returnNew, sort...)
}

func FindAndRemove(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, sort ...string) (err gopkg.CodeError) {
	return std.FindAndRemove(ctx, dbName, collection, result, selector, fields, sort...)
}

func FindAndModifyWithArrayFilters(ctx goctx.Context, dbName, collection string, result, selector, update, fields interface{},
	upsert, returnNew bool, arrayFilters interface{}, sort ...string) (err gopkg.CodeError) {
	return std.FindAndModifyWithArrayFilters(ctx, dbName, collection, result, selector, update, fields, upsert, returnNew, arrayFilters, sort...)
}

func Indexes(ctx goctx.Context, dbName, collection string) (result []map[string]interface{}, err gopkg.CodeError) {
	return std.Indexes(ctx, dbName, collection)
}

func CreateIndex(ctx goctx.Context, dbName, collection string, key []string, sparse, unique bool, name string) (err gopkg.CodeError) {
	return std.CreateIndex(ctx, dbName, collection, key, sparse, unique, name)
}

func CreateTTLIndex(ctx goctx.Context, dbName, collection string, key string, ttlSec int) (err gopkg.CodeError) {
	return std.CreateTTLIndex(ctx, dbName, collection, key, ttlSec)
}

func EnsureIndex(ctx goctx.Context, dbName, collection string, index mongo.IndexModel) (err gopkg.CodeError) {
	return std.EnsureIndex(ctx, dbName, collection, index)
}

func EnsureIndexCompat(ctx goctx.Context, dbName, collection string, indexCompat compat.Index) (err gopkg.CodeError) {
	return std.EnsureIndexCompat(ctx, dbName, collection, indexCompat)
}

func DropIndex(ctx goctx.Context, dbName, collection string, keys []string) (err gopkg.CodeError) {
	return std.DropIndex(ctx, dbName, collection, keys)
}

func DropIndexName(ctx goctx.Context, dbName, collection, name string) (err gopkg.CodeError) {
	return std.DropIndexName(ctx, dbName, collection, name)
}

func CreateCollection(ctx goctx.Context, dbName, collection string, option *options.CreateCollectionOptions) (err gopkg.CodeError) {
	return std.CreateCollection(ctx, dbName, collection, option)
}

func DropCollection(ctx goctx.Context, dbName, collection string) (err gopkg.CodeError) {
	return std.DropCollection(ctx, dbName, collection)
}

func RenameCollection(ctx goctx.Context, dbName, oldName, newName string) (err gopkg.CodeError) {
	return std.RenameCollection(ctx, dbName, oldName, newName)
}

func Pipe(ctx goctx.Context, dbName, collection string, pipeline, result interface{}) (err gopkg.CodeError) {
	return std.Pipe(ctx, dbName, collection, pipeline, result)
}

func Distinct(ctx goctx.Context, dbName, collection string, selector bson.M, field string, result interface{}) (err gopkg.CodeError) {
	return std.Distinct(ctx, dbName, collection, selector, field, result)
}

func GetCursor(ctx goctx.Context, dbName, collection string, selector interface{}, option ...*options.FindOptions) (MongoCursor, gopkg.CodeError) {
	return std.GetCursor(ctx, dbName, collection, selector, option...)
}

func GetBulk(ctx goctx.Context, dbName, collection string, batchSize ...int) *Bulk {
	return std.GetBulk(ctx, dbName, collection, batchSize...)
}
