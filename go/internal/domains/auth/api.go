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

func (a *AuthAPI) CreateUser(ctx context.Context,
	request *generated.CreateUserRequest,
) (*generated.CreateUserResponse, error) {
	name, err := NewName(request.GetName())
	if err != nil {
		return nil, err
	}

	email, err := NewEmail(request.GetEmail())
	if err != nil {
		return nil, err
	}

	username, err := NewUsername(request.GetUsername())
	if err != nil {
		return nil, err
	}

	password, err := NewPassword(request.GetPassword())
	if err != nil {
		return nil, err
	}

	err = a.authService.CreateUser(ctx, &CreateUserArgs{
		Name:     name,
		Email:    email,
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	response := &generated.CreateUserResponse{}
	return response, nil
}

func (a *AuthAPI) Signin(ctx context.Context,
	request *generated.SigninRequest,
) (*generated.SigninResponse, error) {
	password, err := NewPassword(request.GetPassword())
	if err != nil {
		return nil, err
	}

	//nolint:exhaustruct
	args := &SigninArgs{
		Password: password,
	}
	switch request.GetId().(type) {
	case *generated.SigninRequest_Email:
		args.IDKind = SigninIDKindEmail

		email, err := NewEmail(request.GetEmail())
		if err != nil {
			return nil, err
		}
		args.Email = &email

	case *generated.SigninRequest_Username:
		args.IDKind = SigninIDKindUsername

		username, err := NewUsername(request.GetUsername())
		if err != nil {
			return nil, err
		}
		args.Username = &username
	}

	output, err := a.authService.Signin(ctx, args)
	if err != nil {
		return nil, err
	}

	response := &generated.SigninResponse{
		Jwt:    output.JWT,
		UserId: output.UserID,
	}
	return response, nil
}

func (a *AuthAPI) VerifyJWT(ctx context.Context,
	request *generated.VerifyJWTRequest,
) (*generated.VerifyJWTResponse, error) {
	output, err := a.authService.VerifyJWT(ctx, &VerifyJWTArgs{
		JWT: request.GetJwt(),
	})
	if err != nil {
		return nil, err
	}

	response := &generated.VerifyJWTResponse{
		UserId: output.UserID,
	}
	return response, nil
}
