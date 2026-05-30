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

use {crate::domains::auth::error::Error, validator::ValidateEmail};

const NAME_MIN_LENGTH: usize = 3;
const NAME_MAX_LENGTH: usize = 20;

const PASSWORD_MIN_LENGTH: usize = 6;
const PASSWORD_MAX_LENGTH: usize = 30;

pub struct Name(pub String);

impl TryFrom<String> for Name {
  type Error = Error;

  fn try_from(value: String) -> Result<Self, Self::Error> {
    let mut validation_errors = vec![];

    if !((NAME_MIN_LENGTH <= value.len()) && (value.len() <= NAME_MAX_LENGTH)) {
      validation_errors.push(format!(
        "Name length must be between {NAME_MIN_LENGTH} - {NAME_MAX_LENGTH} characters"
      ));
    }

    if !value.chars().all(|character| character.is_alphabetic()) {
      validation_errors.push(String::from("Name must contain only alphabetic characters"));
    }

    if !validation_errors.is_empty() {
      return Err(Error::ValidationFailed(validation_errors));
    }

    Ok(Self(value))
  }
}

pub struct Email(pub String);

impl TryFrom<String> for Email {
  type Error = Error;

  fn try_from(value: String) -> Result<Self, Self::Error> {
    if !value.validate_email() {
      return Err(Error::ValidationFailed(vec![String::from("Email validation failed")]));
    }

    Ok(Self(value))
  }
}

pub struct Username(pub String);

impl TryFrom<String> for Username {
  type Error = Error;

  fn try_from(value: String) -> Result<Self, Self::Error> {
    let mut validation_errors = vec![];

    let mut contains_alphabet = false;

    for character in value.chars() {
      if !(character.is_alphanumeric() || (character == '_') || (character == '.')) {
        validation_errors.push(String::from(
          "Username must contain only alphanumeric, underscore and dot characters"
        ));

        break;
      }

      if character.is_alphabetic() {
        contains_alphabet = true;
      }
    }

    if !contains_alphabet {
      validation_errors
        .push(String::from("Username must contain atleast one alphabetic character"));
    }

    if !validation_errors.is_empty() {
      return Err(Error::ValidationFailed(validation_errors));
    }

    Ok(Self(value))
  }
}

pub struct Password(pub String);

impl TryFrom<String> for Password {
  type Error = Error;

  fn try_from(value: String) -> Result<Self, Self::Error> {
    if !((PASSWORD_MIN_LENGTH <= value.len()) && (value.len() <= PASSWORD_MAX_LENGTH)) {
      return Err(Error::ValidationFailed(vec![format!("Password length must be between {PASSWORD_MIN_LENGTH} - {PASSWORD_MAX_LENGTH} characters")]));
    }

    Ok(Self(value))
  }
}
