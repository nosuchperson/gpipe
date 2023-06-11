package main

import (
	"context"
	"fmt"
	"github.com/goccy/go-graphviz"
	"github.com/nosuchperson/gpipe"
	"os"
	"time"
)

func main() {
	fd, err := os.Open("examples/num_pipeline/config.yml")
	if err != nil {
		panic(err)
	}
	eng := gpipe.NewEngine()
	gpipe.RegisterModule(NewPrintModule())
	gpipe.RegisterModule(NewRandomGenModule())
	gpipe.RegisterModule(NewDoubleModule())
	gpipe.RegisterModule(NewTrebleModule())
	if err := eng.Run(context.Background(), fd); err != nil {
		panic(err)
	}
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second * 5)
		if str, err := eng.GraphState(graphviz.XDOT); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(str)
		}
	}
	eng.Stop()
	time.Sleep(time.Hour)
}
