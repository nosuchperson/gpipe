package main

import (
	"context"
	"fmt"
	"github.com/nosuchperson/gpipe"
	_ "github.com/nosuchperson/gpipe/plugins/interval"
	_ "github.com/nosuchperson/gpipe/plugins/kafka"
	_ "github.com/nosuchperson/gpipe/plugins/printer"
	"log"
	"strings"
	"time"
)

type logger struct {
}

func (l logger) ModuleStarted(ctx gpipe.ModuleContext) {
	log.Printf("module started: %s", ctx.Name())
}

func (l logger) ModuleStopped(ctx gpipe.ModuleContext) {
	log.Printf("module stopped: %s", ctx.Name())
}

func (l logger) Info(moduleContext gpipe.ModuleContext, s string, i ...interface{}) {
	log.Printf("module %s: %s", moduleContext.Name(), fmt.Sprintf(s, i...))
}

func (l logger) Warn(moduleContext gpipe.ModuleContext, s string, i ...interface{}) {
	log.Printf("module %s: %s", moduleContext.Name(), fmt.Sprintf(s, i...))
}

func (l logger) Error(moduleContext gpipe.ModuleContext, s string, i ...interface{}) {
	log.Printf("module %s: %s", moduleContext.Name(), fmt.Sprintf(s, i...))
}

func (l logger) Trace(moduleContext gpipe.ModuleContext, s string, i ...interface{}) {
	log.Printf("module %s: %s", moduleContext.Name(), fmt.Sprintf(s, i...))
}

func main() {
	cfg := `
engine:
  source:
    module: interval
    parent: [ ]
    queueSize: 1
    parallels: 1
    config:
      interval: 1000

  generate:
    module: random
    parent: 
      - source
    queueSize: 1
    parallels: 1
    config: {}

  sink:
    module: kafka-producer
    parent:
      - generate
    queueSize: 1
    parallels: 1
    config:
      topic: gpw-test
      async: false
      rdKafka:
        "bootstrap.servers": "127.0.0.1:9092"
        "client.id": gpw-test
        "fetch.message.max.bytes": 1000000000
        "message.max.bytes": 67108864
`

	gpipe.RegisterModule(gpipe.NewSimpleModule("random", func(name string, config interface{}) (gpipe.ModuleInstance, error) {
		return gpipe.NewSimpleModuleInstance("random", name, func(ctx context.Context, modCtx gpipe.ModuleContext) error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case _ = <-modCtx.MessageQueue():
					modCtx.Collect([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
				}
			}
		}), nil
	}))

	engine := gpipe.NewEngine(gpipe.EngineWithLogger(&logger{}))
	if err := engine.Run(context.Background(), strings.NewReader(cfg)); err != nil {
		panic(err)
	}
	time.Sleep(time.Hour * 24 * 365)
}
