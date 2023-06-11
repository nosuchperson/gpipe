package gpipe

import "fmt"

type WorkNodeConfig struct {
	Module    string      `yaml:"module"`
	Parent    []string    `yaml:"parent"`
	QueueSize int         `yaml:"queueSize"`
	Parallels int         `yaml:"parallels"`
	Config    interface{} `yaml:"config"`
}
type Config struct {
	Engine map[string]*WorkNodeConfig `yaml:"engine"`
}

func (cfg *Config) Valid() error {
	if err := cfg.workerNameIsValid(); err != nil {
		return err
	} else if err := cfg.hasInvalidParent(); err != nil {
		return err
	} else if err := cfg.hasCycle(); err != nil {
		return err
	}
	return nil
}

func (cfg *Config) workerNameIsValid() error {
	// 检查 WorkName 的唯一性
	workerNameMap := map[string]bool{}
	for name := range cfg.Engine {
		if workerNameMap[name] {
			return newGPWError(fmt.Sprintf("duplicate worker name: %s", name))
		}
		workerNameMap[name] = true
	}
	return nil
}

// hasInvalidParent 检查是否有不存在的父节点
func (cfg *Config) hasInvalidParent() error {
	for name, nodeCfg := range cfg.Engine {
		for _, parent := range nodeCfg.Parent {
			if _, exists := cfg.Engine[parent]; !exists {
				return newGPWError(fmt.Sprintf("worker %s has invalid parent %s", name, parent))
			}
		}
	}
	return nil
}

func (cfg *Config) hasCycle() error {
	// 注意该检测只能最后最后一项检测
	// 构造完整图
	type detectNode struct {
		name     string
		indegree int
		cfg      *WorkNodeConfig
		next     []*detectNode
	}
	nodes := []*detectNode{}
	nodeMap := map[string]*detectNode{}

	// 构造所有点
	for name, nodeCfg := range cfg.Engine {
		node := &detectNode{
			name:     name,
			cfg:      nodeCfg,
			indegree: len(nodeCfg.Parent),
			next:     []*detectNode{}}
		nodes = append(nodes, node)
		nodeMap[name] = node
	}
	// 构造边
	for _, node := range nodes {
		for _, parentName := range node.cfg.Parent {
			nodeMap[parentName].next = append(nodeMap[parentName].next, node)
		}
	}

	getAndPopIndegreeZeroNode := func() *detectNode {
		remain := []*detectNode{}
		var res *detectNode = nil

		for _, node := range nodes {
			if res == nil && node.indegree == 0 {
				res = node
			} else {
				remain = append(remain, node)
			}
		}
		nodes = remain
		return res
	}

	for {
		node := getAndPopIndegreeZeroNode()
		if node == nil {
			break
		}
		// 将该点对应的边删除
		for _, next := range node.next {
			next.indegree--
		}
	}
	if len(nodes) > 0 {
		return newGPWError("has cycle")
	}
	return nil
}
