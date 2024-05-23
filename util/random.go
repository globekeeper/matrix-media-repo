package util

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"io"
	"strings"
)

type RandStringLength int

const encode = "abcdefghijklmnopqrstuvwxyz234567"

func GenerateRandomString(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	hasher := sha1.New()
	hasher.Write(b)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func RandStringN(n RandStringLength) (string, error) {
	builder := strings.Builder{}
	enc := base32.NewEncoder(base32.NewEncoding(encode).WithPadding(-1), &builder)
	written, err := io.CopyN(enc, rand.Reader, int64(n))
	if written != int64(n) {
		return "", errors.New("foo")
	}
	return builder.String(), err
}
