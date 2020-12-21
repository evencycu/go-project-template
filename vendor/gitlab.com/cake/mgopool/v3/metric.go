package mgopool

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/cake/gopkg"
)

const (
	prometheusNamespace = "mgopool"
	prometheusSubsystem = "pool"
)

const (
	labelPoolName  = "pool_name"
	labelSuccess   = "success"
	labelType      = "type"
	labelOperation = "operation"
)

const (
	operationLabelSucceed = "true"
	operationLabelFailed  = "false"
)

const (
	operationLabelBulkInsert    = "bulk_insert"
	operationLabelBulkUpdate    = "bulk_update"
	operationLabelBulkReplace   = "bulk_replace"
	operationLabelBulkUpsert    = "bulk_upsert"
	operationLabelBulkUpdateAll = "bulk_update_all"
	operationLabelBulkRemove    = "bulk_remove"
	operationLabelBulkRemoveAll = "bulk_remove_all"
	operationLabelBulkRun       = "bulk_run"
)

const (
	operationLabelCollectionCount          = "collection_count"
	operationLabelRun                      = "run"
	operationLabelDBRun                    = "db_run"
	operationLabelInsert                   = "insert"
	operationLabelDistinct                 = "distinct"
	operationLabelRemove                   = "remove"
	operationLabelRemoveAll                = "remove_all"
	operationLabelReplaceOne               = "replace_one"
	operationLabelUpdate                   = "update"
	operationLabelUpdateAll                = "update_all"
	operationLabelUpdateId                 = "update_id"
	operationLabelUpdateWithArrayFilters   = "update_with_array_filters"
	operationLabelUpsert                   = "upsert"
	operationLabelPoolBulkInsert           = "pool_bulk_insert"
	operationLabelPoolBulkUpsert           = "pool_bulk_upsert"
	operationLabelPoolBulkUpdate           = "pool_bulk_update"
	operationLabelPoolBulkInsertInterfaces = "pool_bulk_insert_interfaces"
	operationLabelPoolBulkUpsertInterfaces = "pool_bulk_upsert_interfaces"
	operationLabelPoolBulkUpdateInterfaces = "pool_bulk_update_interfaces"
	operationLabelPoolBulkDelete           = "pool_bulk_delete"
	operationLabelQueryCountWithOptions    = "query_count_with_options"
	operationLabelQueryAllWithOptions      = "query_all_with_options"
	operationLabelQueryOne                 = "query_one"
	operationLabelFindAndModify            = "find_and_modify"
	operationLabelFindAndReplace           = "find_and_replace"
	operationLabelFindAndRemove            = "find_and_remove"
	operationLabelIndexes                  = "indexes"
	operationLabelCreateIndex              = "create_index"
	operationLabelCreateTTLIndex           = "create_ttl_index"
	operationLabelEnsureIndex              = "ensure_index"
	operationLabelEnsureIndexCompat        = "ensure_index_compat"
	operationLabelDropIndex                = "drop_index"
	operationLabelDropIndexName            = "drop_index_name"
	operationLabelCreateCollection         = "create_collection"
	operationLabelDropCollection           = "drop_collection"
	operationLabelRenameCollection         = "rename_collection"
	operationLabelPipe                     = "pipe"
)

var (
	poolTimeoutCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "get_count",
			Help:      "pool.get timeout count",
		},
		[]string{
			labelPoolName,
		},
	)

	operationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "operation_duration",
			Help:      "how long does it take to operate",
			Buckets: []float64{
				1.0,
				10.0,
				50.0,
				100.0,
				300.0,
				500.0,
				800.0,
				1100.0,
				1500.0,
				2000.0,
				2500.0,
				3000.0,
				4000.0,
				5000.0,
			},
		},
		[]string{
			labelType,
		},
	)

	operationCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "operation_count",
			Help:      "how many time of operations",
		},
		[]string{
			labelSuccess,
			labelType,
		},
	)
)

func updateMetrics(operationLabel string, start time.Time, err *gopkg.CodeError) {
	if *err != nil {
		operationCount.WithLabelValues(operationLabelFailed, operationLabel).Inc()
	} else {
		operationCount.WithLabelValues(operationLabelSucceed, operationLabel).Inc()
		processElapsed := time.Since(start)
		operationDuration.WithLabelValues(operationLabel).Observe(float64(processElapsed.Milliseconds()))
	}
}

func init() {
	prometheus.MustRegister(poolTimeoutCounter)
	prometheus.MustRegister(operationDuration)
	prometheus.MustRegister(operationCount)
}
