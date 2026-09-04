package main

import (
	"fmt"
	"time"

	runner2 "github.com/xxgh/basic-cases-go/algorithm/learn-go-concurrent/runner"
)

func creatTask() func(int) {
	return func(id int) {
		time.Sleep(time.Second * 1)
		fmt.Printf("Task completed #%d\n", id)
	}
}

func main() {
	runner := runner2.New(4 * time.Second)
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
