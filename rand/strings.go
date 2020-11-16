package rand

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

func GenerateByteSlice(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
