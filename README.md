# gPipeWorker

基于 Go 实现一个类似于当前 pyEngine 的方案

运行结构类似于树

```
             / - Bo -> D -> F -> G
  /- Ao -> B
A            \ - Bo - E
  \                   /
   \                 /
    \- Ao -> C <-Eo-/
             ^
    / - Zo --|
  /
Z -- Zo -> X
```

A 的输出会复制两份传输给 B 和 C

B 的输出会复制两份传给 D 和 E

E 的输出会传输给 C

C 同时接受 A 和 E 和 Z 的数据

# 配置

同一套配置内，可以有多个 Root

```yaml
engine:
  起一个代号名:
    module: RandomGen # 需要使用 engine.Add() 添加过
    parent: [ ]        # 上级节点的「代号名」，如果为空则代表是起点
    queueSize: 1       # 这个节点的 InputQueue Size， 0 为完全阻塞，即不存在缓存队列
    parallels: 1       # 这个节点的并行数，所有的并行是基于同一个 Instance 的，并分享相同的 InputQueue
    config: 
      name: "随便写一下，这个地方的配置取决于 module.Config 咋配置的"

  TrebleIt: # 也是随便一个名字
    module: Treble # 模块名
    parent:
      - 起一个代号名 # 该节点的输入是 「起一个代号名」
    queueSize: 1
    parallels: 1
    config: {} # 也可以没有内容，但是要有这个 key

```

# module

自带的 module 为

|   moduleName   |                       desc                       |
|:--------------:|:------------------------------------------------:|
|   blackhole    |                 黑洞, 所有进来的消息都被丢弃                  |
|    interval    | 定时信号，根据 config 内的 interval（单位 ms） 来休眠，休眠后向下游发送信号 |
| kafka-consumer |                    kafka 消费者                     |
| kafka-producer |                    kafka 生产者                     |

# TODO

1. ~~一个 node 有多个上级时，两个上级的 output 应该会合并进 同一个 node inst 内，而不是现在这样反直觉~~ Done

2. ~~提供基于 node 的 perfcounter(time, in/out) 之类的~~ Done