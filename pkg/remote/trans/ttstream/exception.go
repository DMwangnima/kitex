/*
 * Copyright 2024 CloudWeGo Authors
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

package ttstream

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/streaming"
)

const (
	errApplicationExceptionTypeId = 12001
	// protocol related
	errUnexpectedHeaderTypeId = 12002
	errIllegalBizErrTypeId    = 12003
	errIllegalFrameTypeId     = 12004
	errIllegalOperationTypeId = 12005
	errTransportTypeId        = 12006
	// cancel related
	errBizCancelTypeId          = 12007
	errBizCancelWithCauseTypeId = 12008
	errDownstreamCancelTypeId   = 12009
	errUpstreamCancelTypeId     = 12010
	errInternalCancelTypeId     = 12011
	errBizHandlerReturnTypeId   = 12012
	errConnectionClosedTypeId   = 12013
	// timeout related
	errRecvTimeoutTypeId   = 12014
	errStreamTimeoutTypeId = 12015
	// other
	errStreamInternalTypeId = 12016
)

var (
	errApplicationException = newException("application exception", nil, errApplicationExceptionTypeId)
	errUnexpectedHeader     = newException("unexpected header frame", kerrors.ErrStreamingProtocol, errUnexpectedHeaderTypeId)
	errIllegalBizErr        = newException("illegal bizErr", kerrors.ErrStreamingProtocol, errIllegalBizErrTypeId)
	errIllegalFrame         = newException("illegal frame", kerrors.ErrStreamingProtocol, errIllegalFrameTypeId)
	errIllegalOperation     = newException("illegal operation", kerrors.ErrStreamingProtocol, errIllegalOperationTypeId)
	errTransport            = newException("transport is closing", kerrors.ErrStreamingProtocol, errTransportTypeId)

	errBizCancel = newException("user code invoking stream RPC with context processed by context.WithCancel or context.WithTimeout, then invoking cancel() actively",
		kerrors.ErrStreamingCanceled, errBizCancelTypeId)
	errBizCancelWithCause     = newException("user code canceled with cancelCause(error)", kerrors.ErrStreamingCanceled, errBizCancelWithCauseTypeId)
	errDownstreamCancel       = newException("canceled by downstream", kerrors.ErrStreamingCanceled, errDownstreamCancelTypeId)
	errUpstreamCancel         = newException("canceled by upstream", kerrors.ErrStreamingCanceled, errUpstreamCancelTypeId)
	errInternalCancel         = newException("internal canceled", kerrors.ErrStreamingCanceled, errInternalCancelTypeId)
	errBizHandlerReturnCancel = newException("canceled by business handler returning", kerrors.ErrStreamingCanceled, errBizHandlerReturnTypeId)
	errConnectionClosedCancel = newException("canceled by connection closed", kerrors.ErrStreamingCanceled, errConnectionClosedTypeId)
)

var errServerSideBizHandlerReturnCancel = errBizHandlerReturnCancel.newBuilder().withSide(serverSide)

func NewStreamRecvTimeoutException(cfg streaming.TimeoutConfig, isClient bool) *Exception {
	side := clientSide
	if !isClient {
		side = serverSide
	}
	return newException(fmt.Sprintf("stream Recv timeout, timeout config=%+v", cfg), kerrors.ErrStreamingTimeout, errRecvTimeoutTypeId).withSide(side)
}

func newStreamRecvTimeoutException(tm time.Duration) *Exception {
	return newException(fmt.Sprintf("stream Recv timeout, timeout=%+v", tm), kerrors.ErrStreamingTimeout, errRecvTimeoutTypeId).withSide(clientSide)
}

func newStreamTimeoutException(tm time.Duration) *Exception {
	var msg string
	if tm > 0 {
		msg = fmt.Sprintf("stream timeout, timeout in ctx: %+v", tm)
	} else {
		msg = "stream timeout, no explicit timeout set in stream ctx"
	}
	return newException(msg, kerrors.ErrStreamingTimeout, errStreamTimeoutTypeId).withSide(clientSide)
}

func NewStreamInternalException(desc string, isClient bool) *Exception {
	side := clientSide
	if !isClient {
		side = serverSide
	}
	// stream internal exception does not have specified parent err, just pass nil
	return newException(desc, nil, errStreamInternalTypeId).withSide(side)
}

const (
	setSide = 1 << iota
	setCancelPath
	setCause
)

type Exception struct {
	// basic information, align with thrift ApplicationException
	message string
	typeId  int32

	// extended information, for better troubleshooting experience
	side       sideType
	cancelPath string

	// error hierarchy
	parent error
	// when cause is set, replace message to cause.Error() when displaying error information
	cause error

	bitSet uint8
}

func newException(message string, parent error, typeId int32) *Exception {
	return &Exception{message: message, parent: parent, typeId: typeId}
}

// newBuilder shallow-copy a new Exception.
// this func should be invoked before building a new Exception from pre-defined Exceptions
func (e *Exception) newBuilder() *Exception {
	newEx := *e
	return &newEx
}

func (e *Exception) Error() string {
	var strBuilder strings.Builder
	strBuilder.WriteString(fmt.Sprintf("[ttstream error, code=%d] ", e.typeId))

	if e.isSideSet() {
		switch e.side {
		case clientSide:
			strBuilder.WriteString("[client-side stream] ")
		case serverSide:
			strBuilder.WriteString("[server-side stream] ")
		}
	}

	if e.isCancelPathSet() {
		strBuilder.WriteString("[canceled path: ")
		strBuilder.WriteString(formatCancelPath(e.cancelPath))
		strBuilder.WriteString("] ")
	}

	if e.isCauseSet() {
		strBuilder.WriteString(e.cause.Error())
	} else {
		strBuilder.WriteString(e.message)
	}

	return strBuilder.String()
}

func (e *Exception) withCause(cause error) *Exception {
	if cause != nil {
		e.cause = cause
		e.bitSet |= setCause
	}
	return e
}

func (e *Exception) withCauseAndTypeId(cause error, typeId int32) *Exception {
	e.cause = cause
	e.typeId = typeId
	e.bitSet |= setCause
	return e
}

func (e *Exception) isCauseSet() bool {
	return e.bitSet&setCause != 0
}

func (e *Exception) withSide(side sideType) *Exception {
	e.side = side
	e.bitSet |= setSide
	return e
}

func (e *Exception) isSideSet() bool {
	return e.bitSet&setSide != 0
}

func (e *Exception) setOrAppendCancelPath(cancelPath string) *Exception {
	e.cancelPath = appendCancelPath(e.cancelPath, cancelPath)
	e.bitSet |= setCancelPath
	return e
}

func (e *Exception) isCancelPathSet() bool {
	return e.bitSet&setCancelPath != 0
}

func (e *Exception) Is(target error) bool {
	if rawEx, ok := target.(*Exception); ok {
		return rawEx.message == e.message
	}
	return target == e || errors.Is(e.parent, target) || errors.Is(e.cause, target)
}

func (e *Exception) getMessage() string {
	if e.isCauseSet() {
		return e.cause.Error()
	}
	return e.message
}

func (e *Exception) TypeId() int32 {
	return e.typeId
}

// appendCancelPath is a common util func to process cancelPath metadata in Rst Frame and Exception
func appendCancelPath(oriCp, node string) string {
	if len(oriCp) > 0 {
		return strings.Join([]string{oriCp, node}, ",")
	}
	return node
}

func formatCancelPath(cancelPath string) string {
	if cancelPath == "" {
		return cancelPath
	}
	parts := strings.Split(cancelPath, ",")
	return strings.Join(parts, " -> ")
}
