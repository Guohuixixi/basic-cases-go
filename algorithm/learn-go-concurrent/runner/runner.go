package runner

import (
	"errors"
	"os"
	"os/signal"
	"time"
)

var (
	ErrInterrupted = errors.New("received interrupted from OS")
	ErrTimeout     = errors.New("cannot finish the tasks within timeout")
)

type SimpleRunner struct {
	Interrupt chan os.Signal
	Timeout   <-chan time.Time // 用于计时
	Complete  chan error       // 是否完成
	Tasks     []func(int)
}

func New(t time.Duration) *SimpleRunner {
	return &SimpleRunner{
		Interrupt: make(chan os.Signal, 1),
		Timeout:   time.After(t),
		Complete:  make(chan error),
		Tasks:     make([]func(int), 0),
	}
}
func (r *SimpleRunner) Run() error {
	for id, task := range r.Tasks {
		select {
		case <-r.Interrupt:
			signal.Stop(r.Interrupt)
			return ErrInterrupted
		default:
			task(id)
		}
	}
	return nil
}

func (r *SimpleRunner) Start() error {
	// relay interrupt from os
	signal.Notify(r.Interrupt, os.Interrupt)
	go func() {
		r.Complete <- r.Run()
	}()
	select {
	case err := <-r.Complete:
		return err
	case <-r.Timeout:
		return ErrTimeout
	}
}

func (r *SimpleRunner) AddTasks(tasks ...func(int)) {
	r.Tasks = append(r.Tasks, tasks...)
}
