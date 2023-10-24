package crone

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
)

type Matches struct {
	minute  []int
	hour    []int
	day     []int
	month   []int
	weekday []int
}

func (m Matches) Match(t time.Time) bool {
	contains := func(arr []int, val int) bool {
		for _, v := range arr {
			if v == val {
				return true
			}
		}
		return false
	}

	return contains(m.minute, t.Minute()) &&
		contains(m.hour, t.Hour()) &&
		contains(m.day, t.Day()) &&
		contains(m.month, int(t.Month())) &&
		contains(m.weekday, int(t.Weekday()))
}

type Cronexpr struct {
	expr     string
	matches  Matches
	accurate time.Duration
}

func NewExpr(expr string) *Cronexpr {
	return &Cronexpr{
		expr:     expr,
		matches:  newMatches(expr),
		accurate: time.Minute,
	}
}

func newMatches(expr string) Matches {
	splits := strings.Split(expr, " ")
	if len(splits) != 5 {
		return Matches{}
	}

	mustParse := func(s string, f field) []int {
		matches, err := parse(s, f)
		if err != nil {
			panic(err)
		}
		return matches
	}

	return Matches{
		minute:  mustParse(splits[0], minute),
		hour:    mustParse(splits[1], hour),
		day:     mustParse(splits[2], day),
		month:   mustParse(splits[3], month),
		weekday: mustParse(splits[4], weekday),
	}
}

type field int

const (
	minute field = iota
	hour
	day
	month
	weekday
)

func (f field) limit() (int, int) {
	switch f {
	case minute:
		return 0, 59
	case hour:
		return 0, 23
	case day:
		return 1, 31
	case month:
		return 1, 12
	case weekday:
		return 0, 6
	}
	return 0, 0
}

func parse(rule string, f field) ([]int, error) {
	if len(rule) == 0 {
		return nil, errors.New("empty spec")
	}

	specs := strings.Split(rule, ",")
	matches := make([]int, 0)
	low, high := f.limit()

	for _, spec := range specs {
		if spec == "*" {
			for i := low; i < high; i++ {
				matches = append(matches, i)
			}
		} else if strings.Contains(spec, "/") {
			parts := strings.Split(spec, "/")
			if len(parts) != 2 {
				return nil, errors.New("invalid spec")
			}
			step, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, errors.New("invalid spec")
			}
			for i := low; i < high; i += step {
				matches = append(matches, i)
			}
		} else if strings.Contains(spec, "-") {
			minuteRange := strings.Split(spec, "-")
			if len(minuteRange) != 2 {
				return nil, errors.New("invalid spec")
			}

			start, err := strconv.Atoi(minuteRange[0])
			if err != nil {
				return nil, errors.New("invalid spec")
			}
			end, err := strconv.Atoi(minuteRange[1])
			if err != nil {
				return nil, errors.New("invalid spec")
			}

			if start > end || start < low || end > high {
				return nil, errors.New("invalid spec")
			}

			for i := start; i <= end; i++ {
				matches = append(matches, i)
			}
		} else {
			val, err := strconv.Atoi(spec)
			if err != nil {
				return nil, errors.New("invalid spec")
			}
			if val < low || val > high {
				return nil, errors.New("invalid spec")
			}
			matches = append(matches, val)
		}
	}

	return matches, nil
}

func (e *Cronexpr) Expr() string {
	return e.expr
}

var END = time.Now().AddDate(20, 0, 0) // 20 years is enough

func (e *Cronexpr) Next(z time.Time) time.Time {
	return e.next1(z)
}

func (e *Cronexpr) NextN(z time.Time, n int) []time.Time {
	return e.nextN(z, n)
}

func (e *Cronexpr) nextN(z time.Time, n int) []time.Time {
	ts := make([]time.Time, 0, n)
	lt := z

	for i := 0; i < n; i++ {
		n1 := e.next1(lt)
		ts = append(ts, n1)
		lt = n1
	}

	return ts
}

func (e *Cronexpr) next1(z time.Time) time.Time {
	for t := z.Add(e.accurate); t.Before(END); t = t.Add(e.accurate) {
		if e.matches.Match(t) {
			return t
		}
	}

	return END
}

func (e *Cronexpr) Notify(ctx context.Context, out chan<- time.Time) {
	ticker := time.NewTicker(e.accurate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			if e.matches.Match(t) {
				out <- t
			}
		}
	}
}
