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

	"openmedia.io/internal/domains/auth/users/repository"
	"openmedia.io/internal/errors"
	"openmedia.io/internal/passwordhasher"
	"openmedia.io/internal/token"
)

type Service struct {
	passwordHasher  passwordhasher.PasswordHasher
	usersRepository repository.Repository
	tokenService    token.TokenService
}

func NewAuthService(
	passwordHasher passwordhasher.PasswordHasher,
	usersRepository repository.Repository,
	tokenService token.TokenService,
) *Service {
	return &Service{
		passwordHasher,
		usersRepository,
		tokenService,
	}
}

func (s *Service) CreateUser(ctx context.Context, args *CreateUserArgs) error {
	// Hash the password, using the Argon2ID algorithm.
	hashedPassword, err := s.passwordHasher.Hash(string(args.Password))
	if err != nil {
		return err
	}

	// Try to create the user in the database.
	_, err = s.usersRepository.Create(ctx, &repository.CreateArgs{
		Name:     string(args.Name),
		Email:    string(args.Email),
		Username: string(args.Username),

		HashedPassword: hashedPassword,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Signin(ctx context.Context, args *SigninArgs) (*SigninOutput, error) {
	// Try finding the user from the database.
	var (
		userDetails *repository.FindByOutput
		err         error
	)
	switch args.IDKind {
	case SigninIDKindEmail:
		userDetails, err = s.usersRepository.FindByEmail(ctx, string(*args.Email))

	case SigninIDKindUsername:
		userDetails, err = s.usersRepository.FindByUsername(ctx, string(*args.Username))
	}
	if err != nil {
		return nil, err
	}

	// Verify that the provided password is correct.
	err = s.passwordHasher.Verify(string(args.Password), userDetails.HashedPassword)
	if err != nil {
		return nil, err
	}

	// Generate JWT.
	jwt, err := s.tokenService.Issue(userDetails.ID)
	if err != nil {
		return nil, err
	}

	output := &SigninOutput{
		UserID: userDetails.ID,
		JWT:    jwt,
	}
	return output, nil
}

func (s *Service) VerifyJWT(ctx context.Context,
	args *VerifyJWTArgs,
) (*VerifyJWTOutput, error) {
	// Try to retriece the user ID from the JWT.
	userID, err := s.tokenService.GetUserIDFrom(args.JWT)
	if err != nil {
		return nil, err
	}

	// Verify that the user exists.
	userExists, err := s.usersRepository.Exists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !userExists {
		return nil, errors.ErrUserNotFound
	}

	output := &VerifyJWTOutput{
		UserID: userID,
	}
	return output, nil
}
