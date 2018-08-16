package mocks

import "github.com/stretchr/testify/mock"

import "github.com/3dsim/workflow-goclient/models"

type Client struct {
	mock.Mock
}

// StartWorkflow provides a mock function with given fields: _a0
func (_m *Client) StartWorkflow(_a0 *models.PostWorkflow) (string, error) {
	ret := _m.Called(_a0)

	var r0 string
	if rf, ok := ret.Get(0).(func(*models.PostWorkflow) string); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*models.PostWorkflow) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CancelWorkflow provides a mock function with given fields: workflowID
func (_m *Client) CancelWorkflow(workflowID string) error {
	ret := _m.Called(workflowID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(workflowID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateActivity provides a mock function with given fields: workflowID, activity
func (_m *Client) UpdateActivity(workflowID string, activity *models.Activity) (*models.Activity, error) {
	ret := _m.Called(workflowID, activity)

	var r0 *models.Activity
	if rf, ok := ret.Get(0).(func(string, *models.Activity) *models.Activity); ok {
		r0 = rf(workflowID, activity)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Activity)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, *models.Activity) error); ok {
		r1 = rf(workflowID, activity)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateActivityPercentComplete provides a mock function with given fields: workflowID, activityID, percentComplete
func (_m *Client) UpdateActivityPercentComplete(workflowID string, activityID string, percentComplete int) (*models.Activity, error) {
	ret := _m.Called(workflowID, activityID, percentComplete)

	var r0 *models.Activity
	if rf, ok := ret.Get(0).(func(string, string, int) *models.Activity); ok {
		r0 = rf(workflowID, activityID, percentComplete)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Activity)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, int) error); ok {
		r1 = rf(workflowID, activityID, percentComplete)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CompleteSuccessfulActivity provides a mock function with given fields: workflowID, activityID, result
func (_m *Client) CompleteSuccessfulActivity(workflowID string, activityID string, result interface{}) (*models.Activity, error) {
	ret := _m.Called(workflowID, activityID, result)

	var r0 *models.Activity
	if rf, ok := ret.Get(0).(func(string, string, interface{}) *models.Activity); ok {
		r0 = rf(workflowID, activityID, result)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Activity)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, interface{}) error); ok {
		r1 = rf(workflowID, activityID, result)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CompleteCancelledActivity provides a mock function with given fields: workflowID, activityID, reason, details
func (_m *Client) CompleteCancelledActivity(workflowID string, activityID string, reason string, details string) (*models.Activity, error) {
	ret := _m.Called(workflowID, activityID, reason, details)

	var r0 *models.Activity
	if rf, ok := ret.Get(0).(func(string, string, string, string) *models.Activity); ok {
		r0 = rf(workflowID, activityID, reason, details)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Activity)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string) error); ok {
		r1 = rf(workflowID, activityID, reason, details)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CompleteFailedActivity provides a mock function with given fields: workflowID, activityID, reason, details
func (_m *Client) CompleteFailedActivity(workflowID string, activityID string, reason string, details string) (*models.Activity, error) {
	ret := _m.Called(workflowID, activityID, reason, details)

	var r0 *models.Activity
	if rf, ok := ret.Get(0).(func(string, string, string, string) *models.Activity); ok {
		r0 = rf(workflowID, activityID, reason, details)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Activity)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string) error); ok {
		r1 = rf(workflowID, activityID, reason, details)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HeartbeatActivity provides a mock function with given fields: workflowID, activityID
func (_m *Client) HeartbeatActivity(workflowID string, activityID string) (*models.Heartbeat, error) {
	ret := _m.Called(workflowID, activityID)

	var r0 *models.Heartbeat
	if rf, ok := ret.Get(0).(func(string, string) *models.Heartbeat); ok {
		r0 = rf(workflowID, activityID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Heartbeat)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(workflowID, activityID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HeartbeatActivityWithToken provides a mock function with given fields: taskToken, activityID, details
func (_m *Client) HeartbeatActivityWithToken(taskToken string, activityID string, details string) (*models.Heartbeat, error) {
	ret := _m.Called(taskToken, activityID, details)

	var r0 *models.Heartbeat
	if rf, ok := ret.Get(0).(func(string, string, string) *models.Heartbeat); ok {
		r0 = rf(taskToken, activityID, details)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Heartbeat)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(taskToken, activityID, details)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
