package gpipe

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestEngine(t *testing.T) {
	rand.Seed(time.Now().UnixMilli())
	eng := NewEngine()
	testsCount := int(rand.Uint32() % 10000)
	genVal := []int{}
	recvVal := []int{}
	if err := RegisterModule(NewSimpleModule("gen", func(name string, config interface{}) (ModuleInstance, error) {
		return NewSimpleModuleInstance("gen", name, func(ctx context.Context, modCtx ModuleContext) error {
			for i := 0; i < testsCount; i++ {
				val := rand.Int()
				genVal = append(genVal, val)
				modCtx.Collect(val)
			}
			return nil
		}), nil
	})); err != nil {
		t.Fatal(err)
	}
	if err := RegisterModule(NewSimpleModule("recv", func(name string, config interface{}) (ModuleInstance, error) {
		return NewSimpleModuleInstance("recv", name, func(ctx context.Context, modCtx ModuleContext) error {
			for i := 0; i < testsCount; i++ {
				recvVal = append(recvVal, (<-modCtx.MessageQueue()).(int))
			}
			return nil
		}), nil
	})); err != nil {
		t.Fatal(err)
	}

	if err := eng.Run(context.Background(), strings.NewReader(`
engine:
  Gen:
    module: gen
    parent: [ ]
    queueSize: 1
    parallels: 1
    config: {}

  Test:
    module: recv
    parent:
    - Gen
    queueSize: 1
    parallels: 1
    config: {}

`)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 5) // 1w 次内 5秒总该结束了。。。
	if len(genVal) != len(recvVal) {
		t.Fatal("len(genVal) != len(recvVal)")
	}
	for i, v := range genVal {
		if v != recvVal[i] {
			t.Fatalf("genVal[%d](%d) != recvVal[%d](%d)", i, v, i, recvVal[i])
		}
	}
}

func TestEngine_AddModule(t *testing.T) {
	modName := uuid.NewString()
	if err := RegisterModule(NewSimpleModule(modName, func(name string, config interface{}) (ModuleInstance, error) {
		return NewSimpleModuleInstance(modName, name, func(ctx context.Context, modCtx ModuleContext) error {
			return nil
		}), nil
	})); err != nil {
		t.Fatal(err)
	}
	if _, ok := moduleRegister[modName]; !ok {
		t.Fatal("Add module failed")
	}
}

func TestEngine_Stop(t *testing.T) {
	modName := uuid.NewString()
	mod1Count := &atomic.Int64{}
	eng := NewEngine()
	if err := RegisterModule(NewSimpleModule(modName, func(name string, config interface{}) (ModuleInstance, error) {
		return NewSimpleModuleInstance(modName, name, func(ctx context.Context, modCtx ModuleContext) error {
			mod1Count.Add(1)
			defer mod1Count.Add(-1)
			select {
			case _ = <-ctx.Done():
				return nil
			}
		}), nil
	})); err != nil {
		t.Fatal(err)
	} else if err := eng.Run(context.Background(), strings.NewReader(fmt.Sprintf(`
engine:
    Hello:
        module: %s
        parent: []
        queueSize: 1
        parallels: 1
        config: { }
`, modName))); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	eng.Stop()
	time.Sleep(time.Second)
	if mod1Count.Load() != 0 {
		t.Fatal("work node not exit ..")
	}
}

func TestEngine_WithMultiRoot(t *testing.T) {
	config := `
engine:
  Gen1:
    module: test/numberGen
    parent: []
    queueSize: 1
    parallels: 1
    config: 
      start: 1000
      step: 1000
  Gen2:
    module: test/numberGen
    parent: []
    queueSize: 1
    parallels: 1
    config:
      start: 5000
      step: 1000

  Recv:
    module: test/actual
    parent:
    - Gen1
    - Gen2
    queueSize: 1
    parallels: 1
    config: {}
`
	eng := NewEngine()
	nodeInstanceMap := map[interface{}]int{}
	if err := RegisterModule(NewSimpleModule("test/numberGen", func(name string, config interface{}) (ModuleInstance, error) {
		type numberGenConfig struct {
			Start int `json:"start"`
			Step  int `json:"step"`
		}
		configmap, err := ConfigMapUnmarshal(config, &numberGenConfig{})
		if err != nil {
			log.Panic(err)
		}
		inst := NewSimpleModuleInstance("test/numberGen", name, func(ctx context.Context, modCtx ModuleContext) error {
			for i := configmap.Start; i < configmap.Start+configmap.Step; i++ {
				modCtx.Collect(i)
			}
			modCtx.Collect(nil)
			return nil
		})
		nodeInstanceMap[inst]++
		return inst, nil
	})); err != nil {
		log.Panic(err)
	}

	numRecv := map[int]int{}
	endSignal := make(chan bool)
	if err := RegisterModule(NewSimpleModule("test/actual", func(name string, config interface{}) (ModuleInstance, error) {
		inst := NewSimpleModuleInstance("test/actual", name, func(ctx context.Context, modCtx ModuleContext) error {
			nilCount := 0
		LOOP:
			for msg := range modCtx.MessageQueue() {
				switch msg {
				case nil:
					nilCount++
					if nilCount == 2 {
						break LOOP
					}
				default:
					numRecv[msg.(int)]++
				}
			}
			endSignal <- true
			return nil
		})
		nodeInstanceMap[inst]++
		return inst, nil
	})); err != nil {
		log.Panic(err)
	}
	if err := eng.Run(context.Background(), strings.NewReader(config)); err != nil {
		log.Panic(err)
	}
	<-endSignal
	// ensure data
	for _, v := range numRecv {
		assert.Equal(t, 1, v)
	}
	assert.Equal(t, 2000, len(numRecv))
	// ensure all module only create one instance
	assert.Equal(t, 3, len(nodeInstanceMap))
	for _, v := range nodeInstanceMap {
		assert.Equal(t, 1, v)
	}
}

func TestEngine_WithThreeRoot(t *testing.T) {
	config := `
engine:
  Gen1:
    module: test/numberGen_r3
    parent: []
    queueSize: 1
    parallels: 1
    config: 
      start: 1000
      step: 1000
  Gen2:
    module: test/numberGen_r3
    parent: []
    queueSize: 1
    parallels: 1
    config:
      start: 5000
      step: 1000
  Gen3:
    module: test/numberGen_r3
    parent: []
    queueSize: 1
    parallels: 1
    config:
      start: 10000
      step: 1000

  Recv:
    module: test/actual_r3
    parent:
    - Gen1
    - Gen2
    - Gen3
    queueSize: 1
    parallels: 1
    config: {}
`
	eng := NewEngine()
	nodeInstanceMap := map[interface{}]int{}
	if err := RegisterModule(NewSimpleModule("test/numberGen_r3", func(name string, config interface{}) (ModuleInstance, error) {
		type numberGenConfig struct {
			Start int `json:"start"`
			Step  int `json:"step"`
		}
		configmap, err := ConfigMapUnmarshal(config, &numberGenConfig{})
		if err != nil {
			log.Panic(err)
		}
		inst := NewSimpleModuleInstance("test/numberGen_r3", name, func(ctx context.Context, modCtx ModuleContext) error {
			for i := configmap.Start; i < configmap.Start+configmap.Step; i++ {
				modCtx.Collect(i)
			}
			modCtx.Collect(nil)
			return nil
		})
		nodeInstanceMap[inst]++
		return inst, nil
	})); err != nil {
		log.Panic(err)
	}

	numRecv := map[int]int{}
	endSignal := make(chan bool)
	if err := RegisterModule(NewSimpleModule("test/actual_r3", func(name string, config interface{}) (ModuleInstance, error) {
		inst := NewSimpleModuleInstance("test/actual_r3", name, func(ctx context.Context, modCtx ModuleContext) error {
			nilCount := 0
		LOOP:
			for msg := range modCtx.MessageQueue() {
				switch msg {
				case nil:
					nilCount++
					if nilCount == 3 {
						break LOOP
					}
				default:
					numRecv[msg.(int)]++
				}
			}
			endSignal <- true
			return nil
		})
		nodeInstanceMap[inst]++
		return inst, nil
	})); err != nil {
		log.Panic(err)
	}
	if err := eng.Run(context.Background(), strings.NewReader(config)); err != nil {
		log.Panic(err)
	}
	<-endSignal
	// ensure data
	for _, v := range numRecv {
		assert.Equal(t, 1, v)
	}
	assert.Equal(t, 3000, len(numRecv))
	// ensure all module only create one instance
	assert.Equal(t, 4, len(nodeInstanceMap))
	for _, v := range nodeInstanceMap {
		assert.Equal(t, 1, v)
	}
}

func BenchmarkEngine_Send(b *testing.B) {
	receiver := func(done *sync.WaitGroup, max int) ModuleFactory {
		return NewSimpleModule("receiver", func(name string, config interface{}) (ModuleInstance, error) {
			return NewSimpleModuleInstance("receiver", name, func(ctx context.Context, modCtx ModuleContext) error {
				defer done.Done()
				cnt := 0
				for _ = range modCtx.MessageQueue() {
					cnt++
					if cnt == max {
						break
					}
				}
				return nil
			}), nil
		})
	}
	sender := func(max int) ModuleFactory {
		return NewSimpleModule("sender", func(name string, config interface{}) (ModuleInstance, error) {
			return NewSimpleModuleInstance("sender", name, func(ctx context.Context, modCtx ModuleContext) error {
				b.ResetTimer()
				for i := 0; i < max; i++ {
					modCtx.Collect(i)
				}
				return nil
			}), nil
		})
	}

	eng := NewEngine()
	wg := sync.WaitGroup{}
	wg.Add(1)
	if err := RegisterModule(receiver(&wg, b.N)); err != nil {
		b.Fatal(err)
	}
	if err := RegisterModule(sender(b.N)); err != nil {
		b.Fatal(err)
	}
	startAt := time.Now()
	b.ResetTimer()
	b.Log(fmt.Sprintf("BenchmarkEngine_Send with %d items", b.N))
	err := eng.Run(context.Background(), strings.NewReader(`
engine:
  benchmarkSender:
    module: sender
    parent: []
    queueSize: 1
    parallels: 1
    config: {}

  benchmarkReceiver:
    module: receiver
    parent:
    - benchmarkSender
    queueSize: 1
    parallels: 1
    config: {}
`))
	if err != nil {
		b.Fatal(err)
	}
	wg.Wait()
	endAt := time.Now()
	b.Log(fmt.Sprintf("Send %d msg in %s throughput %0.2f/s", b.N, endAt.Sub(startAt).String(), float64(b.N)/(endAt.Sub(startAt).Seconds())))
}
