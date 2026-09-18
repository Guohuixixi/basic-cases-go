package main

import (
	"context"
	"fmt"
	"time"

	runner2 "github.com/xxgh/basic-cases-go/algorithm/learn-go-concurrent/runner"
)

func creatTask() runner2.Task {
	return func(ctx context.Context, id int) error {
		timer := time.NewTimer(1 * time.Second)
		select {
		case <-timer.C:
			fmt.Printf("Task completed #%d\n", id)
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func main() {
	runner := runner2.New(3 * time.Second)
	runner.AddTasks(creatTask(), creatTask(), creatTask())
	err := runner.Start()
	switch err {
	case runner2.ErrInterrupted:
		fmt.Println("Interrupted")
	case runner2.ErrTimeout:
		fmt.Println("Timeout")
	default:
		fmt.Println("All tasks finished")
	}
}
