package crone

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestCronexprNext(t *testing.T) {
	cron := NewExpr("* * * * *")
	ti := cron.Next(time.Now())
	fmt.Printf("next1: %v\n", ti)

	ts := cron.NextN(time.Now(), 2)
	fmt.Printf("next2: %v\n", ts)
}

func TestCronexprTicker(t *testing.T) {
	cron := NewExpr("* * * * *")
	ch := make(chan time.Time)
	defer close(ch)
	ctx, cancel := context.WithCancel(context.Background())

	go cron.Notify(ctx, ch)

	go func() {
		select {
		case <-ctx.Done():
			fmt.Printf("cancel: %v\n", ctx.Err())
			return
		case t := <-ch:
			fmt.Printf("t: %v\n", t)
		}
	}()

	time.Sleep(61 * time.Second)
	cancel()
	time.Sleep(2 * time.Second)
}
