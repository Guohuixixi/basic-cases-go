package main

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrInvalidRequest    = errors.New("请求号不能为空且扣减数量必须大于零")
	ErrQuantityConflict  = errors.New("同一成功请求不能更改扣减数量")
	ErrInsufficientStock = errors.New("库存不足")
)

// Inventory 保存单进程内的库存和成功请求；使用后不能复制。
type Inventory struct {
	mu        sync.Mutex
	stock     int
	succeeded map[string]int
}

// NewInventory 创建库存，调用方保证初始库存非负。
func NewInventory(stock int) *Inventory {
	return &Inventory{stock: stock, succeeded: make(map[string]int)}
}

// Reserve 原子地完成查重和扣减，只记录成功请求。
func (i *Inventory) Reserve(requestID string, qty int) error {
	if requestID == "" || qty <= 0 {
		return ErrInvalidRequest
	}

	// 库存和成功记录共同组成业务状态，必须由同一把锁保护。
	i.mu.Lock()
	defer i.mu.Unlock()

	// 先判断历史成功记录，确保库存耗尽后重试仍能返回成功。
	if previousQty, ok := i.succeeded[requestID]; ok {
		if previousQty != qty {
			return ErrQuantityConflict
		}
		return nil
	}
	if i.stock < qty {
		return ErrInsufficientStock
	}

	i.stock -= qty
	i.succeeded[requestID] = qty
	return nil
}

// Remaining 返回读取这一时刻的库存快照。
func (i *Inventory) Remaining() int {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.stock
}

func main() {
	inventory := NewInventory(100)
	fmt.Println("首次扣减：", inventory.Reserve("request-1", 10))
	fmt.Println("相同请求重试：", inventory.Reserve("request-1", 10))
	fmt.Println("更改数量重试：", inventory.Reserve("request-1", 20))
	fmt.Println("剩余库存：", inventory.Remaining())
}
