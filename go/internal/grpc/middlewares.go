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

package grpc

import (
	"context"
	"log/slog"

	"openmedia.io/internal/errors"
	"openmedia.io/internal/logger"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Returns a gRPC request logger (which under the hood invokes slog).
// TODO : Filter out critical fields (like password).
func newGRPCRequestLogger(slogLogger *slog.Logger) logging.Logger {
	return logging.LoggerFunc(
		func(ctx context.Context, logLevel logging.Level, message string, fields ...any) {
			slogLogger.Log(ctx, slog.Level(logLevel), message, fields...)
		},
	)
}

// Constructs and returns unary server interceptor for error handling.
// The error handler interceptor converts any error (returned from the usecases layer) to gRPC
// specific error.
func errorHandlerUnaryServerInterceptor(
	toGRPCErrorStatusCodeFn ToGRPCErrorStatusCodeFn,
) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		request any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		response, err := handler(ctx, request)
		return response, toGRPCError(ctx, err, toGRPCErrorStatusCodeFn)
	}
}

// Constructs and returns stream server interceptor for error handling.
// The error handler interceptor converts any error (returned from the usecases layer) to gRPC
// specific error.
func errorHandlerStreamServerInterceptor(
	toGRPCErrorStatusCodeFn ToGRPCErrorStatusCodeFn,
) grpc.StreamServerInterceptor {
	return func(
		server any,
		stream grpc.ServerStream,
		_ *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		err := handler(server, stream)
		return toGRPCError(stream.Context(), err, toGRPCErrorStatusCodeFn)
	}
}

// Converts any error (returned from the usecases layer) to gRPC specific error.
// If the error is unexpected (not of type APIError), then that gets logged.
func toGRPCError(ctx context.Context,
	err error,
	toGRPCErrorStatusCodeFn ToGRPCErrorStatusCodeFn,
) error {
	if err == nil {
		return nil
	}

	switch err.(type) {
	case errors.APIError:
		return status.Error(toGRPCErrorStatusCodeFn(err), err.Error())

	default:
		// Log unexpected error.
		slog.ErrorContext(ctx, "Unexpected error occurred", logger.Error(err))

		return status.Error(codes.Internal, errors.ErrInternalServer.Error())
	}
}
