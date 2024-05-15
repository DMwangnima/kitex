// Code generated by Kitex v0.9.1. DO NOT EDIT.

package serverreflection

import (
	"context"
	"errors"
	"fmt"
	client "github.com/cloudwego/kitex/client"
	reflection_v1alpha "github.com/cloudwego/kitex/pkg/reflection/grpc/reflection_v1alpha"
	kitex "github.com/cloudwego/kitex/pkg/serviceinfo"
	streaming "github.com/cloudwego/kitex/pkg/streaming"
	proto "google.golang.org/protobuf/proto"
)

var errInvalidMessageType = errors.New("invalid message type for service method handler")

var serviceMethods = map[string]kitex.MethodInfo{
	"ServerReflectionInfo": kitex.NewMethodInfo(
		serverReflectionInfoHandler,
		newServerReflectionInfoArgs,
		newServerReflectionInfoResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingBidirectional),
	),
}

var (
	serverReflectionServiceInfo                = NewServiceInfo()
	serverReflectionServiceInfoForClient       = NewServiceInfoForClient()
	serverReflectionServiceInfoForStreamClient = NewServiceInfoForStreamClient()
)

// for server
func serviceInfo() *kitex.ServiceInfo {
	return serverReflectionServiceInfo
}

// for client
func serviceInfoForStreamClient() *kitex.ServiceInfo {
	return serverReflectionServiceInfoForStreamClient
}

// for stream client
func serviceInfoForClient() *kitex.ServiceInfo {
	return serverReflectionServiceInfoForClient
}

// NewServiceInfo creates a new ServiceInfo containing all methods
func NewServiceInfo() *kitex.ServiceInfo {
	return newServiceInfo(true, true, true)
}

// NewServiceInfo creates a new ServiceInfo containing non-streaming methods
func NewServiceInfoForClient() *kitex.ServiceInfo {
	return newServiceInfo(false, false, true)
}
func NewServiceInfoForStreamClient() *kitex.ServiceInfo {
	return newServiceInfo(true, true, false)
}

func newServiceInfo(hasStreaming bool, keepStreamingMethods bool, keepNonStreamingMethods bool) *kitex.ServiceInfo {
	serviceName := "ServerReflection"
	handlerType := (*reflection_v1alpha.ServerReflection)(nil)
	methods := map[string]kitex.MethodInfo{}
	for name, m := range serviceMethods {
		if m.IsStreaming() && !keepStreamingMethods {
			continue
		}
		if !m.IsStreaming() && !keepNonStreamingMethods {
			continue
		}
		methods[name] = m
	}
	extra := map[string]interface{}{
		"PackageName": "grpc.reflection.v1alpha",
	}
	if hasStreaming {
		extra["streaming"] = hasStreaming
	}
	svcInfo := &kitex.ServiceInfo{
		ServiceName:     serviceName,
		HandlerType:     handlerType,
		Methods:         methods,
		PayloadCodec:    kitex.Protobuf,
		KiteXGenVersion: "v0.9.1",
		Extra:           extra,
	}
	return svcInfo
}

func serverReflectionInfoHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	streamingArgs, ok := arg.(*streaming.Args)
	if !ok {
		return errInvalidMessageType
	}
	st := streamingArgs.Stream
	stream := &serverReflectionServerReflectionInfoServer{st}
	return handler.(reflection_v1alpha.ServerReflection).ServerReflectionInfo(stream)
}

type serverReflectionServerReflectionInfoClient struct {
	streaming.Stream
}

func (x *serverReflectionServerReflectionInfoClient) DoFinish(err error) {
	if finisher, ok := x.Stream.(streaming.WithDoFinish); ok {
		finisher.DoFinish(err)
	} else {
		panic(fmt.Sprintf("streaming.WithDoFinish is not implemented by %T", x.Stream))
	}
}
func (x *serverReflectionServerReflectionInfoClient) Send(m *reflection_v1alpha.ServerReflectionRequest) error {
	return x.Stream.SendMsg(m)
}
func (x *serverReflectionServerReflectionInfoClient) Recv() (*reflection_v1alpha.ServerReflectionResponse, error) {
	m := new(reflection_v1alpha.ServerReflectionResponse)
	return m, x.Stream.RecvMsg(m)
}

type serverReflectionServerReflectionInfoServer struct {
	streaming.Stream
}

func (x *serverReflectionServerReflectionInfoServer) Send(m *reflection_v1alpha.ServerReflectionResponse) error {
	return x.Stream.SendMsg(m)
}

func (x *serverReflectionServerReflectionInfoServer) Recv() (*reflection_v1alpha.ServerReflectionRequest, error) {
	m := new(reflection_v1alpha.ServerReflectionRequest)
	return m, x.Stream.RecvMsg(m)
}

func newServerReflectionInfoArgs() interface{} {
	return &ServerReflectionInfoArgs{}
}

func newServerReflectionInfoResult() interface{} {
	return &ServerReflectionInfoResult{}
}

type ServerReflectionInfoArgs struct {
	Req *reflection_v1alpha.ServerReflectionRequest
}

func (p *ServerReflectionInfoArgs) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetReq() {
		p.Req = new(reflection_v1alpha.ServerReflectionRequest)
	}
	return p.Req.FastRead(buf, _type, number)
}

func (p *ServerReflectionInfoArgs) FastWrite(buf []byte) (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.FastWrite(buf)
}

func (p *ServerReflectionInfoArgs) Size() (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.Size()
}

func (p *ServerReflectionInfoArgs) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
		return out, nil
	}
	return proto.Marshal(p.Req)
}

func (p *ServerReflectionInfoArgs) Unmarshal(in []byte) error {
	msg := new(reflection_v1alpha.ServerReflectionRequest)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Req = msg
	return nil
}

var ServerReflectionInfoArgs_Req_DEFAULT *reflection_v1alpha.ServerReflectionRequest

func (p *ServerReflectionInfoArgs) GetReq() *reflection_v1alpha.ServerReflectionRequest {
	if !p.IsSetReq() {
		return ServerReflectionInfoArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *ServerReflectionInfoArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *ServerReflectionInfoArgs) GetFirstArgument() interface{} {
	return p.Req
}

type ServerReflectionInfoResult struct {
	Success *reflection_v1alpha.ServerReflectionResponse
}

var ServerReflectionInfoResult_Success_DEFAULT *reflection_v1alpha.ServerReflectionResponse

func (p *ServerReflectionInfoResult) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetSuccess() {
		p.Success = new(reflection_v1alpha.ServerReflectionResponse)
	}
	return p.Success.FastRead(buf, _type, number)
}

func (p *ServerReflectionInfoResult) FastWrite(buf []byte) (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.FastWrite(buf)
}

func (p *ServerReflectionInfoResult) Size() (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.Size()
}

func (p *ServerReflectionInfoResult) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
		return out, nil
	}
	return proto.Marshal(p.Success)
}

func (p *ServerReflectionInfoResult) Unmarshal(in []byte) error {
	msg := new(reflection_v1alpha.ServerReflectionResponse)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Success = msg
	return nil
}

func (p *ServerReflectionInfoResult) GetSuccess() *reflection_v1alpha.ServerReflectionResponse {
	if !p.IsSetSuccess() {
		return ServerReflectionInfoResult_Success_DEFAULT
	}
	return p.Success
}

func (p *ServerReflectionInfoResult) SetSuccess(x interface{}) {
	p.Success = x.(*reflection_v1alpha.ServerReflectionResponse)
}

func (p *ServerReflectionInfoResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *ServerReflectionInfoResult) GetResult() interface{} {
	return p.Success
}

type kClient struct {
	c client.Client
}

func newServiceClient(c client.Client) *kClient {
	return &kClient{
		c: c,
	}
}

func (p *kClient) ServerReflectionInfo(ctx context.Context) (ServerReflection_ServerReflectionInfoClient, error) {
	streamClient, ok := p.c.(client.Streaming)
	if !ok {
		return nil, fmt.Errorf("client not support streaming")
	}
	res := new(streaming.Result)
	err := streamClient.Stream(ctx, "ServerReflectionInfo", nil, res)
	if err != nil {
		return nil, err
	}
	stream := &serverReflectionServerReflectionInfoClient{res.Stream}
	return stream, nil
}
