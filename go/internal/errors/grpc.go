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

import "google.golang.org/grpc/codes"

// Returns suitable gRPC error status code, based on the given error.
func GetGRPCErrorStatusCode(err error) codes.Code {
	apiErr, ok := err.(APIError)
	if !ok {
		return codes.Internal
	}

	if _, ok = apiErr.(ValidationErrors); ok {
		return codes.InvalidArgument
	}

	switch apiErr {
	case ErrDuplicateEmail, ErrDuplicateUsername:
		return codes.AlreadyExists

	case ErrWrongPassword, ErrInvalidJWT, ErrExpiredJWT:
		return codes.Unauthenticated

	case ErrUserNotFound:
		return codes.NotFound

	default:
		return codes.Unknown
	}
}
