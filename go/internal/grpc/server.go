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
	"fmt"
	"log/slog"
	"net"

	"openmedia.io/internal/assert"
	"openmedia.io/internal/healthcheck"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/oops"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/health/grpc_health_v1"

	"google.golang.org/grpc/reflection"

	grpcmetricsmiddleware "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"

	// This is necessary to avoid ambiguous import error.
	// REFER : https://github.com/open-telemetry/opentelemetry-collector/issues/10476
	_ "google.golang.org/genproto/googleapis/type/date"

	/*
	  The WASI preview 1 specification has partial support for socket networking, preventing a large
	  class of Go applications from running when compiled to WebAssembly with GOOS=wasip1. Extensions
	  to the base specifications have been implemented by runtimes to enable a wider range of
	  programs to be run as WebAssembly modules.

	  Where possible, the package offers the ability to automatically configure the network stack via
	  init functions called on package imports.

	  When imported, this package alter the default configuration to install a dialer function
	  implemented on top of the WASI socket extensions. When compiled to other targets, the import
	  of those packages does nothing.

	  REFER : https://github.com/dev-wasm/dev-wasm-go.
	*/
	_ "github.com/stealthrocket/net/http"
)

type GRPCServer struct {
	*grpc.Server
}

type (
	NewGRPCServerArgs struct {
		DevModeEnabled bool

		Healthcheckables []healthcheck.Healthcheckable

		ToGRPCErrorStatusCodeFn ToGRPCErrorStatusCodeFn
	}

	ToGRPCErrorStatusCodeFn = func(error) codes.Code
)

// Creates and returns a gRPC server.
func NewGRPCServer(ctx context.Context, args NewGRPCServerArgs) *GRPCServer {
	var (
		requestLogger = newGRPCRequestLogger(slog.Default())

		serverMetrics = grpcmetricsmiddleware.NewServerMetrics()
	)

	server := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),

		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(requestLogger),

			errorHandlerUnaryServerInterceptor(args.ToGRPCErrorStatusCodeFn),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(requestLogger),

			errorHandlerStreamServerInterceptor(args.ToGRPCErrorStatusCodeFn),
		),
	)

	serverMetrics.InitializeMetrics(server)
	prometheus.DefaultRegisterer.MustRegister(serverMetrics)

	if args.DevModeEnabled {
		reflection.Register(server)
	}

	grpc_health_v1.RegisterHealthServer(server, &HealthcheckService{
		healthcheckables: args.Healthcheckables,
	})

	return &GRPCServer{server}
}

// Creates a TCP listener at the given address and uses it to run the gRPC server.
func (server *GRPCServer) MustRun(ctx context.Context, port int) error {
	address := fmt.Sprintf("0.0.0.0:%d", port)

	tcpListener, err := net.Listen("tcp", address)
	assert.AssertErrNil(ctx, err, "Failed creating TCP listener", slog.Int("port", port))

	slog.DebugContext(ctx, "Running gRPC server....", slog.Int("port", port))
	if err := server.Serve(tcpListener); err != nil {
		return oops.Wrapf(err, "Failed running gRPC server")
	}

	return nil
}

// Stops the gRPC server from accepting new connections and RPC requests.
// Then, waits for the RPCs which are currently being processed, to finish.
func (server *GRPCServer) GracefulShutdown() {
	server.GracefulStop()
	slog.Debug("Shut down gRPC server")
}
