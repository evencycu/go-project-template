package mgopool

const (
	dollar = "$"
)

// Error Code
const (
	UnknownError       = 1030000
	APIConnectDatabase = 1030001
	MongoPoolClosed    = 1030014
	// MongoDataNodeOp    = 1030015

	// Data Logic Error
	NotFound               = 1030301
	CollectionNotFound     = 1030302
	DocumentConflict       = 1030305
	CollectionConflict     = 1030306
	QueryInputArray        = 1030309
	UpdateInputArray       = 1030310
	IncrementNumeric       = 1030311
	RegexString            = 1030312
	DotField               = 1030313
	Timeout                = 1030314
	StringIndexTooLong     = 1030315
	BadUpdateOperatorUsage = 1030316
	TypeNotSupported       = 1030317
	PoolTimeout            = 1030318
	BadInputDoc            = 1030319
	BadUpdatePairs         = 1030320
	NotPtrInput            = 1030321
)

// Mongo Error Message
const (
	MongoMsgEOF                  = "EOF"
	MongoMsgKernelEOF            = "End of file"
	MongoMsgClose                = "Closed explicitly"
	MongoMsgwiredTigerIndex      = "WiredTigerIndex"
	MongoMsgNotFound             = "not found"
	MongoMsgCollectionConflict   = "already exists"
	MongoMsgNsNotFound           = "ns not found"
	MongoMsgCursorNotFound       = "cursor not found"
	MongoMsgNamespaceNotFound    = "source namespace does not exist"
	MongoMsgCollectionNotFound   = "no collection"
	MongoMsgDocumentsNotFound    = "mongo: no documents in result"
	MongoMsgNonEmpty             = "$and/$or/$nor must be a nonempty array"
	MongoMsgArray                = "to be an array"
	MongoMsgInArray              = "$in needs an array"
	MongoMsgEachArray            = "The argument to $each"
	MongoMsgPullAllArray         = "$pullAll requires an array argument"
	MongoMsgEmptySet             = "'$set' is empty"
	MongoMsgEmptyUnset           = "'$unset' is empty"
	MongoMsgBadModifier          = "Modifiers operate on fields"
	MongoMsgEmptyInc             = "'$inc' is empty"
	MongoMsgEmptyRename          = "'$rename' is empty"
	MongoMsgIncrement            = "Cannot increment with non-numeric argument"
	MongoMsgE11000               = "E11000"
	MongoMsgDuplicateKey         = "(DuplicateKey)"
	MongoMsgUnknown              = "Unknown"
	MongoMsgBulk                 = "multiple errors in bulk operation"
	MongoMsgTimeout              = "i/o timeout"
	MongoMsgWriteUnavailable     = "write results unavailable"
	MongoMsgWriteTCP             = "read tcp"
	MongoMsgReadTCP              = "write tcp"
	MongoMsgContextTimeout       = "context deadline exceeded"
	MongoMsgNoReachableServers   = "no reachable servers"
	MongoMsgNoHost               = "could not find host matching read preference"
	MongoMsgNoHost2              = "None of the hosts"
	MongoMsgRegexString          = "$regex has to be a string"
	MongoMsgDotField             = "The dotted field"
	MongoMsgNotMaster            = "not master"
	MongoMsgEmptyDollarKey       = "update document must contain key beginning with '$'"
	MongoMsgGetPoolTimeout       = "timed out while checking out a connection from connection pool"
	MongoMsgDocNil               = "document is nil"
	MongoMsgServerSelectionError = "server selection error"
)

// wrapped error msg

const (
	MgopoolBadConnection = "bad connection/cluster unhealthy"
	MgopoolOpTimeout     = "operation timeout"
)

const (
	MongoDriverDuplicateCode = 11000
)

const (
	ObjId = "_id"

	OpWhere     = "$where"
	OpSet       = "$set"
	OpUnset     = "$unset"
	OpSlice     = "$slice"
	OpEqual     = "$eq"
	OpNotEqual  = "$ne"
	OpPush      = "$push"
	OpPull      = "$pull"
	OpIn        = "$in"
	OpNin       = "$nin"
	OpExists    = "$exists"
	OpAnd       = "$and"
	OpOr        = "$or"
	OpNor       = "$nor"
	OpRegex     = "$regex"
	OpInc       = "$inc"
	OpElemMatch = "$elemMatch"
	OpGt        = "$gt"
	OpGte       = "$gte"
	OpLt        = "$lt"
	OpLte       = "$lte"
	OpMax       = "$max"
	OpMin       = "$min"
	OpMul       = "$mul"
	OpOptions   = "$options"
	OpSize      = "$size"
	OpEach      = "$each"
	OpPosition  = "$position"

	OpAddToSet    = "$addToSet"
	OpCurrentDate = "$currentDate"
	OpExpr        = "$expr"
	OpConcat      = "$concat"
	OpMatch       = "$match"
	OpProject     = "$project"
	OpSkip        = "$skip"
	OpLimit       = "$limit"
	OpSort        = "$sort"
	OpAddFields   = "$addFields"
	OpGroup       = "$group"
	OpSum         = "$sum"
	OpCount       = "$count"
	OpOut         = "$out"
	OpFacet       = "$facet"
	OpBucket      = "$bucket"
	OpBucketAuto  = "$bucketAuto"
	OpCollStats   = "$collStats"
	OpIndexStats  = "$indexStats"
	OpUnwind      = "$unwind"
	OpFirst       = "$first"
	OpLast        = "$last"
	OpSwitch      = "$switch"
	OpCond        = "$cond"
	OpPullAll     = "$pullAll"
	OpPop         = "$pop"

	OpSetOnInsert = "$setOnInsert"
)
