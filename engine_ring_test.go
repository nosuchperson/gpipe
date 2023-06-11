package gpipe

import (
	"context"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func newOnceTriggerModule(modName string) ModuleFactory {
	return NewSimpleModule(modName, func(name string, config interface{}) (ModuleInstance, error) {
		return NewSimpleModuleInstance(modName, name, func(ctx context.Context, modCtx ModuleContext) error {
			modCtx.Collect(1)
			return nil
		}), nil
	})
}

func newTestDoubleItModule(modName string, lastVal *int) ModuleFactory {
	return NewSimpleModule(modName, func(name string, config interface{}) (ModuleInstance, error) {
		return NewSimpleModuleInstance(modName, name, func(ctx context.Context, modCtx ModuleContext) error {
			cnt := 0
			for i := range modCtx.MessageQueue() {
				cnt += 1
				if cnt > 3 {
					break
				}
				*lastVal = i.(int)
				modCtx.Collect(i.(int) * 2)
			}
			return nil
		}), nil
	})
}

func TestEngine_Ring(t *testing.T) {
	eng := NewEngine()
	modALastVal := 0
	modBLastVal := 0
	assert.NoError(t, RegisterModule(newOnceTriggerModule("trigger")))
	assert.NoError(t, RegisterModule(newTestDoubleItModule("doubleA", &modALastVal)))
	assert.NoError(t, RegisterModule(newTestDoubleItModule("doubleB", &modBLastVal)))
	config := `
engine:
  Entry:
    module: trigger
    parent: [ ]
    queueSize: 1
    parallels: 1
    config: {}
  A:
    module: doubleA
    parent:
      - Entry
      - B
    queueSize: 1
    parallels: 1
    config: {}

  B:
    module: doubleB
    parent:
    - A
    queueSize: 1
    parallels: 1
    config: {}
`
	assert.Error(t, &GPWError{"has cycle"}, eng.Run(context.Background(), strings.NewReader(config)))
}
