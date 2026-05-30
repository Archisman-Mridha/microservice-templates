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
    domains::auth::{
      self,
      api::{AuthAPI, auth_api_service_server::AuthApiServiceServer}
    },
    healthcheck::Healthcheckable
  },
  axum::http::{HeaderMap, Request},
  opentelemetry::{global::get_text_map_propagator, propagation::Extractor},
  std::{net::SocketAddr, time::Duration},
  tokio::time::sleep,
  tonic::{body::Body, codec::CompressionEncoding, transport::Server},
  tonic_health::server::HealthReporter,
  tower::ServiceBuilder,
  tower_http::trace::TraceLayer,
  tracing::{Span, info_span, warn},
  tracing_opentelemetry::OpenTelemetrySpanExt
};

pub async fn start(address: SocketAddr,
                   auth_api: AuthAPI,
                   healthcheckables: Vec<impl Healthcheckable>)
                   -> anyhow::Result<()> {
  let (health_reporter, health_service) = tonic_health::server::health_reporter();

  let auth_api_service =
    AuthApiServiceServer::new(auth_api).send_compressed(CompressionEncoding::Gzip)
                                       .accept_compressed(CompressionEncoding::Gzip);

  let reflection_service = tonic_reflection::server::Builder::configure()
    .register_encoded_file_descriptor_set(auth::api::FILE_DESCRIPTOR_SET)
    .build_v1()?;

  let trace_propagation_layer =
    ServiceBuilder::new().layer(TraceLayer::new_for_grpc().make_span_with(make_span))
                         .map_request(link_parent_trace);

  tokio::spawn(set_health_status(health_reporter, healthcheckables));

  let router = Server::builder().layer(trace_propagation_layer)
                                .add_service(auth_api_service)
                                .add_service(health_service)
                                .add_service(reflection_service);
  router.serve(address).await?;

  Ok(())
}

pub fn make_span(request: &Request<Body>) -> Span {
  let headers = request.headers();
  info_span!("Incoming request", ?headers, trace_id = tracing::field::Empty)
}

pub fn link_parent_trace(request: Request<Body>) -> Request<Body> {
  let parent_trace_context =
    get_text_map_propagator(|propagator| propagator.extract(&MetadataExtractor(request.headers())));

  if let Err(error) = Span::current().set_parent(parent_trace_context) {
    warn!("Failed linking span with parent : {error}")
  }

  request
}

/// Responsible for extracting trace context from the metadata map of the incoming gRPC request.
struct MetadataExtractor<'metadata_extractor>(&'metadata_extractor HeaderMap);

impl<'metadata_extractor> Extractor for MetadataExtractor<'metadata_extractor> {
  fn keys(&self) -> Vec<&str> { self.0.keys().map(|key| key.as_str()).collect() }

  fn get(&self, key: &str) -> Option<&str> { self.0.get(key).and_then(|value| value.to_str().ok()) }
}

/// Responsible for setting the health status of this gRPC server.
/// Currently, we keep it simple : we check health of every healthcheckable periodically.
/// And, if any of them turns out to be unhealthy, we set all the services as not serving.
async fn set_health_status(health_reporter: HealthReporter,
                           healthcheckables: Vec<impl Healthcheckable + Send + Sync>) {
  // Set all the services as serving, initially.
  health_reporter.set_serving::<AuthApiServiceServer<AuthAPI>>().await;

  loop {
    sleep(Duration::from_secs(1)).await;

    let mut everything_healthy = true;

    for healthcheckable in &healthcheckables {
      if healthcheckable.healthcheck().await.is_err() {
        everything_healthy = false;
      }
    }

    match everything_healthy {
      | true => health_reporter.set_serving::<AuthApiServiceServer<AuthAPI>>().await,
      | _ => health_reporter.set_not_serving::<AuthApiServiceServer<AuthAPI>>().await
    }
  }
}
