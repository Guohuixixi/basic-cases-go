package main

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestSearch(t *testing.T) {
	for _, tc := range []struct {
		name            string
		tasks           []int
		target, workers int
		want            bool
	}{
		{"末尾余数", []int{0, 1, 2, 3, 4}, 4, 2, true},
		{"不存在", []int{1, 2, 3}, 4, 2, false},
		{"重复目标", []int{2, 2, 2, 2}, 2, 4, true},
		{"协程多于元素", []int{1, 2}, 2, 10, true},
		{"空切片", nil, 1, 10, false},
		{"零协程", []int{1}, 1, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := search(context.Background(), tc.tasks, tc.target, tc.workers); got != tc.want {
				t.Fatalf("search = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSearchExpired(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	if search(ctx, []int{1}, 1, 1) {
		t.Fatal("超时后不应继续查找")
	}
}

func TestWorkerCancelsPeers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var found atomic.Bool
	worker(ctx, cancel, []int{1, 2}, 2, &found)
	if !found.Load() || ctx.Err() != context.Canceled {
		t.Fatal("命中后应记录结果并广播取消")
	}
	// 使用独立结果验证另一个协程收到取消后不会再处理目标。
	var peerFound atomic.Bool
	worker(ctx, cancel, []int{2}, 2, &peerFound)
	if peerFound.Load() {
		t.Fatal("取消后其他协程应退出")
	}
}
