// ----------------------------------------------------------------------------
// The code here is about cancelable contexts, to help gracefully stop
// some processes
// ----------------------------------------------------------------------------
package utils

import (
	"context"
	"io"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"
)

// ----------------------------------------------------------------------------
// Cancelable Context
// ----------------------------------------------------------------------------

type CancelableContext interface {
	context.Context
	WithExecDir(...string) CancelableContext
	getExecDir() string
	WithStdOutWriter(io.Writer) CancelableContext
	getStdOutWriter() io.Writer
	WithStdErrWriter(io.Writer) CancelableContext
	getStdErrWriter() io.Writer
	WithErrLogFn(errLogFn) CancelableContext
	getErrLogFn() errLogFn
	WithEnvVars(...string) CancelableContext
	getEnvVars() []string
	WithReRun() CancelableContext
	isReRun() bool
	WithAllowFailure(bool) CancelableContext
	isAllowingFailure() bool
	CancelAll()
	NewChildContext() CancelableContext
}

// ----------------------------------------------------------------------------
// Base Cancelable Context
// ----------------------------------------------------------------------------

type baseCancelableContext struct {
	context.Context
	cancelFn      func()
	execDir       string
	shortCommands bool
	stdoutWriter  io.Writer
	stderrWriter  io.Writer
	errLogFn      errLogFn
	envVars       []string
	reRun         bool
	allowFailure  bool
}

func NewBaseContext() *baseCancelableContext {
	return &baseCancelableContext{
		Context: context.WithoutCancel(context.Background()),
	}
}

func newBaseCancelableContext() *baseCancelableContext {
	ctx, cancelFn := context.WithCancel(context.Background())
	return &baseCancelableContext{ctx, cancelFn, "", false, nil, nil, nil, nil, false, false}
}

func (thisCtx *baseCancelableContext) WithExecDir(dirElems ...string) CancelableContext {
	if len(dirElems) > 0 {
		thisCtx.execDir = path.Join(dirElems...)
	}

	return thisCtx
}

func (thisCtx *baseCancelableContext) getExecDir() string {
	return thisCtx.execDir
}

func (thisCtx *baseCancelableContext) WithStdOutWriter(writer io.Writer) CancelableContext {
	thisCtx.stdoutWriter = writer
	return thisCtx
}

func (thisCtx *baseCancelableContext) getStdOutWriter() io.Writer {
	return thisCtx.stdoutWriter
}

func (thisCtx *baseCancelableContext) WithStdErrWriter(writer io.Writer) CancelableContext {
	thisCtx.stderrWriter = writer
	return thisCtx
}

func (thisCtx *baseCancelableContext) getStdErrWriter() io.Writer {
	return thisCtx.stderrWriter
}

func (thisCtx *baseCancelableContext) WithErrLogFn(errLogFn errLogFn) CancelableContext {
	thisCtx.errLogFn = errLogFn
	return thisCtx
}

func (thisCtx *baseCancelableContext) getErrLogFn() errLogFn {
	if thisCtx.errLogFn != nil {
		return thisCtx.errLogFn
	}

	return Fatal
}

func (thisCtx *baseCancelableContext) WithEnvVars(envVars ...string) CancelableContext {
	thisCtx.envVars = envVars
	return thisCtx
}

func (thisCtx *baseCancelableContext) getEnvVars() []string {
	return thisCtx.envVars
}

func (thisCtx *baseCancelableContext) WithReRun() CancelableContext {
	thisCtx.reRun = true
	return thisCtx
}

func (thisCtx *baseCancelableContext) isReRun() bool {
	return thisCtx.reRun
}

func (thisCtx *baseCancelableContext) WithAllowFailure(allowFailure bool) CancelableContext {
	thisCtx.allowFailure = allowFailure
	return thisCtx
}

func (thisCtx *baseCancelableContext) isAllowingFailure() bool {
	return thisCtx.allowFailure
}

func (thisCtx *baseCancelableContext) CancelAll() {
	thisCtx.cancelFn()
}

// checking this base implem does satisfy the interface above
var _ CancelableContext = (*baseCancelableContext)(nil)

// ----------------------------------------------------------------------------
// Aldev Context
// ----------------------------------------------------------------------------

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
	loopCtx    *baseCancelableContext // context used for the loop run by aldev
	exitWaitMs int                    // time waited right after cancelling the loop
	stopFn     func()                 // funtion called when the user stops the program
	children   []CancelableContext    // children contexts
	mx         sync.Mutex             // mutex to protect the children contexts
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
	aldevCtx.stopFn()
	aldevCtx.loopCtx.CancelAll()
	for _, childCtx := range aldevCtx.children {
		childCtx.CancelAll()
	}
	Info("Waiting for some cleanup...")
	time.Sleep(time.Duration(aldevCtx.exitWaitMs) * time.Millisecond) // TODO waiting
	aldevCtx.cancelFn()
}

func InitAldevContext(waitTimeMs int, stopFn func()) *aldevContext {
	stopFunction := stopFn
	if stopFn == nil {
		stopFunction = func() {}
	}

	// init
	aldevCtx := &aldevContext{
		baseCancelableContext: *newBaseCancelableContext(),
		loopCtx:               newBaseCancelableContext(),
		exitWaitMs:            waitTimeMs,
		stopFn:                stopFunction,
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

// ----------------------------------------------------------------------------
// Children contexts
// ----------------------------------------------------------------------------

func (aldevCtx *aldevContext) NewChildContext() CancelableContext {
	aldevCtx.mx.Lock()
	defer aldevCtx.mx.Unlock()
	childCtx := newBaseCancelableContext()
	aldevCtx.children = append(aldevCtx.children, childCtx)
	return childCtx
}

func (baseCancelableContext *baseCancelableContext) NewChildContext() CancelableContext {
	panic("No child context on children contexts allowed")
}
