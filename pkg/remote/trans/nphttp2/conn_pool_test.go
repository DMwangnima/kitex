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

package nphttp2

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/grpc"
)

// mockClientTransport implements grpc.ClientTransport for testing
type mockClientTransport struct {
	activeStreams int
	isActive      bool
	closed        bool
}

func (m *mockClientTransport) Write(s *grpc.Stream, hdr, data []byte, opts *grpc.Options) error {
	return nil
}

func (m *mockClientTransport) NewStream(ctx context.Context, callHdr *grpc.CallHdr) (*grpc.Stream, error) {
	return nil, nil
}

func (m *mockClientTransport) CloseStream(stream *grpc.Stream, err error) {
}

func (m *mockClientTransport) Error() <-chan struct{} {
	return nil
}

func (m *mockClientTransport) GoAway() <-chan struct{} {
	return nil
}

func (m *mockClientTransport) GetGoAwayReason() grpc.GoAwayReason {
	return grpc.GoAwayInvalid
}

func (m *mockClientTransport) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
}

func (m *mockClientTransport) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9090}
}

func (m *mockClientTransport) GracefulClose() {
	m.closed = true
}

func (m *mockClientTransport) Close(err error) error {
	m.closed = true
	return nil
}

func (m *mockClientTransport) ActiveStreams(tag string) int {
	// 简化测试：如果 tag 为空或匹配，返回 activeStreams
	// 在实际场景中，应该按 tag 过滤
	if tag == "" || tag == "matched" {
		return m.activeStreams
	}
	return 0
}

func (m *mockClientTransport) IsActive() bool {
	return m.isActive
}

var _ grpc.ClientTransport = (*mockClientTransport)(nil)
var _ grpc.IsActive = (*mockClientTransport)(nil)

func TestTransports_ActiveNums(t *testing.T) {
	// Test with empty transports
	trans := &transports{
		size:          3,
		cliTransports: make([]grpc.ClientTransport, 3),
	}

	if got := trans.activeNums("matched"); got != 0 {
		t.Errorf("activeNums() on empty transports = %d, want 0", got)
	}

	// Test with one transport
	trans.cliTransports[0] = &mockClientTransport{activeStreams: 5, isActive: true}
	if got := trans.activeNums("matched"); got != 5 {
		t.Errorf("activeNums() with one transport = %d, want 5", got)
	}

	// Test with multiple transports
	trans.cliTransports[1] = &mockClientTransport{activeStreams: 10, isActive: true}
	trans.cliTransports[2] = &mockClientTransport{activeStreams: 3, isActive: true}
	if got := trans.activeNums("matched"); got != 18 {
		t.Errorf("activeNums() with multiple transports = %d, want 18", got)
	}

	// Test with nil transport in the middle
	trans.cliTransports[1] = nil
	if got := trans.activeNums("matched"); got != 8 {
		t.Errorf("activeNums() with nil transport = %d, want 8", got)
	}

	// Test with non-matching tag
	if got := trans.activeNums("non-matched"); got != 0 {
		t.Errorf("activeNums() with non-matching tag = %d, want 0", got)
	}
}

func TestTransports_ActiveNums_Concurrent(t *testing.T) {
	trans := &transports{
		size:          3,
		cliTransports: make([]grpc.ClientTransport, 3),
	}

	// Initialize with mock transports
	for i := 0; i < 3; i++ {
		trans.cliTransports[i] = &mockClientTransport{activeStreams: i + 1, isActive: true}
	}

	var wg sync.WaitGroup
	// Concurrent reads with different tags
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			tag := "matched"
			if idx%2 == 0 {
				tag = "non-matched"
			}
			_ = trans.activeNums(tag)
		}(i)
	}

	// Concurrent writes (put)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			trans.put(&mockClientTransport{activeStreams: idx, isActive: true})
		}(i)
	}

	wg.Wait()
}

func TestConnPool_ActiveStreams(t *testing.T) {
	pool := NewConnPool("test-service", 3, grpc.ConnectOptions{})

	// Test with non-existent address (no connections stored)
	if got := pool.ActiveStreams("matched"); got != 0 {
		t.Errorf("ActiveStreams() for non-existent connections = %d, want 0", got)
	}

	// Test with one connection address (simulating conn to proxy1)
	trans1 := &transports{
		size:          3,
		cliTransports: make([]grpc.ClientTransport, 3),
	}
	trans1.cliTransports[0] = &mockClientTransport{activeStreams: 5, isActive: true}
	trans1.cliTransports[1] = &mockClientTransport{activeStreams: 3, isActive: true}
	trans1.cliTransports[2] = &mockClientTransport{activeStreams: 2, isActive: true}

	pool.conns.Store("proxy1", trans1)

	// Now the behavior changed: it searches ALL connections for matching tag
	if got := pool.ActiveStreams("matched"); got != 10 {
		t.Errorf("ActiveStreams() for matched tag = %d, want 10", got)
	}

	// Test with multiple connection addresses (simulating conns to proxy1 and proxy2)
	trans2 := &transports{
		size:          2,
		cliTransports: make([]grpc.ClientTransport, 2),
	}
	trans2.cliTransports[0] = &mockClientTransport{activeStreams: 7, isActive: true}
	trans2.cliTransports[1] = &mockClientTransport{activeStreams: 8, isActive: true}

	pool.conns.Store("proxy2", trans2)

	// Important: Now it aggregates across ALL connections
	// trans1 has 10 matched streams, trans2 has 15 matched streams
	if got := pool.ActiveStreams("matched"); got != 25 {
		t.Errorf("ActiveStreams() aggregated across connections = %d, want 25", got)
	}

	// Test with non-matching tag
	if got := pool.ActiveStreams("non-matched"); got != 0 {
		t.Errorf("ActiveStreams() for non-matched tag = %d, want 0", got)
	}
}

func TestConnPool_ActiveStreams_Concurrent(t *testing.T) {
	pool := NewConnPool("test-service", 3, grpc.ConnectOptions{})

	// Setup initial state with multiple connection addresses
	trans1 := &transports{
		size:          3,
		cliTransports: make([]grpc.ClientTransport, 3),
	}
	for i := 0; i < 3; i++ {
		trans1.cliTransports[i] = &mockClientTransport{activeStreams: i + 1, isActive: true}
	}
	pool.conns.Store("proxy1", trans1)

	trans2 := &transports{
		size:          2,
		cliTransports: make([]grpc.ClientTransport, 2),
	}
	for i := 0; i < 2; i++ {
		trans2.cliTransports[i] = &mockClientTransport{activeStreams: i + 5, isActive: true}
	}
	pool.conns.Store("proxy2", trans2)

	var wg sync.WaitGroup
	// Concurrent reads with matched tag
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = pool.ActiveStreams("matched")
		}()
	}

	// Concurrent reads with non-matched tag
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = pool.ActiveStreams("non-matched")
		}()
	}

	wg.Wait()
}

func TestTransports_Get_WithLock(t *testing.T) {
	trans := &transports{
		size:          3,
		cliTransports: make([]grpc.ClientTransport, 3),
	}

	// Initialize transports
	for i := 0; i < 3; i++ {
		trans.cliTransports[i] = &mockClientTransport{activeStreams: i, isActive: true}
	}

	var wg sync.WaitGroup
	// Concurrent get operations
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tr := trans.get()
			if tr == nil {
				t.Errorf("get() returned nil")
			}
		}()
	}

	// Concurrent activeNums operations (read lock) with different tags
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			tag := "matched"
			if idx%3 == 0 {
				tag = "non-matched"
			}
			_ = trans.activeNums(tag)
		}(i)
	}

	wg.Wait()
}

func TestConnPool_ActiveStreams_WithNilTransport(t *testing.T) {
	trans := &transports{
		size:          3,
		cliTransports: make([]grpc.ClientTransport, 3),
	}

	// Only set some transports
	trans.cliTransports[0] = &mockClientTransport{activeStreams: 5, isActive: true}
	// trans.cliTransports[1] is nil
	trans.cliTransports[2] = &mockClientTransport{activeStreams: 3, isActive: true}

	// Should handle nil transport gracefully
	got := trans.activeNums("matched")
	want := 8 // 5 + 0 + 3
	if got != want {
		t.Errorf("activeNums() with nil transport = %d, want %d", got, want)
	}
}

func BenchmarkTransports_ActiveNums(b *testing.B) {
	trans := &transports{
		size:          10,
		cliTransports: make([]grpc.ClientTransport, 10),
	}

	for i := 0; i < 10; i++ {
		trans.cliTransports[i] = &mockClientTransport{activeStreams: i, isActive: true}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trans.activeNums("matched")
	}
}

func BenchmarkConnPool_ActiveStreams(b *testing.B) {
	pool := NewConnPool("test-service", 3, grpc.ConnectOptions{})

	trans := &transports{
		size:          3,
		cliTransports: make([]grpc.ClientTransport, 3),
	}
	for i := 0; i < 3; i++ {
		trans.cliTransports[i] = &mockClientTransport{activeStreams: i + 1, isActive: true}
	}
	pool.conns.Store("addr1", trans)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.ActiveStreams("matched")
	}
}

func BenchmarkTransports_ActiveNums_Parallel(b *testing.B) {
	trans := &transports{
		size:          10,
		cliTransports: make([]grpc.ClientTransport, 10),
	}

	for i := 0; i < 10; i++ {
		trans.cliTransports[i] = &mockClientTransport{activeStreams: i, isActive: true}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			trans.activeNums("matched")
		}
	})
}
