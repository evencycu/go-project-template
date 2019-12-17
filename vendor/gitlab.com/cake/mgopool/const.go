package mgopool

// Error Code
const (
	UnknownError       = 1030000
	APIConnectDatabase = 1030001
	APIFullResource    = 1030013
	MongoPoolClosed    = 1030014
	MongoDataNodeOp    = 1030015
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
)

// Mongo Error Message
const (
	MongoMsgEOF                = "EOF"
	MongoMsgKernelEOF          = "End of file"
	MongoMsgClose              = "Closed explicitly"
	MongoMsgwiredTigerIndex    = "WiredTigerIndex"
	MongoMsgNotFound           = "not found"
	MongoMsgCollectionConflict = "already exists"
	MongoMsgNsNotFound         = "ns not found"
	MongoMsgNamespaceNotFound  = "source namespace does not exist"
	MongoMsgCollectionNotFound = "no collection"
	MongoMsgNonEmpty           = "$and/$or/$nor must be a nonempty array"
	MongoMsgArray              = "needs an array"
	MongoMsgEachArray          = "The argument to $each"
	MongoMsgPullAllArray       = "$pullAll requires an array argument"
	MongoMsgEmptySet           = "'$set' is empty"
	MongoMsgEmptyUnset         = "'$unset' is empty"
	MongoMsgBadModifier        = "Modifiers operate on fields"
	MongoMsgEmptyInc           = "'$inc' is empty"
	MongoMsgEmptyRename        = "'$rename' is empty"
	MongoMsgIncrement          = "Cannot increment with non-numeric argument"
	MongoMsgE11000             = "E11000"
	MongoMsgUnknown            = "Unknown"
	MongoMsgBulk               = "multiple errors in bulk operation"
	MongoMsgTimeout            = "i/o timeout "
	MongoMsgWriteUnavailable   = "write results unavailable"
	MongoMsgWriteTCP           = "read tcp"
	MongoMsgReadTCP            = "write tcp"
	MongoMsgNoReachableServers = "no reachable servers"
	MongoMsgNoHost             = "could not find host matching read preference"
	MongoMsgNoHost2            = "None of the hosts"
	MongoMsgRegexString        = "$regex has to be a string"
	MongoMsgDotField           = "The dotted field"
	MongoMsgNotMaster          = "not master"
)
