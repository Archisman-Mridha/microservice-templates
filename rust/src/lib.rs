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
  crate::{
    config::Config,
    connectors::PostgreSQLConnector,
    domains::auth::{api::AuthAPI, service::AuthService, token::JWTService},
    observability::setup_observability
  },
  std::str::FromStr,
  tracing::Level
};

pub mod config;
mod connectors;
mod domains;
mod healthcheck;
mod observability;
mod server;

const SERVICE_NAME: &str = "openmedia-backend";

pub async fn run(config: Config) -> anyhow::Result<()> {
  // Setup observability.
  let log_level = Level::from_str(&config.telemetry.logger.level)?;
  setup_observability(SERVICE_NAME.to_string(), log_level, &config.telemetry.otlp_collector_url)?;

  // Create connectors.

  let postgresql_connector = PostgreSQLConnector::new(&config.postgresql.url).await?;

  // Create services.

  let jwt_service = JWTService::new(config.jwt.issuer,
                                    config.jwt.audiences,
                                    config.jwt.expires_after,
                                    config.jwt.encoding_secret);

  let auth_service = AuthService::new(postgresql_connector.get_connection(), jwt_service);

  // Create APIs.

  let auth_api = AuthAPI::new(auth_service);

  // Create and run the gRPC server.

  let healthcheckables = vec![postgresql_connector];

  let address = config.server.address.parse()?;
  server::start(address, auth_api, healthcheckables).await?;

  Ok(())
}
