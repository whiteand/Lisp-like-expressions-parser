package main

import (
	"fmt"
	"testing"
	"time"
)

func assume(t *testing.T, condition bool, msg string) {
	if !condition {
		t.Error(msg)
	}
}
func assumeF(t *testing.T, condition bool, format string, rest ...interface{}) {
	assume(t, condition, fmt.Sprintf(format, rest...))
}
func getValueOfEvaluate(in string) (res int, err error) {
	resChan := make(chan int)
	errChan := make(chan interface{})
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				errChan <- err
			}
		}()
		resChan <- evaluate(in)
	}()
	select {
	case <-time.After(100 * time.Millisecond):
		return -10, fmt.Errorf("So long for %s", in)
	case r := <-resChan:
		return r, nil
	case err := <-errChan:
		return -10, err.(error)
	}
}
func TestEvaluate(t *testing.T) {
	testsEvaluate := []struct {
		inCode string
		out    int
	}{
		{"1", 1},
		{"-1", -1},
		{"0", 0},
		{"(add 1 2)", 3},
		{"(mult 3 (add 2 3))", 15},
		{"(let x 1 x)", 1},
		{"(let x 1 2)", 2},
		{"(let x 2 (mult x 5))", 10},
		{"(let x 2 (mult x (let x 3 y 4 (add x y))))", 14},
		{"(let x 3 x 2 x)", 2},
		{"(let x 1 y 2 x (add x y) (add x y))", 5},
		{"(let x 2 (add (let x 3 (let x 4 x)) x))", 6},
		{"(let a1 3 b2 (add a1 1) b2) ", 4},
		{"(let a (add 1 2) b (mult a 3) c 4 d (add a b) (mult d d))", 144},
	}
	for _, test := range testsEvaluate {
		res, err := getValueOfEvaluate(test.inCode)
		if err != nil {
			t.Errorf("error while try %s --> %s", test.inCode, err.Error())
			continue
		}
		// res := evaluate(test.inCode)
		assumeF(t, res == test.out, "%s -> %d, but %d expected", test.inCode, res, test.out)
		if res == test.out {
			defer t.Logf("OK: %q -> %d", test.inCode, res)
		}
	}
}
