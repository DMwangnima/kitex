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
	"strings"

	"github.com/cloudwego/kitex/pkg/kerrors"
)

var (
	errApplicationException = newException("application exception", nil, 12001)
	errUnexpectedHeader     = newException("unexpected header frame", kerrors.ErrStreamingProtocol, 12002)
	errIllegalBizErr        = newException("illegal bizErr", kerrors.ErrStreamingProtocol, 12003)
	errIllegalFrame         = newException("illegal frame", kerrors.ErrStreamingProtocol, 12004)
	errIllegalOperation     = newException("illegal operation", kerrors.ErrStreamingProtocol, 12005)
	errTransport            = newException("transport is closing", kerrors.ErrStreamingProtocol, 12006)

	errBizCancel = newException("user code invoking stream RPC with context processed by context.WithCancel or context.WithTimeout, then invoking cancel() actively",
		kerrors.ErrStreamingCanceled, 12007)
	errBizCancelWithCause     = newException("user code canceled with cancelCause(error)", kerrors.ErrStreamingCanceled, 12008)
	errDownstreamCancel       = newException("canceled by downstream", kerrors.ErrStreamingCanceled, 12009)
	errUpstreamCancel         = newException("canceled by upstream", kerrors.ErrStreamingCanceled, 12010)
	errInternalCancel         = newException("internal canceled", kerrors.ErrStreamingCanceled, 12011)
	errBizHandlerReturnCancel = newException("canceled by business handler returning", kerrors.ErrStreamingCanceled, 12012)
	errConnectionClosedCancel = newException("canceled by connection closed", kerrors.ErrStreamingCanceled, 12013)
)

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

	bitSet       uint8
	cachedErrStr string // cached error string to avoid repeated allocations
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
	// Return cached error string if available
	if e.cachedErrStr != "" {
		return e.cachedErrStr
	}

	// Build error string with optimized allocations
	var strBuilder strings.Builder
	// Pre-allocate estimated capacity to avoid reallocations
	strBuilder.Grow(128)

	// Avoid fmt.Sprintf, use WriteString directly
	strBuilder.WriteString("[ttstream error, code=")
	strBuilder.WriteString(itoa(int(e.typeId)))
	strBuilder.WriteString("] ")

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

	// Cache the result
	e.cachedErrStr = strBuilder.String()
	return e.cachedErrStr
}

// itoa is a simple integer to string conversion to avoid fmt.Sprintf
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	neg := i < 0
	if neg {
		i = -i
	}
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
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
		// Avoid strings.Join allocation, use strings.Builder instead
		var buf strings.Builder
		buf.Grow(len(oriCp) + len(node) + 1)
		buf.WriteString(oriCp)
		buf.WriteByte(',')
		buf.WriteString(node)
		return buf.String()
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
