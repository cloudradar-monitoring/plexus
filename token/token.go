package token

import (
	"crypto/rand"
	"math/big"
)

var (
	tokenCharacters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func NewAuth() string {
	return New(20)
}

func New(length int) string {
	res := make([]byte, length)
	for i := range res {
		index := randIntn(len(tokenCharacters))
		res[i] = tokenCharacters[index]
	}
	return string(res)
}

func randIntn(n int) int {
	max := big.NewInt(int64(n))
	res, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic("random source is not available")
	}
	return int(res.Int64())
}
