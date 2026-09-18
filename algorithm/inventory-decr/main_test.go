package main

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

func TestReserveRules(t *testing.T) {
	cases := []struct {
		name      string
		id        string
		qty       int
		want      error
		remaining int
	}{
		{"空请求号", "", 1, ErrInvalidRequest, 100},
		{"零数量", "invalid", 0, ErrInvalidRequest, 100},
		{"负数量", "invalid", -1, ErrInvalidRequest, 100},
		{"无效请求后可重试", "invalid", 1, nil, 99},
		{"库存不足", "retry", 100, ErrInsufficientStock, 99},
		{"失败请求不占用请求号", "retry", 9, nil, 90},
		{"重复成功", "retry", 9, nil, 90},
		{"数量冲突", "retry", 10, ErrQuantityConflict, 90},
		{"冲突不修改成功记录", "retry", 9, nil, 90},
		{"恰好扣完", "last", 90, nil, 0},
		{"耗尽后原请求仍成功", "last", 90, nil, 0},
		{"耗尽后数量冲突", "last", 91, ErrQuantityConflict, 0},
		{"耗尽后新请求失败", "new", 1, ErrInsufficientStock, 0},
	}
	inventory := NewInventory(100)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := inventory.Reserve(tc.id, tc.qty); !errors.Is(err, tc.want) {
				t.Fatalf("错误 = %v，期望 %v", err, tc.want)
			}
			if got := inventory.Remaining(); got != tc.remaining {
				t.Fatalf("库存 = %d，期望 %d", got, tc.remaining)
			}
		})
	}
}

func TestZeroStock(t *testing.T) {
	inventory := NewInventory(0)
	if err := inventory.Reserve("zero", 1); !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("零库存扣减错误 = %v", err)
	}
	if inventory.Remaining() != 0 {
		t.Fatal("失败后库存发生变化")
	}
}

// 同时放行请求，避免依赖休眠碰运气制造并发。
func concurrently(n int, f func(int)) {
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(n)
	for worker := 0; worker < n; worker++ {
		go func(worker int) {
			defer wg.Done()
			<-start
			f(worker)
		}(worker)
	}
	close(start)
	wg.Wait()
}

func TestConcurrentUniqueRequests(t *testing.T) {
	inventory := NewInventory(100)
	var successful atomic.Int64
	concurrently(500, func(worker int) {
		err := inventory.Reserve(fmt.Sprintf("request-%d", worker), 1)
		switch {
		case err == nil:
			successful.Add(1)
		case errors.Is(err, ErrInsufficientStock):
		default:
			t.Errorf("意外错误：%v", err)
		}
		// 在扣减仍可能进行时读取库存，交给竞态检测器检查读写同步。
		if got := inventory.Remaining(); got < 0 || got > 100 {
			t.Errorf("库存越界：%d", got)
		}
	})
	if successful.Load() != 100 || inventory.Remaining() != 0 {
		t.Fatalf("成功数 = %d，库存 = %d", successful.Load(), inventory.Remaining())
	}
}

func TestConcurrentSameRequest(t *testing.T) {
	inventory := NewInventory(100)
	concurrently(200, func(_ int) {
		if err := inventory.Reserve("same", 7); err != nil {
			t.Errorf("同号同数量应全部返回成功：%v", err)
		}
	})
	if got := inventory.Remaining(); got != 93 {
		t.Fatalf("应只扣一次，库存 = %d", got)
	}
}

func TestConcurrentConflictingQuantities(t *testing.T) {
	inventory := NewInventory(100)
	results := make([]error, 200)
	concurrently(len(results), func(worker int) {
		results[worker] = inventory.Reserve("same", 3+worker%2)
	})
	// 不假设哪一个请求先拿到锁，只检查最终状态及每个返回值一致。
	winnerQty := 100 - inventory.Remaining()
	if winnerQty != 3 && winnerQty != 4 {
		t.Fatalf("应只扣 3 或 4，实际扣减 %d", winnerQty)
	}
	for worker, err := range results {
		var want error
		if 3+worker%2 != winnerQty {
			want = ErrQuantityConflict
		}
		if !errors.Is(err, want) {
			t.Errorf("请求 %d 错误 = %v，期望 %v", worker, err, want)
		}
	}
}

func TestConcurrentBatchAtomicity(t *testing.T) {
	inventory := NewInventory(10)
	var successful atomic.Int64
	concurrently(2, func(worker int) {
		err := inventory.Reserve(fmt.Sprintf("batch-%d", worker), 6)
		if err == nil {
			successful.Add(1)
		} else if !errors.Is(err, ErrInsufficientStock) {
			t.Errorf("意外错误：%v", err)
		}
	})
	if successful.Load() != 1 || inventory.Remaining() != 4 {
		t.Fatalf("批量扣减必须全有或全无：成功数 = %d，库存 = %d", successful.Load(), inventory.Remaining())
	}
}
