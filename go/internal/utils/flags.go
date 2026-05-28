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

package utils

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"strings"

	"openmedia.io/internal/assert"
)

type GetFlagOrEnvFn = func(f *flag.Flag)

// Usage : flagSet.VisitAll(getFlagOrEnvValue("AUTH_MICROSERVICE_"))
func CreateGetFlagOrEnvValueFn(envPrefix string) GetFlagOrEnvFn {
	/*
		When a flag isn't set, we try to get its value from the corresponding environment variable.

		For example, when the flag name is config-file, the corresponding environment variable is
		AUTH_MICROSERVICE_CONFIG_FILE. AUTH_MICROSERVICE_ here is the env-prefix.

		NOTE : Panics, if both the flag and environment variable aren't set and a default value isn't
		       set for the flag.
	*/
	getFlagOrEnvFn := func(f *flag.Flag) {
		ctx := context.Background()

		if len(f.Value.String()) > 0 {
			return
		}

		// Since the flag is not set, we'll try to get the value from the corresponding environment
		// variable.
		envName := envPrefix + strings.ReplaceAll(strings.ToUpper(f.Name), "-", "_")
		envValue, envFound := os.LookupEnv(envName)
		if envFound {
			err := f.Value.Set(envValue)
			assert.AssertErrNil(ctx, err,
				"Failed setting flag value to corresponding env value",
				slog.String("flag", f.Name),
			)

			return
		}

		assert.Assert(ctx, (len(f.DefValue) == 0),
			"Neither flag nor corresponding env was set",
			slog.String("flag", f.Name),
		)
	}

	return getFlagOrEnvFn
}
