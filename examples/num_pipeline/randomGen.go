package main

import (
	"context"
	"github.com/nosuchperson/gpipe"
	"math/rand"
	"time"
)

const (
	randomGenName = "randomGen"
)

func NewRandomGenModule() gpipe.ModuleFactory {
	return gpipe.NewSimpleModule(randomGenName, func(name string, config interface{}) (gpipe.ModuleInstance, error) {
		return gpipe.NewSimpleModuleInstance(randomGenName, name, func(ctx context.Context, modCtx gpipe.ModuleContext) error {
			rand.Seed(time.Now().Unix())
			for {
				select {
				case _ = <-time.After(time.Second * 1):
					val := int64(rand.Int31())
					modCtx.Collect(val)
				case _ = <-ctx.Done():
					return nil
				}
			}
		}), nil
	})
}
