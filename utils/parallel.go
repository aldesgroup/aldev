// ----------------------------------------------------------------------------
// The code here is about running stuff in parallel
// ----------------------------------------------------------------------------
package utils

import "sync"

// characterises the functions run during an aldev command or subcommand
type AldevTask func(ctx CancelableContext, cfg *AldevConfig)

func goRoutine(wg *sync.WaitGroup, task AldevTask, ctx CancelableContext, config *AldevConfig) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		task(ctx, config)
	}()
}
