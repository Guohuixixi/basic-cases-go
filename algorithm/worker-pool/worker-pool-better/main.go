package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	workerCount = 3
	taskTimeout = 5 * time.Second
	taskCount   = 10
)

type Task struct {
	ID    int
	Value int
}

type Result struct {
	TaskID   int
	Value    int
	Err      error
	WorkerID int
}

type WorkerExit struct {
	WorkerID int
	Timeout  bool
}

// calculate 必须配合 context 取消，否则无法保证 worker 在 5 秒后退出
func calculate(ctx context.Context, task Task) (int, error) {
	// 模拟部分任务执行超过 5 秒
	duration := time.Second
	if task.ID%3 == 0 {
		duration = 6 * time.Second
	}

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		return task.Value * task.Value, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func worker(
	workerID int,
	jobs <-chan Task,
	results chan<- Result,
	exits chan<- WorkerExit,
) {
	timedOut := false

	defer func() {
		exits <- WorkerExit{
			WorkerID: workerID,
			Timeout:  timedOut,
		}
	}()

	for task := range jobs {
		// 每个任务都有自己独立的 5 秒超时时间
		ctx, cancel := context.WithTimeout(
			context.Background(),
			taskTimeout,
		)

		value, err := calculate(ctx, task)
		cancel()

		results <- Result{
			TaskID:   task.ID,
			Value:    value,
			Err:      err,
			WorkerID: workerID,
		}

		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf(
				"worker %d 处理任务 %d 超时，即将退出\n",
				workerID,
				task.ID,
			)

			timedOut = true
			return
		}
	}
}

func main() {
	jobs := make(chan Task, taskCount)
	results := make(chan Result, taskCount)
	exits := make(chan WorkerExit, workerCount)

	// 提前放入全部任务，之后不再产生新任务
	for i := 0; i < taskCount; i++ {
		jobs <- Task{
			ID:    i,
			Value: i,
		}
	}
	close(jobs)

	nextWorkerID := 1
	activeWorkers := 0

	startWorker := func() {
		workerID := nextWorkerID
		nextWorkerID++
		activeWorkers++

		fmt.Printf("启动 worker %d\n", workerID)

		go worker(
			workerID,
			jobs,
			results,
			exits,
		)
	}

	// 初始只启动 3 个 worker
	for i := 0; i < workerCount; i++ {
		startWorker()
	}

	// master 负责监督 worker
	for activeWorkers > 0 {
		exit := <-exits
		activeWorkers--

		if exit.Timeout {
			// 旧 worker 已经结束任务处理，补充一个新 worker
			startWorker()
		}
	}

	close(results)

	successCount := 0
	timeoutCount := 0

	for result := range results {
		if errors.Is(result.Err, context.DeadlineExceeded) {
			timeoutCount++
			continue
		}

		if result.Err == nil {
			successCount++
		}
	}

	fmt.Printf(
		"处理完成：成功 %d，超时 %d\n",
		successCount,
		timeoutCount,
	)
}
