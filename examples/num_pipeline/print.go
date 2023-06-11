package main

import (
	"context"
	"github.com/nosuchperson/gpipe"
)

const (
	printModName = "print"
)

func NewPrintModule() gpipe.ModuleFactory {
	return gpipe.NewSimpleModule(printModName, func(name string, config interface{}) (gpipe.ModuleInstance, error) {
		return gpipe.NewSimpleModuleInstance(printModName, name, func(ctx context.Context, moduleContext gpipe.ModuleContext) error {
			for {
				select {
				case _ = <-ctx.Done():
					return nil
				case val := <-moduleContext.MessageQueue():
					moduleContext.Logger().Info(moduleContext, "%s: Dump value %v", moduleContext.Name(), val)
				}
			}
		}), nil
	})
}
