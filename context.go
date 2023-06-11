package gpipe

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// ModuleInstance 具体运行的上下文
type ModuleInstance interface {
	Core(ctx context.Context, modCtx ModuleContext) error
}

// ModuleFactory 用于描述一个 Module，并可以构造 ModuleInstance
type ModuleFactory interface {
	Name() string
	New(name string, config interface{}) (ModuleInstance, error)
}

// ModuleContext 用于公开暴露 moduleContext 这个私有结构的接口
type ModuleContext interface {
	Name() string
	Logger() Logger
	MessageQueue() chan interface{}
	Collect(v interface{})
	GetModuleFactory() ModuleFactory
	GetModuleInstance() ModuleInstance
}

// moduleContext 用于描述每个 Instance 在运行时的状态, 并可对其进行部分控制
type moduleContext struct {
	engine          *Engine
	name            string               // 名称, module 的实例名称
	ctx             context.Context      // 该 Node 的主 ctx
	stop            context.CancelFunc   // 该 Node 的停止信号
	module          ModuleFactory        // 关联的模块
	moduleInst      ModuleInstance       // 关联的实例
	engCfg          *WorkNodeConfig      // 节点的 Engine 配置
	input           chan interface{}     // 该节点的输入口, 输出口为该 Node 的下游接口, 该节点退出应该就自动释放
	parallelsCancel []context.CancelFunc // ctrl -> 控制每个 goroutine 是否退出, 主要是 parallels 的控制, 每个 parallels 可以独立控制是否退出，便于动态扩容起停
	downstream      []*moduleContext     // 下游节点
	// ======= 统计计数器 =======
	isRunning   bool
	recvCount   atomic.Uint64
	sendCount   atomic.Uint64
	qpsOverflow bool
	qpsSeek     int
	recvQPS     []uint64
	sendQPS     []uint64
}

// Init 用于初始化一些帮助线程
func (m *moduleContext) Init() {
	m.recvCount.Swap(0)
	m.sendCount.Swap(0)
	m.qpsOverflow = false
	m.qpsSeek = 0
	m.recvQPS = make([]uint64, m.engine.qpsArrayCap)
	m.sendQPS = make([]uint64, m.engine.qpsArrayCap)
	go m.qpsMonitor()
}

func (m *moduleContext) qpsMonitor() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	lastSendCount := uint64(0)
	lastRecvCount := uint64(0)
	for {
		select {
		case _ = <-m.ctx.Done():
			return
		case _ = <-ticker.C:
			curSendCount := m.sendCount.Load()
			curRecvCount := m.recvCount.Load()
			m.recvQPS[m.qpsSeek] = curRecvCount - lastRecvCount
			m.sendQPS[m.qpsSeek] = curSendCount - lastSendCount
			lastRecvCount = curRecvCount
			lastSendCount = curSendCount
			m.qpsSeek++
			// 不 reset 6 年多就嗝屁了
			if m.qpsSeek >= m.engine.qpsArrayCap {
				m.qpsOverflow = true
				m.qpsSeek = 0
			}
		}
	}
}
func (m *moduleContext) GetQPS() (recv, sent []uint64) {
	size := m.qpsSeek
	if size == 0 {
		return []uint64{}, []uint64{}
	}
	if m.qpsOverflow {
		size = m.engine.qpsArrayCap
	}
	recv = make([]uint64, size)
	sent = make([]uint64, size)
	copy(recv, m.recvQPS[:size-1])
	copy(sent, m.sendQPS[:size-1])
	return
}

func (m *moduleContext) Name() string {
	return m.name
}

func (m *moduleContext) Logger() Logger {
	return m.engine.logger
}

func (m *moduleContext) Collect(v interface{}) {
	startAt := time.Now()
	for _, down := range m.downstream {
		down.input <- v
		consume := time.Now().Sub(startAt)
		startAt = startAt.Add(consume)
		if m.engine.slowThreshold > 0 && consume > m.engine.slowThreshold {
			m.Logger().Warn(m, fmt.Sprintf("Detect backpress: %s --[%v ms]--> %s", m.Name(), consume.Milliseconds(), down.Name()))
		}
		down.recvCount.Add(1)
	}
	m.sendCount.Add(1)
}

func (m *moduleContext) MessageQueue() chan interface{} {
	return m.input
}

func (m *moduleContext) GetModuleFactory() ModuleFactory {
	return m.module
}

func (m *moduleContext) GetModuleInstance() ModuleInstance {
	return m.moduleInst
}
