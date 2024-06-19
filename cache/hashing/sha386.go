package hashing

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"regexp"

	pb "github.com/buchgr/bazel-remote/v2/genproto/build/bazel/remote/execution/v2"
)

func init() {
	hasher := &sha384Hasher{}
	register(hasher)
}

var sha384Regex = regexp.MustCompile("^[a-f0-9]{96}$")

type sha384Hasher struct{}

func (d *sha384Hasher) New() hash.Hash {
	return sha512.New384()
}

func (d *sha384Hasher) Hash(data []byte) string {
	sum := sha512.Sum384(data)
	return hex.EncodeToString(sum[:])
}

func (d *sha384Hasher) DigestFunction() pb.DigestFunction_Value {
	return pb.DigestFunction_SHA384
}

func (d *sha384Hasher) Dir() string {
	return "sha384"
}

func (d *sha384Hasher) Empty() string {
	return "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"
}

func (d *sha384Hasher) Size() int {
	return sha512.Size384
}

func (d *sha384Hasher) Validate(value string) error {
	if d.Size()*2 != len(value) {
		return fmt.Errorf("Invalid sha384 hash length %d: expected %d", len(value), d.Size())
	}
	if !sha384Regex.MatchString(value) {
		return errors.New("Malformed sha384 hash " + value)
	}
	return nil
}

func (d *sha384Hasher) ValidateDigest(hash string, size int64) error {
	if size == int64(0) {
		if hash == d.Empty() {
			return nil
		}
		return fmt.Errorf("Invalid zero-length %s hash", d.DigestFunction())
	}
	return d.Validate(hash)
}
