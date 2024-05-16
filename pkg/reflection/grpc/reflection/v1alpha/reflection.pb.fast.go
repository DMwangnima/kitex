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

// Code generated by Fastpb v0.0.2. DO NOT EDIT.

package v1alpha

import (
	fmt "fmt"
	fastpb "github.com/cloudwego/fastpb"
)

var (
	_ = fmt.Errorf
	_ = fastpb.Skip
)

func (x *ServerReflectionRequest) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 1:
		offset, err = x.fastReadField1(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 3:
		offset, err = x.fastReadField3(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 4:
		offset, err = x.fastReadField4(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 5:
		offset, err = x.fastReadField5(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 6:
		offset, err = x.fastReadField6(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 7:
		offset, err = x.fastReadField7(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_ServerReflectionRequest[number], err)
}

func (x *ServerReflectionRequest) fastReadField1(buf []byte, _type int8) (offset int, err error) {
	x.Host, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *ServerReflectionRequest) fastReadField3(buf []byte, _type int8) (offset int, err error) {
	var ov ServerReflectionRequest_FileByFilename
	x.MessageRequest = &ov
	ov.FileByFilename, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *ServerReflectionRequest) fastReadField4(buf []byte, _type int8) (offset int, err error) {
	var ov ServerReflectionRequest_FileContainingSymbol
	x.MessageRequest = &ov
	ov.FileContainingSymbol, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *ServerReflectionRequest) fastReadField5(buf []byte, _type int8) (offset int, err error) {
	var ov ServerReflectionRequest_FileContainingExtension
	x.MessageRequest = &ov
	var v ExtensionRequest
	offset, err = fastpb.ReadMessage(buf, _type, &v)
	if err != nil {
		return offset, err
	}
	ov.FileContainingExtension = &v
	return offset, nil
}

func (x *ServerReflectionRequest) fastReadField6(buf []byte, _type int8) (offset int, err error) {
	var ov ServerReflectionRequest_AllExtensionNumbersOfType
	x.MessageRequest = &ov
	ov.AllExtensionNumbersOfType, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *ServerReflectionRequest) fastReadField7(buf []byte, _type int8) (offset int, err error) {
	var ov ServerReflectionRequest_ListServices
	x.MessageRequest = &ov
	ov.ListServices, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *ExtensionRequest) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 1:
		offset, err = x.fastReadField1(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 2:
		offset, err = x.fastReadField2(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_ExtensionRequest[number], err)
}

func (x *ExtensionRequest) fastReadField1(buf []byte, _type int8) (offset int, err error) {
	x.ContainingType, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *ExtensionRequest) fastReadField2(buf []byte, _type int8) (offset int, err error) {
	x.ExtensionNumber, offset, err = fastpb.ReadInt32(buf, _type)
	return offset, err
}

func (x *ServerReflectionResponse) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 1:
		offset, err = x.fastReadField1(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 2:
		offset, err = x.fastReadField2(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 4:
		offset, err = x.fastReadField4(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 5:
		offset, err = x.fastReadField5(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 6:
		offset, err = x.fastReadField6(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 7:
		offset, err = x.fastReadField7(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_ServerReflectionResponse[number], err)
}

func (x *ServerReflectionResponse) fastReadField1(buf []byte, _type int8) (offset int, err error) {
	x.ValidHost, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *ServerReflectionResponse) fastReadField2(buf []byte, _type int8) (offset int, err error) {
	var v ServerReflectionRequest
	offset, err = fastpb.ReadMessage(buf, _type, &v)
	if err != nil {
		return offset, err
	}
	x.OriginalRequest = &v
	return offset, nil
}

func (x *ServerReflectionResponse) fastReadField4(buf []byte, _type int8) (offset int, err error) {
	var ov ServerReflectionResponse_FileDescriptorResponse
	x.MessageResponse = &ov
	var v FileDescriptorResponse
	offset, err = fastpb.ReadMessage(buf, _type, &v)
	if err != nil {
		return offset, err
	}
	ov.FileDescriptorResponse = &v
	return offset, nil
}

func (x *ServerReflectionResponse) fastReadField5(buf []byte, _type int8) (offset int, err error) {
	var ov ServerReflectionResponse_AllExtensionNumbersResponse
	x.MessageResponse = &ov
	var v ExtensionNumberResponse
	offset, err = fastpb.ReadMessage(buf, _type, &v)
	if err != nil {
		return offset, err
	}
	ov.AllExtensionNumbersResponse = &v
	return offset, nil
}

func (x *ServerReflectionResponse) fastReadField6(buf []byte, _type int8) (offset int, err error) {
	var ov ServerReflectionResponse_ListServicesResponse
	x.MessageResponse = &ov
	var v ListServiceResponse
	offset, err = fastpb.ReadMessage(buf, _type, &v)
	if err != nil {
		return offset, err
	}
	ov.ListServicesResponse = &v
	return offset, nil
}

func (x *ServerReflectionResponse) fastReadField7(buf []byte, _type int8) (offset int, err error) {
	var ov ServerReflectionResponse_ErrorResponse
	x.MessageResponse = &ov
	var v ErrorResponse
	offset, err = fastpb.ReadMessage(buf, _type, &v)
	if err != nil {
		return offset, err
	}
	ov.ErrorResponse = &v
	return offset, nil
}

func (x *FileDescriptorResponse) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 1:
		offset, err = x.fastReadField1(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_FileDescriptorResponse[number], err)
}

func (x *FileDescriptorResponse) fastReadField1(buf []byte, _type int8) (offset int, err error) {
	var v []byte
	v, offset, err = fastpb.ReadBytes(buf, _type)
	if err != nil {
		return offset, err
	}
	x.FileDescriptorProto = append(x.FileDescriptorProto, v)
	return offset, err
}

func (x *ExtensionNumberResponse) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 1:
		offset, err = x.fastReadField1(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 2:
		offset, err = x.fastReadField2(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_ExtensionNumberResponse[number], err)
}

func (x *ExtensionNumberResponse) fastReadField1(buf []byte, _type int8) (offset int, err error) {
	x.BaseTypeName, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *ExtensionNumberResponse) fastReadField2(buf []byte, _type int8) (offset int, err error) {
	offset, err = fastpb.ReadList(buf, _type,
		func(buf []byte, _type int8) (n int, err error) {
			var v int32
			v, offset, err = fastpb.ReadInt32(buf, _type)
			if err != nil {
				return offset, err
			}
			x.ExtensionNumber = append(x.ExtensionNumber, v)
			return offset, err
		})
	return offset, err
}

func (x *ListServiceResponse) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 1:
		offset, err = x.fastReadField1(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_ListServiceResponse[number], err)
}

func (x *ListServiceResponse) fastReadField1(buf []byte, _type int8) (offset int, err error) {
	var v ServiceResponse
	offset, err = fastpb.ReadMessage(buf, _type, &v)
	if err != nil {
		return offset, err
	}
	x.Service = append(x.Service, &v)
	return offset, nil
}

func (x *ServiceResponse) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 1:
		offset, err = x.fastReadField1(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_ServiceResponse[number], err)
}

func (x *ServiceResponse) fastReadField1(buf []byte, _type int8) (offset int, err error) {
	x.Name, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *ErrorResponse) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 1:
		offset, err = x.fastReadField1(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 2:
		offset, err = x.fastReadField2(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_ErrorResponse[number], err)
}

func (x *ErrorResponse) fastReadField1(buf []byte, _type int8) (offset int, err error) {
	x.ErrorCode, offset, err = fastpb.ReadInt32(buf, _type)
	return offset, err
}

func (x *ErrorResponse) fastReadField2(buf []byte, _type int8) (offset int, err error) {
	x.ErrorMessage, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *ServerReflectionRequest) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField1(buf[offset:])
	offset += x.fastWriteField3(buf[offset:])
	offset += x.fastWriteField4(buf[offset:])
	offset += x.fastWriteField5(buf[offset:])
	offset += x.fastWriteField6(buf[offset:])
	offset += x.fastWriteField7(buf[offset:])
	return offset
}

func (x *ServerReflectionRequest) fastWriteField1(buf []byte) (offset int) {
	if x.Host == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 1, x.GetHost())
	return offset
}

func (x *ServerReflectionRequest) fastWriteField3(buf []byte) (offset int) {
	if x.GetFileByFilename() == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 3, x.GetFileByFilename())
	return offset
}

func (x *ServerReflectionRequest) fastWriteField4(buf []byte) (offset int) {
	if x.GetFileContainingSymbol() == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 4, x.GetFileContainingSymbol())
	return offset
}

func (x *ServerReflectionRequest) fastWriteField5(buf []byte) (offset int) {
	if x.GetFileContainingExtension() == nil {
		return offset
	}
	offset += fastpb.WriteMessage(buf[offset:], 5, x.GetFileContainingExtension())
	return offset
}

func (x *ServerReflectionRequest) fastWriteField6(buf []byte) (offset int) {
	if x.GetAllExtensionNumbersOfType() == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 6, x.GetAllExtensionNumbersOfType())
	return offset
}

func (x *ServerReflectionRequest) fastWriteField7(buf []byte) (offset int) {
	if x.GetListServices() == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 7, x.GetListServices())
	return offset
}

func (x *ExtensionRequest) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField1(buf[offset:])
	offset += x.fastWriteField2(buf[offset:])
	return offset
}

func (x *ExtensionRequest) fastWriteField1(buf []byte) (offset int) {
	if x.ContainingType == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 1, x.GetContainingType())
	return offset
}

func (x *ExtensionRequest) fastWriteField2(buf []byte) (offset int) {
	if x.ExtensionNumber == 0 {
		return offset
	}
	offset += fastpb.WriteInt32(buf[offset:], 2, x.GetExtensionNumber())
	return offset
}

func (x *ServerReflectionResponse) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField1(buf[offset:])
	offset += x.fastWriteField2(buf[offset:])
	offset += x.fastWriteField4(buf[offset:])
	offset += x.fastWriteField5(buf[offset:])
	offset += x.fastWriteField6(buf[offset:])
	offset += x.fastWriteField7(buf[offset:])
	return offset
}

func (x *ServerReflectionResponse) fastWriteField1(buf []byte) (offset int) {
	if x.ValidHost == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 1, x.GetValidHost())
	return offset
}

func (x *ServerReflectionResponse) fastWriteField2(buf []byte) (offset int) {
	if x.OriginalRequest == nil {
		return offset
	}
	offset += fastpb.WriteMessage(buf[offset:], 2, x.GetOriginalRequest())
	return offset
}

func (x *ServerReflectionResponse) fastWriteField4(buf []byte) (offset int) {
	if x.GetFileDescriptorResponse() == nil {
		return offset
	}
	offset += fastpb.WriteMessage(buf[offset:], 4, x.GetFileDescriptorResponse())
	return offset
}

func (x *ServerReflectionResponse) fastWriteField5(buf []byte) (offset int) {
	if x.GetAllExtensionNumbersResponse() == nil {
		return offset
	}
	offset += fastpb.WriteMessage(buf[offset:], 5, x.GetAllExtensionNumbersResponse())
	return offset
}

func (x *ServerReflectionResponse) fastWriteField6(buf []byte) (offset int) {
	if x.GetListServicesResponse() == nil {
		return offset
	}
	offset += fastpb.WriteMessage(buf[offset:], 6, x.GetListServicesResponse())
	return offset
}

func (x *ServerReflectionResponse) fastWriteField7(buf []byte) (offset int) {
	if x.GetErrorResponse() == nil {
		return offset
	}
	offset += fastpb.WriteMessage(buf[offset:], 7, x.GetErrorResponse())
	return offset
}

func (x *FileDescriptorResponse) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField1(buf[offset:])
	return offset
}

func (x *FileDescriptorResponse) fastWriteField1(buf []byte) (offset int) {
	if len(x.FileDescriptorProto) == 0 {
		return offset
	}
	for i := range x.GetFileDescriptorProto() {
		offset += fastpb.WriteBytes(buf[offset:], 1, x.GetFileDescriptorProto()[i])
	}
	return offset
}

func (x *ExtensionNumberResponse) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField1(buf[offset:])
	offset += x.fastWriteField2(buf[offset:])
	return offset
}

func (x *ExtensionNumberResponse) fastWriteField1(buf []byte) (offset int) {
	if x.BaseTypeName == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 1, x.GetBaseTypeName())
	return offset
}

func (x *ExtensionNumberResponse) fastWriteField2(buf []byte) (offset int) {
	if len(x.ExtensionNumber) == 0 {
		return offset
	}
	offset += fastpb.WriteListPacked(buf[offset:], 2, len(x.GetExtensionNumber()),
		func(buf []byte, numTagOrKey, numIdxOrVal int32) int {
			offset := 0
			offset += fastpb.WriteInt32(buf[offset:], numTagOrKey, x.GetExtensionNumber()[numIdxOrVal])
			return offset
		})
	return offset
}

func (x *ListServiceResponse) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField1(buf[offset:])
	return offset
}

func (x *ListServiceResponse) fastWriteField1(buf []byte) (offset int) {
	if x.Service == nil {
		return offset
	}
	for i := range x.GetService() {
		offset += fastpb.WriteMessage(buf[offset:], 1, x.GetService()[i])
	}
	return offset
}

func (x *ServiceResponse) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField1(buf[offset:])
	return offset
}

func (x *ServiceResponse) fastWriteField1(buf []byte) (offset int) {
	if x.Name == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 1, x.GetName())
	return offset
}

func (x *ErrorResponse) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField1(buf[offset:])
	offset += x.fastWriteField2(buf[offset:])
	return offset
}

func (x *ErrorResponse) fastWriteField1(buf []byte) (offset int) {
	if x.ErrorCode == 0 {
		return offset
	}
	offset += fastpb.WriteInt32(buf[offset:], 1, x.GetErrorCode())
	return offset
}

func (x *ErrorResponse) fastWriteField2(buf []byte) (offset int) {
	if x.ErrorMessage == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 2, x.GetErrorMessage())
	return offset
}

func (x *ServerReflectionRequest) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField1()
	n += x.sizeField3()
	n += x.sizeField4()
	n += x.sizeField5()
	n += x.sizeField6()
	n += x.sizeField7()
	return n
}

func (x *ServerReflectionRequest) sizeField1() (n int) {
	if x.Host == "" {
		return n
	}
	n += fastpb.SizeString(1, x.GetHost())
	return n
}

func (x *ServerReflectionRequest) sizeField3() (n int) {
	if x.GetFileByFilename() == "" {
		return n
	}
	n += fastpb.SizeString(3, x.GetFileByFilename())
	return n
}

func (x *ServerReflectionRequest) sizeField4() (n int) {
	if x.GetFileContainingSymbol() == "" {
		return n
	}
	n += fastpb.SizeString(4, x.GetFileContainingSymbol())
	return n
}

func (x *ServerReflectionRequest) sizeField5() (n int) {
	if x.GetFileContainingExtension() == nil {
		return n
	}
	n += fastpb.SizeMessage(5, x.GetFileContainingExtension())
	return n
}

func (x *ServerReflectionRequest) sizeField6() (n int) {
	if x.GetAllExtensionNumbersOfType() == "" {
		return n
	}
	n += fastpb.SizeString(6, x.GetAllExtensionNumbersOfType())
	return n
}

func (x *ServerReflectionRequest) sizeField7() (n int) {
	if x.GetListServices() == "" {
		return n
	}
	n += fastpb.SizeString(7, x.GetListServices())
	return n
}

func (x *ExtensionRequest) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField1()
	n += x.sizeField2()
	return n
}

func (x *ExtensionRequest) sizeField1() (n int) {
	if x.ContainingType == "" {
		return n
	}
	n += fastpb.SizeString(1, x.GetContainingType())
	return n
}

func (x *ExtensionRequest) sizeField2() (n int) {
	if x.ExtensionNumber == 0 {
		return n
	}
	n += fastpb.SizeInt32(2, x.GetExtensionNumber())
	return n
}

func (x *ServerReflectionResponse) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField1()
	n += x.sizeField2()
	n += x.sizeField4()
	n += x.sizeField5()
	n += x.sizeField6()
	n += x.sizeField7()
	return n
}

func (x *ServerReflectionResponse) sizeField1() (n int) {
	if x.ValidHost == "" {
		return n
	}
	n += fastpb.SizeString(1, x.GetValidHost())
	return n
}

func (x *ServerReflectionResponse) sizeField2() (n int) {
	if x.OriginalRequest == nil {
		return n
	}
	n += fastpb.SizeMessage(2, x.GetOriginalRequest())
	return n
}

func (x *ServerReflectionResponse) sizeField4() (n int) {
	if x.GetFileDescriptorResponse() == nil {
		return n
	}
	n += fastpb.SizeMessage(4, x.GetFileDescriptorResponse())
	return n
}

func (x *ServerReflectionResponse) sizeField5() (n int) {
	if x.GetAllExtensionNumbersResponse() == nil {
		return n
	}
	n += fastpb.SizeMessage(5, x.GetAllExtensionNumbersResponse())
	return n
}

func (x *ServerReflectionResponse) sizeField6() (n int) {
	if x.GetListServicesResponse() == nil {
		return n
	}
	n += fastpb.SizeMessage(6, x.GetListServicesResponse())
	return n
}

func (x *ServerReflectionResponse) sizeField7() (n int) {
	if x.GetErrorResponse() == nil {
		return n
	}
	n += fastpb.SizeMessage(7, x.GetErrorResponse())
	return n
}

func (x *FileDescriptorResponse) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField1()
	return n
}

func (x *FileDescriptorResponse) sizeField1() (n int) {
	if len(x.FileDescriptorProto) == 0 {
		return n
	}
	for i := range x.GetFileDescriptorProto() {
		n += fastpb.SizeBytes(1, x.GetFileDescriptorProto()[i])
	}
	return n
}

func (x *ExtensionNumberResponse) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField1()
	n += x.sizeField2()
	return n
}

func (x *ExtensionNumberResponse) sizeField1() (n int) {
	if x.BaseTypeName == "" {
		return n
	}
	n += fastpb.SizeString(1, x.GetBaseTypeName())
	return n
}

func (x *ExtensionNumberResponse) sizeField2() (n int) {
	if len(x.ExtensionNumber) == 0 {
		return n
	}
	n += fastpb.SizeListPacked(2, len(x.GetExtensionNumber()),
		func(numTagOrKey, numIdxOrVal int32) int {
			n := 0
			n += fastpb.SizeInt32(numTagOrKey, x.GetExtensionNumber()[numIdxOrVal])
			return n
		})
	return n
}

func (x *ListServiceResponse) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField1()
	return n
}

func (x *ListServiceResponse) sizeField1() (n int) {
	if x.Service == nil {
		return n
	}
	for i := range x.GetService() {
		n += fastpb.SizeMessage(1, x.GetService()[i])
	}
	return n
}

func (x *ServiceResponse) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField1()
	return n
}

func (x *ServiceResponse) sizeField1() (n int) {
	if x.Name == "" {
		return n
	}
	n += fastpb.SizeString(1, x.GetName())
	return n
}

func (x *ErrorResponse) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField1()
	n += x.sizeField2()
	return n
}

func (x *ErrorResponse) sizeField1() (n int) {
	if x.ErrorCode == 0 {
		return n
	}
	n += fastpb.SizeInt32(1, x.GetErrorCode())
	return n
}

func (x *ErrorResponse) sizeField2() (n int) {
	if x.ErrorMessage == "" {
		return n
	}
	n += fastpb.SizeString(2, x.GetErrorMessage())
	return n
}

var fieldIDToName_ServerReflectionRequest = map[int32]string{
	1: "Host",
	3: "FileByFilename",
	4: "FileContainingSymbol",
	5: "FileContainingExtension",
	6: "AllExtensionNumbersOfType",
	7: "ListServices",
}

var fieldIDToName_ExtensionRequest = map[int32]string{
	1: "ContainingType",
	2: "ExtensionNumber",
}

var fieldIDToName_ServerReflectionResponse = map[int32]string{
	1: "ValidHost",
	2: "OriginalRequest",
	4: "FileDescriptorResponse",
	5: "AllExtensionNumbersResponse",
	6: "ListServicesResponse",
	7: "ErrorResponse",
}

var fieldIDToName_FileDescriptorResponse = map[int32]string{
	1: "FileDescriptorProto",
}

var fieldIDToName_ExtensionNumberResponse = map[int32]string{
	1: "BaseTypeName",
	2: "ExtensionNumber",
}

var fieldIDToName_ListServiceResponse = map[int32]string{
	1: "Service",
}

var fieldIDToName_ServiceResponse = map[int32]string{
	1: "Name",
}

var fieldIDToName_ErrorResponse = map[int32]string{
	1: "ErrorCode",
	2: "ErrorMessage",
}
