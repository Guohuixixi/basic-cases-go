package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	pool2 "github.com/xxgh/basic-cases-go/algorithm/learn-go-concurrent-pool/pool"
)

type DBConn struct {
	Id int32
}

// 关闭DB连接
func (D DBConn) Close() error {
	// 模拟随机时间关闭
	t := rand.Intn(5) + 1
	time.Sleep(time.Duration(t) * time.Second)
	fmt.Println("Closed DBConn #" + fmt.Sprint(D.Id))
	return nil
}

var (
	counter int32
	wg      sync.WaitGroup
)

// factory创建新的DB连接
func factory() (io.Closer, error) {
	atomic.AddInt32(&counter, 1)
	return DBConn{Id: counter}, nil
}

func doQuery(id int, p *pool2.Pool) {
	resource, err := p.AquireResource()
	if err != nil {
		fmt.Println("mark once wrong", err)
		return
	}
	// 模拟执行查询操作
	t := rand.Intn(2) + 1
	time.Sleep(time.Duration(t) * time.Second)
	fmt.Println("查询工作已完成")
	// 释放资源
	p.ReleaseResource(resource)

}

func main() {

	pool, err := pool2.New(5, factory)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	num := 10
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func(i int, pool *pool2.Pool) {
			// 这里可以模拟一个随机时间开启的查询任务，来体现资源池的实际工作中减少新资源创建开销
			time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)
			doQuery(i, pool)
			wg.Done()
		}(i, pool)
	}

	wg.Wait()

	fmt.Println("All queryWork have done")

}
