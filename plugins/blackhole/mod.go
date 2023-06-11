package blackhole

import (
	"context"
	"github.com/nosuchperson/gpipe"
)

const (
	blackHoleModuleName = "sink/blackhole"
)

func init() {
	gpipe.RegisterModule(NewBlackHoleModule())
}

func NewBlackHoleModule() gpipe.ModuleFactory {
	return gpipe.NewSimpleModule(blackHoleModuleName, func(name string, config interface{}) (gpipe.ModuleInstance, error) {
		return gpipe.NewSimpleModuleInstance(blackHoleModuleName, name, func(ctx context.Context, modCtx gpipe.ModuleContext) error {
			for _ = range modCtx.MessageQueue() {
			}
			return nil
		}), nil
	})
}
