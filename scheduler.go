package crone

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Job struct {
	name string
	rule string
	fn   func()
}

func newJob(name, rule string, fn func()) *Job {
	return &Job{
		name: name,
		rule: rule,
		fn:   fn,
	}
}

func (job *Job) Run() {
	job.fn()
}

type Scheduler struct {
	jobs      []*Job
	stopCh    chan struct{}
	cancelMap map[string]context.CancelFunc
}

func NewSchduler() *Scheduler {
	return &Scheduler{
		jobs:      make([]*Job, 0),
		stopCh:    make(chan struct{}),
		cancelMap: make(map[string]context.CancelFunc),
	}
}

func (s *Scheduler) Add(name, rule string, fn func()) {
	s.jobs = append(s.jobs, newJob(name, rule, fn))
}

func (s *Scheduler) Start() {
	for _, job := range s.jobs {
		ctx, cancel := context.WithCancel(context.Background())
		s.cancelMap[job.name] = cancel
		go s.startJob(ctx, job)
	}
}

func (s *Scheduler) StartWithSignalListen() {
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		log.Printf("received signal, shutting down")
		s.Stop()
	}()

	s.Start()
}

func (s *Scheduler) startJob(ctx context.Context, job *Job) {
	cron := NewExpr(job.rule)
	ch := make(chan time.Time)
	defer close(ch)

	go cron.Notify(ctx, ch)

	for {
		select {
		case <-ch:
			go job.Run()
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scheduler) Stop() {
	for k, c := range s.cancelMap {
		c()
		log.Printf("%s canceled", k)
	}

	s.stopCh <- struct{}{}
}

func (s *Scheduler) StopJob(name string) {
	s.cancelMap[name]()
}

func (s *Scheduler) Wait() {
	<-s.stopCh
}
