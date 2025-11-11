# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Kitex is a high-performance, extensible Go RPC framework developed by ByteDance (CloudWeGo). It supports multiple protocols (Thrift, Protobuf, gRPC), provides comprehensive service governance features, and uses Netpoll for high-performance networking.

**Key characteristics:**
- Performance-critical codebase with extensive optimization (object pooling, zero-copy, fast codecs)
- Highly extensible through interfaces and options pattern
- Multi-protocol support with protocol-agnostic transport layer
- Supports both code-generated and generic (reflection-based) service calls

## Development Commands

### Testing

```bash
# Run all unit tests with race detection
go test -race ./...

# Run tests for a specific package
go test -race ./client/...
go test -race ./pkg/loadbalance/...

# Run a single test
go test -race -run TestClientCall ./client/

# Run tests with coverage
go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Run benchmarks (development only, not for actual benchmarking)
go test -bench=. -benchmem -run=none ./... -benchtime=100ms
```

### Linting & Formatting

```bash
# Run golangci-lint (required before PR)
golangci-lint run

# Format code with gofumpt (stricter than gofmt)
gofumpt -l -w .

# Organize imports with goimports
goimports -local github.com/cloudwego/kitex -w .
```

### Code Generation

```bash
# Install the kitex tool
go install ./tool/cmd/kitex

# Install thriftgo (required for Thrift IDL)
go install github.com/cloudwego/thriftgo@main

# Generate code from Thrift IDL
kitex -module <module_name> <idl_file.thrift>

# Generate code from Protobuf IDL
kitex -type protobuf -module <module_name> <idl_file.proto>

# Generate with custom options
kitex -module <module_name> -service <service_name> <idl_file.thrift>
```

### Building

```bash
# Build the kitex tool
go build -o kitex ./tool/cmd/kitex

# Install dependencies
go mod tidy
```

### Mock Generation

To generate mocks for testing:

1. Install gomock: `go install github.com/golang/mock/mockgen@latest`
2. Edit `internal/mocks/update.sh` to add your interface
3. Run: `sh internal/mocks/update.sh`

Use [Mockey](https://github.com/bytedance/mockey) for mocking functions (see examples in tests).

## High-Level Architecture

Kitex follows a **layered onion architecture**:

```
Client/Server API Layer
    ↓
Endpoint & Middleware Layer (extensibility point)
    ↓
Transport Layer (protocol-agnostic handlers)
    ↓
Codec Layer (Thrift/Protobuf/gRPC encoding)
    ↓
Network Layer (Netpoll)
```

### Key Subsystems

#### 1. Client & Server (`/client`, `/server`)
- **Entry points**: `Client.Call()` and `Server.RegisterService()`
- Use **options pattern** for configuration (every feature is opt-in via `WithXXX()` options)
- Both build **middleware chains** around core endpoint handlers
- Client manages connection pool, load balancing, retry, circuit breaking
- Server manages service registry, request routing, concurrency limits

#### 2. Endpoint & Middleware (`/pkg/endpoint`)
- **Core abstraction**: `type Endpoint func(ctx, req, resp) error`
- **Middleware**: `func(Endpoint) Endpoint` - decorator pattern for extensibility
- Middleware execution order matters (built in reverse, executed in forward order)
- Three types: universal, unary-specific, stream-specific

**Client middleware order** (outermost → innermost):
```
Service CB → XDS Router → RPC Timeout → Custom Unary → Context →
Custom → ACL → Resolve → Instance CB → Proxy → IO Error Handler
```

**Server middleware order**:
```
Stream → Timeout → Custom → Core (ACL + Error Handler) → Business Handler
```

#### 3. Transport Layer (`/pkg/remote`)
- **Protocol-agnostic** transport abstraction inspired by Netty
- **TransHandler**: Interface for I/O operations with lifecycle callbacks
- **TransPipeline**: Chain of InboundHandlers (reads) and OutboundHandlers (writes)
- **BoundHandler**: Extension point for adding custom processing (like Netty's ChannelHandler)

Transport implementations:
- `netpoll`: Default high-performance transport using Netpoll
- `nphttp2/grpc`: HTTP2 and gRPC support
- `ttstream`: TTHeader streaming
- `netpollmux`: Multiplexed connections

#### 4. Codec Layer (`/pkg/remote/codec`)
- **Pluggable codecs** for different protocols
- Default codec auto-detects protocol via magic bytes
- **Thrift codecs** (priority): Frugal (fastest) → FastCodec → Apache (fallback)
- **Protobuf**: Supports both Kitex Protobuf (with TTHeader) and gRPC
- **Meta-Codec pattern**: Separate encoding for headers and payload

#### 5. Generic Calls (`/pkg/generic`)
- Call services **without generated code** using runtime IDL parsing
- Types: BinaryThrift, MapThrift, JSONThrift, HTTPThrift, JSONPb, BinaryPb
- Uses descriptor providers to parse IDL files at runtime
- Enables dynamic proxy scenarios and API gateways

#### 6. Service Discovery & Load Balancing (`/pkg/discovery`, `/pkg/loadbalance`)
- **Resolver**: Service discovery abstraction (DNS, Consul, Nacos, etc.)
- **Loadbalancer**: Returns a Picker for instance selection
- **LBCache** (`/pkg/loadbalance/lbcache`): Caches pickers per destination, updates via event bus
- Built-in algorithms: Weighted RR, Weighted Random, Consistent Hash, Interleaved WRR

#### 7. Retry & Circuit Breaker (`/pkg/retry`, `/pkg/circuitbreak`)
- **Retry types**: Failure retry, backup request, mixed
- Policies are **method-level** using container pattern
- Circuit breaker integration prevents retry storms
- **RPCInfo pooling** - careful lifecycle management during retries to avoid data races

#### 8. ServiceInfo & Code Generation (`/pkg/serviceinfo`, `/tool`)
- **ServiceInfo**: Metadata about service (methods, codec, streaming mode)
- **kitex tool**: Standalone CLI or plugin for thriftgo/protoc
- Generates: service interfaces, client/server stubs, ServiceInfo metadata, fast serialization code

### Critical Patterns

1. **Options Pattern Everywhere**: All features use `WithXXX()` options for extensibility
2. **Dual Middleware Systems**: Separate chains for client vs server, unary vs streaming
3. **Object Pooling**: RPCInfo, ByteBuffer, connections - reused for performance
4. **Event Bus**: Dynamic configuration changes propagate via event-driven architecture
5. **Immutable Views**: RPCInfo exposed as read-only externally, mutable only internally
6. **Protocol Purification**: `purifyProtocol()` normalizes mixed flags before processing

## Code Organization

```
/client          - Client implementation and options
/server          - Server implementation and options
/pkg             - Public APIs (use these for extensions)
  /endpoint      - Middleware abstraction
  /remote        - Transport layer (codec, bound handlers, trans handlers)
  /generic       - Generic call support
  /loadbalance   - Load balancing algorithms
  /discovery     - Service discovery
  /retry         - Retry policies
  /circuitbreak  - Circuit breaker
  /rpcinfo       - RPC metadata
  /serviceinfo   - Service metadata
  /streaming     - Streaming support
/internal        - Framework internals (not for external use)
  /client        - Internal client logic
  /server        - Internal server logic
  /mocks         - Generated mocks for testing
/tool            - Code generation tool (kitex CLI)
  /cmd/kitex     - Main CLI entry point
/transport       - Deprecated (use /pkg/remote)
```

**Generated code location**: `kitex_gen/[package]/` (excluded from linting)

## Testing Guidelines

- **Unit tests**: Use `*_test.go` files alongside implementation
- **Mocking**: Use [gomock](https://github.com/golang/mock) for interfaces, [Mockey](https://github.com/bytedance/mockey) for functions
- **Race detection**: Always run with `-race` flag
- **Coverage**: Aim for meaningful coverage, not just high percentages
- **Benchmark tests**: Include for performance-critical code, but mark as non-critical (they verify code compiles)

## Contribution Guidelines

Based on CONTRIBUTING.md:

1. **Branch**: Create feature branch from `main`: `git checkout -b my-fix-branch main`
2. **Commit messages**: Follow [AngularJS commit conventions](https://docs.google.com/document/d/1QrDFcIiPjSLDn3EL15IJygNPiHORgU1_OOAqWjiDU5Y/edit)
3. **Code style**:
   - Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
   - Use `gofumpt` (stricter than `gofmt`)
   - Use `goimports` with local prefix: `github.com/cloudwego/kitex`
4. **Linting**: Must pass `golangci-lint run` before PR
5. **Tests**: Include appropriate test cases for all changes
6. **PR target**: Submit PR to `kitex:main`

## Important Notes for Development

### When Adding New Features

1. **Use options pattern**: Add `WithXXX()` option functions, never break existing APIs
2. **Add middleware**: For cross-cutting concerns, add via `endpoint.Middleware`
3. **Extend interfaces**: Prefer interface extension over modifying existing interfaces
4. **Object pooling**: For hot paths, consider object pooling to reduce GC pressure
5. **Avoid breaking changes**: Kitex prioritizes backward compatibility

### When Modifying Core Paths

- **Client call path**: Be extremely careful - this is hot path affecting all RPCs
- **Codec layer**: Ensure protocol compatibility, test with all protocol types
- **Middleware chain**: Order matters - understand execution sequence before modifying
- **RPCInfo lifecycle**: Improper handling can cause data races during retry

### Protocol Implementation

- **Thrift**: Three codec implementations (Frugal, Fast, Apache) - test all
- **Protobuf**: Two modes (Kitex Protobuf + TTHeader, gRPC + HTTP2) - ensure both work
- **TTHeader**: Used for service governance metadata - preserve backward compatibility
- **Streaming**: Special handling in middleware and codecs - test all streaming modes

### Performance Considerations

- Use object pools for frequently allocated objects
- Prefer zero-copy where possible (Netpoll's LinkBufferNocopy)
- Benchmark before/after for performance-critical changes
- Avoid allocations in hot paths
- Be mindful of closure captures in middleware

### Common Pitfalls

1. **Middleware order**: Adding middleware in wrong position breaks features
2. **RPCInfo modification during retry**: Can cause data races if not pooled correctly
3. **Codec selection**: Wrong codec flags can cause protocol mismatch
4. **Connection pool lifecycle**: Improper handling causes connection leaks
5. **Streaming args**: Require special `streaming.Args` wrapper, not direct endpoint call

## Documentation

- **Official docs**: https://www.cloudwego.io/docs/kitex/
- **Examples**: https://github.com/cloudwego/kitex-examples
- **Extensions**: https://github.com/kitex-contrib

## Related Projects

- **Netpoll**: High-performance network library (Kitex's network layer)
- **kitex-contrib**: Extension libraries (discovery, tracing, config, etc.)
- **kitex-examples**: Feature examples and tutorials
- **kitex-tests**: Integration test suite
