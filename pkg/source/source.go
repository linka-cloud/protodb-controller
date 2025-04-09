/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package source

import (
	"context"
	"fmt"

	"k8s.io/client-go/util/workqueue"
)

// TypedSource is a generic source of events (e.g. Create, Update, Delete operations on Kubernetes Objects, Webhook callbacks, etc)
// which should be processed by event.EventHandlers to enqueue a request.
//
// * Use Kind for events originating in the cluster (e.g. Pod Create, Pod Update, Deployment Update).
//
// * Use Channel for events originating outside the cluster (e.g. GitHub Webhook callback, Polling external urls).
//
// Users may build their own Source implementations.
type TypedSource[request comparable] interface {
	// Start is internal and should be called only by the Controller to start the source.
	// Start must be non-blocking.
	Start(context.Context, workqueue.TypedRateLimitingInterface[request]) error
}

// TypedSyncingSource is a source that needs syncing prior to being usable. The controller
// will call its WaitForSync prior to starting workers.
type TypedSyncingSource[request comparable] interface {
	TypedSource[request]
	WaitForSync(ctx context.Context) error
}

// TypedFunc is a function that implements Source.
type TypedFunc[request comparable] func(context.Context, workqueue.TypedRateLimitingInterface[request]) error

// Start implements Source.
func (f TypedFunc[request]) Start(ctx context.Context, queue workqueue.TypedRateLimitingInterface[request]) error {
	return f(ctx, queue)
}

func (f TypedFunc[request]) String() string {
	return fmt.Sprintf("func source: %p", f)
}
