package gpipe

import "time"

type EngineOptions func(engine *Engine)

func EngineWithLogger(logger Logger) EngineOptions {
	return func(engine *Engine) {
		engine.logger = logger
	}
}

func EngineWithSlowThresholdMs(thresholdMs time.Duration) EngineOptions {
	return func(engine *Engine) {
		engine.slowThreshold = thresholdMs
	}
}
