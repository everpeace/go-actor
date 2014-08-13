package actor

import "runtime"

func forever(fun func() bool) {
	for {
		// if fun() returns true, forever ends.
		if fun() {
			break
		}
		// force give back control to goroutine scheduler.
		runtime.Gosched()
	}
}
