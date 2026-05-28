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
	"log/slog"

	"openmedia.io/internal/assert"

	govalidator "github.com/go-playground/validator/v10"
	gononstandardvalidtors "github.com/go-playground/validator/v10/non-standard/validators"
)

func NewValidator(ctx context.Context) *govalidator.Validate {
	validator := govalidator.New(govalidator.WithRequiredStructEnabled())

	err := validator.RegisterValidation("notblank", gononstandardvalidtors.NotBlank)
	assert.AssertErrNil(ctx, err, "Failed registering notblank validator")

	return validator
}

type CustomFieldValidators = map[string]govalidator.Func

func RegisterCustomFieldValidators(
	validator *govalidator.Validate,
	customFieldValidators CustomFieldValidators,
) {
	ctx := context.Background()

	for id, customFieldValidator := range customFieldValidators {
		err := validator.RegisterValidation(id, customFieldValidator, false)
		assert.AssertErrNil(ctx, err,
			"Failed registering custom field validator",
			slog.String("id", id),
		)
	}
}
