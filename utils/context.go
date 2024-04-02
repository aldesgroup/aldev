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

type CancelableContext interface {
	context.Context
	CancelAll()
}

type baseCancelableContext struct {
	context.Context
	cancelFn func()
}

func newBaseCancelableContext() *baseCancelableContext {
	ctx, cancelFn := context.WithCancel(context.Background())
	return &baseCancelableContext{ctx, cancelFn}
}

func (thisCtx *baseCancelableContext) CancelAll() {
	thisCtx.cancelFn()
}

// checking this base implem does satisfy the interface above
var _ CancelableContext = (*baseCancelableContext)(nil)

// an Aldev Context has a loop and allows to restart it
type AldevContext interface {
	CancelableContext
	GetLoopCtx() CancelableContext
	RestartLoop()
}

// an Aldev Context consists in a base context that can be canceled,
// + a cancelable context for the loop over the files watched by Aldev directly
type aldevContext struct {
	baseCancelableContext
	loopCtx *baseCancelableContext
}

func (aldevCtx *aldevContext) GetLoopCtx() CancelableContext {
	return aldevCtx.loopCtx
}

func (aldevCtx *aldevContext) RestartLoop() {
	// Cancel the current loop, and thus any function (which works with context) running inside
	aldevCtx.loopCtx.cancelFn()

	// Recreate the context
	aldevCtx.loopCtx = newBaseCancelableContext()
}

// method override to cancel the loop context as well
func (aldevCtx *aldevContext) CancelAll() {
	aldevCtx.loopCtx.CancelAll()
	Info("Waiting for some cleanup...")
	time.Sleep(2000 * time.Millisecond) // TODO waiting
	aldevCtx.cancelFn()
}

func InitAldevContext() *aldevContext {
	// init
	aldevCtx := &aldevContext{
		baseCancelableContext: *newBaseCancelableContext(),
		loopCtx:               newBaseCancelableContext(),
	}

	// Initialize a context that can be interrupted:

	// Initialize channel to receive signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	// Start a goroutine to handle signals
	go func() {
		sig := <-signalCh
		Warn("Received signal: %v", sig)

		// Cancel the context on signal received
		aldevCtx.CancelAll()
	}()

	return aldevCtx
}
