/*
 *
 * Copyright 2014 gRPC authors.
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
 *
 * This file may have been modified by CloudWeGo authors. All CloudWeGo
 * Modifications are Copyright 2021 CloudWeGo Authors.
 */

package errors

import (
	"github.com/cloudwego/kitex/pkg/kerrors"
)

// this package contains all the errors suitable for Kitex errors model

var (
	errStream               = kerrors.ErrStreaming.WithChild("gRPC stream")
	ErrHTTP2Stream          = errStream.WithChild("HTTP2Stream error when parsing frame")
	ErrClosedWithoutTrailer = errStream.WithChild("")
	ErrMiddleHeader         = errStream.WithChild("")
	ErrDecodeHeader         = errStream.WithChild("")
	ErrProxyTrailer         = errStream.WithChild("")
	ErrRecvRstStream        = errStream.WithChild("")
	ErrRecvProxyRstStream   = errStream.WithChild("")
	ErrStreamDrain          = errStream.WithChild("")
	ErrStreamFlowControl    = errStream.WithChild("")

	errConnection          = kerrors.ErrStreaming.WithChild("gRPC connection")
	ErrHTTP2Connection     = errConnection.WithChild("read HTTP2 Frame failed")
	ErrEstablishConnection = errConnection.WithChild("establish connection failed")
	ErrHandleGoAway        = errConnection.WithChild("handles GoAway Frame failed")
	ErrKeepAlive           = errConnection.WithChild("keepalive related")
	ErrOperateHeaders      = errConnection.WithChild("operate Header Frame failed")
	ErrNoActiveStream      = errConnection.WithChild("no active stream")
)
