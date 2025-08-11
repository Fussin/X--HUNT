# Core Proxy Engine

This is the core proxy engine for SentinelX. It is a high-performance, extensible MITM proxy for intercepting and modifying network traffic.

## Features

*   High-performance TCP listener with async multiplexing.
*   TLS interception with on-the-fly certificate generation.
*   (Planned) Support for HTTP/1.1, HTTP/2, HTTP/3, WebSocket, and gRPC.
*   (Planned) Traffic transformation engine.
*   (Planned) Session storage and replay.
*   (Planned) Prometheus metrics and OpenTelemetry tracing.

## Getting Started

To build and run the proxy server, from the `core-proxy` directory:

```sh
go run ./cmd/sentinelx-proxy
```

The proxy will start on port 8080.
