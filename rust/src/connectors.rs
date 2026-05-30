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
  crate::{domains::auth::user, healthcheck::Healthcheckable},
  sea_orm::{Database, DatabaseConnection},
  std::sync::Arc
};

pub struct PostgreSQLConnector {
  connection: Arc<DatabaseConnection>
}

impl PostgreSQLConnector {
  pub async fn new(url: &str) -> anyhow::Result<Self> {
    let connection = Database::connect(url).await?;

    // Synchronizes database schema with entity definitions.
    connection.get_schema_builder().register(user::Entity).sync(&connection).await?;

    Ok(Self { connection: Arc::new(connection) })
  }

  pub fn get_connection(&self) -> Arc<DatabaseConnection> { self.connection.clone() }
}

#[async_trait::async_trait]
impl Healthcheckable for PostgreSQLConnector {
  async fn healthcheck(&self) -> anyhow::Result<()> {
    self.connection.ping().await?;
    Ok(())
  }
}
