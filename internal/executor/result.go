package executor

import (
	"time"

	"github.com/vmsfigueredo/gflow/internal/module"
)

type ResultStatus string

const (
	StatusOK     ResultStatus = "ok"
	StatusError  ResultStatus = "error"
	StatusDryRun ResultStatus = "dry-run"
	StatusSkip   ResultStatus = "skip"
)

// Result captures the outcome of an operation on one module.
type Result struct {
	Module   *module.Module
	Action   string
	Status   ResultStatus
	Output   string
	Err      error
	Duration time.Duration
}

func (r Result) OK() bool { return r.Status == StatusOK || r.Status == StatusDryRun }
