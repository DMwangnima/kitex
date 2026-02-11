/*
 * Copyright 2026 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package meta

import (
	"context"

	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/metadata"
	"github.com/cloudwego/kitex/pkg/remote/transmeta"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
)

var HTTP2MetaClientEventHandler = rpcinfo.ClientStreamEventHandler{
	HandleCreateStreamEvent: HTTP2MetaHandleCreateStreamEvent,
}

var HTTP2MetaServerEventHandler = rpcinfo.ServerStreamEventHandler{
	HandleAcceptStreamEvent: HTTP2MetaHandleAcceptStreamEvent,
}

func HTTP2MetaHandleCreateStreamEvent(ctx context.Context, ri rpcinfo.RPCInfo) (context.Context, error) {
	// append more meta into current metadata
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	metaAppend(md, transmeta.HTTPDestService, ri.To().ServiceName())
	metaAppend(md, transmeta.HTTPDestMethod, ri.To().Method())
	metaAppend(md, transmeta.HTTPSourceService, ri.From().ServiceName())
	metaAppend(md, transmeta.HTTPSourceMethod, ri.From().Method())
	return metadata.NewOutgoingContext(ctx, md), nil
}

func HTTP2MetaHandleAcceptStreamEvent(ctx context.Context, ri rpcinfo.RPCInfo) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, nil
	}
	ci := rpcinfo.AsMutableEndpointInfo(ri.From())
	if ci != nil {
		if v := md[transmeta.HTTPSourceService]; len(v) != 0 {
			ci.SetServiceName(v[0])
		}
		if v := md[transmeta.HTTPSourceMethod]; len(v) != 0 {
			ci.SetMethod(v[0])
		}
	}
	return ctx, nil
}

// same as md.Append without strings.ToLower coz k must always be lower cases
func metaAppend(m metadata.MD, k, v string) {
	m[k] = append(m[k], v)
}
