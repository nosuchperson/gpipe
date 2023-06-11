package gpipe

var (
	moduleRegister = map[string]ModuleFactory{}
)

func RegisterModule(m ModuleFactory) error {
	if _, exists := moduleRegister[m.Name()]; exists {
		return newGPWError("module %s already exists", m.Name())
	}
	moduleRegister[m.Name()] = m
	return nil
}

func GetModuleByName(name string) (ModuleFactory, error) {
	if m, exists := moduleRegister[name]; exists {
		return m, nil
	}
	return nil, newGPWError("module %s not exists", name)
}
