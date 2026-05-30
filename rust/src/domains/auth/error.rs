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

use {tonic::Status, tracing::error};

const ERROR_SEPARATOR: &str = "; ";

pub enum Error {
  ValidationFailed(Vec<String>),

  DuplicateEmail,
  DuplicateUsername,

  UserNotFound,
  WrongPassword,
  DecodingAccessTokenFailed,
  AccessTokenExpired,

  Unexpected(anyhow::Error)
}

impl Into<Status> for Error {
  fn into(self) -> Status {
    match self {
      | Self::ValidationFailed(validation_errors) =>
        Status::invalid_argument(validation_errors.join(ERROR_SEPARATOR)),

      | Self::DuplicateEmail => Status::already_exists("Email already exists"),
      | Self::DuplicateUsername => Status::already_exists("Username alredy exists"),

      | Self::UserNotFound => Status::unauthenticated("User not found"),
      | Self::WrongPassword => Status::unauthenticated("Wrong password"),
      | Self::DecodingAccessTokenFailed => Status::unauthenticated("Decoding access token failed"),
      | Self::AccessTokenExpired => Status::unauthenticated("Access token expired"),

      | Self::Unexpected(error) => {
        error!("Unexpected error occurred : {error}");

        Status::unknown("Unexpected error occurred")
      }
    }
  }
}
