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

package token

import (
	goerrors "errors"
	"strconv"
	"time"

	"openmedia.io/internal/config"
	"openmedia.io/internal/errors"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/samber/oops"
)

const JWT_VALIDITIY_PERIOD = 24 * time.Hour

type (
	JWTService struct {
		config.JWTConfig
	}

	// JSON web tokens (JWTs) claims are pieces of information asserted about a subject.
	JWTClaims struct {
		// Registered claims are standard claims registered with the Internet Assigned Numbers
		// Authority (IANA) and defined by the JWT specification to ensure interoperability with
		// third-party, or external, applications.
		gojwt.RegisteredClaims

		/*
		  Custom claims consist of non-registered public or private claims.

		    (1) Public claims : You can create custom claims for public consumption, which might
		        contain generic information like name and email. If you create public claims, you must
		        either register them or use collision-resistant names through namespacing and take
		        reasonable precautions to make sure you are in control of the namespace you use.

		    (2) Private claims : You can create private custom claims to share information specific to
		        your application. For example, while a public claim might contain generic information
		        like name and email, private claims would be more specific, such as employee ID and
		        department name.
		*/
	}
)

func NewJWTService(jwtConfig config.JWTConfig) *JWTService {
	return &JWTService{jwtConfig}
}

func (j *JWTService) Issue(userID int32) (string, error) {
	jwtSigner := gojwt.NewWithClaims(gojwt.SigningMethodHS256, JWTClaims{
		//nolint:exhaustruct
		RegisteredClaims: gojwt.RegisteredClaims{
			// The "iss" (issuer) claim identifies the principal that issued the JWT. The "iss" value is
			// a case-sensitive string containing a StringOrURI value.
			Issuer: j.Issuer,

			// The "sub" (subject) claim identifies the principal that is the subject of the JWT. The
			// claims in a JWT are normally statements about the subject.
			Subject: strconv.Itoa(int(userID)),

			// The "aud" (audience) claim identifies the recipients that the JWT is intended for. Each
			// principal intended to process the JWT MUST identify itself with a value in the audience
			// claim. If the principal processing the claim does not identify itself with a value in the
			// "aud" claim when this claim is present, then the JWT MUST be rejected. In the general
			// case, the "aud" value is an array of case-sensitive strings, each containing a StringOrURI
			// value. In the special case when the JWT has one audience, the "aud" value MAY be a single
			// case-sensitive string containing a StringOrURI value.
			Audience: gojwt.ClaimStrings(j.Audiences),

			IssuedAt: gojwt.NewNumericDate(time.Now()),

			// The "exp" (expiration time) claim identifies the expiration time on or after which the JWT
			// MUST NOT be accepted for processing. The processing of the "exp" claim requires that the
			// current date/time MUST be before the expiration date/time listed in the "exp" claim.
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(JWT_VALIDITIY_PERIOD)),
		},
	})

	jwt, err := jwtSigner.SignedString([]byte(j.SigningKey))
	if err != nil {
		return "", oops.Wrapf(err, "Failed generating JWT")
	}
	return jwt, nil
}

func (j *JWTService) GetUserIDFromToken(jwt string) (int32, error) {
	parsedJWT, err := gojwt.Parse(jwt,
		func(_ *gojwt.Token) (any, error) { return []byte(j.SigningKey), nil },
		gojwt.WithExpirationRequired(),
	)
	if err != nil {
		if goerrors.Is(err, gojwt.ErrTokenExpired) {
			return 0, errors.ErrExpiredJWT
		}

		return 0, errors.ErrInvalidJWT
	}

	subject, err := parsedJWT.Claims.GetSubject()
	if err != nil {
		return 0, errors.ErrInvalidJWT
	}

	userID, err := strconv.ParseInt(subject, 10, 32)
	if err != nil {
		return 0, errors.ErrInvalidJWT
	}
	return int32(userID), nil
}
