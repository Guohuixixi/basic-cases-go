package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup

	msgCh := make(chan int, 5)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rand.Seed(time.Now().UnixNano())

	// 👉 单独一个 wg 控制生产者
	var producerWg sync.WaitGroup

	// 生产者
	producer := func(ctx context.Context, out chan<- int, idx int) {
		defer producerWg.Done()
		for {
			num := rand.Intn(500)

			select {
			case <-ctx.Done():
				fmt.Printf("producer %d finished\n", idx)
				return
			case out <- num:
				fmt.Printf("producer: %d, msg: %d\n", idx, num)
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			}
		}
	}

	// 消费者
	consumer := func(in <-chan int, idx int) {
		defer wg.Done()
		for num := range in { // ✅ 自动消费直到 channel 关闭
			fmt.Printf("consumer %d, msg: %d\n", idx, num)
		}
		fmt.Printf("consumer %d finished\n", idx)
	}

	// 启动生产者
	for i := 0; i < 5; i++ {
		producerWg.Add(1)
		go producer(ctx, msgCh, i+1)
	}

	// 启动消费者
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go consumer(msgCh, i+1)
	}

	// 👉 等生产者全部退出
	producerWg.Wait()

	// 👉 关闭 channel（关键！！）
	close(msgCh)

	// 👉 等消费者消费完退出
	wg.Wait()

	fmt.Println("all finished")
}
