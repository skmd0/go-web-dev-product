package internal

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"hash"
)

// NewHMAC is a constructor for HMAC struct.
func NewHMAC(key string) HMAC {
	h := hmac.New(sha256.New, []byte(key))
	return HMAC{hmac: h}
}

// HMAC is a wrapper over hash.Hash field
type HMAC struct {
	hmac hash.Hash
}

// Hash outputs a hashed token.
// It is used for passwords and cookie remember tokens.
func (h HMAC) Hash(input string) string {
	h.hmac.Reset()
	h.hmac.Write([]byte(input))
	b := h.hmac.Sum(nil)
	return base64.URLEncoding.EncodeToString(b)
}
