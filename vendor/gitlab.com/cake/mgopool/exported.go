package mgopool

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"gitlab.com/cake/goctx"
	"gitlab.com/cake/gopkg"
)

var std *Pool

// Initialize init mongo instance
func Initialize(dbi *DBInfo) (err error) {
	std, err = NewSessionPool(dbi)
	return
}

func Close() {
	std.Close()
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

func Mode() mgo.Mode {
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
func RemoveAll(ctx goctx.Context, dbName, collection string, selector interface{}) (info *mgo.ChangeInfo, err gopkg.CodeError) {
	return std.RemoveAll(ctx, dbName, collection, selector)
}
func Update(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (err gopkg.CodeError) {
	return std.Update(ctx, dbName, collection, selector, update)
}
func UpdateAll(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (info *mgo.ChangeInfo, err gopkg.CodeError) {
	return std.UpdateAll(ctx, dbName, collection, selector, update)
}

func UpdateId(ctx goctx.Context, dbName, collection string, id interface{}, update interface{}) (err gopkg.CodeError) {
	return std.UpdateId(ctx, dbName, collection, id, update)
}

func Upsert(ctx goctx.Context, dbName, collection string, selector interface{}, update interface{}) (info *mgo.ChangeInfo, err gopkg.CodeError) {
	return std.Upsert(ctx, dbName, collection, selector, update)
}

func BulkInsert(ctx goctx.Context, dbName, collection string, documents []bson.M) (err gopkg.CodeError) {
	return std.BulkInsert(ctx, dbName, collection, documents)
}

func BulkUpsert(ctx goctx.Context, dbName, collection string, selectors, documents []bson.M) (result *mgo.BulkResult, err gopkg.CodeError) {
	return std.BulkUpsert(ctx, dbName, collection, selectors, documents)
}

func BulkDelete(ctx goctx.Context, dbName, collection string, documents []bson.M) (result *mgo.BulkResult, err gopkg.CodeError) {
	return std.BulkDelete(ctx, dbName, collection, documents)
}

func GetBulk(ctx goctx.Context, dbName, collection string) (*Bulk, gopkg.CodeError) {
	return std.GetBulk(ctx, dbName, collection)
}

func QueryCount(ctx goctx.Context, dbName, collection string, selector interface{}) (n int, err gopkg.CodeError) {
	return std.QueryCount(ctx, dbName, collection, selector)
}

func QueryAll(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, skip, limit int, sort ...string) (err gopkg.CodeError) {
	return std.QueryAll(ctx, dbName, collection, result, selector, fields, skip, limit, sort...)
}

func QueryOne(ctx goctx.Context, dbName, collection string, result, selector, fields interface{}, skip, limit int, sort ...string) (err gopkg.CodeError) {
	return std.QueryOne(ctx, dbName, collection, result, selector, fields, skip, limit, sort...)
}

// FindAndModify can only update one doc
func FindAndModify(ctx goctx.Context, dbName, collection string, result, selector, update, fields interface{},
	skip, limit int, upsert, returnNew bool, sort ...string) (info *mgo.ChangeInfo, err gopkg.CodeError) {
	return std.FindAndModify(ctx, dbName, collection, result, selector, update, fields, skip, limit, upsert, returnNew, sort...)
}

func FindAndRemove(ctx goctx.Context, dbName, collection string, result, selector, fields interface{},
	skip, limit int, sort ...string) (info *mgo.ChangeInfo, err gopkg.CodeError) {
	return std.FindAndRemove(ctx, dbName, collection, result, selector, fields, skip, limit, sort...)

}

func Indexes(ctx goctx.Context, dbName, collection string) (result []mgo.Index, err gopkg.CodeError) {
	return std.Indexes(ctx, dbName, collection)
}

func CreateIndex(ctx goctx.Context, dbName, collection string, key []string, sparse, unique bool, name string) (err gopkg.CodeError) {
	return std.CreateIndex(ctx, dbName, collection, key, sparse, unique, name)
}

func EnsureIndex(ctx goctx.Context, dbName, collection string, index mgo.Index) (err gopkg.CodeError) {
	return std.EnsureIndex(ctx, dbName, collection, index)
}

func DropIndex(ctx goctx.Context, dbName, collection string, keys []string) (err gopkg.CodeError) {
	return std.DropIndex(ctx, dbName, collection, keys)
}

func DropIndexName(ctx goctx.Context, dbName, collection, name string) (err gopkg.CodeError) {
	return std.DropIndexName(ctx, dbName, collection, name)
}

func CreateCollection(ctx goctx.Context, dbName, collection string, info *mgo.CollectionInfo) (err gopkg.CodeError) {
	return std.CreateCollection(ctx, dbName, collection, info)
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
