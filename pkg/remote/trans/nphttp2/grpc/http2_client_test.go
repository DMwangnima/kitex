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

package grpc

import (
	"context"
	"sync"
	"testing"
)

// TestHTTP2Client_ActiveStreams tests the ActiveStreams method
func TestHTTP2Client_ActiveStreams(t *testing.T) {
	// Create a minimal http2Client for testing
	client := &http2Client{
		ctx:           context.Background(),
		activeStreams: make(map[uint32]*Stream),
	}

	// Test with no active streams
	if got := client.ActiveStreams(); got != 0 {
		t.Errorf("ActiveStreams() with no streams = %d, want 0", got)
	}

	// Add one stream
	client.activeStreams[1] = &Stream{id: 1}
	if got := client.ActiveStreams(); got != 1 {
		t.Errorf("ActiveStreams() with 1 stream = %d, want 1", got)
	}

	// Add multiple streams
	client.activeStreams[3] = &Stream{id: 3}
	client.activeStreams[5] = &Stream{id: 5}
	client.activeStreams[7] = &Stream{id: 7}
	if got := client.ActiveStreams(); got != 4 {
		t.Errorf("ActiveStreams() with 4 streams = %d, want 4", got)
	}

	// Remove a stream
	delete(client.activeStreams, 3)
	if got := client.ActiveStreams(); got != 3 {
		t.Errorf("ActiveStreams() after removing 1 stream = %d, want 3", got)
	}

	// Remove all streams
	client.activeStreams = make(map[uint32]*Stream)
	if got := client.ActiveStreams(); got != 0 {
		t.Errorf("ActiveStreams() after removing all streams = %d, want 0", got)
	}
}

// TestHTTP2Client_ActiveStreams_Concurrent tests concurrent access to ActiveStreams
func TestHTTP2Client_ActiveStreams_Concurrent(t *testing.T) {
	client := &http2Client{
		ctx:           context.Background(),
		activeStreams: make(map[uint32]*Stream),
	}

	// Initialize with some streams
	for i := uint32(1); i <= 10; i += 2 {
		client.activeStreams[i] = &Stream{id: i}
	}

	var wg sync.WaitGroup
	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count := client.ActiveStreams()
			if count < 0 {
				t.Errorf("ActiveStreams() returned negative value: %d", count)
			}
		}()
	}

	wg.Wait()
}

// TestHTTP2Client_ActiveStreams_ThreadSafety verifies that the method is thread-safe
func TestHTTP2Client_ActiveStreams_ThreadSafety(t *testing.T) {
	client := &http2Client{
		ctx:           context.Background(),
		activeStreams: make(map[uint32]*Stream),
	}
	client.mu = sync.Mutex{}

	var wg sync.WaitGroup

	// Goroutines calling ActiveStreams
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = client.ActiveStreams()
			}
		}()
	}

	// Goroutines modifying activeStreams (simulating stream creation/deletion)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				streamID := uint32(idx*100 + j)
				client.mu.Lock()
				if j%2 == 0 {
					client.activeStreams[streamID] = &Stream{id: streamID}
				} else {
					delete(client.activeStreams, streamID)
				}
				client.mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Final check - should not panic and should return a valid count
	count := client.ActiveStreams()
	if count < 0 {
		t.Errorf("Final ActiveStreams() returned negative value: %d", count)
	}
}

// TestHTTP2Client_ActiveStreams_LargeNumber tests with a large number of streams
func TestHTTP2Client_ActiveStreams_LargeNumber(t *testing.T) {
	client := &http2Client{
		ctx:           context.Background(),
		activeStreams: make(map[uint32]*Stream),
	}

	// Add 10000 streams
	const numStreams = 10000
	for i := uint32(1); i <= numStreams; i++ {
		client.activeStreams[i] = &Stream{id: i}
	}

	if got := client.ActiveStreams(); got != numStreams {
		t.Errorf("ActiveStreams() with %d streams = %d, want %d", numStreams, got, numStreams)
	}
}

// TestHTTP2Client_ActiveStreams_NilMap verifies behavior with nil map
func TestHTTP2Client_ActiveStreams_NilMap(t *testing.T) {
	client := &http2Client{
		ctx:           context.Background(),
		activeStreams: nil,
	}

	// Should return 0 for nil map
	if got := client.ActiveStreams(); got != 0 {
		t.Errorf("ActiveStreams() with nil map = %d, want 0", got)
	}
}

// BenchmarkHTTP2Client_ActiveStreams benchmarks the ActiveStreams method
func BenchmarkHTTP2Client_ActiveStreams(b *testing.B) {
	client := &http2Client{
		ctx:           context.Background(),
		activeStreams: make(map[uint32]*Stream),
	}

	// Add some streams
	for i := uint32(1); i <= 100; i++ {
		client.activeStreams[i] = &Stream{id: i}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.ActiveStreams()
	}
}

// BenchmarkHTTP2Client_ActiveStreams_Parallel benchmarks concurrent access
func BenchmarkHTTP2Client_ActiveStreams_Parallel(b *testing.B) {
	client := &http2Client{
		ctx:           context.Background(),
		activeStreams: make(map[uint32]*Stream),
	}

	// Add some streams
	for i := uint32(1); i <= 100; i++ {
		client.activeStreams[i] = &Stream{id: i}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client.ActiveStreams()
		}
	})
}
