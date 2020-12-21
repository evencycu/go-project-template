package compat

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Index struct {
	Key []string

	Unique        bool     // Prevent two documents from having the same index key
	Background    bool     // Build index in background and return immediately
	Sparse        bool     // Only index documents containing the Key fields
	PartialFilter bson.M   // Partial index filter expression

	// If ExpireAfter is defined the server will periodically delete
	// documents with indexed time.Time older than the provided delta.
	ExpireAfter time.Duration

	// Name holds the stored index name. On creation if this field is unset it is
	// computed by EnsureIndex based on the index key.
	Name string

	Min, Max float64
	BucketSize int
	Bits       int

	// Properties for text indexes.
	DefaultLanguage  string
	LanguageOverride string

	// Weights defines the significance of provided fields relative to other
	// fields in a text index. The score for a given word in a document is derived
	// from the weighted sum of the frequency for each of the indexed fields in
	// that document. The default field weight is 1.
	Weights map[string]int

	// Collation defines the collation to use for the index.
	Collation *options.Collation
}
