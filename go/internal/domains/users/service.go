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

package users

import (
	"context"

	"golang.org/x/crypto/bcrypt"
	"openmedia.io/internal/domains/users/repository"
)

type Service struct {
	usersRepository repository.Repository
}

func NewService(usersRepository repository.Repository) *Service {
	return &Service{
		usersRepository,
	}
}

type CreateInput struct {
	Name,
	Email,
	Username,
	Password string
}

func (s *Service) Create(ctx context.Context, input *CreateInput) (int32, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	userID, err := s.usersRepository.Create(ctx, &repository.CreateInput{
		Name:     input.Name,
		Email:    input.Email,
		Username: input.Username,

		HashedPassword: string(hashedPassword),
	})
	if err != nil {
		return 0, err
	}

	return userID, nil
}

type FindByOutput struct {
	ID             int32
	HashedPassword string
}

func (s *Service) FindByEmail(ctx context.Context, email string) (*FindByOutput, error) {
	output, err := s.usersRepository.FindByEmail(ctx, email)
	return (*FindByOutput)(output), err
}

func (s *Service) FindByUsername(ctx context.Context, username string) (*FindByOutput, error) {
	output, err := s.usersRepository.FindByUsername(ctx, username)
	return (*FindByOutput)(output), err
}

func (s *Service) Exists(ctx context.Context, id int32) (bool, error) {
	return s.usersRepository.Exists(ctx, id)
}
