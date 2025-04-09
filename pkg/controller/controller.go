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

package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"go.linka.cloud/grpc-toolkit/logger"
	"k8s.io/client-go/util/workqueue"

	"go.linka.cloud/protodb-controller/pkg/internal/controller"
	"go.linka.cloud/protodb-controller/pkg/reconcile"
	"go.linka.cloud/protodb-controller/pkg/source"
)

// TypedOptions are the arguments for creating a new Controller.
type TypedOptions[request comparable] struct {
	// SkipNameValidation allows skipping the name validation that ensures that every controller name is unique.
	// Unique controller names are important to get unique metrics and logs for a controller.
	// Defaults to the Controller.SkipNameValidation setting from the Manager if unset.
	// Defaults to false if Controller.SkipNameValidation setting from the Manager is also unset.
	SkipNameValidation *bool

	// MaxConcurrentReconciles is the maximum number of concurrent Reconciles which can be run. Defaults to 1.
	MaxConcurrentReconciles int

	// CacheSyncTimeout refers to the time limit set to wait for syncing caches.
	// Defaults to 2 minutes if not set.
	CacheSyncTimeout time.Duration

	// RecoverPanic indicates whether the panic caused by reconcile should be recovered.
	// Defaults to the Controller.RecoverPanic setting from the Manager if unset.
	// Defaults to true if Controller.RecoverPanic setting from the Manager is also unset.
	RecoverPanic *bool

	// NeedLeaderElection indicates whether the controller needs to use leader election.
	// Defaults to true, which means the controller will use leader election.
	NeedLeaderElection *bool

	// Reconciler reconciles an object
	Reconciler reconcile.TypedReconciler[request]

	// RateLimiter is used to limit how frequently requests may be queued.
	// Defaults to MaxOfRateLimiter which has both overall and per-item rate limiting.
	// The overall is a token bucket and the per-item is exponential.
	RateLimiter workqueue.TypedRateLimiter[request]

	// NewQueue constructs the queue for this controller once the controller is ready to start.
	// With NewQueue a custom queue implementation can be used, e.g. a priority queue to prioritize with which
	// priority/order objects are reconciled (e.g. to reconcile objects with changes first).
	// This is a func because the standard Kubernetes work queues start themselves immediately, which
	// leads to goroutine leaks if something calls controller.New repeatedly.
	// The NewQueue func gets the controller name and the RateLimiter option (defaulted if necessary) passed in.
	// NewQueue defaults to NewRateLimitingQueueWithConfig.
	//
	// NOTE: LOW LEVEL PRIMITIVE!
	// Only use a custom NewQueue if you know what you are doing.
	NewQueue func(controllerName string, rateLimiter workqueue.TypedRateLimiter[request]) workqueue.TypedRateLimitingInterface[request]

	// LogConstructor is used to construct a logger used for this controller and passed
	// to each reconciliation via the context field.
	LogConstructor func(request *request) logr.Logger
}

// TypedController implements an API.
type TypedController[request comparable] interface {
	// Reconciler is called to reconcile an object by Namespace/Name
	reconcile.TypedReconciler[request]

	// Watch watches the provided Source.
	Watch(src source.TypedSource[request]) error

	// Start starts the controller.  Start blocks until the context is closed or a
	// controller has an error starting.
	Start(ctx context.Context) error

	// GetLogger returns this controller logger prefilled with basic information.
	GetLogger() logr.Logger
}

// NewTypedUnmanaged returns a new typed controller without adding it to the manager.
//
// The name must be unique as it is used to identify the controller in metrics and logs.
func NewTypedUnmanaged[request comparable](name string, options TypedOptions[request]) (TypedController[request], error) {
	if options.Reconciler == nil {
		return nil, fmt.Errorf("must specify Reconciler")
	}

	if len(name) == 0 {
		return nil, fmt.Errorf("must specify Name for Controller")
	}

	if options.SkipNameValidation == nil || !*options.SkipNameValidation {
		if err := checkName(name); err != nil {
			return nil, err
		}
	}

	if options.LogConstructor == nil {
		log := logger.StandardLogger().Logr().WithValues(
			"controller", name,
		)
		options.LogConstructor = func(in *request) logr.Logger {
			return log
		}
	}

	if options.MaxConcurrentReconciles <= 0 {
		options.MaxConcurrentReconciles = 1
	}

	if options.CacheSyncTimeout == 0 {

		options.CacheSyncTimeout = 2 * time.Minute
	}

	if options.RateLimiter == nil {
		options.RateLimiter = workqueue.DefaultTypedControllerRateLimiter[request]()
	}

	if options.NewQueue == nil {
		options.NewQueue = func(controllerName string, rateLimiter workqueue.TypedRateLimiter[request]) workqueue.TypedRateLimitingInterface[request] {
			return workqueue.NewTypedRateLimitingQueueWithConfig(rateLimiter, workqueue.TypedRateLimitingQueueConfig[request]{
				Name: controllerName,
			})
		}
	}

	// Create controller with dependencies set
	return &controller.Controller[request]{
		Do:                      options.Reconciler,
		RateLimiter:             options.RateLimiter,
		NewQueue:                options.NewQueue,
		MaxConcurrentReconciles: options.MaxConcurrentReconciles,
		CacheSyncTimeout:        options.CacheSyncTimeout,
		Name:                    name,
		LogConstructor:          options.LogConstructor,
		RecoverPanic:            options.RecoverPanic,
		LeaderElected:           options.NeedLeaderElection,
	}, nil
}

// ReconcileIDFromContext gets the reconcileID from the current context.
var ReconcileIDFromContext = controller.ReconcileIDFromContext
