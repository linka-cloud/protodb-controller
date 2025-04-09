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

package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"
	"go.linka.cloud/grpc-toolkit/logger"
	"go.linka.cloud/protodb"
	"go.linka.cloud/protodb/typed"

	controller "go.linka.cloud/protodb-controller"
	"go.linka.cloud/protodb-controller/example/pb"
)

//go:generate buf generate

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.C(ctx)

	pdb, err := protodb.Open(ctx, protodb.WithInMemory(true))
	if err != nil {
		log.Fatal(err)
	}
	defer pdb.Close()
	db := typed.NewStore[pb.Resource](pdb)

	k := controller.KeyFunc[*pb.Resource, string](func(r *pb.Resource) string {
		return r.GetID()
	})
	c, err := controller.New("noop", db.Raw(), k, controller.Options[string]{
		Reconciler: controller.ReconcilerFunc[string](func(ctx context.Context, req string) (controller.Result, error) {
			log := controller.LoggerFrom(ctx)
			log.Info("Reconciling resource")
			rs, _, err := db.Get(ctx, &pb.Resource{ID: req})
			if err != nil {
				return controller.Result{}, err
			}
			if len(rs) == 0 {
				log.Info("Resource not found")
				return controller.Result{}, nil
			}
			if rs[0].GetStatus().GetMessage() == "ok" {
				log.Info("Resource up to date")
				return controller.Result{}, nil
			}
			time.Sleep(time.Second)
			if rand.IntN(2)%2 == 0 {
				return controller.Result{}, fmt.Errorf("fake error")
			}
			rs[0].Status = &pb.Status{
				Message: "ok",
			}
			if _, err := db.Set(ctx, rs[0]); err != nil {
				log.Error(err, "failed to update resource status")
				return controller.Result{}, err
			}
			log.Info("Resource reconciled")
			return controller.Result{}, nil
		}),
	})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		logger.C(ctx).Info("creating resource")
		if _, err := db.Set(ctx, &pb.Resource{ID: uuid.NewString()}); err != nil {
			logger.C(ctx).WithError(err).Error("failed to set resource")
		}

		tk := time.NewTicker(5 * time.Second)
		defer tk.Stop()
		for {
			select {
			case <-tk.C:
				if rand.IntN(2)%2 == 0 {
					logger.C(ctx).Info("creating resource")
					if _, err := db.Set(ctx, &pb.Resource{ID: uuid.NewString()}); err != nil {
						logger.C(ctx).WithError(err).Error("failed to set resource")
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	if err := c.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
