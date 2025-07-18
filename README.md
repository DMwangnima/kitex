# CloudWeGo-Kitex

English | [中文](README_cn.md)

[![Release](https://img.shields.io/github/v/release/cloudwego/kitex)](https://github.com/cloudwego/kitex/releases)
[![WebSite](https://img.shields.io/website?up_message=cloudwego&url=https%3A%2F%2Fwww.cloudwego.io%2F)](https://www.cloudwego.io/)
[![License](https://img.shields.io/github/license/cloudwego/kitex)](https://github.com/cloudwego/kitex/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudwego/kitex)](https://goreportcard.com/report/github.com/cloudwego/kitex)
[![OpenIssue](https://img.shields.io/github/issues/cloudwego/kitex)](https://github.com/cloudwego/kitex/issues)
[![ClosedIssue](https://img.shields.io/github/issues-closed/cloudwego/kitex)](https://github.com/cloudwego/kitex/issues?q=is%3Aissue+is%3Aclosed)
![Stars](https://img.shields.io/github/stars/cloudwego/kitex)
![Forks](https://img.shields.io/github/forks/cloudwego/kitex)

Kitex [kaɪt'eks] is a **high-performance** and **strong-extensibility** Go RPC framework that helps developers build microservices. If the performance and extensibility are the main concerns when you develop microservices, Kitex can be a good choice.

## Basic Features

- **High Performance**

Kitex integrates [Netpoll](https://github.com/cloudwego/netpoll), a high-performance network library, which offers significant performance advantage over [go net](https://pkg.go.dev/net).

- **Extensibility**

Kitex provides many interfaces with default implementation for users to customize. You can extend or inject them into Kitex to fulfill your needs (please refer to the framework extension section below).

- **Multi-message Protocol**

Kitex is designed to be extensible to support multiple RPC messaging protocols. The initial release contains support for **Thrift**, **Kitex Protobuf** and **gRPC**, in which Kitex Protobuf is a Kitex custom Protobuf messaging protocol with a protocol format similar to Thrift. Kitex also supports developers extending their own messaging protocols.

- **Multi-transport Protocol**

For service governance, Kitex supports **TTHeader** and **HTTP2**. TTHeader can be used in conjunction with Thrift and Kitex Protobuf.

- **Multi-message Type**

Kitex supports **PingPong**, **One-way**, and **Bidirectional Streaming**. Among them, One-way currently only supports Thrift protocol.

- **Service Governance**

Kitex integrates service governance modules such as service registry, service discovery, load balancing, circuit breaker, rate limiting, retry, monitoring, tracing, logging, diagnosis, etc. Most of these have been provided with default extensions, giving users the option to integrate them as desired.

- **Code Generation**

Kitex has built-in code generation tools that support generating **Thrift**, **Protobuf**, and scaffold code.

## Documentation

- [**Getting Started**](https://www.cloudwego.io/docs/kitex/getting-started/)

- **User Guide**

  - **Basic Features**
  
    Including Message Type, Supported Protocols, Directly Invoke, Connection Pool, Timeout Control, Request Retry, LoadBalancer, Circuit Breaker, Rate Limiting, Instrumentation Control, Logging and HttpResolver.[[more]](https://www.cloudwego.io/docs/kitex/tutorials/basic-feature/)
    
  - **Governance Features**
  
    Supporting Service Discovery, Monitoring, Tracing and Customized Access Control.[[more]](https://www.cloudwego.io/docs/kitex/tutorials/service-governance/)
    
  - **Advanced Features**
  
    Supporting Generic Call and Server SDK Mode.[[more]](https://www.cloudwego.io/docs/kitex/tutorials/advanced-feature/)
    
  - **Code Generation**
  
    Including Code Generation Tool and Combined Service.[[more]](https://www.cloudwego.io/docs/kitex/tutorials/code-gen/)
    
  - **Framework Extension**
  
    Providing Middleware Extensions, Suite Extensions, Service Registry, Service Discovery, Customize LoadBalancer, Monitoring, Logging, Codec, Transport Module, Transport Pipeline, Metadata Transparent Transmission, Diagnosis Module.[[more]](https://www.cloudwego.io/docs/kitex/tutorials/framework-exten/)
  
- **Reference**

  - For Transport Protocol, Exception Instruction and Version Specification, please refer to [doc](https://www.cloudwego.io/docs/kitex/reference/).

- **Best Practice**
  - Kitex best practices in production, such as graceful shutdown, error handling, integration testing. [More](https://www.cloudwego.io/docs/kitex/best-practice/)

- **FAQ**

  - Please refer to [FAQ](https://www.cloudwego.io/docs/kitex/faq/).

## Performance

Performance benchmark can only provide limited reference. In production, there are many factors can affect actual performance.

We provide the [kitex-benchmark](https://github.com/cloudwego/kitex-benchmark) project to track and compare the performance of Kitex and other frameworks under different conditions for reference.

## Related Projects

- [Netpoll](https://github.com/cloudwego/netpoll): A high-performance network library.
- [kitex-contrib](https://github.com/kitex-contrib): A partial extension library of Kitex, which users can integrate into Kitex through options according to their needs.
- [kitex-examples](https://github.com/cloudwego/kitex-examples): Examples of Kitex showcasing various features.
- [biz-demo](https://github.com/cloudwego/biz-demo): Business demos using Kitex.

## Blogs
- [Enhancing Performance in Microservice Architecture with Kitex](https://www.cloudwego.io/blog/2024/01/29/enhancing-performance-in-microservice-architecture-with-kitex/)
- [CloudWeGo: A leading practice for building enterprise cloud native middleware!](https://www.cloudwego.io/blog/2023/06/15/cloudwego-a-leading-practice-for-building-enterprise-cloud-native-middleware/)
- [Kitex: Unifying Open Source Practice for a High-Performance RPC Framework](https://www.cloudwego.io/blog/2022/09/30/kitex-unifying-open-source-practice-for-a-high-performance-rpc-framework/)
- [Performance Optimization on Kitex](https://www.cloudwego.io/blog/2021/09/23/performance-optimization-on-kitex/)
- [ByteDance Practice on Go Network Library](https://www.cloudwego.io/blog/2020/05/24/bytedance-practices-on-go-network-library/)
- [Getting Started With Kitex's Practice: Performance Testing Guide](https://www.cloudwego.io/blog/2021/11/24/getting-started-with-kitexs-practice-performance-testing-guide/)

## Contributing

Contributor guide: [Contributing](https://github.com/cloudwego/kitex/blob/develop/CONTRIBUTING.md).

## License

Kitex is distributed under the [Apache License, version 2.0](https://github.com/cloudwego/kitex/blob/develop/LICENSE). The licenses of third party dependencies of Kitex are explained [here](https://github.com/cloudwego/kitex/blob/develop/licenses).

## Community
- Email: [conduct@cloudwego.io](conduct@cloudwego.io)
- How to become a member: [COMMUNITY MEMBERSHIP](https://github.com/cloudwego/community/blob/main/COMMUNITY_MEMBERSHIP.md)
- Issues: [Issues](https://github.com/cloudwego/kitex/issues)
- Discord: Join community with [Discord Channel](https://discord.gg/jceZSE7DsW).
- Lark: Scan the QR code below with [Lark](https://www.larksuite.com/zh_cn/download) to join our CloudWeGo/kitex user group.

  ![LarkGroup](images/lark_group.png)

## Landscapes

<p align="center">
<img src="https://landscape.cncf.io/images/cncf-landscape-horizontal-color.svg" width="150"/>&nbsp;&nbsp;<img src="https://www.cncf.io/wp-content/uploads/2023/04/cncf-main-site-logo.svg" width="200"/>
<br/><br/>
CloudWeGo enriches the <a href="https://landscape.cncf.io/">CNCF CLOUD NATIVE Landscape</a>.
</p>
