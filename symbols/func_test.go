package symbols

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestA(t *testing.T) {
	task := make(chan int, 1)
	wg := sync.WaitGroup{}
	wg.Add(1)
	max := 0
	// read
	go func() {
		defer wg.Done()
		for {
			select {
			case j := <-task:
				if max != 0 && max != j {
					continue
				}
				fmt.Println("read", j)
				if j == 1 {
					time.Sleep(10 * time.Second)
				}
				time.Sleep(3 * time.Second)
			}
		}

	}()
	wg.Add(1)
	// write
	go func() {
		defer wg.Done()
		t1 := time.NewTicker(2 * time.Second)
		i := 0
		for {
			i++
			select {
			case <-t1.C:
				fmt.Println("write", i)
				max = i
				task <- i
			}
		}
	}()
	wg.Wait()
}
