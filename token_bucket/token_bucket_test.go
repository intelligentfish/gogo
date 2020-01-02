package token_bucket

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestTokenBucket(t *testing.T) {
	tokenBucket := NewTokenBucket(10)
	tokenBucket.Start()
	var wg sync.WaitGroup
	wg.Add(1000)
	start := time.Now()
	for i := 0; i < 1000; i++ {
		go func(i int) {
			defer wg.Done()
			tokenBucket.WithToken(func() {
				fmt.Println(time.Now(), i, "run...")
			})
		}(i)
	}
	wg.Wait()
	fmt.Println(time.Now().Sub(start))
	time.Sleep(10 * time.Second)
	tokenBucket.Stop()
}
