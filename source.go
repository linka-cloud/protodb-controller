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
	"fmt"

	"go.linka.cloud/protodb"
	"go.linka.cloud/protodb/typed"
	"k8s.io/client-go/util/workqueue"
)

type src[T any, PT Message[T], K comparable] struct {
	db  typed.Store[T, PT]
	key func(PT) K
}

func (s *src[T, PT, K]) String() string {
	var z T
	return fmt.Sprintf("protodb/*%T", z)
}

func (s *src[T, PT, K]) Start(ctx context.Context, w workqueue.TypedRateLimitingInterface[K]) error {
	var z T
	ch, err := s.db.Watch(ctx, &z)
	if err != nil {
		return err
	}
	rs, _, err := s.db.Get(ctx, &z)
	if err != nil {
		return err
	}
	for _, v := range rs {
		w.Add(s.key(v))
	}
	go func() {
		defer w.ShutDown()
		for {
			select {
			case e, ok := <-ch:
				if !ok {
					return
				}
				if e == nil {
					continue
				}
				if e.Err() != nil {
					continue
				}
				switch e.Type() {
				case protodb.EventTypeEnter:
					w.Add(s.key(e.New()))
				case protodb.EventTypeUpdate:
					w.Add(s.key(e.New()))
				case protodb.EventTypeLeave:
					w.Add(s.key(e.Old()))
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}
