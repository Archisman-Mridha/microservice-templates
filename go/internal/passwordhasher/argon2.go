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

package passwordhasher

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/samber/oops"
	"golang.org/x/crypto/argon2"
	"openmedia.io/internal/errors"
)

const (
	ARGON2_PASSWORD_HASHER_SALT_LENGTH = 16

	ARGON2_PASSWORD_HASHING_ITERATION_COUNT = 1
	ARGON2_PASSWORD_HASHING_MEMORY_QUOTA    = 64 * 1024
	ARGON2_PASSWORD_HASHING_THREADS_QUOTA   = 4
	ARGON2_PASSWORD_HASHING_HASH_LENGTH     = 32
)

type Argon2PasswordHasher struct{}

func NewArgon2PasswordHasher() PasswordHasher {
	return &Argon2PasswordHasher{}
}

func (a *Argon2PasswordHasher) Hash(password string) (string, error) {
	// In cryptography, a salt is random data fed as an additional input to a
	// one-way function that hashes data, a password or passphrase. Salting
	// helps defend against attacks that use precomputed tables (e.g.
	// rainbow tables), by vastly growing the size of table needed for a
	// successful attack. It also helps protect passwords that occur multiple
	// times in a database, as a new salt is used for each password
	// instance. Additionally, salting does not place any burden on users.
	salt := make([]byte, ARGON2_PASSWORD_HASHER_SALT_LENGTH)
	_, err := rand.Read(salt)
	if err != nil {
		return "", oops.Wrapf(err, "Failed generating salt")
	}

	// Derive the key from the password.
	key := argon2.IDKey([]byte(password), salt,
		ARGON2_PASSWORD_HASHING_ITERATION_COUNT,
		ARGON2_PASSWORD_HASHING_MEMORY_QUOTA,
		ARGON2_PASSWORD_HASHING_THREADS_QUOTA,
		ARGON2_PASSWORD_HASHING_HASH_LENGTH,
	)

	// Hashed password is composed of the salt and the key.

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedKey := base64.RawStdEncoding.EncodeToString(key)

	hashedPassword := fmt.Sprintf("%s$%s", encodedSalt, encodedKey)
	return hashedPassword, nil
}

func (a *Argon2PasswordHasher) Verify(providedPassword, storedHashedPassword string) error {
	// Extract the salt and key from the stored hashed password.

	parts := strings.Split(storedHashedPassword, "$")
	if len(parts) != 2 {
		return oops.New("Stored hashed password is invalid")
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[0])
	if err != nil {
		return oops.Wrapf(err, "Failed Base64 decoding salt")
	}

	storedKey, err := base64.RawStdEncoding.Strict().DecodeString(parts[1])
	if err != nil {
		return oops.Wrapf(err, "Failed Base64 decoding stored key")
	}

	// Derive the key from the provided password.
	key := argon2.IDKey([]byte(providedPassword), salt,
		ARGON2_PASSWORD_HASHING_ITERATION_COUNT,
		ARGON2_PASSWORD_HASHING_MEMORY_QUOTA,
		ARGON2_PASSWORD_HASHING_THREADS_QUOTA,
		ARGON2_PASSWORD_HASHING_HASH_LENGTH,
	)

	// The key extrated from the stored hashed password should match with
	// the key dervied from the provided password.
	if subtle.ConstantTimeCompare(storedKey, key) != 1 {
		return errors.ErrWrongPassword
	}
	return nil
}
