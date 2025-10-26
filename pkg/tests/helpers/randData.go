package helpers

import (
	"math/rand"
	"sync"
	"time"
)

var (
	randSource = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	mu         sync.Mutex
)

func RandStrLimit(n int) string {
	mu.Lock()
	defer mu.Unlock()

	letters := []rune("abcdefghijklmnopqrstuvwxyz") // для универсальности пусть будут только буквы нижнего регистра
	lettersLen := len(letters)                      // count runes
	b := make([]rune, n)

	for i := range b {
		randomIdx := randSource.Intn(lettersLen) // 0 - (lettersLen-1)
		b[i] = letters[randomIdx]
	}

	return string(b)
}

func RandStr() string {
	return RandStrLimit(10)
}

func RandEmail() string {
	return RandStr() + "@example.com"
}

func RandIntByRange(minSrc, maxSrc int) int {
	mu.Lock()
	defer mu.Unlock()
	return randSource.Intn(maxSrc-minSrc) + minSrc
}
