package shortener

import (
	"crypto/rand"
	"math/big"
)

const (
	CodeLength = 10
	// 63 символа длина
	alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_"
)

func Generate() (string, error) {
	shortURL := make([]byte, CodeLength)
	maximum := big.NewInt(int64(len(alphabet)))
	for i := 0; i < len(shortURL); i++ {
		n, err := rand.Int(rand.Reader, maximum)
		if err != nil {
			return "", err
		}
		// TODO: Почему такой алгоритм? не вылезет n за пределы 63? не понимаю
		shortURL[i] = alphabet[n.Int64()]
	}
	result := string(shortURL)
	return result, nil
}
