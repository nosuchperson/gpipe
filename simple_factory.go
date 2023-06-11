package gpipe

import (
	"context"
)

// simpleModule 一个简易的通用模块，使用 NewSimpleModule 输入 func 即可构造，省点事
type simpleModule struct {
	name    string
	newFunc func(name string, config interface{}) (ModuleInstance, error)
}

func (s *simpleModule) Name() string {
	return s.name
}

func (s *simpleModule) New(name string, config interface{}) (ModuleInstance, error) {
	return s.newFunc(name, config)
}

func NewSimpleModule(name string, newFunc func(name string, config interface{}) (ModuleInstance, error)) ModuleFactory {
	return &simpleModule{
		name:    name,
		newFunc: newFunc,
	}
}

// simpleModuleInstance 用于构造一个状态无关的 instance，便于一些简单的 instance 开发
type simpleModuleInstance struct {
	modName string
	name    string
	core    func(ctx context.Context, modCtx ModuleContext) error
}

func (s simpleModuleInstance) Core(ctx context.Context, modCtx ModuleContext) error {
	return s.core(ctx, modCtx)
}

func NewSimpleModuleInstance(modName, name string, callFunc func(ctx context.Context, modCtx ModuleContext) error) ModuleInstance {
	return &simpleModuleInstance{
		modName: modName,
		name:    name,
		core:    callFunc,
	}
}
