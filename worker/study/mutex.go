package study

import (
	"fmt"
	"sync"
)

func Mutex() {
	var counter int
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			mu.Lock()
			counter++ // nhiều goroutines cùng sửa biến này
			mu.Unlock()
			wg.Done()
		}()
	}

	wg.Wait()
	fmt.Println("Counter:", counter)
}
