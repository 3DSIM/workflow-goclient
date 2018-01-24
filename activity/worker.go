package activity

import (
	"context"
	"fmt"
	"time"

	"github.com/3dsim/workflow-goclient/workflow"
	log "github.com/inconshreveable/log15"
)

// Log is a github.com/inconshreveable/log15.Logger.  Log is exposed so that users of this library can set
// their own log handler.  By default this Log uses the DiscardHandler, which discards log statements.
// See: https://godoc.org/github.com/inconshreveable/log15#hdr-Library_Use
//
// To set a different log handler do something like this:
//
// 		Log.SetHandler(log.LvlFilterHandler(log.LvlInfo, log.CallerFileHandler(log.StdoutHandler)))
var Log = log.New()

const (
	defaultHeartbeatInterval   = 1 * time.Minute
	defaultCancellationTimeout = 1 * time.Minute
	timeoutErrorMessage        = "Work cancelled after timeout"
)

func init() {
	Log.SetHandler(log.DiscardHandler())
}

// Worker handles executing work and reporting status and progress to the workflow API via the WorkflowClient field.
type Worker struct {
	WorkflowClient    workflow.Client
	HeartbeatInterval time.Duration
	// Time to wait for a cancellation before forcefully exiting.  If not set, default is 1 min
	CancellationTimeout time.Duration
}

// WorkerFunc is a function that can be passed into Worker.Do to do work.  It should
// listen for context cancellations and stop/cleanup/exit accordingly.  The channel given to the function should be used to
// report back percent complete as an integer (e.g. send 5 on the channel when operation is 5% complete).
type WorkerFunc func(ctx context.Context, percentCompleteChan chan<- int) (result interface{}, err error)

// Do executes the given function and reports back status and progress to the workflow API.  It takes
// care of heartbeating at the interval given by Worker.HeartbeatInterval or defaults to 1 min.
// If the given WorkflowFunc returns a non-nil error, then this will report a failure to the
// API.  Otherwise it will return a success back to the API.  If a heartbeat
// returns that a cancellation has been requested, then this function will handle closing
// the parent context and reporting the cancellation back to the workflow.  WorkflowFunc should
// listen for context closing and cleanup/exit accordingly.
func (w *Worker) Do(ctx context.Context, workflowID, activityID, taskToken string, f WorkerFunc) {
	workLog := Log.New("workflowID", workflowID, "activityID", activityID)
	pc := make(chan int)
	ec := make(chan error)
	rc := make(chan interface{})
	stop := make(chan struct{})
	childCtx, cancelFunc := context.WithCancel(ctx)

	go w.heartbeat(workLog, taskToken, activityID, cancelFunc, stop)
	go w.updatePercentComplete(workflowID, activityID, workLog, pc)

	go func() {
		result, err := f(childCtx, pc)
		if err != nil {
			// work has failed
			ec <- err
			return
		}
		// work has succeeded
		rc <- result
	}()

	select {
	case <-childCtx.Done():
		w.handleCancellation(workflowID, activityID, workLog, ec, rc)
	case err := <-ec:
		// Work has failed
		workLog.Info("Sending failure message to workflow API", "error", err)
		_, err = w.WorkflowClient.CompleteFailedActivity(workflowID, activityID, err.Error(), "")
		if err != nil {
			workLog.Error("Problem sending failure message", "error", err)
		}
	case result := <-rc:
		// Work has succeeded
		workLog.Info("Sending success message to workflow API", "result", result)
		_, err := w.WorkflowClient.CompleteSuccessfulActivity(workflowID, activityID, result)
		if err != nil {
			workLog.Error("Problem sending success message", "error", err)
		}
	}
	// Stop heartbeating
	stop <- struct{}{}
}

func (w *Worker) heartbeat(workLog log.Logger, taskToken, activityID string, cancelFunc context.CancelFunc, stop <-chan struct{}) {
	heartbeatInterval := defaultHeartbeatInterval
	if w.HeartbeatInterval > 0 {
		heartbeatInterval = w.HeartbeatInterval
	}
	heartbeats := time.NewTicker(heartbeatInterval)
	defer heartbeats.Stop()
	for {
		select {
		case <-heartbeats.C:
			workLog.Debug("Sending heartbeat")
			details := fmt.Sprintf("Heartbeat for activity %v", activityID)
			hb, err := w.WorkflowClient.HeartbeatActivityWithToken(taskToken, activityID, details)
			if err != nil {
				workLog.Error("Problem sending heartbeat", "error", err, "taskToken", taskToken)
			}
			if hb != nil && hb.Cancelled {
				workLog.Info("Cancellation requested via heartbeat")
				cancelFunc()
			}
		case <-stop:
			return
		}
	}

}

func (w *Worker) updatePercentComplete(workflowID, activityID string, workLog log.Logger, pc <-chan int) {
	lastPercentComplete := -1
	for percentComplete := range pc {
		if percentComplete != lastPercentComplete {
			_, err := w.WorkflowClient.UpdateActivityPercentComplete(workflowID, activityID, percentComplete)
			if err != nil {
				workLog.Error("Problem updating percent complete", "error", err, "percentComplete", percentComplete)
			}
			lastPercentComplete = percentComplete
		}
	}
}

func (w *Worker) handleCancellation(workflowID, activityID string, workLog log.Logger, ec <-chan error, rc <-chan interface{}) {
	workLog.Debug("Child context has been closed")
	cancellationTimeout := defaultCancellationTimeout
	if w.CancellationTimeout > 0 {
		cancellationTimeout = w.CancellationTimeout
	}
	select {
	case err := <-ec:
		_, err = w.WorkflowClient.CompleteCancelledActivity(workflowID, activityID, err.Error())
		if err != nil {
			workLog.Error("Problem sending completed via cancellation message", "error", err)
		}
	case <-rc:
		_, err := w.WorkflowClient.CompleteCancelledActivity(workflowID, activityID, "")
		if err != nil {
			workLog.Error("Problem sending completed via cancellation message", "error", err)
		}
	case <-time.After(cancellationTimeout):
		_, err := w.WorkflowClient.CompleteCancelledActivity(workflowID, activityID, timeoutErrorMessage)
		if err != nil {
			workLog.Error("Problem sending completed via cancellation message", "error", err)
		}
	}
}
