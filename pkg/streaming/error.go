/*
 * Copyright 2025 CloudWeGo Authors
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

package streaming

import "time"

//// WireError 需要同时支持框架内的错误处理，以及和 Mesh Proxy 间的错误处理
//type WireError struct {
//	// real gRPC error
//	err *status.Error
//	// real ttstream Exception
//	ex *ttstream.Exception
//}

type WireError interface {
	Error() string
	StatusCode() int32
	IsFromProxy() bool
}

type CancelError interface {
	Error() string
	StatusCode() int32
	CancelPath() []string
}

type TimeoutError interface {
	Error() string
	StatusCode() int32
	TimeoutType() TimeoutType
	Timeout() time.Duration
}

//// 此处类型需要定义
//// 此外 Code 的称呼太过通用，如果使用接口处理可能会造成误伤
//func (wErr *WireError) Code() int32 {
//	if wErr.err != nil {
//		st := wErr.err.GRPCStatus()
//		return int32(st.Code())
//	}
//	if wErr.ex != nil {
//		return wErr.ex.TypeId()
//	}
//	// 这种情况不可能发生
//	return -1
//}
//
//func (wErr *WireError) String() string {
//	if wErr.err != nil {
//		return wErr.err.Error()
//	}
//	if wErr.ex != nil {
//		return wErr.ex.Error()
//	}
//	return ""
//}
//
//func (wErr *WireError) IsFromProxy() bool {
//
//}

type TimeoutType uint8

const (
	StreamTimeoutType TimeoutType = iota
	StreamRecvTimeoutType
	StreamSendTimeoutType
)

//type TimeoutError struct {
//	err error
//	typ TimeoutType
//}
//
//func NewTimeoutError(originalErr error, typ TimeoutType) TimeoutError {
//	return TimeoutError{
//		err: originalErr,
//		typ: typ,
//	}
//}
//
//func (te TimeoutError) TimeoutType() TimeoutType {
//	return te.typ
//}
//
//func (te TimeoutError) Error() string {
//	return te.err.Error()
//}
//
//func (te TimeoutError) StatusCode() int32 {
//
//}

//type CancelError struct {
//	path []string
//	err  error
//}
//
//func NewCancelError(originalErr error, cancelPath []string) CancelError {
//	return CancelError{
//		path: cancelPath,
//		err:  originalErr,
//	}
//}
//
//func (ce CancelError) TriggeredBy() string {
//	if len(ce.path) <= 0 {
//		return ""
//	}
//	return ce.path[len(ce.path)-1]
//}
//
//func (ce CancelError) CancelPath() []string {
//	return ce.path
//}
//
//func (ce CancelError) FirstService() string {
//	if len(ce.path) <= 0 {
//		return ""
//	}
//	return ce.path[0]
//}
//
//func (ce CancelError) Error() string {
//	return ce.err.Error()
//}
//
//func (ce CancelError) StatusCode() int32 {
//
//}
