package compute

import (
	"github.com/hashicorp/errwrap"
	"github.com/joyent/triton-go/client"
)

// IsBadRequest tests whether err wraps a client.TritonError with
// code BadRequest
func IsBadRequest(err error) bool {
	return isSpecificError(err, "BadRequest")
}

// IsInternalError tests whether err wraps a client.TritonError with
// code InternalError
func IsInternalError(err error) bool {
	return isSpecificError(err, "InternalError")
}

// IsInUseError tests whether err wraps a client.TritonError with
// code InUseError
func IsInUseError(err error) bool {
	return isSpecificError(err, "InUseError")
}

// IsInvalidArgument tests whether err wraps a client.TritonError with
// code InvalidArgument
func IsInvalidArgument(err error) bool {
	return isSpecificError(err, "InvalidArgument")
}

// IsInvalidCredentials tests whether err wraps a client.TritonError with
// code InvalidCredentials
func IsInvalidCredentials(err error) bool {
	return isSpecificError(err, "InvalidCredentials")
}

// IsInvalidHeader tests whether err wraps a client.TritonError with
// code InvalidHeader
func IsInvalidHeader(err error) bool {
	return isSpecificError(err, "InvalidHeader")
}

// IsInvalidVersion tests whether err wraps a client.TritonError with
// code InvalidVersion
func IsInvalidVersion(err error) bool {
	return isSpecificError(err, "InvalidVersion")
}

// IsMissingParameter tests whether err wraps a client.TritonError with
// code MissingParameter
func IsMissingParameter(err error) bool {
	return isSpecificError(err, "MissingParameter")
}

// IsNotAuthorized tests whether err wraps a client.TritonError with
// code NotAuthorized
func IsNotAuthorized(err error) bool {
	return isSpecificError(err, "NotAuthorized")
}

// IsRequestThrottled tests whether err wraps a client.TritonError with
// code RequestThrottled
func IsRequestThrottled(err error) bool {
	return isSpecificError(err, "RequestThrottled")
}

// IsRequestTooLarge tests whether err wraps a client.TritonError with
// code RequestTooLarge
func IsRequestTooLarge(err error) bool {
	return isSpecificError(err, "RequestTooLarge")
}

// IsRequestMoved tests whether err wraps a client.TritonError with
// code RequestMoved
func IsRequestMoved(err error) bool {
	return isSpecificError(err, "RequestMoved")
}

// IsResourceFound tests whether err wraps a client.TritonError with code ResourceFound
func IsResourceFound(err error) bool {
	return isSpecificError(err, "ResourceFound")
}

// IsResourceNotFound tests whether err wraps a client.TritonError with
// code ResourceNotFound
func IsResourceNotFound(err error) bool {
	return isSpecificError(err, "ResourceNotFound")
}

// IsUnknownError tests whether err wraps a client.TritonError with
// code UnknownError
func IsUnknownError(err error) bool {
	return isSpecificError(err, "UnknownError")
}

// isSpecificError checks whether the error represented by err wraps
// an underlying client.TritonError with code errorCode.
func isSpecificError(err error, errorCode string) bool {
	if err == nil {
		return false
	}

	tritonErrorInterface := errwrap.GetType(err.(error), &client.TritonError{})
	if tritonErrorInterface == nil {
		return false
	}

	tritonErr := tritonErrorInterface.(*client.TritonError)
	if tritonErr.Code == errorCode {
		return true
	}

	return false
}
