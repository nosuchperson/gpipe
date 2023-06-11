package interval

import (
	"context"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestCronJob(t *testing.T) {
	afterDone := ""
	if err := gpipe.RegisterModule(gpipe.NewSimpleModule("test", func(name string, config interface{}) (gpipe.ModuleInstance, error) {
		return gpipe.NewSimpleModuleInstance("test", name, func(ctx context.Context, modCtx gpipe.ModuleContext) error {
			select {
			case s := <-modCtx.MessageQueue():
				afterDone = s.(string)
				return nil
			}
			return nil
		}), nil
	})); err != nil {
		t.Fatal(err)
	}
	gpipe.NewEngine().Run(context.Background(), strings.NewReader(`
engine:
  IntervalCall:
    module: timer/cronjob 
    parent: [ ]
    queueSize: 1
    parallels: 1
    config:
      jobs:
      - cronjob: "*/1 * * * * *"
        tag: "hello"

  Test:
    module: test
    parent:
    - IntervalCall
    queueSize: 1
    parallels: 1
    config: { }
`))
	time.Sleep(time.Second * 2)
	assert.Equal(t, "hello", afterDone)
}
