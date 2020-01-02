package token_bucket

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestTokenBucket(t *testing.T) {
	tokenBucket := NewTokenBucket(100)
	tokenBucket.Start()
	var wg sync.WaitGroup
	wg.Add(1000)
	start := time.Now()
	for i := 0; i < 1000; i++ {
		go func(i int) {
			defer wg.Done()
			tokenBucket.WithToken(func() {
				fmt.Println(time.Now(), i, "do work.")
			})
		}(i)
	}
	wg.Wait()
	fmt.Println(time.Now().Sub(start))
	tokenBucket.Stop()
}
