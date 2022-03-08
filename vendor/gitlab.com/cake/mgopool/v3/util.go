package mgopool

import (
	"reflect"
	"strings"

	"gitlab.com/cake/goctx"
	"gitlab.com/cake/gopkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func ParseIndexKey(key []string) (bson.D, string) {
	keys := bson.D{}
	indexName := ""
	for _, k := range key {
		if k == "" {
			continue
		}
		if indexName != "" {
			indexName += "_"
		}
		switch k[0] {
		case '-':
			k = k[1:]
			keys = append(keys, bson.E{Key: k, Value: -1})
			indexName += k + "_-1"
		case '+':
			k = k[1:]
			fallthrough
		default:
			keys = append(keys, bson.E{Key: k, Value: 1})
			indexName += k + "_1"
		}
	}
	return keys, indexName
}

func ParseSortField(sort ...string) bson.D {
	sortField := bson.D{}
	for _, s := range sort {
		if s == "" {
			continue
		}
		switch s[0] {
		case '-':
			s = s[1:]
			sortField = append(sortField, bson.E{Key: s, Value: -1})
		case '+':
			s = s[1:]
			fallthrough
		default:
			sortField = append(sortField, bson.E{Key: s, Value: 1})
		}
	}
	return sortField
}

// EnsureUpsertOp: go mongo driver will check if $set/$inc is in upsert/update operation.
// To compatible with mgo upsert, this function ensures $set in upsert object
func EnsureUpsertOp(val interface{}) bson.M {
	if bsonObj, ok := val.(bson.M); ok {
		if _, exists := bsonObj[OpSet]; exists {
			return bsonObj
		}
		if _, exists := bsonObj[OpSetOnInsert]; exists {
			return bsonObj
		}
	}
	return bson.M{OpSet: val}
}

func EnsureUpdateOp(val interface{}) (opObj bson.M, hasOperator bool) {
	if bsonObj, ok := val.(bson.M); ok {
		for key := range bsonObj {
			if strings.HasPrefix(key, dollar) {
				return bsonObj, true
			}
		}
		return bson.M{OpSet: bsonObj}, false
	}
	return bson.M{OpSet: val}, false
}

func RemoveEmptyOperand(m bson.M) {
	for op, val := range m {
		switch op {
		case OpAnd, OpOr, OpNor:
			v := reflect.ValueOf(val)
			if (v.Kind() == reflect.Slice || v.Kind() == reflect.Array) && v.Len() == 0 {
				delete(m, op)
			}
		default:
			//pass
		}
	}
}

// IsDup returns whether err informs of a duplicate key error because
// a primary key index or a secondary unique index already has an entry
// with the given value.
func IsDup(err error) bool {
	switch e := err.(type) {
	case gopkg.CodeError:
		return e.ErrorCode() == DocumentConflict
	case mongo.WriteException:
		if len(e.WriteErrors) >= 1 {
			return e.WriteErrors[0].Code == 11000 || strings.Contains(e.WriteErrors[0].Message, MongoMsgE11000)
		}
	case mongo.BulkWriteException:
		for _, we := range e.WriteErrors {
			if we.Code == 11000 || strings.Contains(we.Message, MongoMsgE11000) {
				return true
			}
		}
	case mongo.CommandError:
		return e.Name == "DuplicateKey" || strings.Contains(e.Message, MongoMsgE11000)
	}
	return false
}

func IsBadConnectionError(err gopkg.CodeError) bool {
	switch err.ErrorCode() {
	case UnknownError,
		APIConnectDatabase,
		MongoPoolClosed,
		Timeout,
		PoolTimeout:
		return true
	}
	return false
}

func MergeBulkResult(result interface{}, err error) *BulkResult {
	r := &BulkResult{}
	switch result.(type) {
	case *mongo.BulkWriteResult:
		if result.(*mongo.BulkWriteResult) != nil {
			r.BulkWriteResult = *result.(*mongo.BulkWriteResult)
		}
	case *mongo.InsertManyResult:
		if result.(*mongo.InsertManyResult) != nil {
			r.InsertManyResult = *result.(*mongo.InsertManyResult)
		}
	}
	if e, ok := err.(mongo.BulkWriteException); ok {
		r.BulkWriteException = e
	}
	return r
}

func RemoveErrorElements(_ goctx.Context, writeErrors []mongo.BulkWriteError, data interface{}) gopkg.CodeError {
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
		newLen := slice.Len() - len(writeErrors)
		result := reflect.MakeSlice(reflect.SliceOf(slice.Index(0).Type()), newLen, newLen)
		for _, errCase := range writeErrors {
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
