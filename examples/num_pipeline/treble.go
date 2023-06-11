package main

import (
	"context"
	"github.com/nosuchperson/gpipe"
)

const (
	trebleModName = "treble"
)

func NewTrebleModule() gpipe.ModuleFactory {
	return gpipe.NewSimpleModule(trebleModName, func(name string, config interface{}) (gpipe.ModuleInstance, error) {
		return gpipe.NewSimpleModuleInstance(trebleModName, name, func(ctx context.Context, moduleContext gpipe.ModuleContext) error {
			for {
				select {
				case _ = <-ctx.Done():
					return nil
				case val := <-moduleContext.MessageQueue():
					num := val.(int64)
					moduleContext.Collect(num * 3)
				}
			}
		}), nil
	})
}
