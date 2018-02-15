package workflow

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/3dsim/auth0/auth0fakes"
	"github.com/3dsim/workflow-goclient/models"
	"github.com/go-openapi/swag"
	"github.com/gorilla/mux"
	log "github.com/inconshreveable/log15"
	"github.com/stretchr/testify/assert"
)

const (
	gatewayURL          = "apiGatewayURL"
	workflowAPIBasePath = "workflow-api"
	audience            = "test audience"
)

var logger log.Logger

func init() {
	logger = log.New()
	logger.SetHandler(log.LvlFilterHandler(log.LvlDebug, log.CallerFileHandler(log.StdoutHandler)))
}

func TestNewClientExpectsClientReturned(t *testing.T) {
	// arrange
	// act
	client := NewClient(nil, gatewayURL, workflowAPIBasePath, audience, logger)

	// assert
	assert.NotNil(t, client, "Expected new client to not be nil")
}

func TestWorkflow(t *testing.T) {
	// arrange
	workflowID := "my-workflow"
	endpoint := "/" + workflowAPIBasePath + "/workflows/{workflowID}"

	t.Run("WhenSuccessfulExpectsWorkflowReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("token", nil)
		workflowToReturn := &models.Workflow{
			ID:                workflowID,
			State:             "Running",
			WaitingOnCapacity: false,
		}
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			assert.NotEmpty(t, r.Header.Get("Authorization"), "Authorization header should not be empty")
			receivedWorkflowID := mux.Vars(r)["workflowID"]
			assert.EqualValues(t, workflowID, receivedWorkflowID, "Expected workflow id received to match what was passed in")
			bytes, err := json.Marshal(workflowToReturn)
			if err != nil {
				t.Error("Failed to marshal workflow")
			}
			w.Write(bytes)
		})

		// Setup routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		workflow, err := client.Workflow(workflowID)

		// assert
		assert.Nil(t, err, "Expected error to be nil when getting workflow")
		assert.NotNil(t, workflow, "Expected retrieved workflow to not be nil")
		assert.Equal(t, workflowToReturn.ID, workflow.ID, "Expected workflow IDs to match")
		assert.Equal(t, workflowToReturn.State, workflow.State, "Expected workflow states to match")
		assert.Equal(t, workflowToReturn.WaitingOnCapacity, workflow.WaitingOnCapacity, "Expected workflow 'waiting on capacity' to match")
	})

	t.Run("WhenFetcherErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		expectedError := errors.New("Some auth0 error")
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("", expectedError)
		client := NewClient(fakeTokenFetcher, gatewayURL, workflowAPIBasePath, audience, logger)

		// act
		workflow, err := client.Workflow(workflowID)

		// assert
		assert.Nil(t, workflow, "Expected no workflow to be returned due to token error")
		assert.Equal(t, expectedError, err, "Expected an error returned")

	})

	t.Run("WhenAPIErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("Token", nil)

		// return server error from http handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		})

		// set up routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		workflow, err := client.Workflow(workflowID)

		// assert
		assert.Nil(t, workflow, "Expected no workflow to be returned due to API error")
		assert.NotNil(t, err, "Expected an error returned because workflow API sent a 500 error")
	})
}

func TestCancelWorkflow(t *testing.T) {
	// arrange
	workflowID := "my-workflow"
	endpoint := "/" + workflowAPIBasePath + "/workflows/{workflowID}/cancel"

	t.Run("WhenSuccessfulExpectsNothingReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("token", nil)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			assert.NotEmpty(t, r.Header.Get("Authorization"), "Authorization header should not be empty")
			receivedWorkflowID := mux.Vars(r)["workflowID"]
			assert.EqualValues(t, workflowID, receivedWorkflowID, "Expected workflow id received to match what was passed in")
		})

		// Setup routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		err := client.CancelWorkflow(workflowID)

		// assert
		assert.Nil(t, err, "Expected error to be nil when cancelling workflow")
	})

	t.Run("WhenFetcherErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		expectedError := errors.New("Some auth0 error")
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("", expectedError)
		client := NewClient(fakeTokenFetcher, gatewayURL, workflowAPIBasePath, audience, logger)

		// act
		err := client.CancelWorkflow(workflowID)

		// assert
		assert.Equal(t, expectedError, err, "Expected an error returned")

	})

	t.Run("WhenAPIErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("Token", nil)

		// return server error from http handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		})

		// set up routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		err := client.CancelWorkflow(workflowID)

		// assert
		assert.NotNil(t, err, "Expected an error returned because workflow API sent a 500 error")
	})
}

func TestSignalWorkflow(t *testing.T) {
	// arrange
	workflowID := "my-workflow"
	endpoint := "/" + workflowAPIBasePath + "/workflows/{workflowID}/signals"

	t.Run("WhenSuccessfulExpectsNothingReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("token", nil)
		signal := &models.Signal{
			Name:  swag.String("SignalName"),
			Input: "inputjson",
		}
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			assert.NotEmpty(t, r.Header.Get("Authorization"), "Authorization header should not be empty")
			receivedWorkflowID := mux.Vars(r)["workflowID"]
			assert.EqualValues(t, workflowID, receivedWorkflowID, "Expected workflow id received to match what was passed in")
			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			var actualSignal models.Signal
			err = json.Unmarshal(bodyBytes, &actualSignal)
			if err != nil {
				t.Error("Failed to unmarshal signal")
			}
			assert.EqualValues(t, *signal, actualSignal, "Expected signal to be passed in body of request")
		})

		// Setup routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		err := client.SignalWorkflow(workflowID, signal)

		// assert
		assert.Nil(t, err, "Expected error to be nil when signalling workflow")
	})

	t.Run("WhenFetcherErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		expectedError := errors.New("Some auth0 error")
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("", expectedError)
		client := NewClient(fakeTokenFetcher, gatewayURL, workflowAPIBasePath, audience, logger)

		// act
		err := client.SignalWorkflow(workflowID, nil)

		// assert
		assert.Equal(t, expectedError, err, "Expected an error returned")

	})

	t.Run("WhenAPIErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("Token", nil)

		// return server error from http handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		})

		// set up routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		err := client.SignalWorkflow(workflowID, &models.Signal{})

		// assert
		assert.NotNil(t, err, "Expected an error returned because workflow API sent a 500 error")
	})
}

func TestUpdateActivity(t *testing.T) {
	// arrange
	workflowID := "my-workflow"
	activityID := "my-activity"
	endpoint := "/" + workflowAPIBasePath + "/workflows/{workflowID}/activities/{activityID}"
	activityToReturn := &models.Activity{
		ID:              swag.String(activityID),
		Status:          swag.String("Completed"),
		PercentComplete: 100,
		Result:          "json-string",
		Error: &models.ActivityError{
			Reason:  swag.String("reason"),
			Details: "details",
		},
	}

	t.Run("WhenSuccessfulExpectsActivityReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("token", nil)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			assert.NotEmpty(t, r.Header.Get("Authorization"), "Authorization header should not be empty")
			receivedWorkflowID := mux.Vars(r)["workflowID"]
			receivedActivityID := mux.Vars(r)["activityID"]
			assert.EqualValues(t, workflowID, receivedWorkflowID, "Expected workflow id received to match what was passed in")
			assert.EqualValues(t, activityID, receivedActivityID, "Expected activity id received to match what was passed in")
			bytes, err := json.Marshal(activityToReturn)
			if err != nil {
				t.Error("Failed to marshal activity")
			}
			w.Write(bytes)
		})

		// Setup routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		activity, err := client.UpdateActivity(workflowID, activityToReturn)

		// assert
		assert.Nil(t, err, "Expected error to be nil when getting workflow")
		assert.NotNil(t, activity, "Expected retrieved activity to not be nil")
		assert.Equal(t, *activityToReturn.ID, *activity.ID, "Expected activity IDs to match")
		assert.Equal(t, *activityToReturn.Status, *activity.Status, "Expected activity status to match")
		assert.Equal(t, activityToReturn.PercentComplete, activity.PercentComplete, "Expected activity percent complete to match")
		assert.Equal(t, activityToReturn.Result, activity.Result, "Expected activity result to match")
		assert.NotNil(t, activity.Error, "Expected activity error to not be nil")
		assert.Equal(t, *activityToReturn.Error.Reason, *activity.Error.Reason, "Expected activity error reason to not be nil")
		assert.Equal(t, activityToReturn.Error.Details, activity.Error.Details, "Expected activity error details to not be nil")
	})

	t.Run("WhenFetcherErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		expectedError := errors.New("Some auth0 error")
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("", expectedError)
		client := NewClient(fakeTokenFetcher, gatewayURL, workflowAPIBasePath, audience, logger)

		// act
		activity, err := client.UpdateActivity(workflowID, activityToReturn)

		// assert
		assert.Nil(t, activity, "Expected no activity to be returned due to token error")
		assert.Equal(t, expectedError, err, "Expected an error returned")

	})

	t.Run("WhenAPIErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("Token", nil)

		// return server error from http handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		})

		// set up routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		activity, err := client.UpdateActivity(workflowID, activityToReturn)

		// assert
		assert.Nil(t, activity, "Expected no activity to be returned due to API error")
		assert.NotNil(t, err, "Expected an error returned because workflow API sent a 500 error")
	})
}

func TestUpdateActivityPercentComplete(t *testing.T) {
	// arrange
	workflowID := "my-workflow"
	activityID := "my-activity"
	endpoint := "/" + workflowAPIBasePath + "/workflows/{workflowID}/activities/{activityID}"

	t.Run("WhenSuccessfulExpectsUpdatedActivityInRequest", func(t *testing.T) {
		// arrange
		expectedActivity := &models.Activity{
			ID:              swag.String(activityID),
			Status:          swag.String(models.ActivityStatusRunning),
			PercentComplete: 32,
		}
		var actualActivity models.Activity
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("token", nil)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			assert.NotEmpty(t, r.Header.Get("Authorization"), "Authorization header should not be empty")
			receivedWorkflowID := mux.Vars(r)["workflowID"]
			receivedActivityID := mux.Vars(r)["activityID"]
			assert.EqualValues(t, workflowID, receivedWorkflowID, "Expected workflow id received to match what was passed in")
			assert.EqualValues(t, activityID, receivedActivityID, "Expected activity id received to match what was passed in")
			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			err = json.Unmarshal(bodyBytes, &actualActivity)
			if err != nil {
				t.Fatal(err)
			}
			bytes, err := json.Marshal(&models.Activity{})
			if err != nil {
				t.Fatal("Failed to marshal activity " + err.Error())
			}
			w.Write(bytes)
		})

		// Setup routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		activity, err := client.UpdateActivityPercentComplete(workflowID, activityID, int(expectedActivity.PercentComplete))

		// assert
		assert.Equal(t, *expectedActivity.ID, *actualActivity.ID, "Expected activity IDs to match")
		assert.Equal(t, models.ActivityStatusRunning, *actualActivity.Status, "Expected activity status to be: "+models.ActivityStatusRunning)
		assert.EqualValues(t, expectedActivity.PercentComplete, actualActivity.PercentComplete, "Expected percent complete to match what was passed in")
		assert.Nil(t, err, "Expected no error")
		assert.Nil(t, actualActivity.Error, "Expected no activity error")
		assert.NotNil(t, activity, "Expected retrieved activity to not be nil")
	})

	t.Run("WhenFetcherErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		expectedError := errors.New("Some auth0 error")
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("", expectedError)
		client := NewClient(fakeTokenFetcher, gatewayURL, workflowAPIBasePath, audience, logger)

		// act
		activity, err := client.UpdateActivityPercentComplete(workflowID, activityID, 0)

		// assert
		assert.Nil(t, activity, "Expected no activity to be returned due to token error")
		assert.Equal(t, expectedError, err, "Expected an error returned")

	})

	t.Run("WhenAPIErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("Token", nil)

		// return server error from http handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		})

		// set up routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		activity, err := client.UpdateActivityPercentComplete(workflowID, activityID, 0)

		// assert
		assert.Nil(t, activity, "Expected no activity to be returned due to API error")
		assert.NotNil(t, err, "Expected an error returned because workflow API sent a 500 error")
	})
}

func TestCompleteSuccessfulActivity(t *testing.T) {
	// arrange
	workflowID := "my-workflow"
	activityID := "my-activity"
	endpoint := "/" + workflowAPIBasePath + "/workflows/{workflowID}/activities/{activityID}"

	t.Run("WhenSuccessfulExpectsCompletedActivityInRequest", func(t *testing.T) {
		// arrange
		result := struct{ Foo string }{Foo: "Bar"}
		resultBytes, err := json.Marshal(result)
		if err != nil {
			t.Fatal(err)
		}
		expectedActivity := &models.Activity{
			ID:              swag.String(activityID),
			Status:          swag.String(models.ActivityStatusCompleted),
			PercentComplete: 100,
			Result:          string(resultBytes),
		}
		var actualActivity models.Activity
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("token", nil)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			assert.NotEmpty(t, r.Header.Get("Authorization"), "Authorization header should not be empty")
			receivedWorkflowID := mux.Vars(r)["workflowID"]
			receivedActivityID := mux.Vars(r)["activityID"]
			assert.EqualValues(t, workflowID, receivedWorkflowID, "Expected workflow id received to match what was passed in")
			assert.EqualValues(t, activityID, receivedActivityID, "Expected activity id received to match what was passed in")
			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			err = json.Unmarshal(bodyBytes, &actualActivity)
			if err != nil {
				t.Fatal(err)
			}
			bytes, err := json.Marshal(&models.Activity{})
			if err != nil {
				t.Fatal("Failed to marshal activity " + err.Error())
			}
			w.Write(bytes)
		})

		// Setup routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		activity, err := client.CompleteSuccessfulActivity(workflowID, activityID, result)

		// assert
		assert.Equal(t, *expectedActivity.ID, *actualActivity.ID, "Expected activity IDs to match")
		assert.Equal(t, models.ActivityStatusCompleted, *actualActivity.Status, "Expected activity status to be: "+models.ActivityStatusCompleted)
		assert.Equal(t, expectedActivity.Result, actualActivity.Result, "Expected activity result to match")
		assert.EqualValues(t, 100, actualActivity.PercentComplete, "Expected percent complete to be 100")
		assert.Nil(t, err, "Expected no error")
		assert.Nil(t, actualActivity.Error, "Expected no activity error")
		assert.NotNil(t, activity, "Expected retrieved activity to not be nil")
	})

	t.Run("WhenFetcherErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		expectedError := errors.New("Some auth0 error")
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("", expectedError)
		client := NewClient(fakeTokenFetcher, gatewayURL, workflowAPIBasePath, audience, logger)

		// act
		activity, err := client.CompleteSuccessfulActivity(workflowID, activityID, nil)

		// assert
		assert.Nil(t, activity, "Expected no activity to be returned due to token error")
		assert.Equal(t, expectedError, err, "Expected an error returned")

	})

	t.Run("WhenAPIErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("Token", nil)

		// return server error from http handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		})

		// set up routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		activity, err := client.CompleteSuccessfulActivity(workflowID, activityID, nil)

		// assert
		assert.Nil(t, activity, "Expected no activity to be returned due to API error")
		assert.NotNil(t, err, "Expected an error returned because workflow API sent a 500 error")
	})
}

func TestCompleteCancelledActivity(t *testing.T) {
	// arrange
	workflowID := "my-workflow"
	activityID := "my-activity"
	endpoint := "/" + workflowAPIBasePath + "/workflows/{workflowID}/activities/{activityID}"

	t.Run("WhenSuccessfulExpectsCancelledActivityInRequest", func(t *testing.T) {
		// arrange
		expectedActivity := &models.Activity{
			ID:     swag.String(activityID),
			Status: swag.String(models.ActivityStatusCancelled),
			Error:  &models.ActivityError{Reason: swag.String("some reason"), Details: "some cancel details"},
		}
		var actualActivity models.Activity
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("token", nil)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			assert.NotEmpty(t, r.Header.Get("Authorization"), "Authorization header should not be empty")
			receivedWorkflowID := mux.Vars(r)["workflowID"]
			receivedActivityID := mux.Vars(r)["activityID"]
			assert.EqualValues(t, workflowID, receivedWorkflowID, "Expected workflow id received to match what was passed in")
			assert.EqualValues(t, activityID, receivedActivityID, "Expected activity id received to match what was passed in")
			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			err = json.Unmarshal(bodyBytes, &actualActivity)
			if err != nil {
				t.Fatal(err)
			}
			bytes, err := json.Marshal(&models.Activity{})
			if err != nil {
				t.Fatal("Failed to marshal activity " + err.Error())
			}
			w.Write(bytes)
		})

		// Setup routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		activity, err := client.CompleteCancelledActivity(workflowID, activityID, *expectedActivity.Error.Reason, expectedActivity.Error.Details)

		// assert
		assert.Equal(t, *expectedActivity.ID, *actualActivity.ID, "Expected activity IDs to match")
		assert.Equal(t, models.ActivityStatusCancelled, *actualActivity.Status, "Expected activity status to be: "+models.ActivityStatusCancelled)
		assert.NotNil(t, actualActivity.Error, "Expected an activity error")
		assert.Equal(t, expectedActivity.Error.Details, actualActivity.Error.Details, "Expected error details to be passed in")
		assert.Equal(t, *expectedActivity.Error.Reason, *actualActivity.Error.Reason, "Expected error reason to be passed in")
		assert.Nil(t, err, "Expected no error")
		assert.NotNil(t, activity, "Expected retrieved activity to not be nil")
	})

	t.Run("WhenFetcherErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		expectedError := errors.New("Some auth0 error")
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("", expectedError)
		client := NewClient(fakeTokenFetcher, gatewayURL, workflowAPIBasePath, audience, logger)

		// act
		activity, err := client.CompleteCancelledActivity(workflowID, activityID, "", "")

		// assert
		assert.Nil(t, activity, "Expected no activity to be returned due to token error")
		assert.Equal(t, expectedError, err, "Expected an error returned")

	})

	t.Run("WhenAPIErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("Token", nil)

		// return server error from http handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		})

		// set up routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		activity, err := client.CompleteCancelledActivity(workflowID, activityID, "", "")

		// assert
		assert.Nil(t, activity, "Expected no activity to be returned due to API error")
		assert.NotNil(t, err, "Expected an error returned because workflow API sent a 500 error")
	})
}

func TestCompleteFailedActivity(t *testing.T) {
	// arrange
	workflowID := "my-workflow"
	activityID := "my-activity"
	endpoint := "/" + workflowAPIBasePath + "/workflows/{workflowID}/activities/{activityID}"

	t.Run("WhenSuccessfulExpectsFailedActivityInRequest", func(t *testing.T) {
		// arrange
		expectedActivity := &models.Activity{
			ID:     swag.String(activityID),
			Status: swag.String(models.ActivityStatusFailed),
			Error:  &models.ActivityError{Reason: swag.String("some reason"), Details: "some failure details"},
		}
		var actualActivity models.Activity
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("token", nil)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			assert.NotEmpty(t, r.Header.Get("Authorization"), "Authorization header should not be empty")
			receivedWorkflowID := mux.Vars(r)["workflowID"]
			receivedActivityID := mux.Vars(r)["activityID"]
			assert.EqualValues(t, workflowID, receivedWorkflowID, "Expected workflow id received to match what was passed in")
			assert.EqualValues(t, activityID, receivedActivityID, "Expected activity id received to match what was passed in")
			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			err = json.Unmarshal(bodyBytes, &actualActivity)
			if err != nil {
				t.Fatal(err)
			}
			bytes, err := json.Marshal(&models.Activity{})
			if err != nil {
				t.Fatal("Failed to marshal activity " + err.Error())
			}
			w.Write(bytes)
		})

		// Setup routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		activity, err := client.CompleteFailedActivity(workflowID, activityID, *expectedActivity.Error.Reason, expectedActivity.Error.Details)

		// assert
		assert.Equal(t, *expectedActivity.ID, *actualActivity.ID, "Expected activity IDs to match")
		assert.Equal(t, models.ActivityStatusFailed, *actualActivity.Status, "Expected activity status to be: "+models.ActivityStatusCancelled)
		assert.NotNil(t, actualActivity.Error, "Expected an activity error")
		assert.Equal(t, *expectedActivity.Error.Reason, *actualActivity.Error.Reason, "Expected error reason to be passed in")
		assert.Equal(t, expectedActivity.Error.Details, actualActivity.Error.Details, "Expected error details to be passed in")
		assert.Nil(t, err, "Expected no error")
		assert.NotNil(t, activity, "Expected retrieved activity to not be nil")
	})

	t.Run("WhenFetcherErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		expectedError := errors.New("Some auth0 error")
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("", expectedError)
		client := NewClient(fakeTokenFetcher, gatewayURL, workflowAPIBasePath, audience, logger)

		// act
		activity, err := client.CompleteFailedActivity(workflowID, activityID, "", "")

		// assert
		assert.Nil(t, activity, "Expected no activity to be returned due to token error")
		assert.Equal(t, expectedError, err, "Expected an error returned")

	})

	t.Run("WhenAPIErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("Token", nil)

		// return server error from http handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		})

		// set up routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		activity, err := client.CompleteFailedActivity(workflowID, activityID, "", "")

		// assert
		assert.Nil(t, activity, "Expected no activity to be returned due to API error")
		assert.NotNil(t, err, "Expected an error returned because workflow API sent a 500 error")
	})
}

func TestHeartbeatActivity(t *testing.T) {
	// arrange
	workflowID := "my-workflow"
	activityID := "my-activity"
	taskToken := "token"
	heartbeatDetails := "details"
	endpoint := "/" + workflowAPIBasePath + "/workflows/{workflowID}/activities/{activityID}/heartbeat"
	heartbeatToReturn := &models.Heartbeat{
		ActivityID: swag.String(activityID),
		TaskToken:  swag.String(taskToken),
		Cancelled:  false,
		Details:    heartbeatDetails,
	}

	t.Run("WhenSuccessfulExpectsHeartbeatReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("token", nil)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			assert.NotEmpty(t, r.Header.Get("Authorization"), "Authorization header should not be empty")
			receivedWorkflowID := mux.Vars(r)["workflowID"]
			receivedActivityID := mux.Vars(r)["activityID"]
			assert.EqualValues(t, workflowID, receivedWorkflowID, "Expected workflow id received to match what was passed in")
			assert.EqualValues(t, activityID, receivedActivityID, "Expected activity id received to match what was passed in")
			bytes, err := json.Marshal(heartbeatToReturn)
			if err != nil {
				t.Error("Failed to marshal heartbeat")
			}
			w.Write(bytes)
		})

		// Setup routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		heartbeat, err := client.HeartbeatActivity(workflowID, activityID)

		// assert
		assert.Nil(t, err, "Expected error to be nil when getting workflow")
		assert.NotNil(t, heartbeat, "Expected retrieved heartbeat to not be nil")
		assert.Equal(t, *heartbeatToReturn.ActivityID, *heartbeat.ActivityID, "Expected heartbeat activity IDs to match")
		assert.Equal(t, *heartbeatToReturn.TaskToken, *heartbeat.TaskToken, "Expected heartbeat tokens to match")
		assert.Equal(t, heartbeatToReturn.Cancelled, heartbeat.Cancelled, "Expected heartbeat cancelled field to match")
		assert.Equal(t, heartbeatToReturn.Details, heartbeat.Details, "Expected heartbeat details to match")
	})

	t.Run("WhenFetcherErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		expectedError := errors.New("Some auth0 error")
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("", expectedError)
		client := NewClient(fakeTokenFetcher, gatewayURL, workflowAPIBasePath, audience, logger)

		// act
		heartbeat, err := client.HeartbeatActivity(workflowID, activityID)

		// assert
		assert.Nil(t, heartbeat, "Expected no heartbeat to be returned due to token error")
		assert.Equal(t, expectedError, err, "Expected an error returned")
	})

	t.Run("WhenAPIErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("Token", nil)

		// return server error from http handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		})

		// set up routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		heartbeat, err := client.HeartbeatActivity(workflowID, activityID)

		// assert
		assert.Nil(t, heartbeat, "Expected no heartbeat to be returned due to API error")
		assert.NotNil(t, err, "Expected an error returned because workflow API sent a 500 error")
	})
}

func TestHeartbeatActivityWithToken(t *testing.T) {
	// arrange
	activityID := "my-activity"
	taskToken := "token"
	heartbeatDetails := "details"
	endpoint := "/" + workflowAPIBasePath + "/heartbeats"
	heartbeatToReturn := &models.Heartbeat{
		ActivityID: swag.String(activityID),
		TaskToken:  swag.String(taskToken),
		Cancelled:  false,
		Details:    heartbeatDetails,
	}

	t.Run("WhenSuccessfulExpectsHeartbeatReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("token", nil)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			assert.NotEmpty(t, r.Header.Get("Authorization"), "Authorization header should not be empty")
			receivedBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				assert.Fail(t, "Message did not include heartbeat")
			}
			receivedHeartbeat := &models.Heartbeat{}
			if err := json.Unmarshal(receivedBytes, receivedHeartbeat); err != nil {
				assert.Fail(t, "Unable to unmarshal heartbeat")
			}
			assert.EqualValues(t, heartbeatToReturn, receivedHeartbeat, "Expected received heartbeat to match what was passed in")
			w.Write(receivedBytes)
		})

		// Setup routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		heartbeat, err := client.HeartbeatActivityWithToken(taskToken, activityID, heartbeatDetails)

		// assert
		assert.Nil(t, err, "Expected error to be nil when getting workflow")
		assert.NotNil(t, heartbeat, "Expected retrieved heartbeat to not be nil")
		assert.Equal(t, *heartbeatToReturn.ActivityID, *heartbeat.ActivityID, "Expected heartbeat activity IDs to match")
		assert.Equal(t, *heartbeatToReturn.TaskToken, *heartbeat.TaskToken, "Expected heartbeat tokens to match")
		assert.Equal(t, heartbeatToReturn.Cancelled, heartbeat.Cancelled, "Expected heartbeat cancelled field to match")
		assert.Equal(t, heartbeatToReturn.Details, heartbeat.Details, "Expected heartbeat details to match")
	})

	t.Run("WhenAuthTokenFetcherErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		expectedError := errors.New("Some auth0 error")
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("", expectedError)
		client := NewClient(fakeTokenFetcher, gatewayURL, workflowAPIBasePath, audience, logger)

		// act
		heartbeat, err := client.HeartbeatActivityWithToken(taskToken, activityID, heartbeatDetails)

		// assert
		assert.Nil(t, heartbeat, "Expected no heartbeat to be returned due to token error")
		assert.Equal(t, expectedError, err, "Expected an error returned")
	})

	t.Run("WhenAPIErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("Token", nil)

		// return server error from http handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		})

		// set up routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		heartbeat, err := client.HeartbeatActivityWithToken(taskToken, activityID, heartbeatDetails)

		// assert
		assert.Nil(t, heartbeat, "Expected no heartbeat to be returned due to API error")
		assert.NotNil(t, err, "Expected an error returned because workflow API sent a 500 error")
	})
}

func TestStartWorkflow(t *testing.T) {
	// arrange
	entityID := int32(200)
	orgID := int32(10)
	workflowType := models.PostWorkflowWorkflowTypeAssumedStrain
	workflowID := "sim-200"
	endpoint := "/" + workflowAPIBasePath + "/workflows"
	post := &models.PostWorkflow{
		EntityID:                  swag.Int32(entityID),
		OrganizationID:            swag.Int32(orgID),
		WorkflowType:              swag.String(workflowType),
		RunDistortionCompensation: false,
		RunSupportOptimization:    false,
	}

	t.Run("WhenSuccessfulExpectsWorkflowIDReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("token", nil)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			assert.NotEmpty(t, r.Header.Get("Authorization"), "Authorization header should not be empty")
			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				assert.Fail(t, "Message body did not include workflow")
			}
			receivedWorkflow := &models.PostWorkflow{}
			if err := json.Unmarshal(bodyBytes, receivedWorkflow); err != nil {
				assert.Fail(t, "Unable to unmarshal workflow")
			}
			assert.EqualValues(t, post, receivedWorkflow, "Expected recieved workflow to match what was passed in")
			workflowIDBytes, err := json.Marshal(workflowID)
			if err != nil {
				assert.Fail(t, "Failed to marshal workflow ID")
			}
			w.Write(workflowIDBytes)
		})

		// Setup routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		returnedWorkflowID, err := client.StartWorkflow(post)

		// assert
		assert.Nil(t, err, "Expected error to be nil when getting workflow")
		assert.Equal(t, workflowID, returnedWorkflowID, "Expected returned workflow ID to match response value")
	})

	t.Run("WhenAuthTokenFetcherErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		expectedError := errors.New("Some auth0 error")
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("", expectedError)
		client := NewClient(fakeTokenFetcher, gatewayURL, workflowAPIBasePath, audience, logger)

		// act
		workflowID, err := client.StartWorkflow(post)

		// assert
		assert.Empty(t, workflowID, "Expected no workflow ID to be returned due to token error")
		assert.Equal(t, expectedError, err, "Expected an error returned")
	})

	t.Run("WhenAPIErrorsExpectsErrorReturned", func(t *testing.T) {
		// arrange
		fakeTokenFetcher := &auth0fakes.FakeTokenFetcher{}
		fakeTokenFetcher.TokenReturns("Token", nil)

		// return server error from http handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		})

		// set up routes
		r := mux.NewRouter()
		r.HandleFunc(endpoint, handler)
		testServer := httptest.NewServer(r)
		defer testServer.Close()
		client := NewClient(fakeTokenFetcher, testServer.URL, workflowAPIBasePath, audience, logger)

		// act
		workflowID, err := client.StartWorkflow(post)

		// assert
		assert.Empty(t, workflowID, "Expected no workflow ID to be returned due to api error")
		assert.NotNil(t, err, "Expected an error returned because workflow API sent a 500 error")
	})
}
