package interval

import (
	"context"
	"fmt"
	"github.com/nosuchperson/gpipe"
	"strings"
	"testing"
	"time"
)

func TestInterval(t *testing.T) {
	TEST_INTERVAL := 1000
	eng := gpipe.NewEngine()
	if err := gpipe.RegisterModule(gpipe.NewSimpleModule("test", func(name string, config interface{}) (gpipe.ModuleInstance, error) {
		return gpipe.NewSimpleModuleInstance("test", name, func(ctx context.Context, modCtx gpipe.ModuleContext) error {
			select {
			case _ = <-time.After(time.Duration(TEST_INTERVAL) * time.Millisecond * 2): // 两倍时间等待
				t.Fatal("can't recv event")
			case _ = <-modCtx.MessageQueue():
				return nil
			}
			return nil
		}), nil
	})); err != nil {
		t.Fatal(err)
	}
	cfg := fmt.Sprintf(`
engine:
  IntervalCall:
    module: timer/interval 
    parent: [ ]
    queueSize: 1
    parallels: 1

  Test:
    module: test
    parent:
    - IntervalCall
    queueSize: 1
    parallels: 1

config:
  IntervalCall:
    interval: %d

  Test: {}
`, TEST_INTERVAL)
	if err := eng.Run(context.Background(), strings.NewReader(cfg)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond * time.Duration(TEST_INTERVAL) * 5) // 5 倍原始时间等 fatal 就好
}
