package types

// Generic wrapper for all API responses, including the error. The data can be empty.
type APIResponse[T any] struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Data       T      `json:"data,omitempty"`
}
