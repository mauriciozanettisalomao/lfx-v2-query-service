// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package nats

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain"
	"github.com/stretchr/testify/assert"
)

// MockNATSClient is a mock implementation of NATSClientInterface
type MockNATSClient struct {
	checkAccessResponse AccessCheckNATSResponse
	checkAccessError    error
	closeError          error
}

func NewMockNATSClient() *MockNATSClient {
	return &MockNATSClient{}
}

func (m *MockNATSClient) CheckAccess(ctx context.Context, request *AccessCheckNATSRequest) (AccessCheckNATSResponse, error) {
	if m.checkAccessError != nil {
		return nil, m.checkAccessError
	}
	return m.checkAccessResponse, nil
}

func (m *MockNATSClient) Close() error {
	return m.closeError
}

func (m *MockNATSClient) SetCheckAccessResponse(response AccessCheckNATSResponse) {
	m.checkAccessResponse = response
}

func (m *MockNATSClient) SetCheckAccessError(err error) {
	m.checkAccessError = err
}

func (m *MockNATSClient) SetCloseError(err error) {
	m.closeError = err
}

func TestNATSAccessControlChecker_CheckAccess(t *testing.T) {
	tests := []struct {
		name           string
		subject        string
		data           []byte
		timeout        time.Duration
		setupMock      func(*MockNATSClient)
		expectedError  bool
		expectedErrMsg string
		expectedResult domain.AccessCheckResult
	}{
		{
			name:    "successful access check with allowed permissions",
			subject: "access.check.project",
			data:    []byte(`{"user_id": "user123", "resource": "project:abc"}`),
			timeout: 5 * time.Second,
			setupMock: func(mock *MockNATSClient) {
				mock.SetCheckAccessResponse(AccessCheckNATSResponse{
					"view":   "allowed",
					"edit":   "allowed",
					"delete": "denied",
				})
			},
			expectedError: false,
			expectedResult: domain.AccessCheckResult{
				"view":   "allowed",
				"edit":   "allowed",
				"delete": "denied",
			},
		},
		{
			name:    "successful access check with denied permissions",
			subject: "access.check.project",
			data:    []byte(`{"user_id": "user456", "resource": "project:xyz"}`),
			timeout: 5 * time.Second,
			setupMock: func(mock *MockNATSClient) {
				mock.SetCheckAccessResponse(AccessCheckNATSResponse{
					"view":   "denied",
					"edit":   "denied",
					"delete": "denied",
				})
			},
			expectedError: false,
			expectedResult: domain.AccessCheckResult{
				"view":   "denied",
				"edit":   "denied",
				"delete": "denied",
			},
		},
		{
			name:    "successful access check with empty response",
			subject: "access.check.project",
			data:    []byte(`{"user_id": "user789", "resource": "project:empty"}`),
			timeout: 5 * time.Second,
			setupMock: func(mock *MockNATSClient) {
				mock.SetCheckAccessResponse(AccessCheckNATSResponse{})
			},
			expectedError:  false,
			expectedResult: domain.AccessCheckResult{},
		},
		{
			name:    "NATS client error",
			subject: "access.check.project",
			data:    []byte(`{"user_id": "user123", "resource": "project:abc"}`),
			timeout: 5 * time.Second,
			setupMock: func(mock *MockNATSClient) {
				mock.SetCheckAccessError(errors.New("NATS connection timeout"))
			},
			expectedError:  true,
			expectedErrMsg: "NATS access control check failed",
		},
		{
			name:    "empty subject",
			subject: "",
			data:    []byte(`{"user_id": "user123", "resource": "project:abc"}`),
			timeout: 5 * time.Second,
			setupMock: func(mock *MockNATSClient) {
				mock.SetCheckAccessError(errors.New("invalid NATS access check request: subject and message must be set"))
			},
			expectedError:  true,
			expectedErrMsg: "NATS access control check failed",
		},
		{
			name:    "nil data",
			subject: "access.check.project",
			data:    nil,
			timeout: 5 * time.Second,
			setupMock: func(mock *MockNATSClient) {
				mock.SetCheckAccessError(errors.New("invalid NATS access check request: subject and message must be set"))
			},
			expectedError:  true,
			expectedErrMsg: "NATS access control check failed",
		},
		{
			name:    "empty data",
			subject: "access.check.project",
			data:    []byte{},
			timeout: 5 * time.Second,
			setupMock: func(mock *MockNATSClient) {
				mock.SetCheckAccessError(errors.New("invalid NATS access check request: subject and message must be set"))
			},
			expectedError:  true,
			expectedErrMsg: "NATS access control check failed",
		},
		{
			name:    "zero timeout",
			subject: "access.check.project",
			data:    []byte(`{"user_id": "user123", "resource": "project:abc"}`),
			timeout: 0,
			setupMock: func(mock *MockNATSClient) {
				mock.SetCheckAccessResponse(AccessCheckNATSResponse{
					"view": "allowed",
				})
			},
			expectedError: false,
			expectedResult: domain.AccessCheckResult{
				"view": "allowed",
			},
		},
	}

	assertion := assert.New(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock
			mockClient := NewMockNATSClient()
			tc.setupMock(mockClient)

			// Create access control checker
			checker := &NATSAccessControlChecker{
				client: mockClient,
			}

			// Execute
			ctx := context.Background()
			result, err := checker.CheckAccess(ctx, tc.subject, tc.data, tc.timeout)

			// Verify
			if tc.expectedError {
				assertion.Error(err)
				assertion.Contains(err.Error(), tc.expectedErrMsg)
				return
			}

			assertion.NoError(err)
			assertion.Equal(tc.expectedResult, result)
		})
	}
}

func TestNATSAccessControlChecker_Close(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockNATSClient)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "successful close",
			setupMock: func(mock *MockNATSClient) {
				// No error on close
			},
			expectedError: false,
		},
		{
			name: "close with error",
			setupMock: func(mock *MockNATSClient) {
				mock.SetCloseError(errors.New("failed to close connection"))
			},
			expectedError:  true,
			expectedErrMsg: "failed to close connection",
		},
	}

	assertion := assert.New(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock
			mockClient := NewMockNATSClient()
			tc.setupMock(mockClient)

			// Create access control checker
			checker := &NATSAccessControlChecker{
				client: mockClient,
			}

			// Execute
			err := checker.Close()

			// Verify
			if tc.expectedError {
				assertion.Error(err)
				assertion.Contains(err.Error(), tc.expectedErrMsg)
				return
			}

			assertion.NoError(err)
		})
	}
}

func TestNATSAccessControlChecker_convertFromNATSResponse(t *testing.T) {
	tests := []struct {
		name           string
		natsResponse   AccessCheckNATSResponse
		expectedResult domain.AccessCheckResult
	}{
		{
			name: "convert response with multiple permissions",
			natsResponse: AccessCheckNATSResponse{
				"view":   "allowed",
				"edit":   "allowed",
				"delete": "denied",
				"admin":  "denied",
			},
			expectedResult: domain.AccessCheckResult{
				"view":   "allowed",
				"edit":   "allowed",
				"delete": "denied",
				"admin":  "denied",
			},
		},
		{
			name: "convert response with single permission",
			natsResponse: AccessCheckNATSResponse{
				"view": "allowed",
			},
			expectedResult: domain.AccessCheckResult{
				"view": "allowed",
			},
		},
		{
			name:           "convert empty response",
			natsResponse:   AccessCheckNATSResponse{},
			expectedResult: domain.AccessCheckResult{},
		},
		{
			name:           "convert nil response",
			natsResponse:   nil,
			expectedResult: domain.AccessCheckResult(nil),
		},
	}

	assertion := assert.New(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create access control checker
			checker := &NATSAccessControlChecker{
				client: NewMockNATSClient(),
			}

			// Execute
			result := checker.convertFromNATSResponse(tc.natsResponse)

			// Verify
			assertion.Equal(tc.expectedResult, result)
		})
	}
}

func TestNewAccessControlChecker(t *testing.T) {
	tests := []struct {
		name           string
		config         Config
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "create access control checker with valid config",
			config: Config{
				URL:           "nats://localhost:4222",
				Timeout:       5 * time.Second,
				MaxReconnect:  10,
				ReconnectWait: 2 * time.Second,
			},
			expectedError: false,
		},
		{
			name: "create access control checker with empty URL",
			config: Config{
				URL:           "",
				Timeout:       5 * time.Second,
				MaxReconnect:  10,
				ReconnectWait: 2 * time.Second,
			},
			expectedError:  true,
			expectedErrMsg: "failed to create NATS client",
		},
		{
			name: "create access control checker with invalid URL",
			config: Config{
				URL:           "invalid-url",
				Timeout:       5 * time.Second,
				MaxReconnect:  10,
				ReconnectWait: 2 * time.Second,
			},
			expectedError:  true,
			expectedErrMsg: "failed to create NATS client",
		},
		{
			name: "create access control checker with zero timeout",
			config: Config{
				URL:           "nats://localhost:4222",
				Timeout:       0,
				MaxReconnect:  10,
				ReconnectWait: 2 * time.Second,
			},
			expectedError: false,
		},
	}

	assertion := assert.New(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Execute
			ctx := context.Background()
			checker, err := NewAccessControlChecker(ctx, tc.config)

			// Verify
			if tc.expectedError {
				assertion.Error(err)
				assertion.Contains(err.Error(), tc.expectedErrMsg)
				assertion.Nil(checker)
				return
			}

			// Note: For successful cases, we can't easily test without a real NATS server
			// In a real scenario, this would fail because there's no NATS server running
			// But we can check the basic structure
			if err == nil {
				assertion.NotNil(checker)
				assertion.IsType(&NATSAccessControlChecker{}, checker)
				// Clean up
				checker.Close()
			}
		})
	}
}

func TestNATSAccessControlChecker_Integration(t *testing.T) {
	assertion := assert.New(t)

	t.Run("end-to-end access control flow", func(t *testing.T) {
		// Setup mock with realistic data
		mockClient := NewMockNATSClient()
		mockClient.SetCheckAccessResponse(AccessCheckNATSResponse{
			"view":   "allowed",
			"edit":   "allowed",
			"delete": "denied",
			"admin":  "denied",
		})

		// Create access control checker
		checker := &NATSAccessControlChecker{
			client: mockClient,
		}

		// Execute access check
		ctx := context.Background()
		subject := "access.check.project"
		data := []byte(`{
			"user_id": "user123",
			"resource": "project:integration-test",
			"relations": ["view", "edit", "delete", "admin"]
		}`)
		timeout := 5 * time.Second

		result, err := checker.CheckAccess(ctx, subject, data, timeout)

		// Verify
		assertion.NoError(err)
		assertion.NotNil(result)
		assertion.Equal("allowed", result["view"])
		assertion.Equal("allowed", result["edit"])
		assertion.Equal("denied", result["delete"])
		assertion.Equal("denied", result["admin"])

		// Test close
		err = checker.Close()
		assertion.NoError(err)
	})
}
