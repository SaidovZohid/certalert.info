package utils

import (
	"crypto/rand"
	"math/big"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
)

func GenerateRandomPassword(length int) string {
	link := make([]byte, length)
	letterLen := big.NewInt(int64(len(letterBytes)))

	for i := 0; i < length; i++ {
		index, _ := rand.Int(rand.Reader, letterLen)
		link[i] = letterBytes[index.Int64()]
	}

	return string(link)
}
