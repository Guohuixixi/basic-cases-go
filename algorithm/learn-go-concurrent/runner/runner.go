package runner

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"time"
)

var (
	ErrInterrupted = errors.New("received interrupted from OS")
	ErrTimeout     = errors.New("cannot finish the tasks within timeout")
)

type Task func(ctx context.Context, id int) error
type SimpleRunner struct {
	Timeout time.Duration // 用于计时
	Tasks   []Task
}

func New(t time.Duration) *SimpleRunner {
	return &SimpleRunner{
		Timeout: t,
		Tasks:   make([]Task, 0),
	}
}
func (r *SimpleRunner) Run(ctx context.Context) error {
	for id, task := range r.Tasks {

		if err := ctx.Err(); err != nil {
			return ConvertContextErr(err)
		}
		if err := task(ctx, id); err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ConvertContextErr(ctxErr)
			}
			return err
		}
	}
	return nil
}

func (r *SimpleRunner) Start() error {
	// relay interrupt from os
	ctx, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stopSignal()
	ctx, cancelFunc := context.WithTimeout(ctx, r.Timeout)
	defer cancelFunc()

	return r.Run(ctx)

}

func (r *SimpleRunner) AddTasks(tasks ...Task) {
	r.Tasks = append(r.Tasks, tasks...)
}

func ConvertContextErr(err error) error {
	switch err {
	case context.DeadlineExceeded:
		return ErrTimeout
	case context.Canceled:
		return ErrInterrupted
	default:
		return err
	}
}
