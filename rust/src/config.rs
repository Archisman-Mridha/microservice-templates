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
  anyhow::anyhow,
  config::{Environment, File},
  serde::Deserialize
};

#[derive(Deserialize)]
pub struct Config {
  #[serde(default)]
  pub server: ServerConfig,

  pub postgresql: PostgreSQLConfig,

  pub jwt: JWTConfig,

  pub telemetry: TelemetryConfig
}

#[derive(Deserialize)]
pub struct ServerConfig {
  pub address: String
}

impl Default for ServerConfig {
  fn default() -> Self { Self { address: "127.0.0.1:4000".to_string() } }
}

#[derive(Deserialize)]
pub struct PostgreSQLConfig {
  pub url: String
}

#[derive(Deserialize)]
pub struct JWTConfig {
  #[serde(default = "default_jwt_issuer")]
  pub issuer: String,

  #[serde(default = "default_jwt_audiences")]
  pub audiences: String,

  #[serde(default = "default_jwt_expires_after")]
  pub expires_after: usize,

  pub encoding_secret: String
}

#[derive(Deserialize)]
pub struct TelemetryConfig {
  pub otlp_collector_url: String,

  #[serde(default)]
  pub logger: LoggerConfig
}

#[derive(Deserialize)]
pub struct LoggerConfig {
  pub level: String
}

impl Default for LoggerConfig {
  fn default() -> Self { Self { level: "INFO".to_string() } }
}

fn default_jwt_issuer() -> String { "openmedia-backend".to_string() }

fn default_jwt_audiences() -> String { "openmedia-client".to_string() }

fn default_jwt_expires_after() -> usize { 24 * 60 * 60 } // 24 hours

impl Config {
  pub fn new(config_file_path: &str) -> anyhow::Result<Self> {
    config::Config::builder().add_source(File::with_name(config_file_path))
                             .add_source(Environment::default().separator("__"))
                             .build()?
                             .try_deserialize()
                             .map_err(|error| anyhow!("Failed deserializing config file : {error}"))
  }
}
