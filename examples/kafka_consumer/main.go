package main

import (
	"context"
	"github.com/nosuchperson/gpipe"
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
	log.Printf("module %s: %s", moduleContext.Name(), s)
}

func (l logger) Warn(moduleContext gpipe.ModuleContext, s string, i ...interface{}) {
	log.Printf("module %s: %s", moduleContext.Name(), s)
}

func (l logger) Error(moduleContext gpipe.ModuleContext, s string, i ...interface{}) {
	log.Printf("module %s: %s", moduleContext.Name(), s)
}

func (l logger) Trace(moduleContext gpipe.ModuleContext, s string, i ...interface{}) {
	log.Printf("module %s: %s", moduleContext.Name(), s)
}

func main() {
	cfg := `
engine:
  source:
    module: kafka-consumer
    parent: [ ]
    queueSize: 1
    parallels: 1
    config:
      topics:
        - gpw-test
      pollMs: 10
      config:
        "bootstrap.servers": "127.0.0.1:9092"
        "client.id": gpw-test
        "group.id": gpw-test
        "enable.auto.commit": true
        "auto.offset.reset": earliest
        "fetch.message.max.bytes": 1000000000

  sink:
    module: "sink/printer"
    parent:
      - source
    queueSize: 1
    parallels: 1
    config: {}
`

	engine := gpipe.NewEngine(gpipe.EngineWithLogger(&logger{}))
	if err := engine.Run(context.Background(), strings.NewReader(cfg)); err != nil {
		panic(err)
	}
	time.Sleep(time.Hour * 24 * 365)
}
