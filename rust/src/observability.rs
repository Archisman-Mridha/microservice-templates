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
  opentelemetry::global,
  opentelemetry_appender_tracing::layer::OpenTelemetryTracingBridge,
  opentelemetry_otlp::WithExportConfig,
  opentelemetry_sdk::{
    Resource,
    logs::{SdkLoggerProvider, SimpleLogProcessor},
    metrics::{SdkMeterProvider, Temporality},
    propagation::TraceContextPropagator,
    trace::SdkTracerProvider
  },
  tracing::{Level, level_filters::LevelFilter},
  tracing_subscriber::{EnvFilter, Layer, layer::SubscriberExt}
};

const DEFAULT_TRACER_NAME: &str = "default";

pub fn setup_observability(service_name: String,
                           log_level: Level,
                           otlp_collector_url: &str)
                           -> anyhow::Result<()> {
  // Construct the OTEL resource that represents this backend server.
  let resource = Resource::builder().with_service_name(service_name).build();

  // Setup trace exporter.

  global::set_text_map_propagator(TraceContextPropagator::new());

  let trace_exporter = opentelemetry_otlp::SpanExporter::builder().with_tonic()
                                                                  .with_endpoint(otlp_collector_url)
                                                                  .build()?;
  let tracer_provider = SdkTracerProvider::builder().with_resource(resource.clone())
                                                    .with_batch_exporter(trace_exporter)
                                                    .build();
  global::set_tracer_provider(tracer_provider.clone());

  let tracer = global::tracer(DEFAULT_TRACER_NAME);

  let trace_emitter_layer = tracing_opentelemetry::layer().with_tracer(tracer);

  // Setup metric exporter.

  let metric_exporter =
    opentelemetry_otlp::MetricExporterBuilder::new().with_temporality(Temporality::Delta)
                                                    .with_tonic()
                                                    .with_endpoint(otlp_collector_url)
                                                    .build()?;
  let meter_provider = SdkMeterProvider::builder().with_resource(resource.clone())
                                                  .with_periodic_exporter(metric_exporter)
                                                  .build();
  global::set_meter_provider(meter_provider);

  // Setup log exporter.

  let log_exporter = opentelemetry_stdout::LogExporter::default();
  let log_processor = SimpleLogProcessor::new(log_exporter);
  let logger_provider =
    SdkLoggerProvider::builder().with_resource(resource).with_log_processor(log_processor).build();

  // To prevent a telemetry-induced-telemetry loop, OpenTelemetry's own internal
  // logging is properly suppressed. However, logs emitted by external
  // components (such as reqwest, tonic, etc.) are not suppressed as they do
  // not propagate OpenTelemetry context. Until this issue is addressed
  // (https://github.com/open-telemetry/opentelemetry-rust/issues/2877),
  // filtering like this is the best way to suppress such logs.
  //
  // NOTE : This filtering will also drop logs from these components even when
  // they are used outside of the OTLP Exporter.
  let log_filter = EnvFilter::new("info").add_directive("hyper=off".parse()?)
                                         .add_directive("tonic=off".parse()?)
                                         .add_directive("h2=off".parse()?)
                                         .add_directive("reqwest=off".parse()?);

  let log_emitter_layer = OpenTelemetryTracingBridge::new(&logger_provider).with_filter(log_filter);

  tracing_subscriber::registry().with(LevelFilter::from_level(log_level))
                                .with(EnvFilter::from_default_env())
                                .with(trace_emitter_layer)
                                .with(log_emitter_layer);

  Ok(())
}
