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
	"go.linka.cloud/protodb-controller/pkg/controller"
	"go.linka.cloud/protodb-controller/pkg/log"
	"go.linka.cloud/protodb-controller/pkg/reconcile"
)

// Result contains the result of a Reconciler invocation.
type Result = reconcile.Result

type Reconciler[request comparable] = reconcile.TypedReconciler[request]

type ReconcilerFunc[request comparable] = reconcile.TypedFunc[request]

type Options[request comparable] = controller.TypedOptions[request]

var (
	// Log is the base logger used by controller-runtime.  It delegates
	// to another logr.Logger.  You *must* call SetLogger to
	// get any actual logging.
	Log = log.Log

	// LoggerFrom returns a logger with predefined values from a context.Context.
	// The logger, when used with controllers, can be expected to contain basic information about the object
	// that's being reconciled like:
	// - `reconciler group` and `reconciler kind` coming from the For(...) object passed in when building a controller.
	// - `name` and `namespace` from the reconciliation request.
	//
	// This is meant to be used with the context supplied in a struct that satisfies the Reconciler interface.
	LoggerFrom = log.FromContext

	// LoggerInto takes a context and sets the logger as one of its keys.
	//
	// This is meant to be used in reconcilers to enrich the logger within a context with additional values.
	LoggerInto = log.IntoContext

	// SetLogger sets a concrete logging implementation for all deferred Loggers.
	SetLogger = log.SetLogger
)
