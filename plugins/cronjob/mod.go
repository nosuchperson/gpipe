package interval

import (
	"context"
	"errors"
	"fmt"
	"github.com/nosuchperson/gpipe"
	"github.com/robfig/cron/v3"
)

const (
	moduleName = "timer/cronjob"
)

type cronjobConfiguration struct {
	Jobs []struct {
		Cronjob string `yaml:"cronjob"`
		Tag     string `yaml:"tag"`
	} `yaml:"jobs"`
}

func init() {
	gpipe.RegisterModule(func() gpipe.ModuleFactory {
		return gpipe.NewSimpleModule(moduleName, func(name string, config interface{}) (gpipe.ModuleInstance, error) {
			// 解析 config
			if configMap, err := gpipe.ConfigMapUnmarshal(config, &cronjobConfiguration{}); err != nil {
				return nil, errors.New(fmt.Sprintf("%s: invalid configuration, err = %v", moduleName, err))
			} else {
				return gpipe.NewSimpleModuleInstance(moduleName, name, func(ctx context.Context, modCtx gpipe.ModuleContext) error {
					cronParser := cron.New(cron.WithParser(cron.NewParser(
						cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
					)))

					for _, job := range configMap.Jobs {
						if entityId, err := cronParser.AddJob(job.Cronjob, cron.FuncJob(func() {
							handleCronJob(ctx, modCtx, job.Cronjob, job.Tag)
						})); err != nil {
							modCtx.Logger().Error(modCtx, "failed to add cronjob, crontab: [%s], tag: [%s], err = %v", job.Cronjob, job.Tag, err)
						} else {
							modCtx.Logger().Trace(modCtx, "added cronjob, crontab: [%s], tag: [%s], entityId: [%d]", job.Cronjob, job.Tag, entityId)
						}
					}
					cronParser.Start()
					<-ctx.Done()
					<-cronParser.Stop().Done()
					return nil
				}), nil
			}
		})
	}())
}

func handleCronJob(ctx context.Context, modCtx gpipe.ModuleContext, cronjob string, tag string) {
	modCtx.Logger().Trace(modCtx, "cronjob triggered, crontab: [%s], tag: [%s]", cronjob, tag)
	modCtx.Collect(tag)
}
