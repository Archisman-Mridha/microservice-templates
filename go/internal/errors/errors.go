// MIT License
//
// Copyright (c) 2026 OpenMedia
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package errors

import "errors"

/*
When the presentation layer (in this case, gRPC server) gets :

	(1) an error of type APIError, it straight away sends that to the client.
	    The gRPC error status code can be decided based on the concrete error type (like
	    ErrUserNotFound).


	(2) any other type of (unexpected) error, it logs that error and sends the ErrInternalServer back
	    to the client.
	    The gRPC error status code will always be codes.Internal.
*/
type APIError error

// Returns an APIError, constructed using the given error message.
func NewAPIError(message string) APIError {
	//nolint:errcheck
	return errors.New(message).(APIError)
}

// API errors.
var (
	ErrInvalidEmail    = NewAPIError("invalid email")
	ErrInvalidUsername = NewAPIError("invalid username")

	ErrDuplicateEmail    = NewAPIError("email already exists")
	ErrDuplicateUsername = NewAPIError("username already exists")

	ErrInvalidJWT = NewAPIError("invalid JWT")
	ErrExpiredJWT = NewAPIError("expired JWT")

	ErrUserNotFound = NewAPIError("user not found")
)

var ErrInternalServer = errors.New("internal server error")
