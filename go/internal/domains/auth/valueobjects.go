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
	stderrors "errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode"

	"openmedia.io/internal/errors"
)

const (
	NAME_MIN_LENGTH = 3
	NAME_MAX_LENGTH = 20

	PASSWORD_MIN_LENGTH = 6
	PASSWORD_MAX_LENGTH = 30
)

type Name string

func NewName(value string) (Name, errors.ValidationErrors) {
	value = strings.TrimSpace(value)

	validationErrors := []error{}

	if !((NAME_MIN_LENGTH <= len(value)) && (len(value) <= NAME_MAX_LENGTH)) {
		validationErrors = append(validationErrors,
			fmt.Errorf("name length must be between %d - %d characters",
				NAME_MIN_LENGTH,
				NAME_MAX_LENGTH,
			),
		)
	}

	for _, character := range value {
		if !(unicode.IsLetter(character) || (character == ' ')) {
			validationErrors = append(validationErrors,
				stderrors.New("name must contain only alphabetic characters"))
		}
	}

	if len(validationErrors) > 0 {
		return Name(""), errors.ValidationErrors(stderrors.Join(validationErrors...))
	}
	return Name(value), nil
}

type Email string

func NewEmail(value string) (Email, errors.ValidationErrors) {
	value = strings.TrimSpace(value)

	_, err := mail.ParseAddress(value)
	if err != nil {
		return Email(""), errors.ValidationErrors(fmt.Errorf("couldn't parse email : %v", err))
	}

	return Email(value), nil
}

type Username string

func NewUsername(value string) (Username, errors.ValidationErrors) {
	value = strings.TrimSpace(value)

	validationErrors := []error{}

	containsLetter := false

	for _, character := range value {
		if !(unicode.IsLetter(character) || (character == '_') || (character == '.')) {
			validationErrors = append(validationErrors,
				stderrors.New(
					"username must contain only alphanumeric, underscore and dot characters",
				),
			)

			break
		}

		if unicode.IsLetter(character) {
			containsLetter = true
		}
	}

	if !containsLetter {
		validationErrors = append(validationErrors,
			stderrors.New("username must contain atleast one alphabetic character"))
	}

	if len(validationErrors) > 0 {
		return Username(""), errors.ValidationErrors(stderrors.Join(validationErrors...))
	}
	return Username(value), nil
}

type Password string

func NewPassword(value string) (Password, errors.ValidationErrors) {
	value = strings.TrimSpace(value)

	if !((PASSWORD_MIN_LENGTH <= len(value)) && (len(value) <= PASSWORD_MAX_LENGTH)) {
		return Password(""), errors.ValidationErrors(fmt.Errorf(
			"password length must be between %d - %d characters",
			PASSWORD_MIN_LENGTH,
			PASSWORD_MAX_LENGTH,
		))
	}

	return Password(value), nil
}
