package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	taskSlice := make([]int, 1000)
	for i := range taskSlice {
		taskSlice[i] = i
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if search(ctx, taskSlice, 989, 10) {
		fmt.Println("Yes")
	} else {
		fmt.Println("No")
	}
}

// search 在返回前等待所有查找协程退出，父上下文负责控制超时。
func search(parent context.Context, taskSlice []int, targetNum, workNum int) bool {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	if len(taskSlice) == 0 || ctx.Err() != nil {
		return false
	}
	if workNum <= 0 {
		workNum = 1
	}
	if workNum > len(taskSlice) {
		workNum = len(taskSlice)
	}
	var wg sync.WaitGroup
	var found atomic.Bool
	for i := 0; i < workNum; i++ {
		// 使用左闭右开区间，让不能整除时的剩余元素也被覆盖。
		start := i * len(taskSlice) / workNum
		end := (i + 1) * len(taskSlice) / workNum
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			worker(ctx, cancel, taskSlice[start:end], targetNum, &found)
		}(start, end)
	}
	// 取消只发送通知，等待才能保证返回时所有协程已经退出。
	wg.Wait()
	return found.Load()
}

func worker(ctx context.Context, cancel context.CancelFunc, tasks []int, targetNum int, found *atomic.Bool) {
	for _, num := range tasks {
		select {
		case <-ctx.Done():
			return
		default:
			// 未取消时继续查找，不能阻塞在取消信号上。
		}
		if num == targetNum {
			// 原子变量避免多个协程同时命中时产生数据竞争。
			found.Store(true)
			cancel()
			return
		}
	}
}
