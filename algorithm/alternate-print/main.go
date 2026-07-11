package main

import (
	"fmt"
	"sync"
)

func main() {
	// 三个无缓冲通道像接力棒：拿到信号才能打印，打印后交给下一个协程。
	ch1 := make(chan struct{})
	ch2 := make(chan struct{})
	ch3 := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		for i := 1; i <= 10; i++ {
			<-ch1
			fmt.Println("a")
			ch2 <- struct{}{}
		}
	}()

	go func() {
		defer wg.Done()
		for i := 1; i <= 10; i++ {
			<-ch2
			fmt.Println("b")
			ch3 <- struct{}{}
		}
	}()

	go func() {
		defer wg.Done()
		for i := 1; i <= 10; i++ {
			<-ch3
			fmt.Println("c")
			// 最后一轮已经不用唤醒 a，继续发送会因为无人接收而卡住。
			if i < 10 {
				ch1 <- struct{}{}
			}
		}
	}()

	ch1 <- struct{}{}
	wg.Wait()
	fmt.Println("end")
}
