// Copyright 2025 Linka Cloud  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"go.linka.cloud/grpc-toolkit/logger"
	"go.linka.cloud/protodb"
	"go.linka.cloud/protodb/typed"
	"google.golang.org/protobuf/proto"

	"go.linka.cloud/protodb-controller/pkg/controller"
)

type Message[T any] interface {
	proto.Message
	*T
}

type Key[T any, K comparable] interface {
	Key(m T) K
}

type KeyFunc[T any, K comparable] func(m T) K

func (fn KeyFunc[T, K]) Key(m T) K {
	return fn(m)
}

type Controller interface {
	Start(ctx context.Context) error
}

type ctrl[T any, PT Message[T], K comparable] struct {
	c controller.TypedController[K]
	s *src[T, PT, K]
}

func New[T any, PT Message[T], K comparable](name string, db protodb.Client, fn Key[PT, K], options Options[K]) (Controller, error) {
	var z PT
	t := z.ProtoReflect().Descriptor().FullName()
	if db == nil {
		return nil, errors.New("db is required")
	}
	if fn == nil {
		return nil, errors.New("fn is required")
	}
	if options.LogConstructor == nil {
		options.LogConstructor = func(in *K) logr.Logger {
			var k any = "unknown"
			if in != nil {
				k = *in
			}
			return logger.StandardLogger().Logr().WithValues(
				"controller", name,
				"key", fmt.Sprintf("%s/%s", t, k),
			)
		}
	}
	c, err := controller.NewTypedUnmanaged[K](name, options)
	if err != nil {
		return nil, err
	}
	return &ctrl[T, PT, K]{s: &src[T, PT, K]{db: typed.NewStore[T, PT](db), key: fn.Key}, c: c}, nil
}

func (c *ctrl[T, PT, K]) Start(ctx context.Context) error {
	if err := c.c.Watch(c.s); err != nil {
		return err
	}
	return c.c.Start(ctx)
}
