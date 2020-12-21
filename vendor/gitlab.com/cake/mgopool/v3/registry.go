package mgopool

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/mgocompat"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// package level because I think whole project should use same registry
// not by pool object level
var (
	registry          = mgocompat.Registry
	collectionOptions = &options.CollectionOptions{
		Registry: registry,
	}
	databaseOptions = &options.DatabaseOptions{
		Registry: registry,
	}
)

func GetBsonRegistry() *bsoncodec.Registry {
	return registry
}

func SetBsonRegistry(r *bsoncodec.Registry) {
	registry = r
	collectionOptions.Registry = registry
	databaseOptions.Registry = registry
}

func Marshal(val interface{}) ([]byte, error) {
	return bson.MarshalWithRegistry(registry, val)
}

func Unmarshal(data []byte, val interface{}) error {
	return bson.UnmarshalWithRegistry(registry, data, val)
}
