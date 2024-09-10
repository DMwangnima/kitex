package stream_v2

var ServerTpl = `// Code generated by Kitex {{.Version}}. DO NOT EDIT.
package {{ToLower .ServiceName}}

import (
	{{- range $path, $aliases := .Imports}}
		{{- if not $aliases}}
			"{{$path}}"
		{{- else}}
			{{- range $alias, $is := $aliases}}
				{{$alias}} "{{$path}}"
			{{- end}}
		{{- end}}
	{{- end}}
)
{{- $protocol := .Protocol | ToString | ToLower}}

type ClientStreamingServer[Req, Res any] streamx.ClientStreamingServer[{{$protocol}}.Header, {{$protocol}}.Trailer, Req, Res]
type ServerStreamingServer[Res any] streamx.ServerStreamingServer[{{$protocol}}.Header, {{$protocol}}.Trailer, Res]
type BidiStreamingServer[Req, Res any] streamx.BidiStreamingServer[{{$protocol}}.Header, {{$protocol}}.Trailer, Req, Res]

type ServerInterface interface {
{{- range .AllMethods}}
{{- $unary := and (not .ServerStreaming) (not .ClientStreaming)}}
{{- $clientSide := and .ClientStreaming (not .ServerStreaming)}}
{{- $serverSide := and (not .ClientStreaming) .ServerStreaming}}
{{- $bidiSide := and .ClientStreaming .ServerStreaming}}
{{- $arg := index .Args 0}}
    {{.Name}}{{- if $unary}}(ctx context.Context, req {{$arg.Type}}) ({{.Resp.Type}}, error)
             {{- else if $clientSide}}(ctx context.Context, stream ClientStreamingServer[{{NotPtr $arg.Type}}, {{NotPtr .Resp.Type}}]) ({{.Resp.Type}}, error)
             {{- else if $serverSide}}(ctx context.Context, req {{$arg.Type}}, stream ServerStreamingServer[{{NotPtr .Resp.Type}}]) error
             {{- else if $bidiSide}}(ctx context.Context, stream BidiStreamingServer[{{NotPtr $arg.Type}}, {{NotPtr .Resp.Type}}]) error
             {{- end}}
{{- end}}
}

func NewServer(handler ServerInterface, opts ...streamxserver.ServerOption) (streamxserver.Server, error) {
	var options []streamxserver.ServerOption
	options = append(options, opts...)

	sp, err := {{$protocol}}.NewServerProvider(serviceInfo)
	if err != nil {
		return nil, err
	}
	options = append(options, streamxserver.WithProvider(sp))
	svr := streamxserver.NewServer(options...)
	if err := svr.RegisterService(serviceInfo, handler); err != nil {
		return nil, err
	}
	return svr, nil
}

func RegisterService(svr server.Server, handler {{call .ServiceTypeName}}, opts ...server.RegisterOption) error {
	return svr.RegisterService(serviceInfo, handler, opts...)
}
`
