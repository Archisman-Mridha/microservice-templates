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

	"openmedia.io/internal/healthcheck"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type HealthcheckService struct {
	healthcheckables []healthcheck.Healthcheckable
}

func (h *HealthcheckService) Check(ctx context.Context,
	request *grpc_health_v1.HealthCheckRequest,
) (*grpc_health_v1.HealthCheckResponse, error) {
	response := &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_UNKNOWN,
	}

	err := healthcheck.Healthcheck(h.healthcheckables)
	if err != nil {
		response.Status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
		return response, err
	}

	response.Status = grpc_health_v1.HealthCheckResponse_SERVING
	return response, nil
}

func (h *HealthcheckService) List(ctx context.Context,
	request *grpc_health_v1.HealthListRequest,
) (*grpc_health_v1.HealthListResponse, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (h *HealthcheckService) Watch(
	request *grpc_health_v1.HealthCheckRequest,
	responseStream grpc.ServerStreamingServer[grpc_health_v1.HealthCheckResponse],
) error {
	return status.Error(codes.Unimplemented, "unimplemented")
}
