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

package thrift

import (
	"errors"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/remote/codec/perrors"
)

const (
	EnableSkipDecoder CodecType = 0b10000

	// defaultBufSize is the default buffer size
	defaultBufSize = 128
)

// nextCopyBuffer wraps the remote.ByteBuffer and appends the read data to buf
// when ByteBufferNext is called
type nextCopyBuffer struct {
	remote.ByteBuffer
	buf []byte
}

func (b *nextCopyBuffer) Next(n int) ([]byte, error) {
	buf, err := b.ByteBuffer.Next(n)
	if err == nil {
		b.buf = append(b.buf, buf...)
	}
	return buf, err
}

func newNextCopyBuffer(bb remote.ByteBuffer) *nextCopyBuffer {
	return &nextCopyBuffer{
		ByteBuffer: bb,
		buf:        make([]byte, 0, defaultBufSize),
	}
}

// skipDecoder is used to parse the input byte-by-byte and skip the thrift payload
// for making use of Frugal and FastCodec in standard Thrift Binary Protocol scenario.
type skipDecoder struct {
	tprot *BinaryProtocol
	ncb   *nextCopyBuffer
}

func newSkipDecoder(tprot *BinaryProtocol) *skipDecoder {
	ncb := newNextCopyBuffer(tprot.trans)
	return &skipDecoder{
		tprot: NewBinaryProtocol(ncb),
		ncb:   ncb,
	}
}

func (sd *skipDecoder) skipString() error {
	size, err := sd.tprot.ReadI32()
	if err != nil {
		return err
	}
	if size < 0 {
		return perrors.InvalidDataLength
	}
	_, err = sd.tprot.next(int(size))
	return err
}

func (sd *skipDecoder) skipMap() (err error) {
	keyTypeId, valTypeId, size, err := sd.tprot.ReadMapBegin()
	if err != nil {
		return err
	}
	for i := 0; i < size; i++ {
		if err = sd.skipElem(keyTypeId); err != nil {
			return err
		}
		if err = sd.skipElem(valTypeId); err != nil {
			return err
		}
	}
	return nil
}

func (sd *skipDecoder) skipList() (err error) {
	elemTypeId, size, err := sd.tprot.ReadListBegin()
	if err != nil {
		return err
	}
	for i := 0; i < size; i++ {
		if err = sd.skipElem(elemTypeId); err != nil {
			return err
		}
	}
	return nil
}

func (sd *skipDecoder) skipSet() (err error) {
	return sd.skipList()
}

func (sd *skipDecoder) skipElem(typeId thrift.TType) (err error) {
	switch typeId {
	case thrift.BOOL, thrift.BYTE:
		if _, err = sd.tprot.next(1); err != nil {
			return
		}
	case thrift.I16:
		if _, err = sd.tprot.next(2); err != nil {
			return
		}
	case thrift.I32:
		if _, err = sd.tprot.next(4); err != nil {
			return
		}
	case thrift.I64, thrift.DOUBLE:
		if _, err = sd.tprot.next(8); err != nil {
			return
		}
	case thrift.STRING:
		if err = sd.skipString(); err != nil {
			return
		}
	case thrift.STRUCT:
		// todo(DMwangnima): limit the skip depth
		if err = sd.skipStruct(); err != nil {
			return
		}
	case thrift.MAP:
		if err = sd.skipMap(); err != nil {
			return
		}
	case thrift.SET:
		if err = sd.skipSet(); err != nil {
			return
		}
	case thrift.LIST:
		if err = sd.skipList(); err != nil {
			return
		}
	default:
		return thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA, errors.New(fmt.Sprintf("Unknown data type %d", typeId)))
	}
	return nil
}

func (sd *skipDecoder) skipStruct() (err error) {
	var fieldTypeId thrift.TType

	for {
		_, fieldTypeId, _, err = sd.tprot.ReadFieldBegin()
		if err != nil {
			return err
		}
		if fieldTypeId == thrift.STOP {
			return err
		}
		if err = sd.skipElem(fieldTypeId); err != nil {
			return err
		}
	}
}

func (sd *skipDecoder) buffer() []byte {
	return sd.ncb.buf
}

func (sd *skipDecoder) Recycle() {
	sd.tprot.Recycle()
}
