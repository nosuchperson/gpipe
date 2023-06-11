package interval

import (
	"context"
	"errors"
	"fmt"
	"github.com/nosuchperson/gpipe"
	"time"
)

const (
	intervalModuleName = "timer/interval"
)

type intervalConfig struct {
	Interval int `yaml:"interval"`
}

func init() {
	gpipe.RegisterModule(NewIntervalModule())
}

func NewIntervalModule() gpipe.ModuleFactory {
	return gpipe.NewSimpleModule(intervalModuleName, func(name string, config interface{}) (gpipe.ModuleInstance, error) {
		// 解析 config
		if configMap, err := gpipe.ConfigMapUnmarshal(config, &intervalConfig{}); err != nil {
			return nil, errors.New(fmt.Sprintf("%s: invalid configuration, err = %v", intervalModuleName, err))
		} else {
			return gpipe.NewSimpleModuleInstance(intervalModuleName, name, func(ctx context.Context, modCtx gpipe.ModuleContext) error {
				delay := time.Millisecond * time.Duration(configMap.Interval)
				sign := time.After(delay)
				for {
					select {
					case _ = <-ctx.Done():
						return nil
					case _ = <-sign:
						modCtx.Collect(nil)
						sign = time.After(delay)
					}
				}
			}), nil
		}
	})
}
