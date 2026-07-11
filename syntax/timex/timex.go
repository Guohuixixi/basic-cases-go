package timex

import (
	"context"
	"fmt"
	"time"

	cron "github.com/robfig/cron/v3"
)

// RunTimer 每秒触发一次，直到外部传入的 context 结束。
func RunTimer(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			fmt.Println(now.Format(time.RFC3339))
		case <-ctx.Done():
			return
		}
	}
}

// StartCron 注册并启动一个每秒执行的任务，返回的 stop 用来安全停掉调度器。
func StartCron(job func()) (stop func(), err error) {
	scheduler := cron.New(cron.WithSeconds())
	if _, err = scheduler.AddFunc("@every 1s", job); err != nil {
		return nil, err
	}
	scheduler.Start()

	return func() {
		// Stop 会等已经开始的任务收尾，避免程序退出时把任务拦腰截断。
		<-scheduler.Stop().Done()
	}, nil
}
