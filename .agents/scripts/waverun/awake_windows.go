package main

import (
	"fmt"
	"runtime"
	"syscall"
)

const (
	esContinuous     = 0x80000000
	esSystemRequired = 0x00000001
)

// keepAwake stops Windows from sleeping until the runner exits. A sleep
// freezes the session mid-command (wave G.2, session 6: six hours asleep,
// and a cold prompt cache on waking). The request belongs to the thread that
// makes it, so one goroutine locks a thread, makes it and never returns; the
// request ends with the process. The display may still turn off.
func keepAwake() {
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("SetThreadExecutionState")
	if proc.Find() != nil {
		return
	}
	ok := make(chan bool)
	go func() {
		runtime.LockOSThread()
		r, _, _ := proc.Call(esContinuous | esSystemRequired)
		ok <- r != 0
		select {}
	}()
	if !<-ok {
		fmt.Println("Could not keep the PC awake; a sleep will pause the run.")
	}
}
