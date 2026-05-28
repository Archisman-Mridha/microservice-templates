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

package config

import (
	"context"
	"strings"

	"openmedia.io/internal/assert"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/mcuadros/go-defaults"
)

func MustParseConfig(ctx context.Context,
	configFilePath string,
	validator *validator.Validate,
) *Config {
	k := koanf.New(".")

	// Load configurations from the YAML config file, if provided.
	if configFilePath != "" {
		err := k.Load(file.Provider(configFilePath), yaml.Parser())
		assert.AssertErrNil(ctx, err, "Failed reading config file")
	}

	// Load configurations from environment variables (highest precedance).
	envProvider := env.Provider("", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(s), "__", ".")
	})
	err := k.Load(envProvider, nil)
	assert.AssertErrNil(ctx, err, "Failed loading env vars into config")

	config := new(Config)
	assert.AssertErrNil(ctx, k.Unmarshal("", config), "Failed unmarshalling config")

	// Populate optional fields with corresponding default values.
	defaults.SetDefaults(config)

	// Validate based on struct tags.
	err = validator.Struct(config)
	assert.AssertErrNil(ctx, err, "Config validation failed")

	return config
}
