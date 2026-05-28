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

package assert

import (
	"context"
	"log/slog"
	"os"

	"openmedia.io/internal/logger"
)

// Panics if the given error isn't nil.
func AssertErrNil(ctx context.Context, err error,
	contextualErrorMessage string,
	attributes ...any,
) {
	if err == nil {
		return
	}

	attributes = append(attributes, logger.Error(err))
	slog.ErrorContext(ctx, contextualErrorMessage, attributes...)
	os.Exit(1)
}

// Panics if the condition is false.
func Assert(ctx context.Context, condition bool, contextualErrorMessage string, attributes ...any) {
	if condition {
		return
	}

	slog.ErrorContext(ctx, contextualErrorMessage, attributes...)
	os.Exit(1)
}
