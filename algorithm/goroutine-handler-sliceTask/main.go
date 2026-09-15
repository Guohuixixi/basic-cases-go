package main

import (
	"fmt"
	"sync"
)

func main() {
	taskNum := 888
	workerNum := 10
	var wg sync.WaitGroup
	taskList := make([]int, 0, taskNum)
	for i := 1; i <= taskNum; i++ {
		taskList = append(taskList, i)
	}

	for i := 0; i < workerNum; i++ {
		wg.Add(1)
		// 	计算每个协程处理的开始和截止下标
		step := taskNum/workerNum + 1
		starIndex := i * step
		endIndex := starIndex + step - 1

		if endIndex >= taskNum {
			endIndex = taskNum - 1
		}

		go func() {
			defer wg.Done()
			worker(starIndex, endIndex, taskList)
		}()
	}
	wg.Wait()
	fmt.Println(taskList)
}

func worker(starIndex, endIndex int, taskList []int) {
	for i := starIndex; i <= endIndex; i++ {
		taskList[i] = taskList[i] * taskList[i]
	}
}
