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

package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samber/oops"
	"openmedia.io/internal/domains/auth/users/repository"
	"openmedia.io/internal/domains/auth/users/repository/postgres/sqlc/generated"
	apierrors "openmedia.io/internal/errors"
)

type Repository struct {
	queries *generated.Queries
}

func NewRepository(connection *sql.DB) repository.Repository {
	queries := generated.New(connection)

	return &Repository{queries}
}

func (r *Repository) Create(ctx context.Context, args *repository.CreateArgs) (int32, error) {
	userID, err := r.queries.CreateUser(ctx, (*generated.CreateUserParams)(args))
	if err != nil {
		pgErr, ok := err.(*pgconn.PgError)
		if ok && (pgErr.Code == pgerrcode.UniqueViolation) {
			switch pgErr.ColumnName {
			case "email":
				return 0, apierrors.ErrDuplicateEmail

			case "username":
				return 0, apierrors.ErrDuplicateUsername
			}
		}

		return 0, oops.Wrap(err)
	}
	return userID, nil
}

func (r *Repository) FindByEmail(ctx context.Context,
	email string,
) (*repository.FindByOutput, error) {
	userDetails, err := r.queries.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrUserNotFound
		}

		return nil, oops.Wrap(err)
	}
	return (*repository.FindByOutput)(userDetails), nil
}

func (r *Repository) FindByUsername(ctx context.Context,
	username string,
) (*repository.FindByOutput, error) {
	userDetails, err := r.queries.FindUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrUserNotFound
		}

		return nil, oops.Wrap(err)
	}
	return (*repository.FindByOutput)(userDetails), nil
}

func (r *Repository) Exists(ctx context.Context, id int32) (bool, error) {
	_, err := r.queries.FindUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, oops.Wrap(err)
	}
	return true, nil
}
