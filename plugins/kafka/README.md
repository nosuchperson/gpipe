# kafka

kafka 通用模块, 基于 confluent-kafka-go

## kafka-producer

该模块的上游产生消息必须为 bytes, 该模块接受到数据后，将根据配置把 bytes 数据发送至 kafka

### Input

任意 bytes 数据

### Output

无

### 参数说明

#### Example

```yaml
config:
  topic: flow-clustering-dna-values
  async: false
  rdKafka:
    "bootstrap.servers": "kafka-cluster-kafka-bootstrap.chaoscube:9092"
    "client.id": go-flow-clustering-dev
    "group.id": go-flow-clustering-dev
```

|参数名|说明|
|:---:|:---|
|topic|kafka topic|
|async|是否启用异步模式，该模式会有更高吞吐，但是目前有 bug|
|rdKafka|librdkafka 的参数配置，具体参考 https://servaltech.feishu.cn/docx/doxcnHmuyaVxEFTcPJ45WM5XQ0f|


## kafka-consumer

从 kafka topic 中读取数据，并向下游输出数据对应的 bytes

### Input
无

### Output
bytes message


### 参数说明

#### Example

```yaml
config:
  topics:
    - adr-ultrax-event
  pollMs: 10
  config:
    "bootstrap.servers": "kafka-cluster-kafka-bootstrap.chaoscube:9092"
    "client.id": go-flow-clustering
    "group.id": go-flow-clustering
    "enable.auto.commit": true
    "auto.offset.reset": earliest
```

|   参数名   | 说明                                                                                  |
|:-------:|:------------------------------------------------------------------------------------|
| topics  | kafka topic。如果指定多个，则从所有指定的 topic 内获取数据                                              |
| pollMs  | poll 之间的间隔。单位为毫秒。在数据吞吐较高时，可以降低该值                                                    |
| config  | librdkafka 的参数配置，具体参考 https://servaltech.feishu.cn/docx/doxcnHmuyaVxEFTcPJ45WM5XQ0f |

