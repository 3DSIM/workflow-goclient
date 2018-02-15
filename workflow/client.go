//go:generate counterfeiter ./ Client

package workflow

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/3dsim/auth0"
	"github.com/3dsim/workflow-goclient/genclient"
	"github.com/3dsim/workflow-goclient/genclient/operations"
	"github.com/3dsim/workflow-goclient/models"
	"github.com/PuerkitoBio/rehttp"
	openapiclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	log "github.com/inconshreveable/log15"
)

// Client is a wrapper around the generated client found in the "genclient" package.  It provides convenience methods
// for common operations.  If the operation needed is not found in Client, use the "genclient" package using this client
// as an example of how to utilize the genclient.  PRs are welcome if more functionality is wanted in this client package.
type Client interface {
	// StartWorkflow begins a new workflow and returns the workflow ID
	StartWorkflow(*models.PostWorkflow) (string, error)
	Workflow(workflowID string) (*models.Workflow, error)
	CancelWorkflow(workflowID string) error
	SignalWorkflow(workflowID string, signal *models.Signal) error
	UpdateActivity(workflowID string, activity *models.Activity) (*models.Activity, error)
	UpdateActivityPercentComplete(workflowID, activityID string, percentComplete int) (*models.Activity, error)
	CompleteSuccessfulActivity(workflowID, activityID string, result interface{}) (*models.Activity, error)
	CompleteCancelledActivity(workflowID, activityID, reason, details string) (*models.Activity, error)
	CompleteFailedActivity(workflowID, activityID, reason, details string) (*models.Activity, error)
	HeartbeatActivity(workflowID, activityID string) (*models.Heartbeat, error)
	HeartbeatActivityWithToken(taskToken, activityID, details string) (*models.Heartbeat, error)
}

type client struct {
	tokenFetcher auth0.TokenFetcher
	client       *genclient.Workflow
	audience     string
	logger       log.Logger
}

// NewClient creates a client for interacting with the 3DSIM workflow api.  See the auth0 package for how to construct
// the token fetcher.  The apiGatewayURL's are as follows:
//
// 		QA 				= https://3dsim-qa.cloud.tyk.io
//		Prod and Gov 	= https://3dsim.cloud.tyk.io
//
// The audience's are:
//
// 		QA 		= https://workflow-qa.3dsim.com
//		Prod 	= https://workflow.3dsim.com
// 		Gov 	= https://workflow-gov.3dsim.com
//
// The apiBasePath is "/workflow-api".
//
// logger is a github.com/inconshreveable/log15.Logger.  Log is exposed so that users of this client can pass
// their own log handler.  If nil is passed, this logger will be initialized to use the DiscardHandler, which discards log statements.
// See: https://godoc.org/github.com/inconshreveable/log15#hdr-Library_Use
//
func NewClient(tokenFetcher auth0.TokenFetcher, apiGatewayURL, apiBasePath, audience string, logger log.Logger) Client {
	return newClient(tokenFetcher, apiGatewayURL, apiBasePath, audience, nil, openapiclient.DefaultTimeout, logger)
}

// NewClientWithRetry creates the same type of client as NewClient, but allows for retrying any temporary errors or
// any responses with status >= 400 and < 600 for a specified amount of time.
//
// See NewClient for more information
func NewClientWithRetry(tokenFetcher auth0.TokenFetcher, apiGatewayURL, apiBasePath, audience string, retryTimeout time.Duration, logger log.Logger) Client {
	tr := rehttp.NewTransport(
		nil, // will use http.DefaultTransport
		rehttp.RetryAny(rehttp.RetryStatusInterval(400, 600), rehttp.RetryTemporaryErr()),
		rehttp.ExpJitterDelay(1*time.Second, retryTimeout),
	)
	return newClient(tokenFetcher, apiGatewayURL, apiBasePath, audience, tr, retryTimeout, logger)
}

func newClient(tokenFetcher auth0.TokenFetcher, apiGatewayURL, apiBasePath, audience string,
	roundTripper http.RoundTripper, defaultRequestTimeout time.Duration, logger log.Logger) Client {
	if logger == nil {
		logger = log.New()
		logger.SetHandler(log.DiscardHandler())
	}
	if roundTripper == nil {
		logger.Info("Creating workflow client with retry disabled")
	} else {
		logger.Info("Creating workflow client with retry enabled")
	}

	parsedURL, err := url.Parse(apiGatewayURL)
	if err != nil {
		message := "API Gateway URL was invalid!"
		logger.Error(message, "apiGatewayURL", apiGatewayURL)
		panic(message + " " + err.Error())
	}

	workflowTransport := openapiclient.New(parsedURL.Host, apiBasePath, []string{parsedURL.Scheme})
	if roundTripper != nil {
		workflowTransport.Transport = roundTripper
	}
	openapiclient.DefaultTimeout = defaultRequestTimeout
	workflowTransport.Debug = true
	workflowClient := genclient.New(workflowTransport, strfmt.Default)
	return &client{
		tokenFetcher: tokenFetcher,
		client:       workflowClient,
		audience:     audience,
		logger:       logger,
	}
}

// StartWorkflow creates a new workflow and returns the workflow ID
func (c *client) StartWorkflow(workflow *models.PostWorkflow) (workflowID string, err error) {
	token, err := c.tokenFetcher.Token(c.audience)
	if err != nil {
		return "", err
	}
	c.logger.Info("Starting workflow", "type", workflow.WorkflowType, "entityID", *workflow.EntityID)
	params := operations.NewStartWorkflowParams().WithWorkflow(workflow)
	response, err := c.client.Operations.StartWorkflow(params, openapiclient.BearerToken(token))
	if err != nil {
		c.logger.Error("Problem starting workflow", "type", workflow.WorkflowType, "entityID", *workflow.EntityID, "error", err)
		return "", err
	}
	return response.Payload, nil
}

func (c *client) Workflow(workflowID string) (workflow *models.Workflow, err error) {
	token, err := c.tokenFetcher.Token(c.audience)
	if err != nil {
		return nil, err
	}
	c.logger.Info("Getting workflow", "workflowID", workflowID)
	params := operations.NewGetWorkflowParams().WithID(workflowID)
	response, err := c.client.Operations.GetWorkflow(params, openapiclient.BearerToken(token))
	if err != nil {
		c.logger.Error("Problem getting workflow", "workflowID", workflowID, "error", err)
		return nil, err
	}
	return response.Payload, nil
}

func (c *client) CancelWorkflow(workflowID string) error {
	token, err := c.tokenFetcher.Token(c.audience)
	if err != nil {
		return err
	}
	c.logger.Info("Cancelling workflow", "workflowID", workflowID)
	params := operations.NewCancelWorkflowParams().WithID(workflowID)
	_, err = c.client.Operations.CancelWorkflow(params, openapiclient.BearerToken(token))
	if err != nil {
		c.logger.Error("Problem cancelling workflow", "workflowID", workflowID, "error", err)
		return err
	}
	return nil
}

func (c *client) SignalWorkflow(workflowID string, signal *models.Signal) error {
	token, err := c.tokenFetcher.Token(c.audience)
	if err != nil {
		return err
	}
	c.logger.Info("Signaling workflow", "workflowID", workflowID)
	params := operations.NewSignalWorkflowParams().WithID(workflowID).WithSignal(signal)
	_, err = c.client.Operations.SignalWorkflow(params, openapiclient.BearerToken(token))
	if err != nil {
		c.logger.Error("Problem signaling workflow", "workflowID", workflowID, "error", err)
		return err
	}
	return nil
}

func (c *client) UpdateActivity(workflowID string, activity *models.Activity) (*models.Activity, error) {
	token, err := c.tokenFetcher.Token(c.audience)
	if err != nil {
		return nil, err
	}
	c.logger.Info("Updating activity", "workflowID", workflowID, "activityID", *activity.ID)
	params := operations.NewUpdateActivityParams().WithID(workflowID).WithActivityID(*activity.ID).WithActivity(activity)
	response, err := c.client.Operations.UpdateActivity(params, openapiclient.BearerToken(token))
	if err != nil {
		c.logger.Error("Problem updating activity", "workflowID", workflowID, "activityID", *activity.ID, "error", err)
		return nil, err
	}
	return response.Payload, nil
}

func (c *client) UpdateActivityPercentComplete(workflowID, activityID string, percentComplete int) (*models.Activity, error) {
	token, err := c.tokenFetcher.Token(c.audience)
	if err != nil {
		return nil, err
	}
	updatedActivity := &models.Activity{
		ID:              swag.String(activityID),
		Status:          swag.String(models.ActivityStatusRunning),
		PercentComplete: int32(percentComplete),
	}
	c.logger.Info("Updating activity percent complete", "workflowID", workflowID, "activityID", activityID, "percentComplete", percentComplete)
	params := operations.NewUpdateActivityParams().WithID(workflowID).WithActivityID(activityID).WithActivity(updatedActivity)
	response, err := c.client.Operations.UpdateActivity(params, openapiclient.BearerToken(token))
	if err != nil {
		c.logger.Error("Problem updating activity percent complete", "workflowID", workflowID, "activityID", activityID, "error", err)
		return nil, err
	}
	return response.Payload, nil
}

func (c *client) CompleteSuccessfulActivity(workflowID, activityID string, result interface{}) (*models.Activity, error) {
	token, err := c.tokenFetcher.Token(c.audience)
	if err != nil {
		return nil, err
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	completedActivity := &models.Activity{
		ID:              swag.String(activityID),
		Status:          swag.String(models.ActivityStatusCompleted),
		Result:          string(resultBytes),
		PercentComplete: 100,
	}
	c.logger.Info("Completing successful activity", "workflowID", workflowID, "activityID", activityID, "result", result)
	params := operations.NewUpdateActivityParams().WithID(workflowID).WithActivityID(activityID).WithActivity(completedActivity)
	response, err := c.client.Operations.UpdateActivity(params, openapiclient.BearerToken(token))
	if err != nil {
		c.logger.Error("Problem completing successful activity", "workflowID", workflowID, "activityID", activityID, "error", err)
		return nil, err
	}
	return response.Payload, nil
}

// CompleteCancelledActivity will sent an activity with a cancelled status to the workflow API.  workflowID, activityID,
// and reason are required.
func (c *client) CompleteCancelledActivity(workflowID, activityID, reason, details string) (*models.Activity, error) {
	token, err := c.tokenFetcher.Token(c.audience)
	if err != nil {
		return nil, err
	}
	cancelledActivity := &models.Activity{
		ID:     swag.String(activityID),
		Status: swag.String(models.ActivityStatusCancelled),
		Error:  &models.ActivityError{Reason: swag.String(reason), Details: details},
	}
	c.logger.Info("Completing cancelled activity", "workflowID", workflowID, "activityID", activityID)
	params := operations.NewUpdateActivityParams().WithID(workflowID).WithActivityID(activityID).WithActivity(cancelledActivity)
	response, err := c.client.Operations.UpdateActivity(params, openapiclient.BearerToken(token))
	if err != nil {
		c.logger.Error("Problem completing cancelled activity", "workflowID", workflowID, "activityID", activityID, "error", err)
		return nil, err
	}
	return response.Payload, nil
}

func (c *client) CompleteFailedActivity(workflowID, activityID, reason, details string) (*models.Activity, error) {
	token, err := c.tokenFetcher.Token(c.audience)
	if err != nil {
		return nil, err
	}
	failedActivity := &models.Activity{
		ID:     swag.String(activityID),
		Status: swag.String(models.ActivityStatusFailed),
		Error:  &models.ActivityError{Reason: swag.String(reason), Details: details},
	}
	c.logger.Info("Completing failed activity", "workflowID", workflowID, "activityID", activityID)
	params := operations.NewUpdateActivityParams().WithID(workflowID).WithActivityID(activityID).WithActivity(failedActivity)
	response, err := c.client.Operations.UpdateActivity(params, openapiclient.BearerToken(token))
	if err != nil {
		c.logger.Error("Problem completing failed activity", "workflowID", workflowID, "activityID", activityID, "error", err)
		return nil, err
	}
	return response.Payload, nil
}

func (c *client) HeartbeatActivity(workflowID string, activityID string) (*models.Heartbeat, error) {
	token, err := c.tokenFetcher.Token(c.audience)
	if err != nil {
		return nil, err
	}
	c.logger.Debug("Heartbeating activity", "workflowID", workflowID, "activityID", activityID)
	params := operations.NewHeartbeatActivityParams().WithID(workflowID).WithActivityID(activityID)
	response, err := c.client.Operations.HeartbeatActivity(params, openapiclient.BearerToken(token))
	if err != nil {
		c.logger.Error("Problem heartbeating activity", "workflowID", workflowID, "activityID", activityID, "error", err)
		return nil, err
	}
	return response.Payload, nil
}

func (c *client) HeartbeatActivityWithToken(taskToken, activityID, details string) (*models.Heartbeat, error) {
	token, err := c.tokenFetcher.Token(c.audience)
	if err != nil {
		return nil, err
	}
	heartbeat := &models.Heartbeat{
		TaskToken:  swag.String(taskToken),
		ActivityID: swag.String(activityID),
		Details:    details,
		Cancelled:  false,
	}
	c.logger.Debug("Heartbeating activity", "token", taskToken)
	params := operations.NewHeartbeatParams().WithHeartbeat(heartbeat)
	response, err := c.client.Operations.Heartbeat(params, openapiclient.BearerToken(token))
	if err != nil {
		c.logger.Error("Problem heartbeating activity", "token", taskToken, "error", err)
		return nil, err
	}
	return response.Payload, nil
}
