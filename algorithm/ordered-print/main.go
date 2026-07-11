package main

import (
	"fmt"
	"sync"
)

//  同时开启100个协程(分别为1号协程 2号协程 ... 100号协程，
//  1号协程只打印尾数为1的数字，2号协程只打印尾数为2的数，
//   以此类推)，请顺序打印1-1000整数以及对应的协程号；

func main() {
	s := make(chan struct{})

	// 通过 map 的 key 来保证协程的顺序
	m := make(map[int]chan int, 100)

	var wg sync.WaitGroup

	// 填充 map，初始化 channel
	for i := 1; i <= 100; i++ {
		m[i] = make(chan int)
	}

	// 开启 100 个协程，死循环打印
	// go func() { 这个协程不加也可以的
	for i := 1; i <= 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// 用range自动退出
			for num := range m[id] {
				fmt.Printf("goroutine %d -> %d\n", id, num)
				s <- struct{}{}
			}
		}(i)
	}
	// }()

	// 循环 1-1000，并把值传递给匹配的 map
	// 然后通过 s 限制顺序打印
	for i := 1; i <= 1000; i++ {
		id := i % 100
		if id == 0 {
			id = 100
		}

		m[id] <- i

		// 通过这个来控制打印顺序，每次遍历一次 i
		// 都通过 s 阻塞协程的打印，最后打印完毕
		<-s
	}
	// 关闭所有chan
	for i := 1; i <= 100; i++ {
		close(m[i])
	}
	wg.Wait()

	fmt.Println("all goroutines exited")
}
