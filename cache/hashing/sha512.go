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
	hasher := &sha512Hasher{}
	register(hasher)
}

var sha512Regex = regexp.MustCompile("^[a-f0-9]{128}$")

type sha512Hasher struct{}

func (d *sha512Hasher) New() hash.Hash {
	return sha512.New()
}

func (d *sha512Hasher) Hash(data []byte) string {
	sum := sha512.Sum512(data)
	return hex.EncodeToString(sum[:])
}

func (d *sha512Hasher) DigestFunction() pb.DigestFunction_Value {
	return pb.DigestFunction_SHA512
}

func (d *sha512Hasher) Dir() string {
	return "sha512"
}

func (d *sha512Hasher) Empty() string {
	return "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"
}

func (d *sha512Hasher) Size() int {
	return sha512.Size
}

func (d *sha512Hasher) Validate(value string) error {
	if d.Size()*2 != len(value) {
		return fmt.Errorf("Invalid sha512 hash length %d: expected %d", len(value), d.Size())
	}
	if !sha512Regex.MatchString(value) {
		return errors.New("Malformed sha512 hash " + value)
	}
	return nil
}

func (d *sha512Hasher) ValidateDigest(hash string, size int64) error {
	if size == int64(0) {
		if hash == d.Empty() {
			return nil
		}
		return fmt.Errorf("Invalid zero-length %s hash", d.DigestFunction())
	}
	return d.Validate(hash)
}
