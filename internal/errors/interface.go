// Package errors provides common error types used across the application.
package errors

// CodedError is an interface for errors that have an error code.
type CodedError interface {
	error
	Code() ErrorCode
}

// GetErrorCode extracts the error code from an error if it implements CodedError.
func GetErrorCode(err error) (ErrorCode, bool) {
	if codedErr, ok := err.(CodedError); ok {
		return codedErr.Code(), true
	}
	return "", false
}
