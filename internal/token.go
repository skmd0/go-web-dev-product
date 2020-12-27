package internal

import (
	"crypto/rand"
	"encoding/base64"
)

// RememberTokenBytes represents the length of our byte slice used to generate remember token, which has implication on
// the security of it.
const RememberTokenBytes = 32

func RememberToken() (string, error) {
	return GenerateRememberToken(RememberTokenBytes)
}

// GenerateRememberToken returns a Base64 URL encoded token
func GenerateRememberToken(numBytes int) (string, error) {
	bs, err := GenerateByteSlice(numBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bs), nil
}

// GenerateByteSlice outputs a byte slice with random bytes as values with an length of n specified by parameter
func GenerateByteSlice(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// NBytes returns the length of the Base64 encoded byte slice
// It is used to validate the length of the remember hash.
func NBytes(base64String string) (int, error) {
	b, err := base64.URLEncoding.DecodeString(base64String)
	if err != nil {
		return -1, err
	}
	return len(b), nil
}
