package schema

import (
	"crypto/md5"
	"encoding/hex"
)

type Registry interface {
	// Returns the schema for the given schema key.
	Get(key string) (string, error)

	// Returns the the key of the schema on success.
	Set(schema string) (string, error)

	// Closes all resources.
	Close() error
}

func Hash(schema string) string {
	md5sum := md5.Sum([]byte(schema))
	return hex.EncodeToString(md5sum[:])
}
