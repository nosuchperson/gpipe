package gpipe

import (
	"bytes"
	"context"
	"fmt"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"io"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	defaultQPSArrayCap = 32
)

type Engine struct {
	isRunning     bool
	slowThreshold time.Duration
	logger        Logger
	dgaRoots      []*moduleContext
	qpsArrayCap   int
}

func NewEngine(opts ...EngineOptions) *Engine {
	engine := &Engine{
		isRunning:     false,
		slowThreshold: -1,
		logger:        nil,
		dgaRoots:      []*moduleContext{},
		qpsArrayCap:   defaultQPSArrayCap,
	}
	for _, opt := range opts {
		opt(engine)
	}
	// default logger
	if engine.logger == nil {
		engine.logger = createDefaultLogger()
	}

	return engine
}

func (e *Engine) Run(ctx context.Context, cfg io.Reader) error {
	if e.isRunning {
		return ErrEngineIsRunning
	}
	configMap, err := e.loadConfig(cfg)
	if err != nil {
		return err
	}

	nodesMap := map[string]*moduleContext{}
	for workerName, workerCfg := range e.listRootNodeMap(configMap.Engine) {
		rootCtx, rootCancel := context.WithCancel(ctx)
		if node, err := e.prepareNode(&nodesMap, rootCtx, rootCancel, configMap.Engine, workerName, workerCfg); err != nil {
			return err
		} else {
			e.dgaRoots = append(e.dgaRoots, node)
		}
	}
	for _, root := range e.dgaRoots {
		e.startNode(root)
	}
	e.isRunning = true
	return nil
}

func (e *Engine) Stop() {
	for _, root := range e.dgaRoots {
		root.stop()
	}
}

func (e *Engine) loadConfig(cfg io.Reader) (*Config, error) {
	configMap := &Config{}
	if err := yaml.NewDecoder(cfg).Decode(&configMap); err != nil {
		return nil, err
	}
	return configMap, configMap.Valid()
}

// listRootNodeMap 获取所有根节点
func (e *Engine) listRootNodeMap(nodes map[string]*WorkNodeConfig) map[string]*WorkNodeConfig {
	ret := map[string]*WorkNodeConfig{}
	for name, cfg := range nodes {
		if len(cfg.Parent) == 0 {
			ret[name] = cfg
		}
	}
	return ret
}

func (e *Engine) prepareNode(nodesMap *map[string]*moduleContext, ctx context.Context, stop context.CancelFunc, fullConfig map[string]*WorkNodeConfig, nodeName string, nodeConfig *WorkNodeConfig) (*moduleContext, error) {
	curNodeName := fmt.Sprintf("%s[%s]", nodeConfig.Module, nodeName)
	var node *moduleContext = nil
	if existsNode, ok := (*nodesMap)[curNodeName]; ok {
		return existsNode, nil
	}

	modFactory, modInst, err := e.genNodeModule(nodeName, nodeConfig)
	if err != nil {
		return nil, err
	}
	zap.S().With("nodesMap", fmt.Sprintf("%p", nodesMap), "nodesMapValue", *nodesMap).Info("prepareNode")

	node = &moduleContext{
		engine:          e,
		name:            curNodeName,
		ctx:             ctx,
		stop:            stop,
		module:          modFactory,
		moduleInst:      modInst,
		engCfg:          nodeConfig,
		input:           make(chan interface{}, nodeConfig.QueueSize),
		parallelsCancel: make([]context.CancelFunc, 0, nodeConfig.Parallels),
		downstream:      make([]*moduleContext, 0, 4),
		isRunning:       false,
		recvCount:       atomic.Uint64{},
		sendCount:       atomic.Uint64{},
	}
	node.Init()
	(*nodesMap)[node.name] = node
	zap.S().With("nodeName", node.Name(), "ptr", fmt.Sprintf("%p", node), "mapPtr", fmt.Sprintf("%p", nodesMap)).Info("Node built")

	for downstreamName, downstreamConfig := range e.getDownstreamNode(fullConfig, nodeName) {
		if nodeContext, err := e.prepareNode(nodesMap, ctx, stop, fullConfig, downstreamName, downstreamConfig); err != nil {
			return nil, err
		} else {
			node.downstream = append(node.downstream, nodeContext)
		}
	}

	return node, nil
}

func (e *Engine) genNodeModule(nodeName string, nodeConfig *WorkNodeConfig) (ModuleFactory, ModuleInstance, error) {
	modFactory, err := GetModuleByName(nodeConfig.Module)
	if err != nil {
		return nil, nil, err
	}
	modInst, err := modFactory.New(nodeName, nodeConfig.Config)
	if err != nil {
		return nil, nil, err
	}
	return modFactory, modInst, nil
}

func (e *Engine) startNode(node *moduleContext) {
	if node.isRunning {
		return
	}
	node.isRunning = true
	for i := 0; i < node.engCfg.Parallels; i++ {
		parallelsContext, parallelCancel := context.WithCancel(node.ctx)
		node.parallelsCancel = append(node.parallelsCancel, parallelCancel)
		go func() {
			if err := node.moduleInst.Core(parallelsContext, node); err != nil {
				e.logger.Error(node, "module [%s] core error: %s", node.name, err)
			}
		}()
	}
	for _, downstream := range node.downstream {
		e.startNode(downstream)
	}
}

// getDownstreamNode 获取所有 parent == parentName 的节点
func (e *Engine) getDownstreamNode(fullConfig map[string]*WorkNodeConfig, nodeName string) map[string]*WorkNodeConfig {
	ret := map[string]*WorkNodeConfig{}
	for name, cfg := range fullConfig {
		for _, parent := range cfg.Parent {
			if parent == nodeName {
				ret[name] = cfg
			}
		}
	}
	return ret
}

func (e *Engine) GraphState(outputFormat graphviz.Format) (string, error) {
	g := graphviz.New()
	defer func() { g.Close() }()
	graph, err := g.Graph()
	if err != nil {
		return "", err
	} else {
		defer func() { graph.Close() }()
	}
	graph.SetRankDir(cgraph.TBRank)

	type tmpNode struct {
		nodeCtx *moduleContext
		node    *cgraph.Node
	}

	nameToNode := map[string]*tmpNode{}
	nodes := append([]*moduleContext{}, e.dgaRoots...)
	for len(nodes) > 0 {
		node := nodes[0]
		nodes = nodes[1:]

		if _, ok := nameToNode[node.name]; ok {
			continue
		}

		if gnode, err := graph.CreateNode(node.name); err != nil {
			return "", err
		} else {
			nameToNode[node.name] = &tmpNode{
				nodeCtx: node,
				node:    gnode,
			}
			gnode.SetShape("record")
			inQPSArray, outQPSArray := node.GetQPS()
			inQPSStr := make([]string, 0, len(inQPSArray))
			outQPSStr := make([]string, 0, len(inQPSArray))
			totalInQPS := uint64(0)
			totalOutQPS := uint64(0)

			for i := 0; i < len(inQPSArray); i++ {
				inQPSStr = append(inQPSStr, strconv.FormatUint(inQPSArray[i], 10))
				outQPSStr = append(outQPSStr, strconv.FormatUint(outQPSArray[i], 10))

				totalInQPS += inQPSArray[i]
				totalOutQPS += outQPSArray[i]
			}
			gnode.SetLabel(fmt.Sprintf(`{ %s | {<c1> Receive | <c2> %d } | {<c1> QueueCap | <c2> %d } | {<c1> QueueSize | <c2> %d } | { <c1> InQPS | <c2> %0.2f } | { <c1> OutQPS | <c2> %0.2f } | { <c1> InQPSArray | <c2> %s } | { <c1> OutQPSArray | <c2> %s } }`,
				node.Name(),
				node.recvCount.Load(),
				cap(node.input),
				len(node.input),
				float64(totalInQPS)/float64(len(inQPSArray)),
				float64(totalOutQPS)/float64(len(outQPSArray)),
				strings.Join(inQPSStr, " , "),
				strings.Join(outQPSStr, " , "),
			))
		}

		for _, downstream := range node.downstream {
			nodes = append(nodes, downstream)
		}
	}
	for _, node := range nameToNode {
		for _, downstream := range node.nodeCtx.downstream {
			nodes = append(nodes, downstream)
			if edge, err := graph.CreateEdge(fmt.Sprintf("Broadcast %s -> %s", node.node.Name(), downstream.name), node.node, nameToNode[downstream.name].node); err != nil {
				return "", err
			} else {
				edge.SetLabel(`Sent: ` + strconv.FormatUint(node.nodeCtx.sendCount.Load(), 10))
			}
		}
	}

	var buf bytes.Buffer
	if err := g.Render(graph, outputFormat, &buf); err != nil {
		return "", err
	} else {
		return buf.String(), nil
	}
}
