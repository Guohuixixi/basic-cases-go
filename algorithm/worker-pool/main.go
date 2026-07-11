package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// master要处理1000个计算任务，开启协程去做，并将计算结果返回给master，要求开启的协程数不超过三个，当超时5s后就会退出协程，如何实现（编程）
type Task struct {
	ID    int
	Retry int
}

type Result struct {
	TaskID int
	Value  int
	Err    error
}

func main() {
	taskNum := 1000
	workerLimit := 3
	maxRetry := 3

	tasks := make(chan Task, taskNum)
	results := make(chan Result, taskNum)

	var wg sync.WaitGroup

	rand.Seed(time.Now().UnixNano())

	// =====================
	// worker 启动函数（支持“死亡 + 重启”）
	// =====================
	var worker func(id int)

	worker = func(id int) {
		wg.Add(1)

		go func() {
			defer wg.Done()

			fmt.Printf("worker %d started\n", id)

			for task := range tasks {

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

				done := make(chan int, 1)

				// 模拟任务执行
				go func(t Task) {
					sleep := time.Duration(rand.Intn(7000)) * time.Millisecond
					time.Sleep(sleep)
					done <- t.ID * 2
				}(task)

				select {
				case val := <-done:
					results <- Result{TaskID: task.ID, Value: val}
					cancel()

				case <-ctx.Done():
					cancel()

					// 👉 超时：worker 自杀
					fmt.Printf("worker %d timeout on task %d, exit\n", id, task.ID)

					// 👉 任务重试
					if task.Retry < maxRetry {
						task.Retry++
						go func(t Task) {
							tasks <- t
						}(task)
					} else {
						results <- Result{
							TaskID: task.ID,
							Err:    fmt.Errorf("failed after %d retries", maxRetry),
						}
					}

					// 🔥 worker 退出（关键）
					go worker(id) // 补一个新的
					return
				}
			}
		}()
	}

	// =====================
	// 启动初始 worker（3个）
	// =====================
	for i := 0; i < workerLimit; i++ {
		worker(i + 1)
	}

	// =====================
	// 投递任务
	// =====================
	go func() {
		for i := 1; i <= taskNum; i++ {
			tasks <- Task{ID: i}
		}
	}()

	// =====================
	// 收集结果（控制退出）
	// =====================
	done := make(chan struct{})

	go func() {
		count := 0
		for res := range results {
			if res.Err != nil {
				fmt.Printf("task %d failed: %v\n", res.TaskID, res.Err)
			} else {
				fmt.Printf("task %d result: %d\n", res.TaskID, res.Value)
			}

			count++
			if count == taskNum {
				close(done)
				return
			}
		}
	}()

	// 等所有任务完成
	<-done

	// 关闭任务通道
	close(tasks)

	// 等所有 worker 退出
	wg.Wait()

	close(results)

	fmt.Println("all tasks finished")
}
