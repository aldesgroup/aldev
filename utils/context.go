// ----------------------------------------------------------------------------
// The code here is about cancelable contexts, to help gracefully stop
// some processes
// ----------------------------------------------------------------------------
package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type mainContext struct {
	ctx          context.Context
	mainCancelFn func()
	cancelLoopFn func()
}

func (thisCtx *mainContext) SetCancelLoopFn(cancelLoopFn func()) {
	thisCtx.cancelLoopFn = cancelLoopFn
}

func (thisCtx *mainContext) cancelAll() {
	thisCtx.cancelLoopFn()
	Info("Waiting for some cleanup...")
	time.Sleep(2000 * time.Millisecond) // TODO waiting
	thisCtx.mainCancelFn()
}

func (thisCtx *mainContext) Done() <-chan struct{} {
	return thisCtx.ctx.Done()
}

func InitMainContext() *mainContext {
	mainContext := &mainContext{}

	// Initialize a cancelable context
	mainContext.ctx, mainContext.mainCancelFn = context.WithCancel(context.Background())

	// Initialize channel to receive signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	// Start a goroutine to handle signals
	go func() {
		sig := <-signalCh
		Warn("Received signal: %v", sig)

		// Cancel the context on signal received
		mainContext.cancelAll()
	}()

	return mainContext
}
