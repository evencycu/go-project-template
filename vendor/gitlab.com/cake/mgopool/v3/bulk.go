package mgopool

import (
	"sync"
	"time"

	"gitlab.com/cake/goctx"
	"gitlab.com/cake/gopkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Bulk struct {
	ctx        goctx.Context
	dbName     string
	collection string
	mtx        *sync.RWMutex
	// batch would auto bulk run if size > batch
	batchSize int
	pool      *Pool
	models    []mongo.WriteModel
	options   *options.BulkWriteOptions
}

func (b *Bulk) GetQueueSize() int {
	return len(b.models)
}

func (b *Bulk) SetBulkOrderred() {
	t := true
	b.options.Ordered = &t
}

func (b *Bulk) SetBulkUnorderred() {
	t := false
	b.options.Ordered = &t
}

// Insert queues up the provided documents for insertion.
func (b *Bulk) Insert(docs ...interface{}) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelBulkInsert, start, &err)

	for _, d := range docs {
		if d == nil {
			continue
		}
		b.mtx.Lock()
		b.models = append(b.models, &mongo.InsertOneModel{
			Document: d,
		})
		b.mtx.Unlock()
	}

	b.batchRun()
	return
}

func (b *Bulk) Update(pairs ...interface{}) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelBulkUpdate, start, &err)

	if len(pairs)%2 != 0 {
		err = gopkg.NewCodeError(BadUpdatePairs, "update pairs not equal")
		return
	}

	b.mtx.Lock()
	defer b.mtx.Unlock()
	for i := 0; i < len(pairs); i += 2 {
		update := pairs[i+1]
		if update == nil {
			continue
		}
		setOp, hasOp := EnsureUpdateOp(update)
		var m mongo.WriteModel
		if hasOp {
			m = &mongo.UpdateOneModel{
				Filter: pairs[i],
				Update: setOp,
			}
		} else {
			m = &mongo.ReplaceOneModel{
				Filter:      pairs[i],
				Replacement: update,
			}
		}
		b.models = append(b.models, m)
	}

	b.batchRun()
	return
}

func (b *Bulk) Replace(pairs ...interface{}) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelBulkReplace, start, &err)

	if len(pairs)%2 != 0 {
		err = gopkg.NewCodeError(BadUpdatePairs, "replace pairs not equal")
		return
	}

	b.mtx.Lock()
	defer b.mtx.Unlock()
	for i := 0; i < len(pairs); i += 2 {
		update := pairs[i+1]
		if update == nil {
			continue
		}
		b.models = append(b.models, &mongo.ReplaceOneModel{
			Filter:      pairs[i],
			Replacement: update,
		})
	}

	b.batchRun()
	return
}

func (b *Bulk) Upsert(pairs ...interface{}) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelBulkUpsert, start, &err)

	if len(pairs)%2 != 0 {
		err = gopkg.NewCodeError(BadUpdatePairs, "update pairs not equal")
		return
	}

	t := true
	b.mtx.Lock()
	defer b.mtx.Unlock()
	for i := 0; i < len(pairs); i += 2 {
		update := pairs[i+1]
		if update == nil {
			continue
		}
		b.models = append(b.models, &mongo.UpdateOneModel{
			Filter: pairs[i],
			Update: EnsureUpsertOp(update),
			Upsert: &t,
		})
	}

	b.batchRun()
	return
}

// UpdateAll queues up the provided pairs of updating instructions.
// The first element of each pair selects which documents must be
// updated, and the second element defines how to update it.
// Each pair updates all documents matching the selector.
func (b *Bulk) UpdateAll(selector []bson.M, updater []interface{}) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelBulkUpdateAll, start, &err)

	if len(selector) != len(updater) {
		err = gopkg.NewCodeError(BadUpdatePairs, "update pairs not equal")
		return
	}

	b.mtx.Lock()
	defer b.mtx.Unlock()
	for i := 0; i < len(selector); i += 2 {
		if updater[i] == nil {
			continue
		}
		setOp, hasOp := EnsureUpdateOp(updater[i])
		var m mongo.WriteModel
		if hasOp {
			m = &mongo.UpdateManyModel{
				Filter: selector[i],
				Update: setOp,
			}
		} else {
			m = &mongo.ReplaceOneModel{
				Filter:      selector[i],
				Replacement: updater[i],
			}
		}
		b.models = append(b.models, m)
	}

	b.batchRun()
	return
}

func (b *Bulk) Remove(selectors ...interface{}) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelBulkRemove, start, &err)

	b.mtx.Lock()
	defer b.mtx.Unlock()
	for _, d := range selectors {
		if d == nil {
			continue
		}
		b.models = append(b.models, &mongo.DeleteOneModel{
			Filter: d,
		})
	}

	b.batchRun()
	return
}

func (b *Bulk) RemoveAll(selectors ...interface{}) (err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelBulkRemoveAll, start, &err)

	b.mtx.Lock()
	defer b.mtx.Unlock()
	for _, d := range selectors {
		if d == nil {
			continue
		}
		b.models = append(b.models, &mongo.DeleteManyModel{
			Filter: d,
		})
	}

	b.batchRun()
	return
}

func (b *Bulk) batchRun() {
	if b.batchSize > 0 && len(b.models) >= b.batchSize {
		_, _ = b.Run()
	}
}

func (b *Bulk) Run() (res *BulkResult, err gopkg.CodeError) {
	start := time.Now()
	defer updateMetrics(operationLabelBulkRun, start, &err)
	if len(b.models) == 0 {
		res = &BulkResult{}
		return
	}
	col := getMongoCollection(b.pool.client, b.dbName, b.collection)
	result, errDB := col.BulkWrite(b.ctx.NativeContext(), b.models, b.options)
	b.models = []mongo.WriteModel{}
	res = MergeBulkResult(result, errDB)
	err = b.pool.resultHandling(errDB, b.ctx)
	return
}
