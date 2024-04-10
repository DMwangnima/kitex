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
	"bytes"
	"context"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/remote/codec/thrift/kitex_gen/baseline"
)

const (
	F_kind_mask = (1 << 5) - 1
)

type (
	GoNameOffset int32
	GoTypeOffset int32
)

type GoType struct {
	Size       uintptr
	PtrData    uintptr
	Hash       uint32
	Flags      uint8
	Align      uint8
	FieldAlign uint8
	KindFlags  uint8
	Equal      func(unsafe.Pointer, unsafe.Pointer) bool
	GCData     *byte
	Str        GoNameOffset
	PtrToSelf  GoTypeOffset
}

type GoPtrType struct {
	GoType
	Elem *GoType
}

func (self *GoType) Kind() reflect.Kind {
	return reflect.Kind(self.KindFlags & F_kind_mask)
}

type GoEface struct {
	Type  *GoType
	Value unsafe.Pointer
}

func UnpackEface(v interface{}) GoEface {
	return *(*GoEface)(unsafe.Pointer(&v))
}

func PtrElem(t *GoType) *GoType {
	if t.Kind() != reflect.Ptr {
		panic("t is not a ptr")
	} else {
		return (*GoPtrType)(unsafe.Pointer(t)).Elem
	}
}

//go:noescape
//go:linkname typedmemclr runtime.typedmemclr
//goland:noinspection GoUnusedParameter
func typedmemclr(typ *GoType, ptr unsafe.Pointer)

func objectmemclr(v interface{}) {
	p := UnpackEface(v)
	typedmemclr(PtrElem(p.Type), p.Value)
}

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())

	samples = []Sample{
		{"small", getSimpleValue(), nil},
		{"medium", getNestingValue(), nil},
		{"large", getNesting2Value(), nil},
	}

	for i, s := range samples {
		v, ok := s.val.(thrift.TStruct)
		if !ok {
			panic("s.val must be thrift.TStruct!")
		}
		mm := thrift.NewTMemoryBuffer()
		mm.Reset()
		if err := v.Write(thrift.NewTBinaryProtocolTransport(mm)); err != nil {
			panic(err)
		}
		samples[i].bytes = mm.Bytes()
		println("sample ", s.name, ", size ", len(samples[i].bytes), "B")
	}

	m.Run()
}

type Sample struct {
	name  string
	val   interface{}
	bytes []byte
}

var samples []Sample

var (
	bytesCount  = 16
	stringCount = 16
	listCount   = 8
	mapCount    = 8
)

func getSamples() []Sample {
	return []Sample{
		{samples[0].name, getSimpleValue(), samples[0].bytes},
		{samples[1].name, getNestingValue(), samples[1].bytes},
		{samples[2].name, getNesting2Value(), samples[2].bytes},
	}
}

func getString() string {
	return strings.Repeat("你好,\b\n\r\t世界", stringCount)
}

func getBytes() []byte {
	return bytes.Repeat([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, bytesCount)
}

func getSimpleValue() *baseline.Simple {
	return &baseline.Simple{
		ByteField:   math.MaxInt8,
		I64Field:    math.MaxInt64,
		DoubleField: math.MaxFloat64,
		I32Field:    math.MaxInt32,
		StringField: getString(),
		BinaryField: getBytes(),
	}
}

func getNestingValue() *baseline.Nesting {
	ret := &baseline.Nesting{
		String_:         getString(),
		ListSimple:      []*baseline.Simple{},
		Double:          math.MaxFloat64,
		I32:             math.MaxInt32,
		ListI32:         []int32{},
		I64:             math.MaxInt64,
		MapStringString: map[string]string{},
		SimpleStruct:    getSimpleValue(),
		MapI32I64:       map[int32]int64{},
		ListString:      []string{},
		Binary:          getBytes(),
		MapI64String:    map[int64]string{},
		ListI64:         []int64{},
		Byte:            math.MaxInt8,
		MapStringSimple: map[string]*baseline.Simple{},
	}

	for i := 0; i < listCount; i++ {
		ret.ListSimple = append(ret.ListSimple, getSimpleValue())
		ret.ListI32 = append(ret.ListI32, math.MinInt32)
		ret.ListI64 = append(ret.ListI64, math.MinInt64)
		ret.ListString = append(ret.ListString, getString())
	}

	for i := 0; i < mapCount; i++ {
		ret.MapStringString[strconv.Itoa(i)] = getString()
		ret.MapI32I64[int32(i)] = math.MinInt64
		ret.MapI64String[int64(i)] = getString()
		ret.MapStringSimple[strconv.Itoa(i)] = getSimpleValue()
	}

	return ret
}

func getNesting2Value() *baseline.Nesting2 {
	ret := &baseline.Nesting2{
		MapSimpleNesting: map[*baseline.Simple]*baseline.Nesting{},
		SimpleStruct:     getSimpleValue(),
		Byte:             math.MaxInt8,
		Double:           math.MaxFloat64,
		ListNesting:      []*baseline.Nesting{},
		I64:              math.MaxInt64,
		NestingStruct:    getNestingValue(),
		Binary:           getBytes(),
		String_:          getString(),
		SetNesting:       []*baseline.Nesting{},
		I32:              math.MaxInt32,
	}
	for i := 0; i < mapCount; i++ {
		ret.MapSimpleNesting[getSimpleValue()] = getNestingValue()
	}
	for i := 0; i < listCount; i++ {
		ret.ListNesting = append(ret.ListNesting, getNestingValue())
		x := getNestingValue()
		x.I64 = int64(i)
		ret.SetNesting = append(ret.SetNesting, x)
	}
	return ret
}

func Benchmark_unmarshalThriftData_standard(b *testing.B) {
	benchmarkUnmarshalThriftData(b, Basic, nextCopyBufferType)
}

func Benchmark_unmarshalThriftData_fastCodec_nextCopyBuffer(b *testing.B) {
	benchmarkUnmarshalThriftData(b, FastRead|EnableSkipDecoder, nextCopyBufferType)
}

func Benchmark_unmarshalThriftData_fastCodec_peekBuffer(b *testing.B) {
	benchmarkUnmarshalThriftData(b, FastRead|EnableSkipDecoder, peekBufferType)
}

func Benchmark_unmarshalThriftData_fastCodec_nextMcacheBuffer(b *testing.B) {
	benchmarkUnmarshalThriftData(b, FastRead|EnableSkipDecoder, nextMcacheBufferType)
}

func Benchmark_unmarshalThriftData_nextCopyBuffer(b *testing.B) {
	benchmarkUnmarshalThriftData(b, FrugalRead|EnableSkipDecoder, nextCopyBufferType)
}

func Benchmark_unmarshalThriftData_peekBuffer(b *testing.B) {
	benchmarkUnmarshalThriftData(b, FrugalRead|EnableSkipDecoder, peekBufferType)
}

func Benchmark_unmarshalThriftData_nextMcacheBuffer(b *testing.B) {
	benchmarkUnmarshalThriftData(b, FrugalRead|EnableSkipDecoder, nextMcacheBufferType)
}

func benchmarkUnmarshalThriftData(b *testing.B, cTyp CodecType, bTyp bufferType) {
	chosenBufferType = bTyp
	for _, sample := range getSamples() {
		b.Run(sample.name, func(b *testing.B) {
			b.SetBytes(int64(len(sample.bytes)))
			buf := sample.bytes
			rTyp := reflect.TypeOf(sample.val).Elem()
			v := reflect.New(rTyp).Interface()
			codec := thriftCodec{cTyp}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				objectmemclr(v)
				tProt := NewBinaryProtocol(remote.NewReaderBuffer(buf))
				_ = codec.unmarshalThriftData(context.Background(), tProt, "", v, 0)
				tProt.Recycle()
			}
		})
	}
}
