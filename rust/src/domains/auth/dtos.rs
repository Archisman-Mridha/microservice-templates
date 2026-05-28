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

use crate::domains::auth::valueobjects::{Email, Name, Password, Username};

pub struct CreateUserArgs {
  pub name:     Name,
  pub email:    Email,
  pub username: Username,
  pub password: Password
}

pub enum SigninID {
  Email(Email),
  Username(Username)
}

pub struct SigninArgs {
  pub id:       SigninID,
  pub password: String
}

pub struct SigninOutput {
  pub user_id:      i64,
  pub access_token: String
}

pub struct VerifyAccessTokenArgs {
  pub access_token: String
}

pub struct VerifyAccessTokenOutput {
  pub user_id: i64
}
