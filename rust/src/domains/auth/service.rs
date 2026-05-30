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

use {
  crate::domains::auth::{
    dtos::{
      CreateUserArgs, SigninArgs, SigninID, SigninOutput, VerifyAccessTokenArgs,
      VerifyAccessTokenOutput
    },
    error::Error,
    token::JWTService,
    user
  },
  anyhow::anyhow,
  argon2::{
    Argon2, PasswordHash, PasswordHasher, PasswordVerifier,
    password_hash::{SaltString, rand_core::OsRng}
  },
  sea_orm::{ActiveValue::Set, DatabaseConnection, DbErr, EntityTrait, RuntimeErr},
  std::sync::Arc,
  tracing::error
};

const UNIQUE_EMAIL_CONSTRAINT_NAME: &str = "users_email_key";
const UNIQUE_USERNAME_CONSTRAINT_NAME: &str = "users_username_key";

pub struct AuthService {
  connection: Arc<DatabaseConnection>,

  jwt_service: JWTService
}

impl AuthService {
  pub fn new(connection: Arc<DatabaseConnection>, jwt_service: JWTService) -> Self {
    Self { connection,
           jwt_service }
  }

  pub async fn create_user(&self, args: CreateUserArgs) -> Result<(), Error> {
    // Hash the password, using the Argon2ID algorithm.

    let password_hasher = Argon2::default();

    // In cryptography, a salt is random data fed as an additional input to a
    // one-way function that hashes data, a password or passphrase. Salting
    // helps defend against attacks that use precomputed tables (e.g.
    // rainbow tables), by vastly growing the size of table needed for a
    // successful attack. It also helps protect passwords that occur multiple
    // times in a database, as a new salt is used for each password
    // instance. Additionally, salting does not place any burden on users.
    let salt = SaltString::generate(&mut OsRng);

    let hashed_password = password_hasher
      .hash_password(args.password.0.as_bytes(), &salt)
      .map_err(|error| Error::Unexpected(anyhow!("Failed hashing password : {error}")))?
      .to_string();

    // Try to create the user in the database.

    let user = user::ActiveModel { name: Set(args.name.0),
                                   email: Set(args.email.0),
                                   username: Set(args.username.0),
                                   hashed_password: Set(hashed_password),

                                   ..Default::default() };

    user::Entity::insert(user).exec(&*self.connection).await.map_err(|error| {
                                                               match error {
      DbErr::Exec(RuntimeErr::SqlxError(error)) => match error.as_database_error() {
        Some(error) => match error.constraint() {
          Some(constraint) => match constraint {
            UNIQUE_EMAIL_CONSTRAINT_NAME => Error::DuplicateEmail,
            UNIQUE_USERNAME_CONSTRAINT_NAME => Error::DuplicateUsername,

            _ => Error::Unexpected(anyhow!("{error}"))
          },
          None => Error::Unexpected(anyhow!("{error}"))
        },
        None => Error::Unexpected(anyhow!("{error}"))
      },
      _ => Error::Unexpected(anyhow!("{error}"))
    }
                                                             })?;

    Ok(())
  }

  pub async fn signin(&self, args: SigninArgs) -> Result<SigninOutput, Error> {
    // Try finding the user from the database.

    let select_query = match args.id {
      | SigninID::Email(email) => user::Entity::find_by_email(email.0),
      | SigninID::Username(username) => user::Entity::find_by_username(username.0)
    };

    let user = select_query.one(&*self.connection)
                           .await
                           .map_err(|error| Error::Unexpected(anyhow!("{error}")))?
                           .ok_or(Error::UserNotFound)?;

    // Check whether the provided password is correct.

    let password_hasher = Argon2::default();

    let hashed_password = PasswordHash::new(&user.hashed_password)
      .map_err(|error| Error::Unexpected(anyhow!("{error}")))?;

    password_hasher.verify_password(args.password.as_bytes(), &hashed_password)
                   .map_err(|error| {
                     error!("Password verification failed : {error}");

                     Error::WrongPassword
                   })?;

    // Generate access token.
    let access_token = self.jwt_service.issue(user.id)?;

    Ok(SigninOutput { user_id: user.id,
                      access_token })
  }

  pub async fn verify_access_token(&self,
                                   args: VerifyAccessTokenArgs)
                                   -> Result<VerifyAccessTokenOutput, Error> {
    // Try to retriece the user ID from the JWT.

    let claims = self.jwt_service.verify(&args.access_token)?;

    let user_id =
      claims.registered.subject.parse().map_err(|error| {
                                          error!("Failed parsing access token subject as user ID : {error}");

                                          Error::DecodingAccessTokenFailed
                                        })?;

    // Verify that the user exists.
    user::Entity::find_by_id(user_id).one(&*self.connection)
                                     .await
                                     .map_err(|error| Error::Unexpected(anyhow!("{error}")))?
                                     .ok_or(Error::UserNotFound)?;

    Ok(VerifyAccessTokenOutput { user_id })
  }
}
