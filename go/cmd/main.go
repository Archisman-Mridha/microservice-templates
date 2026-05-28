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

package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"openmedia.io/internal/config"
	"openmedia.io/internal/connectors"
	"openmedia.io/internal/constants"
	"openmedia.io/internal/domains/auth"
	"openmedia.io/internal/domains/users"
	usersrepositorypostgres "openmedia.io/internal/domains/users/repository/postgres"
	"openmedia.io/internal/errors"
	"openmedia.io/internal/grpc"
	"openmedia.io/internal/healthcheck"
	"openmedia.io/internal/logger"
	"openmedia.io/internal/token"
	"openmedia.io/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
)

var configFilePath string

func parseCLIFlags() {
	// Load environment variables from the .env file.
	_ = godotenv.Load()

	flagSet := flag.NewFlagSet("", flag.ExitOnError)

	flagSet.StringVar(&configFilePath, constants.FLAG_CONFIG_FILE, "", "Config file path")

	cmdArgs := os.Args[1:]
	if err := flagSet.Parse(cmdArgs); err != nil {
		slog.Error("Failed parsing command line flags", logger.Error(err))
		os.Exit(1)
	}

	flagSet.VisitAll(utils.CreateGetFlagOrEnvValueFn(""))
}

func main() {
	parseCLIFlags()

	// When the program receives any interruption / SIGKILL / SIGTERM signal, the cancel function is
	// automatically invoked. The cancel function is responsible for freeing all the resources
	// associated with the context.
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT,
	)

	// Construct validator with custom validators.
	validator := utils.NewValidator(ctx)

	// Get config.
	config := config.MustParseConfig(ctx, configFilePath, validator)

	if err := run(ctx, config, validator); err != nil {
		slog.ErrorContext(ctx, err.Error())

		cancel()

		// Give some time for remaining resources (if any) to be cleaned up.
		time.Sleep(time.Second)

		os.Exit(1)
	}
}

func run(ctx context.Context, config *config.Config, validator *validator.Validate) error {
	// Setup logger.
	logger.SetupLogger(config.DebugLogging)

	waitGroup, ctx := errgroup.WithContext(ctx)

	// Construct connectors.

	postgresConnector := connectors.NewPostgresConnector(ctx, config.Postgres.URL)
	defer postgresConnector.Shutdown()

	// Construct services.

	usersRepository := usersrepositorypostgres.NewRepository(postgresConnector.GetConnection())
	usersService := users.NewService(usersRepository)

	tokenService := token.NewJWTService(config.JWT)

	authService := auth.NewAuthService(usersService, tokenService)

	// Construct and run the gRPC server.

	gRPCServer := grpc.NewGRPCServer(ctx, grpc.NewGRPCServerArgs{
		DevModeEnabled: config.DevMode,

		Healthcheckables: []healthcheck.Healthcheckable{
			postgresConnector,
		},

		ToGRPCErrorStatusCodeFn: getGRPCErrorStatusCode,
	})

	auth.RegisterAuthAPI(gRPCServer.Server, authService)

	waitGroup.Go(func() error {
		return gRPCServer.MustRun(ctx, config.ServerPort)
	})

	/*
		The returned channel gets closed when either of this happens :

			(1) A program termination signal is received, because of which the parent context's done
				  channel gets closed.

			(2) Any of the go-routines registered under the wait-group, finishes running.
	*/
	<-ctx.Done()
	slog.DebugContext(ctx, "Gracefully shutting down program")

	// Gracefully shutdown the gRPC server, ensuring that it finishes ongoing processing of requests.
	gRPCServer.GracefulShutdown()

	return waitGroup.Wait()
}

// Returns suitable gRPC error status code, based on the given error.
func getGRPCErrorStatusCode(err error) codes.Code {
	apiErr, ok := err.(errors.APIError)
	if !ok {
		return codes.Internal
	}

	switch apiErr {
	case errors.ErrInvalidEmail, errors.ErrInvalidUsername:
		return codes.InvalidArgument

	case errors.ErrDuplicateEmail, errors.ErrDuplicateUsername:
		return codes.AlreadyExists

	case errors.ErrInvalidJWT, errors.ErrExpiredJWT:
		return codes.Unauthenticated

	case errors.ErrUserNotFound:
		return codes.NotFound

	default:
		return codes.Unknown
	}
}
