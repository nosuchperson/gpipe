package main

import (
	"context"
	"github.com/nosuchperson/gpipe"
)

const (
	doubleModName = "double"
)

func NewDoubleModule() gpipe.ModuleFactory {
	return gpipe.NewSimpleModule(doubleModName, func(name string, config interface{}) (gpipe.ModuleInstance, error) {
		return gpipe.NewSimpleModuleInstance(doubleModName, name, func(ctx context.Context, modCtx gpipe.ModuleContext) error {
			for {
				select {
				case _ = <-ctx.Done():
					return nil
				case val := <-modCtx.MessageQueue():
					num := val.(int64)
					modCtx.Collect(num * 2)
				}
			}
		}), nil
	})
}
