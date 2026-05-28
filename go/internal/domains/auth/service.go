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

	"openmedia.io/internal/domains/users"
	"openmedia.io/internal/errors"
	"openmedia.io/internal/token"
)

const (
	SigninIDKindEmail SigninIDKind = iota
	SigninIDKindUsername
)

type Service struct {
	usersService *users.Service
	tokenService token.TokenService
}

func NewAuthService(
	usersService *users.Service,
	tokenService token.TokenService,
) *Service {
	return &Service{
		usersService,
		tokenService,
	}
}

type (
	SigninIDKind uint

	SigninInput struct {
		IDKind SigninIDKind
		ID,

		Password string
	}

	SigninOutput struct {
		UserID      int32
		AccessToken string
	}
)

func (s *Service) Signin(ctx context.Context, input *SigninInput) (*SigninOutput, error) {
	var (
		userDetails *users.FindByOutput
		err         error
	)
	switch input.IDKind {
	case SigninIDKindEmail:
		userDetails, err = s.usersService.FindByEmail(ctx, input.ID)

	case SigninIDKindUsername:
		userDetails, err = s.usersService.FindByUsername(ctx, input.ID)
	}
	if err != nil {
		return nil, err
	}

	accessToken, err := s.tokenService.Issue(userDetails.ID)
	if err != nil {
		return nil, err
	}

	output := &SigninOutput{
		UserID:      userDetails.ID,
		AccessToken: accessToken,
	}
	return output, nil
}

func (s *Service) VerifyAccessToken(ctx context.Context, accessToken string) (int32, error) {
	userID, err := s.tokenService.GetUserIDFromToken(accessToken)
	if err != nil {
		return 0, err
	}

	userExists, err := s.usersService.Exists(ctx, userID)
	if err != nil {
		return 0, err
	}
	if !userExists {
		return 0, errors.ErrUserNotFound
	}

	return userID, nil
}
