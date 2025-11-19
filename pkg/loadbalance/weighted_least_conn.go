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

package loadbalance

import (
	"context"
	"math"
	"sync/atomic"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/remote"
)

type weightedLeastConnPicker struct {
	pingPongPicker Picker
	instances      []discovery.Instance
	index          uint32
}

func (p *weightedLeastConnPicker) Next(ctx context.Context, request interface{}) (ins discovery.Instance) {
	cs := remote.GetConnStatistics(ctx)
	if cs == nil {
		// Ping-Pong
		return p.pingPongPicker.Next(ctx, request)
	}

	if len(p.instances) == 0 {
		return nil
	}
	minNormalizedLoad := math.MaxFloat64
	ins = p.instances[0]

	var hasActiveStreams bool
	for i := 0; i < len(p.instances); i++ {
		tmpIns := p.instances[i]
		weight := tmpIns.Weight()
		if weight <= 0 {
			weight = 1
		}
		addr := tmpIns.Address().String()
		activeStreams := cs.ActiveStreams(addr)
		if activeStreams > 0 {
			hasActiveStreams = true
		}
		load := float64(activeStreams) / float64(weight)
		klog.CtxWarnf(ctx, "KITEX: instance iterate: %s, activeStreams: %d, load: %.2f", addr, activeStreams, load)
		if load < minNormalizedLoad {
			minNormalizedLoad = load
			ins = p.instances[i]
		}
	}
	// all activeStreams of instances are 0, using round robin
	if !hasActiveStreams {
		idx := atomic.AddUint32(&p.index, 1)
		ins = p.instances[idx%uint32(len(p.instances))]
	}
	klog.CtxWarnf(ctx, "KITEX: instance pick: %s, all instances: %+v", ins.Address(), p.instances)
	return ins
}

func newWeightedLeastConnPicker(instances []discovery.Instance, isBalance bool) Picker {
	res := &weightedLeastConnPicker{
		instances: instances,
	}
	if isBalance {
		res.pingPongPicker = newRoundRobinPicker(instances)
	} else {
		res.pingPongPicker = newWeightedRoundRobinPicker(instances)
	}
	return res
}
