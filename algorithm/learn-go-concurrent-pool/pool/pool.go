package pool

import (
	"errors"
	"fmt"
	"io"
	"sync"
)

var ErrClosed = errors.New("pool closed")

type Pool struct {
	Size     int                       // 资源池大小
	Resource chan io.Closer            // 可关闭的资源
	Closed   bool                      // 是否关闭
	Factory  func() (io.Closer, error) // 工厂函数创建新资源
	Mtx      sync.Mutex                // 锁
}

func New(size int, factory func() (io.Closer, error)) (*Pool, error) {
	if size <= 0 {
		return nil, errors.New("Invalid size")
	}
	pool := Pool{
		Size:     size,
		Resource: make(chan io.Closer, size),
		Factory:  factory,
	}
	for i := 0; i < size; i++ {
		resource, err := pool.Factory()
		if err != nil {
			// 记录一下，预先创建size大小资源失败，可以后续再慢慢创建资源
			fmt.Println("预先创建资源失败")
			continue
		}
		pool.Resource <- resource
	}
	return &pool, nil
}

func (p *Pool) AquireResource() (io.Closer, error) {
	select {
	case resource, ok := <-p.Resource:
		if !ok {
			return nil, ErrClosed
		}
		return resource, nil
	default:
		fmt.Println("aquire  new resource")
		return p.Factory()
	}
}

func (p *Pool) ReleaseResource(resource io.Closer) {

	p.Mtx.Lock()
	defer p.Mtx.Unlock()

	if p.Closed {
		resource.Close()
		return
	}
	select {
	case p.Resource <- resource: // 放回资源池子了
		fmt.Println("Resource released back to pool")
	default:
		fmt.Println("Release resource closed")
		resource.Close()
	}
}

func (p *Pool) Close() {
	p.Mtx.Lock()
	defer p.Mtx.Unlock()

	if p.Closed {
		return
	}

	p.Closed = true
	close(p.Resource)

	for resource := range p.Resource {
		resource.Close()
	}
}
