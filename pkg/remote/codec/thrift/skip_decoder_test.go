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
	"context"
	"fmt"
	"github.com/bytedance/gopkg/util/gctuner"
	"github.com/cloudwego/frugal"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/apache/thrift/lib/go/thrift"

	"github.com/cloudwego/kitex/internal/test"
	"github.com/cloudwego/kitex/pkg/remote"
)

func TestSkipBuffer(t *testing.T) {
	buf := []byte{1, 2, 3}
	rb := remote.NewReaderBuffer(buf)
	sb := newSkipBuffer(rb)
	peekData, err := sb.Next(3)
	test.Assert(t, err == nil)
	test.DeepEqual(t, buf, peekData[:len(buf)])
	test.Assert(t, sb.readNum == len(buf))
	test.Assert(t, rb.ReadLen() == 0)

	data, err := sb.Buffer()
	test.Assert(t, err == nil)
	test.DeepEqual(t, buf, data[:len(buf)])
	test.Assert(t, rb.ReadLen() == len(buf))
}

func TestSkipDecoder_SkipStruct(t *testing.T) {
	tProt := NewBinaryProtocol(remote.NewReaderWriterBuffer(1024))
	defer tProt.Recycle()
	tProt.WriteStructBegin("testStruct")
	// 1. Byte
	tProt.WriteFieldBegin("Byte", thrift.BYTE, 1)
	tProt.WriteByte('1')
	tProt.WriteFieldEnd()
	// 2. Bool
	tProt.WriteFieldBegin("Bool", thrift.BOOL, 2)
	tProt.WriteBool(true)
	tProt.WriteFieldEnd()
	// 3. I16
	tProt.WriteFieldBegin("I16", thrift.I16, 3)
	tProt.WriteI16(2)
	tProt.WriteFieldEnd()
	// 4. I32
	tProt.WriteFieldBegin("I32", thrift.I32, 4)
	tProt.WriteI32(3)
	tProt.WriteFieldEnd()
	// 5. I64
	tProt.WriteFieldBegin("I64", thrift.I64, 5)
	tProt.WriteI64(4)
	tProt.WriteFieldEnd()
	// 6. Double
	tProt.WriteFieldBegin("Double", thrift.DOUBLE, 6)
	tProt.WriteDouble(5)
	tProt.WriteFieldEnd()
	// 7. String
	tProt.WriteFieldBegin("String", thrift.STRING, 7)
	tProt.WriteString("6")
	tProt.WriteFieldEnd()
	// 8. Map
	tProt.WriteFieldBegin("Map", thrift.MAP, 8)
	tProt.WriteMapBegin(thrift.I32, thrift.I32, 1)
	tProt.WriteI32(7)
	tProt.WriteI32(8)
	tProt.WriteMapEnd()
	tProt.WriteFieldEnd()
	// 9. Set
	tProt.WriteFieldBegin("Set", thrift.SET, 9)
	tProt.WriteSetBegin(thrift.I32, 1)
	tProt.WriteI32(9)
	tProt.WriteSetEnd()
	tProt.WriteFieldEnd()
	// 10. List
	tProt.WriteFieldBegin("List", thrift.LIST, 10)
	tProt.WriteListBegin(thrift.I32, 1)
	tProt.WriteI32(9)
	tProt.WriteListEnd()
	tProt.WriteFieldEnd()

	tProt.WriteFieldStop()
	tProt.WriteStructEnd()

	length := tProt.ByteBuffer().ReadableLen()
	sd := newSkipDecoder(tProt)
	defer sd.Recycle()
	err := sd.SkipStruct()
	test.Assert(t, err == nil)
	test.Assert(t, sd.sb.readNum == length)
	test.Assert(t, sd.sb.ReadLen() == 0)
	test.Assert(t, sd.sb.ReadableLen() == length)
	_, err = sd.Buffer()
	test.Assert(t, err == nil)
	test.Assert(t, sd.sb.ReadLen() == length)
	test.Assert(t, sd.sb.ReadableLen() == 0)
}

const (
	FuzzDebugEnv          = "FuzzDebug"
	MemoryLimitEnv        = "MemLimit"
	KB             uint64 = 1024
	MB             uint64 = 1024 * KB
	GB             uint64 = 1024 * MB
)

func init() {
	file, _ := os.OpenFile("/tmp/fuzz-test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	log.SetOutput(file)
	go func() {
		if os.Getenv(FuzzDebugEnv) == "on" {
			log.Println(http.ListenAndServe("localhost:0", nil))
		}
	}()
}

type CompilerTest struct {
	A bool                   `frugal:"0,default,bool"`
	B int8                   `frugal:"1,default,i8"`
	C float64                `frugal:"2,default,double"`
	D int16                  `frugal:"3,default,i16"`
	E int32                  `frugal:"4,default,i32"`
	F int64                  `frugal:"5,default,i64"`
	G string                 `frugal:"6,default,string"`
	I *CompilerTestSubStruct `frugal:"8,default,CompilerTestSubStruct"`
	J map[string]int64       `frugal:"9,default,map<string:i64>"`
	K []string               `frugal:"10,default,set<string>"`
	L []string               `frugal:"11,default,list<string>"`
	M []byte                 `frugal:"12,default,binary"`
	N []int8                 `frugal:"13,default,set<i8>"`
	O []int8                 `frugal:"14,default,list<i8>"`
	P int64                  `frugal:"16,required,i64"`
}

type CompilerTestSubStruct struct {
	X int64                  `frugal:"0,default,i64"`
	Y *CompilerTestSubStruct `frugal:"1,default,CompilerTestSubStruct"`
}

func FuzzMain(f *testing.F) {
	// avoid OOM
	var limit uint64 = 4 * GB
	if os.Getenv(MemoryLimitEnv) != "" {
		if memGB, err := strconv.ParseUint(os.Getenv(MemoryLimitEnv), 10, 64); err == nil {
			limit = memGB * GB
		}
	}
	threshold := uint64(float64(limit) * 0.7)
	numWorker := uint64(runtime.GOMAXPROCS(0))
	gctuner.Tuning(threshold / numWorker)
	log.Printf("[%d] Memory Limit: %d GB, Memory Threshold: %d MB\n", os.Getpid(), limit/GB, threshold/MB)
	log.Printf("[%d] Memory Threshold Per Worker: %d MB\n", os.Getpid(), threshold/numWorker/MB)

	ct := &CompilerTest{
		I: &CompilerTestSubStruct{Y: &CompilerTestSubStruct{}},
	}
	buf := make([]byte, frugal.EncodedSize(ct))
	_, err := frugal.EncodeObject(buf, nil, ct)
	if err != nil {
		f.Fatal(err)
	}

	codec := thriftCodec{FrugalRead | EnableSkipDecoder}

	f.Add(buf)
	f.Fuzz(func(t *testing.T, data []byte) {
		for i := thrift.BOOL; i < thrift.UTF16; i++ {
			typ, length, err := Check(data, thrift.TType(i))
			if err != nil {
				continue
			}
			if length != len(data) {
				continue
			}
			rt, err := fuzzDynamicStruct(typ)
			if err != nil {
				t.Fatal(err)
			}
			object := reflect.New(rt).Interface()
			// wrap base types or container types with struct
			wrappedData := make([]byte, 0, len(data)+3)
			wrappedData = append(wrappedData, []byte{byte(i), 0x0, 0x0}...)
			wrappedData = append(wrappedData, data...)
			wrappedData = append(wrappedData, 0x0)

			tProt := NewBinaryProtocol(remote.NewReaderBuffer(wrappedData))
			defer tProt.Recycle()
			err = codec.unmarshalThriftData(context.Background(), tProt, "", object, 0)
			if err != nil {
				PrintStructTag(rt)
				t.Fatal(err)
			}

			buf := make([]byte, frugal.EncodedSize(object))
			_, err = frugal.EncodeObject(buf, nil, object)
			if err != nil {
				PrintStructTag(rt)
				t.Fatal(err)
			}
		}
	})
}

func PrintStructTag(rt reflect.Type) {
	fmt.Println("struct {")
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		var strs []string
		strs = append(strs, field.Name)

		switch field.Type.Kind() {
		case reflect.Ptr:
			strs = append(strs, "ptr<"+field.Type.Elem().Name()+">")
		case reflect.Slice:
			strs = append(strs, "slice<"+field.Type.Elem().Name()+">")
		case reflect.Map:
			strs = append(strs, "map<"+field.Type.Key().Name()+"("+field.Type.Key().Kind().String()+")"+", "+field.Type.Elem().Name()+"("+field.Type.Elem().Kind().String()+")"+">")
		default:
			strs = append(strs, field.Type.Kind().String())
		}
		strs = append(strs, field.Tag.Get("frugal"))
		fmt.Println("\t", strings.Join(strs, " "))
	}
	fmt.Println("}")

	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i).Type
		if ft.Kind() == reflect.Ptr && ft.Elem().Kind() == reflect.Struct {
			PrintStructTag(ft.Elem())
		}
	}
}
