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
    api::auth_api_service_server::AuthApiService,
    dtos::{CreateUserArgs, SigninArgs, SigninID, VerifyAccessTokenArgs},
    error::Error,
    service::AuthService,
    valueobjects::{Email, Name, Password, Username}
  },
  tonic::{Request, Response, Status},
  tracing::instrument
};

// Code generated from .proto files.
tonic::include_proto!("auth.api.v1");

// The term “descriptors” refers to models that describe the data types defined
// in Protobuf sources. They resemble an AST (Abstract Syntax Tree) of the
// Protobuf IDL, using Protobuf messages for the tree nodes.
pub const FILE_DESCRIPTOR_SET: &[u8] = tonic::include_file_descriptor_set!("auth_api_v1");

#[allow(clippy::upper_case_acronyms)]
pub struct AuthAPI {
  service: AuthService
}

impl AuthAPI {
  pub fn new(service: AuthService) -> Self { Self { service } }
}

#[tonic::async_trait]
impl AuthApiService for AuthAPI {
  #[instrument(skip(self))]
  async fn ping(&self, _request: Request<()>) -> Result<Response<()>, Status> {
    Ok(Response::new(()))
  }

  #[instrument(skip(self))]
  async fn create_user(&self,
                       request: Request<CreateUserRequest>)
                       -> Result<Response<CreateUserResponse>, Status> {
    let into_status = <Error as Into<Status>>::into;

    let request = request.into_inner();

    let args =
      CreateUserArgs { name:     Name::try_from(request.name).map_err(into_status)?,
                       email:    Email::try_from(request.email).map_err(into_status)?,
                       username: Username::try_from(request.username).map_err(into_status)?,
                       password: Password::try_from(request.password).map_err(into_status)? };

    self.service.create_user(args).await.map_err(into_status)?;

    Ok(Response::new(CreateUserResponse {}))
  }

  #[instrument(skip(self))]
  async fn signin(&self,
                  request: Request<SigninRequest>)
                  -> Result<Response<SigninResponse>, Status> {
    let into_status = <Error as Into<Status>>::into;

    let request = request.into_inner();

    let id = match request.id {
      | Some(signin_request::Id::Email(email)) =>
        SigninID::Email(Email::try_from(email).map_err(into_status)?),

      | Some(signin_request::Id::Username(username)) =>
        SigninID::Username(Username::try_from(username).map_err(into_status)?),

      | None => unreachable!()
    };

    let output = self.service
                     .signin(SigninArgs { id,
                                          password: request.password })
                     .await
                     .map_err(into_status)?;

    Ok(Response::new(SigninResponse { user_id:      output.user_id,
                                      access_token: output.access_token }))
  }

  #[instrument(skip(self))]
  async fn verify_access_token(&self,
                               request: Request<VerifyAccessTokenRequest>)
                               -> Result<Response<VerifyAccessTokenResponse>, Status> {
    let request = request.into_inner();

    let output = self.service
                     .verify_access_token(VerifyAccessTokenArgs { access_token:
                                                                    request.access_token })
                     .await
                     .map_err(<Error as Into<Status>>::into)?;

    Ok(Response::new(VerifyAccessTokenResponse { user_id: output.user_id }))
  }
}
