package executor

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vmsfigueredo/gflow/internal/module"
)

func makeMods(t testing.TB, n int) []*module.Module {
	out := make([]*module.Module, n)
	for i := range out {
		out[i] = &module.Module{Name: "mod", Path: t.TempDir()}
	}
	return out
}


func TestSerialRunsAll(t *testing.T) {
	runner := &serialRunner{}
	mods := makeMods(t, 3)
	results := runner.Run(context.Background(), mods, func(_ interface{}, m *module.Module) Result {
		return Result{Module: m, Action: "test", Status: StatusOK}
	})
	assert.Len(t, results, 3)
	for _, r := range results {
		assert.Equal(t, StatusOK, r.Status)
	}
}

func TestSerialFailFastStops(t *testing.T) {
	runner := &serialRunner{failFast: true}
	mods := makeMods(t, 5)
	var count int32
	results := runner.Run(context.Background(), mods, func(_ interface{}, m *module.Module) Result {
		atomic.AddInt32(&count, 1)
		if count == 2 {
			return Result{Module: m, Action: "test", Status: StatusError, Err: errors.New("boom")}
		}
		return Result{Module: m, Action: "test", Status: StatusOK}
	})
	assert.Equal(t, int32(2), count)
	assert.Len(t, results, 2)
}

func TestParallelRunsAll(t *testing.T) {
	runner := &parallelRunner{workers: 4}
	mods := makeMods(t, 8)
	results := runner.Run(context.Background(), mods, func(_ interface{}, m *module.Module) Result {
		time.Sleep(10 * time.Millisecond) // simulate work
		return Result{Module: m, Action: "test", Status: StatusOK}
	})
	assert.Len(t, results, 8)
	for _, r := range results {
		assert.Equal(t, StatusOK, r.Status)
	}
}

func TestParallelFasterThanSerial(t *testing.T) {
	mods := makeMods(t, 8)
	work := func(_ interface{}, m *module.Module) Result {
		time.Sleep(50 * time.Millisecond)
		return Result{Module: m, Action: "test", Status: StatusOK}
	}

	start := time.Now()
	(&serialRunner{}).Run(context.Background(), mods, work)
	serial := time.Since(start)

	start = time.Now()
	(&parallelRunner{workers: 8}).Run(context.Background(), mods, work)
	parallel := time.Since(start)

	assert.Less(t, parallel, serial/2, "parallel should be significantly faster than serial")
}

func TestResultDurationSet(t *testing.T) {
	runner := &serialRunner{}
	mods := makeMods(t, 1)
	results := runner.Run(context.Background(), mods, func(_ interface{}, m *module.Module) Result {
		time.Sleep(5 * time.Millisecond)
		return Result{Module: m, Action: "test", Status: StatusOK}
	})
	assert.Greater(t, results[0].Duration, time.Duration(0))
}
