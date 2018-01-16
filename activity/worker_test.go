package activity

import (
	"context"
	"errors"
	"github.com/3dsim/workflow-goclient/models"
	"github.com/3dsim/workflow-goclient/workflow/workflowfakes"
	"github.com/go-openapi/swag"
	log "github.com/inconshreveable/log15"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func init() {
	Log.SetHandler(log.LvlFilterHandler(log.LvlDebug, log.CallerFileHandler(log.StdoutHandler)))
}

func TestDoExpectsCompleteFailedActivityCalledWhenErrorOccurs(t *testing.T) {
	// arrange
	fakeWorkflowClient := &workflowfakes.FakeClient{}
	worker := &Worker{WorkflowClient: fakeWorkflowClient}
	activityID := "activity id"
	workflowID := "workflow id"
	taskToken := "token"
	errorReason := "Some error"

	// act
	worker.Do(context.Background(), workflowID, activityID, taskToken, func(context.Context, chan<- int) (interface{}, error) {
		return nil, errors.New(errorReason)
	})

	// assert
	assert.Equal(t, 1, fakeWorkflowClient.CompleteFailedActivityCallCount(), "Expected to call CompleteFailedActivity once")
	actualWorkflowID, actualActivityID, actualErrorReason, actualErrorDetails := fakeWorkflowClient.CompleteFailedActivityArgsForCall(0)
	assert.Equal(t, workflowID, actualWorkflowID, "Expected workflow ID passed to CompleteFailedActivity")
	assert.Equal(t, activityID, actualActivityID, "Expected activity ID passed to CompleteFailedActivity")
	assert.Equal(t, errorReason, actualErrorReason, "Expected error reason passed to CompleteFailedActivity")
	assert.Equal(t, "", actualErrorDetails, "Expected error details passed to CompleteFailedActivity")
}

func TestDoExpectsCompleteSuccessfulActivityCalledWhenNoErrorOccurs(t *testing.T) {
	// arrange
	fakeWorkflowClient := &workflowfakes.FakeClient{}
	worker := &Worker{WorkflowClient: fakeWorkflowClient}
	activityID := "activity id"
	workflowID := "workflow id"
	taskToken := "token"
	result := struct{ SomeField string }{"the result"}

	// act
	worker.Do(context.Background(), workflowID, activityID, taskToken, func(context.Context, chan<- int) (interface{}, error) {
		return result, nil
	})

	// assert
	assert.Equal(t, 1, fakeWorkflowClient.CompleteSuccessfulActivityCallCount(), "Expected to call CompleteSuccessfulActivity once")
	actualWorkflowID, actualActivityID, actualResult := fakeWorkflowClient.CompleteSuccessfulActivityArgsForCall(0)
	assert.Equal(t, workflowID, actualWorkflowID, "Expected workflow ID passed to CompleteSuccessfulActivity")
	assert.Equal(t, activityID, actualActivityID, "Expected activity ID passed to CompleteSuccessfulActivity")
	assert.Equal(t, result, actualResult, "Expected result passed to CompleteSuccessfulActivity")
}

func TestDoExpectsHeartbeatActivityWithTokenCalled(t *testing.T) {
	// arrange
	fakeWorkflowClient := &workflowfakes.FakeClient{}
	worker := &Worker{WorkflowClient: fakeWorkflowClient, HeartbeatInterval: 7 * time.Millisecond}
	activityID := "activity id"
	workflowID := "workflow id"
	taskToken := "token"

	// act
	worker.Do(context.Background(), workflowID, activityID, taskToken, func(context.Context, chan<- int) (interface{}, error) {
		// Wait a little time for heartbeat
		time.Sleep(10 * time.Millisecond)
		return nil, nil
	})

	// assert
	assert.Equal(t, 1, fakeWorkflowClient.HeartbeatActivityWithTokenCallCount(), "Expected to call HeartbeatActivityWithToken once")
	actualTaskToken := fakeWorkflowClient.HeartbeatActivityWithTokenArgsForCall(0)
	assert.Equal(t, taskToken, actualTaskToken, "Expected task token passed to HeartbeatActivityWithToken")
}

func TestDoWhenCancellationRequestedExpectsCompleteCancelledActivityCalled(t *testing.T) {
	// arrange
	fakeWorkflowClient := &workflowfakes.FakeClient{}
	worker := &Worker{WorkflowClient: fakeWorkflowClient, HeartbeatInterval: 7 * time.Millisecond}
	activityID := "activity id"
	workflowID := "workflow id"
	taskToken := "token"

	heartbeatToReturn := &models.Heartbeat{
		ActivityID: swag.String(activityID),
		Cancelled:  true,
	}
	fakeWorkflowClient.HeartbeatActivityWithTokenReturns(heartbeatToReturn, nil)

	// act
	worker.Do(context.Background(), workflowID, activityID, taskToken, func(ctx context.Context, percentCompleteChan chan<- int) (interface{}, error) {
		select {
		case <-ctx.Done():
		case <-time.After(30 * time.Millisecond):
			t.Error("Did not receive the cancellation in time")
		}
		return nil, nil
	})

	// assert
	assert.True(t, fakeWorkflowClient.HeartbeatActivityWithTokenCallCount() >= 1, "Expected to call HeartbeatActivityWithToken at least once")
	actualTaskToken := fakeWorkflowClient.HeartbeatActivityWithTokenArgsForCall(0)
	assert.Equal(t, taskToken, actualTaskToken, "Expected task token passed to HeartbeatActivityWithToken")
	assert.Equal(t, 1, fakeWorkflowClient.CompleteCancelledActivityCallCount(), "Expected to call CompleteCancelledActivity once")
	actualWorkflowID, actualActivityID, actualDetails := fakeWorkflowClient.CompleteCancelledActivityArgsForCall(0)
	assert.Equal(t, workflowID, actualWorkflowID, "Expected workflow ID passed to CompleteCancelledActivity")
	assert.Equal(t, activityID, actualActivityID, "Expected activity ID passed to CompleteCancelledActivity")
	assert.Equal(t, "", actualDetails, "Expected to pass empty details")
}

func TestDoWhenCancellationRequestedAndFunctionErrorsExpectsCompleteCancelledActivityCalled(t *testing.T) {
	// arrange
	fakeWorkflowClient := &workflowfakes.FakeClient{}
	worker := &Worker{WorkflowClient: fakeWorkflowClient, HeartbeatInterval: 7 * time.Millisecond}
	activityID := "activity id"
	workflowID := "workflow id"
	taskToken := "token"
	errMsg := "Some cancellation error"
	heartbeatToReturn := &models.Heartbeat{
		ActivityID: swag.String(activityID),
		Cancelled:  true,
	}
	fakeWorkflowClient.HeartbeatActivityWithTokenReturns(heartbeatToReturn, nil)

	// act
	worker.Do(context.Background(), workflowID, activityID, taskToken, func(ctx context.Context, percentCompleteChan chan<- int) (interface{}, error) {
		select {
		case <-ctx.Done():
		case <-time.After(30 * time.Millisecond):
			t.Error("Did not receive the cancellation in time")
		}
		return nil, errors.New(errMsg)
	})

	// assert
	assert.True(t, fakeWorkflowClient.HeartbeatActivityWithTokenCallCount() >= 1, "Expected to call HeartbeatActivityWithToken at least once")
	actualTaskToken := fakeWorkflowClient.HeartbeatActivityWithTokenArgsForCall(0)
	assert.Equal(t, taskToken, actualTaskToken, "Expected task token passed to HeartbeatActivityWithToken")
	assert.Equal(t, 1, fakeWorkflowClient.CompleteCancelledActivityCallCount(), "Expected to call CompleteCancelledActivity once")
	actualWorkflowID, actualActivityID, actualDetails := fakeWorkflowClient.CompleteCancelledActivityArgsForCall(0)
	assert.Equal(t, workflowID, actualWorkflowID, "Expected workflow ID passed to CompleteCancelledActivity")
	assert.Equal(t, activityID, actualActivityID, "Expected activity ID passed to CompleteCancelledActivity")
	assert.Equal(t, errMsg, actualDetails, "Expected to error details")
}

func TestDoWhenCancellationRequestedAndFunctionBlocksForeverExpectsCompleteCancelledActivityCalledAfterTimeout(t *testing.T) {
	// arrange
	fakeWorkflowClient := &workflowfakes.FakeClient{}
	worker := &Worker{
		WorkflowClient:      fakeWorkflowClient,
		HeartbeatInterval:   7 * time.Millisecond,
		CancellationTimeout: 10 * time.Millisecond,
	}
	activityID := "activity id"
	workflowID := "workflow id"
	taskToken := "token"
	heartbeatToReturn := &models.Heartbeat{
		ActivityID: swag.String(activityID),
		Cancelled:  true,
	}
	fakeWorkflowClient.HeartbeatActivityWithTokenReturns(heartbeatToReturn, nil)

	// act
	worker.Do(context.Background(), workflowID, activityID, taskToken, func(ctx context.Context, percentCompleteChan chan<- int) (interface{}, error) {
		<-ctx.Done()
		time.Sleep(30 * time.Millisecond)
		return nil, errors.New("Unexpected error")
	})

	// assert
	assert.True(t, fakeWorkflowClient.HeartbeatActivityWithTokenCallCount() >= 1, "Expected to call HeartbeatActivityWithToken at least once")
	actualTaskToken := fakeWorkflowClient.HeartbeatActivityWithTokenArgsForCall(0)
	assert.Equal(t, taskToken, actualTaskToken, "Expected task token passed to HeartbeatActivityWithToken")
	assert.Equal(t, 1, fakeWorkflowClient.CompleteCancelledActivityCallCount(), "Expected to call CompleteCancelledActivity once")
	actualWorkflowID, actualActivityID, actualDetails := fakeWorkflowClient.CompleteCancelledActivityArgsForCall(0)
	assert.Equal(t, workflowID, actualWorkflowID, "Expected workflow ID passed to CompleteCancelledActivity")
	assert.Equal(t, activityID, actualActivityID, "Expected activity ID passed to CompleteCancelledActivity")
	assert.Equal(t, timeoutErrorMessage, actualDetails, "Expected to pass empty details")
}

func TestDoExpectsUpdateActivityPercentCompleteCalledWhenProgressIsMade(t *testing.T) {
	// arrange
	fakeWorkflowClient := &workflowfakes.FakeClient{}
	worker := &Worker{WorkflowClient: fakeWorkflowClient}
	activityID := "activity id"
	workflowID := "workflow id"
	taskToken := "token"

	// act
	worker.Do(context.Background(), workflowID, activityID, taskToken, func(ctx context.Context, percentCompleteChan chan<- int) (interface{}, error) {
		percentCompleteChan <- 30
		percentCompleteChan <- 60
		percentCompleteChan <- 100
		return nil, nil
	})

	// assert
	assert.Equal(t, 3, fakeWorkflowClient.UpdateActivityPercentCompleteCallCount(), "Expected to call UpdateActivityPercentComplete once")
	actualWorkflowID, actualActivityID, actualPercentComplete := fakeWorkflowClient.UpdateActivityPercentCompleteArgsForCall(0)
	assert.Equal(t, workflowID, actualWorkflowID, "Expected workflow ID passed to UpdateActivityPercentComplete")
	assert.Equal(t, activityID, actualActivityID, "Expected activity ID passed to UpdateActivityPercentComplete")
	assert.Equal(t, 30, actualPercentComplete, "Expected percent complete passed to UpdateActivityPercentComplete")
}

func TestDoExpectsUpdateActivityPercentCompleteCalledOnceWhenSameValuesAreSentConsecutively(t *testing.T) {
	// arrange
	fakeWorkflowClient := &workflowfakes.FakeClient{}
	worker := &Worker{WorkflowClient: fakeWorkflowClient}
	activityID := "activity id"
	workflowID := "workflow id"
	taskToken := "token"

	// act
	worker.Do(context.Background(), workflowID, activityID, taskToken, func(ctx context.Context, percentCompleteChan chan<- int) (interface{}, error) {
		percentCompleteChan <- 30
		percentCompleteChan <- 30
		percentCompleteChan <- 30
		return nil, nil
	})

	// assert
	assert.Equal(t, 1, fakeWorkflowClient.UpdateActivityPercentCompleteCallCount(), "Expected to call UpdateActivityPercentComplete once")
	actualWorkflowID, actualActivityID, actualPercentComplete := fakeWorkflowClient.UpdateActivityPercentCompleteArgsForCall(0)
	assert.Equal(t, workflowID, actualWorkflowID, "Expected workflow ID passed to UpdateActivityPercentComplete")
	assert.Equal(t, activityID, actualActivityID, "Expected activity ID passed to UpdateActivityPercentComplete")
	assert.Equal(t, 30, actualPercentComplete, "Expected percent complete passed to UpdateActivityPercentComplete")
}
