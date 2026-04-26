package cmd

import (
	"context"
	"sync"
)

var (
	runContextMu sync.RWMutex
	runContext   = context.Background()
)

func SetRunContext(ctx context.Context) {
	runContextMu.Lock()
	defer runContextMu.Unlock()
	if ctx == nil {
		runContext = context.Background()
		return
	}
	runContext = ctx
}

func RunContext() context.Context {
	runContextMu.RLock()
	defer runContextMu.RUnlock()
	return runContext
}
