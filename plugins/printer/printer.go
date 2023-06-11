package printer

import (
	"context"
	"fmt"
	"github.com/nosuchperson/gpipe"
)

const (
	moduleName = "sink/printer"
)

func init() {
	gpipe.RegisterModule(gpipe.NewSimpleModule(moduleName, func(name string, config interface{}) (gpipe.ModuleInstance, error) {
		return gpipe.NewSimpleModuleInstance(moduleName, name, func(ctx context.Context, modCtx gpipe.ModuleContext) error {
			for {
				select {
				case _ = <-ctx.Done():
					return nil
				case msg := <-modCtx.MessageQueue():
					fmt.Printf("%s\n", msg)
				}
			}
		}), nil
	}))
}
