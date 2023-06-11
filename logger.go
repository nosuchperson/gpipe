package gpipe

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	ModuleStarted(ctx ModuleContext)
	ModuleStopped(ctx ModuleContext)
	Info(ModuleContext, string, ...interface{})
	Warn(ModuleContext, string, ...interface{})
	Error(ModuleContext, string, ...interface{})
	Trace(ModuleContext, string, ...interface{})
}

type defaultLogger struct {
	log *zap.Logger
}

func (d *defaultLogger) ModuleStarted(ctx ModuleContext) {
	d.log.Sugar().With("module", ctx.Name()).Info("module started")
}

func (d *defaultLogger) ModuleStopped(ctx ModuleContext) {
	d.log.Sugar().With("module", ctx.Name()).Info("module stopped")
}

func (d *defaultLogger) Info(context ModuleContext, s string, i ...interface{}) {
	d.log.Sugar().With("module", context.Name()).Infof(s, i...)
}

func (d *defaultLogger) Warn(context ModuleContext, s string, i ...interface{}) {
	d.log.Sugar().With("module", context.Name()).Warnf(s, i...)
}

func (d *defaultLogger) Error(context ModuleContext, s string, i ...interface{}) {
	d.log.Sugar().With("module", context.Name()).Errorf(s, i...)
}

func (d *defaultLogger) Trace(context ModuleContext, s string, i ...interface{}) {
	d.log.Sugar().With("module", context.Name()).Debugf(s, i...)
}

func createDefaultLogger() Logger {
	zc := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:       true,
		DisableCaller:     false,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "time",
			NameKey:        "name",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.TimeEncoderOfLayout("2006/01/02 15:04:05"),
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields:    map[string]interface{}{"app": "gpw"},
	}
	if lg, err := zc.Build(); err != nil {
		panic(err)
	} else {
		return &defaultLogger{log: lg}
	}
}
