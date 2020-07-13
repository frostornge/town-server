// Code generated by mockery v1.0.0. DO NOT EDIT.

// Regenerate this file using `make einterfaces-mocks`.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/nonce/town-server/model"
)

// MessageExportInterface is an autogenerated mock type for the MessageExportInterface type
type MessageExportInterface struct {
	mock.Mock
}

// RunExport provides a mock function with given fields: format, since
func (_m *MessageExportInterface) RunExport(format string, since int64) (int64, *model.AppError) {
	ret := _m.Called(format, since)

	var r0 int64
	if rf, ok := ret.Get(0).(func(string, int64) int64); ok {
		r0 = rf(format, since)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string, int64) *model.AppError); ok {
		r1 = rf(format, since)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// StartSynchronizeJob provides a mock function with given fields: ctx, exportFromTimestamp
func (_m *MessageExportInterface) StartSynchronizeJob(ctx context.Context, exportFromTimestamp int64) (*model.Job, *model.AppError) {
	ret := _m.Called(ctx, exportFromTimestamp)

	var r0 *model.Job
	if rf, ok := ret.Get(0).(func(context.Context, int64) *model.Job); ok {
		r0 = rf(ctx, exportFromTimestamp)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Job)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(context.Context, int64) *model.AppError); ok {
		r1 = rf(ctx, exportFromTimestamp)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}
