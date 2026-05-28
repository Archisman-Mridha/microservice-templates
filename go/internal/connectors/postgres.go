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

package connectors

import (
	"context"
	"database/sql"
	"log/slog"

	"openmedia.io/internal/assert"
	"openmedia.io/internal/logger"

	"github.com/samber/oops"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresConnector struct {
	connection *sql.DB
}

func NewPostgresConnector(ctx context.Context, url string) *PostgresConnector {
	// TODO : use a query logger and tracer.
	connection, err := sql.Open("pgx", url)
	assert.AssertErrNil(ctx, err, "Failed connecting to Postgres")

	// Ping the database, verifying that a working connection has been established.
	err = connection.Ping()
	assert.AssertErrNil(ctx, err, "Failed pinging Postgres")

	slog.DebugContext(ctx, "Connected to Postgres")

	return &PostgresConnector{connection}
}

func (p *PostgresConnector) GetConnection() *sql.DB {
	return p.connection
}

func (p *PostgresConnector) Healthcheck() error {
	if err := p.connection.Ping(); err != nil {
		return oops.Wrapf(err, "Failed pinging Postgres")
	}
	return nil
}

func (p *PostgresConnector) Shutdown() {
	if err := p.connection.Close(); err != nil {
		slog.Error("Failed closing Postgres connection", logger.Error(err))
		return
	}
	slog.Debug("Shut down Postgres client")
}
