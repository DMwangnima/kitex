package grpcframe

import (
	"bytes"
	"github.com/cloudwego/netpoll"
	"io"
	"reflect"
	"testing"

	"golang.org/x/net/http2"
)

// mockNetpollReader implements netpoll.Reader.
type mockNetpollReader struct {
	reader io.Reader
}

func (m *mockNetpollReader) Next(n int) (p []byte, err error) {
	p = make([]byte, n)
	_, err = m.reader.Read(p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (m *mockNetpollReader) Peek(n int) (buf []byte, err error)        { return }
func (m *mockNetpollReader) Skip(n int) (err error)                    { return }
func (m *mockNetpollReader) Until(delim byte) (line []byte, err error) { return }
func (m *mockNetpollReader) ReadString(n int) (s string, err error)    { return }
func (m *mockNetpollReader) ReadBinary(n int) (p []byte, err error)    { return }
func (m *mockNetpollReader) ReadByte() (b byte, err error)             { return }
func (m *mockNetpollReader) Slice(n int) (r netpoll.Reader, err error) { return }
func (m *mockNetpollReader) Release() (err error)                      { return }
func (m *mockNetpollReader) Len() (length int)                         { return }

func testFramer() (*Framer, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	return NewFramer(buf, &mockNetpollReader{reader: buf}), buf
}

func TestWriteHeaders(t *testing.T) {
	tests := []struct {
		name      string
		p         http2.HeadersFrameParam
		wantEnc   string
		wantFrame *HeadersFrame
	}{
		{
			"basic",
			http2.HeadersFrameParam{
				StreamID:      42,
				BlockFragment: []byte("abc"),
				Priority:      http2.PriorityParam{},
			},
			"\x00\x00\x03\x01\x00\x00\x00\x00*abc",
			&HeadersFrame{
				FrameHeader: http2.FrameHeader{
					StreamID: 42,
					Type:     http2.FrameHeaders,
					Length:   uint32(len("abc")),
				},
				Priority:      http2.PriorityParam{},
				headerFragBuf: []byte("abc"),
			},
		},
		{
			"basic + end flags",
			http2.HeadersFrameParam{
				StreamID:      42,
				BlockFragment: []byte("abc"),
				EndStream:     true,
				EndHeaders:    true,
				Priority:      http2.PriorityParam{},
			},
			"\x00\x00\x03\x01\x05\x00\x00\x00*abc",
			&HeadersFrame{
				FrameHeader: http2.FrameHeader{
					StreamID: 42,
					Type:     http2.FrameHeaders,
					Flags:    http2.FlagHeadersEndStream | http2.FlagHeadersEndHeaders,
					Length:   uint32(len("abc")),
				},
				Priority:      http2.PriorityParam{},
				headerFragBuf: []byte("abc"),
			},
		},
		{
			"with padding",
			http2.HeadersFrameParam{
				StreamID:      42,
				BlockFragment: []byte("abc"),
				EndStream:     true,
				EndHeaders:    true,
				PadLength:     5,
				Priority:      http2.PriorityParam{},
			},
			"\x00\x00\t\x01\r\x00\x00\x00*\x05abc\x00\x00\x00\x00\x00",
			&HeadersFrame{
				FrameHeader: http2.FrameHeader{
					StreamID: 42,
					Type:     http2.FrameHeaders,
					Flags:    http2.FlagHeadersEndStream | http2.FlagHeadersEndHeaders | http2.FlagHeadersPadded,
					Length:   uint32(1 + len("abc") + 5), // pad length + contents + padding
				},
				Priority:      http2.PriorityParam{},
				headerFragBuf: []byte("abc"),
			},
		},
		{
			"with priority",
			http2.HeadersFrameParam{
				StreamID:      42,
				BlockFragment: []byte("abc"),
				EndStream:     true,
				EndHeaders:    true,
				PadLength:     2,
				Priority: http2.PriorityParam{
					StreamDep: 15,
					Exclusive: true,
					Weight:    127,
				},
			},
			"\x00\x00\v\x01-\x00\x00\x00*\x02\x80\x00\x00\x0f\u007fabc\x00\x00",
			&HeadersFrame{
				FrameHeader: http2.FrameHeader{
					StreamID: 42,
					Type:     http2.FrameHeaders,
					Flags:    http2.FlagHeadersEndStream | http2.FlagHeadersEndHeaders | http2.FlagHeadersPadded | http2.FlagHeadersPriority,
					Length:   uint32(1 + 5 + len("abc") + 2), // pad length + priority + contents + padding
				},
				Priority: http2.PriorityParam{
					StreamDep: 15,
					Exclusive: true,
					Weight:    127,
				},
				headerFragBuf: []byte("abc"),
			},
		},
		{
			"with priority stream dep zero", // golang.org/issue/15444
			http2.HeadersFrameParam{
				StreamID:      42,
				BlockFragment: []byte("abc"),
				EndStream:     true,
				EndHeaders:    true,
				PadLength:     2,
				Priority: http2.PriorityParam{
					StreamDep: 0,
					Exclusive: true,
					Weight:    127,
				},
			},
			"\x00\x00\v\x01-\x00\x00\x00*\x02\x80\x00\x00\x00\u007fabc\x00\x00",
			&HeadersFrame{
				FrameHeader: http2.FrameHeader{
					StreamID: 42,
					Type:     http2.FrameHeaders,
					Flags:    http2.FlagHeadersEndStream | http2.FlagHeadersEndHeaders | http2.FlagHeadersPadded | http2.FlagHeadersPriority,
					Length:   uint32(1 + 5 + len("abc") + 2), // pad length + priority + contents + padding
				},
				Priority: http2.PriorityParam{
					StreamDep: 0,
					Exclusive: true,
					Weight:    127,
				},
				headerFragBuf: []byte("abc"),
			},
		},
		{
			"zero length",
			http2.HeadersFrameParam{
				StreamID: 42,
				Priority: http2.PriorityParam{},
			},
			"\x00\x00\x00\x01\x00\x00\x00\x00*",
			&HeadersFrame{
				FrameHeader: http2.FrameHeader{
					StreamID: 42,
					Type:     http2.FrameHeaders,
					Length:   0,
				},
				Priority:      http2.PriorityParam{},
				headerFragBuf: []byte{},
			},
		},
	}
	for _, tt := range tests {
		fr, buf := testFramer()
		if err := fr.WriteHeaders(tt.p); err != nil {
			t.Errorf("test %q: %v", tt.name, err)
			continue
		}
		if buf.String() != tt.wantEnc {
			t.Errorf("test %q: encoded %q; want %q", tt.name, buf.Bytes(), tt.wantEnc)
		}
		f, err := fr.ReadFrame()
		if err != nil {
			t.Errorf("test %q: failed to read the frame back: %v", tt.name, err)
			continue
		}
		if !reflect.DeepEqual(f, tt.wantFrame) {
			t.Errorf("test %q: mismatch.\n got: %#v\nwant: %#v\n", tt.name, f, tt.wantFrame)
		}
	}
}
