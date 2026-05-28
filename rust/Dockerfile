# Download and cache dependencies.
FROM --platform=$BUILDPLATFORM rust:1.95.0-alpine AS dependencies-planner

RUN cargo install cargo-chef

WORKDIR /app

COPY rust-toolchain.toml Cargo.toml Cargo.lock ./
RUN \
  --mount=type=cache,target=/usr/local/cargo/registry \
  --mount=type=cache,target=/usr/local/cargo/git \
  --mount=type=cache,target=/app/target \
  mkdir src && \
  echo "fn main( ) { }" > src/main.rs && \
  cargo chef prepare --recipe-path recipe.json

# Build the binary.
FROM --platform=$BUILDPLATFORM rust:1.95.0-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM

RUN apk add --no-cache protobuf=31.1-r1 protobuf-dev=31.1-r1 musl-dev=1.2.5-r23

COPY --from=dependencies-planner /usr/local/cargo/bin/cargo-chef /usr/local/cargo/bin/cargo-chef

RUN case "$TARGETPLATFORM" in \
  "linux/amd64") echo "x86_64-unknown-linux-musl" > /target ;; \
  "linux/arm64") echo "aarch64-unknown-linux-musl" > /target ;; \
  *) echo "Unsupported platform: $TARGETPLATFORM" && exit 1 ;; \
  esac

RUN rustup target add "$(cat /target)"

WORKDIR /app

COPY --from=dependencies-planner /app/recipe.json recipe.json
RUN \
  --mount=type=cache,target=/usr/local/cargo/registry \
  --mount=type=cache,target=/usr/local/cargo/git \
  cargo chef cook --release --recipe-path recipe.json

COPY . .

RUN \
  --mount=type=cache,target=/usr/local/cargo/registry \
  --mount=type=cache,target=/usr/local/cargo/git \
  --mount=type=cache,target=/app/target \
  cargo build --release --target "$(cat /target)" && \
  cp "target/$(cat /target)/release/openmedia-backend" /openmedia-backend

# Package the binary.
FROM scratch AS packager

USER 1000

COPY --from=builder /openmedia-backend /openmedia-backend

ENTRYPOINT ["/openmedia-backend"]
