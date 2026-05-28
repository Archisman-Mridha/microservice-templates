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

package auth

import (
	"context"

	"google.golang.org/grpc"
	generated "openmedia.io/prototypes/generated/auth/api/v1"
)

type AuthAPI struct {
	generated.UnimplementedAuthAPIServiceServer

	authService *Service
}

func RegisterAuthAPI(server *grpc.Server, authService *Service) {
	//nolint:exhaustruct
	authAPI := &AuthAPI{
		authService: authService,
	}

	generated.RegisterAuthAPIServiceServer(server, authAPI)
}

func (a *AuthAPI) Signin(ctx context.Context,
	request *generated.SigninRequest,
) (*generated.SigninResponse, error) {
	//nolint:exhaustruct
	input := &SigninInput{
		Password: request.GetPassword(),
	}
	switch request.GetId().(type) {
	case *generated.SigninRequest_Email:
		input.IDKind = SigninIDKindEmail
		input.ID = request.GetEmail()

	case *generated.SigninRequest_Username:
		input.IDKind = SigninIDKindUsername
		input.ID = request.GetUsername()
	}

	output, err := a.authService.Signin(ctx, input)
	if err != nil {
		return nil, err
	}

	response := &generated.SigninResponse{
		AccessToken: output.AccessToken,
		UserId:      output.UserID,
	}
	return response, nil
}

func (a *AuthAPI) ValidateJWT(ctx context.Context,
	request *generated.VerifyAccessTokenRequest,
) (*generated.VerifyAccessTokenResponse, error) {
	userID, err := a.authService.VerifyAccessToken(ctx, request.GetAccessToken())
	if err != nil {
		return nil, err
	}

	response := &generated.VerifyAccessTokenResponse{
		UserId: userID,
	}
	return response, nil
}
