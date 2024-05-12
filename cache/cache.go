package cache

import (
	"context"
	"crypto"
	"crypto/sha256"
	"encoding/hex"
	"io"
)

// EntryKind describes the kind of cache entry
type EntryKind int

const (
	// AC stands for Action Cache.
	AC EntryKind = iota

	// CAS stands for Content Addressable Storage.
	CAS

	// RAW cache items are not validated. Not exposed externally, only
	// used for HTTP when running with the --disable_http_ac_validation
	// commandline flag.
	RAW
)

func (e EntryKind) String() string {
	if e == AC {
		return "ac"
	}
	if e == CAS {
		return "cas"
	}
	return "raw"
}

func (e EntryKind) DirName() string {
	if e == AC {
		return "ac.v2"
	}
	if e == CAS {
		return "cas.v2"
	}
	return "raw.v2"
}

// Logger is designed to be satisfied by log.Logger.
type Logger interface {
	Printf(format string, v ...interface{})
}

// Error is used by Cache implementations to return a structured error.
type Error struct {
	// Corresponds to a http.Status* code
	Code int
	// A human-readable string describing the error
	Text string
}

func (e *Error) Error() string {
	return e.Text
}

// Proxy is the interface that (optional) proxy backends must implement.
// Implementations are expected to be safe for concurrent use.
type Proxy interface {

	// Put makes a reasonable effort to asynchronously upload the cache
	// item identified by `hash` with logical size `logicalSize` and
	// `sizeOnDisk` bytes on disk, whose data is readable from `rc` to
	// the proxy backend. The data available in `rc` is in the same
	// format as used by the disk.Cache instance.
	//
	// This is allowed to fail silently (for example when under heavy load).
	Put(ctx context.Context, kind EntryKind, hash string, logicalSize int64, sizeOnDisk int64, rc io.ReadCloser)

	// Get returns an io.ReadCloser from which the cache item identified by
	// `hash` can be read, its logical size, and an error if something went
	// wrong. The data available from `rc` is in the same format as used by
	// the disk.Cache instance.
	Get(ctx context.Context, kind EntryKind, hash string, size int64) (io.ReadCloser, int64, error)

	// Contains returns whether or not the cache item exists on the
	// remote end, and the size if it exists (and -1 if the size is
	// unknown).
	Contains(ctx context.Context, kind EntryKind, hash string, size int64) (bool, int64)
}

// TransformActionCacheKey takes an ActionCache key and an instance name
// and returns a new ActionCache key to use instead. If the instance name
// is empty, then the original key is returned unchanged.
func TransformActionCacheKey(key, instance string, logger Logger) string {
	if instance == "" {
		return key
	}

	h := sha256.New()
	h.Write([]byte(key))
	h.Write([]byte(instance))
	b := h.Sum(nil)
	newKey := hex.EncodeToString(b[:])

	logger.Printf("REMAP AC HASH %s : %s => %s", key, instance, newKey)

	return newKey
}

func LookupKey(kind EntryKind, hash string) string {
	return kind.String() + "/" + hash
}

var EmptyHashes = map[crypto.Hash]string{
	crypto.MD5:    "d41d8cd98f00b204e9800998ecf8427e",
	crypto.SHA1:   "da39a3ee5e6b4b0d3255bfef95601890afd80709",
	crypto.SHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	crypto.SHA512: "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e",
}

func IsHashTypeSupported(hashType string) bool {
	switch hashType {
	case "md5", "sha1", "sha256", "sha512":
		return true
	}
	return false
}

func GetHashType(hash string) crypto.Hash {
	if len(hash) == 32 {
		return crypto.MD5
	} else if len(hash) == 40 {
		return crypto.SHA1
	} else if len(hash) == 64 {
		return crypto.SHA256
	} else if len(hash) == 128 {
		return crypto.SHA512
	}
	return 0
}

func IsEmptyHash(hashType crypto.Hash, hash string) bool {
	return hash == EmptyHashes[hashType]
}
