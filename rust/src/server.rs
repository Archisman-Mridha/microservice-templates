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
    self,
    api::{AuthAPI, auth_api_service_server::AuthApiServiceServer}
  },
  axum::http::{HeaderMap, Request},
  opentelemetry::{global::get_text_map_propagator, propagation::Extractor},
  std::net::SocketAddr,
  tonic::{body::Body, codec::CompressionEncoding, transport::Server},
  tower::ServiceBuilder,
  tower_http::trace::TraceLayer,
  tracing::{Span, info_span, warn},
  tracing_opentelemetry::OpenTelemetrySpanExt
};

pub async fn start(address: SocketAddr, auth_api: AuthAPI) -> anyhow::Result<()> {
  let auth_api_service =
    AuthApiServiceServer::new(auth_api).send_compressed(CompressionEncoding::Gzip)
                                       .accept_compressed(CompressionEncoding::Gzip);

  let reflection_service = tonic_reflection::server::Builder::configure()
    .register_encoded_file_descriptor_set(auth::api::FILE_DESCRIPTOR_SET)
    .build_v1()?;

  let trace_propagation_layer =
    ServiceBuilder::new().layer(TraceLayer::new_for_grpc().make_span_with(make_span));

  let router = Server::builder().layer(trace_propagation_layer)
                                .add_service(auth_api_service)
                                .add_service(reflection_service);
  router.serve(address).await?;

  Ok(())
}

pub fn make_span(request: &Request<Body>) -> Span {
  let headers = request.headers();
  info_span!("Incoming request", ?headers, trace_id = tracing::field::Empty)
}

pub fn link_parent_trace(request: Request<Body>) -> Request<Body> {
  let parent_ctx =
    get_text_map_propagator(|propagator| propagator.extract(&MetadataExtractor(request.headers())));

  if let Err(error) = Span::current().set_parent(parent_ctx) {
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
