package fmx

import "fmt"

type SafeRoutine struct {
	logwriter func(string)
}

func NewGoRoutine(logfn func(string)) *SafeRoutine {
	return &SafeRoutine{logwriter: logfn}
}

func (t *SafeRoutine) Run(fn func()) {
	if fn == nil {
		return
	}

	go func() {
		defer func() {
			recover()
		}()

		defer func() {
			if err := recover(); err != nil {
				stack := stack(3)
				panicInfo := fmt.Sprintf("PANIC: %s\n%s", err, stack)
				if t.logwriter != nil {
					t.logwriter(panicInfo)
				}
			}
		}()

		fn()
	}()
}
