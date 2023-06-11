package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/nosuchperson/gpipe"
	"gopkg.in/yaml.v3"
	"sync"
	"time"
)

const kafkaProducerModuleName = "kafka/producer"

type kafkaProducerSerializer string

const (
	kafkaProducerSerializerJson kafkaProducerSerializer = "json"
	kafkaProducerSerializerYaml kafkaProducerSerializer = "yaml"
)

type kafkaProducerConfig struct {
	Topic      string                  `yaml:"topic"`
	Async      bool                    `yaml:"async"`
	Serializer kafkaProducerSerializer `yaml:"serializer"`
	// configMap Ref: https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md
	// 突然发现直接从配置撸这个不香么？改啥程序。。。
	RdKafka kafka.ConfigMap `yaml:"rdKafka"`
}

type kafkaProducerModule struct {
	name           string
	kafkaConfigMap *kafkaProducerConfig
	kafkaMsgPool   sync.Pool
}

func init() {
	gpipe.RegisterModule(NewKafkaProducerModule())
}

func NewKafkaProducerModule() gpipe.ModuleFactory {
	return gpipe.NewSimpleModule(kafkaProducerModuleName, func(name string, config interface{}) (gpipe.ModuleInstance, error) {
		if configMap, err := gpipe.ConfigMapUnmarshal(config, &kafkaProducerConfig{}); err != nil {
			return nil, err
		} else {
			return &kafkaProducerModule{
				name:           name,
				kafkaConfigMap: configMap,
				kafkaMsgPool: sync.Pool{New: func() interface{} {
					return &kafka.Message{
						TopicPartition: kafka.TopicPartition{
							Topic:     &configMap.Topic,
							Partition: kafka.PartitionAny,
						},
						Value:         nil,
						Key:           nil,
						Timestamp:     time.Time{},
						TimestampType: 0,
						Opaque:        nil,
						Headers:       nil,
					}
				}},
			}, nil
		}
	})
}

func (k *kafkaProducerModule) Core(ctx context.Context, modCtx gpipe.ModuleContext) error {
	for {
		select {
		case _ = <-ctx.Done():
			return nil
		default:
			k.producer(ctx, modCtx)
		}
	}
}

func (k *kafkaProducerModule) producer(ctx context.Context, modCtx gpipe.ModuleContext) {
	// 注意，这个函数产生的生产者并不是线程安全的，所以必须初始化在工作线程中
	kafkaProducer, err := kafka.NewProducer(&k.kafkaConfigMap.RdKafka)
	if err != nil {
		modCtx.Logger().Error(modCtx, "kafka producer init failed due to %c", err)
	}
	defer func() {
		// 刷新 Kafka 并且安心退出
		kafkaProducer.Flush(1000 * 10)
		kafkaProducer.Close()
	}()
	prodCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	eventCh := kafkaProducer.Events()
	// 启动监控
	go k.monitorEvent(prodCtx, eventCh)

	serializerFunc := k.generateSerializer(modCtx)
	producer := k.generateProducer(kafkaProducer, eventCh)

	for {
		select {
		case _ = <-prodCtx.Done():
			return
		case msg := <-modCtx.MessageQueue():
			if body, err := serializerFunc(msg); err != nil {
				modCtx.Logger().Error(modCtx, "serializing input data failed due to %v", err)
			} else {
				kMsg := k.kafkaMsgPool.Get().(*kafka.Message)
				kMsg.Value = body
				if err := producer(kMsg); err != nil {
					modCtx.Logger().Error(modCtx, "kafkaProducer.producer() failed due to %v", err)
				}
			}
		}
	}
}

func (k *kafkaProducerModule) monitorEvent(ctx context.Context, evCh chan kafka.Event) {
	for {
		select {
		case _ = <-ctx.Done():
			return
		case e := <-evCh:
			switch ev := e.(type) {
			case *kafka.Message:
				// The message delivery report, indicating success or
				// permanent failure after retries have been exhausted.
				// Application level retries won't help since the client
				// is already configured to do that.
				k.kafkaMsgPool.Put(ev)
			case kafka.Error:
				// Generic client instance-level errors, such as
				// broker connection failures, authentication issues, etc.
				//
				// These errors should generally be considered informational
				// as the underlying client will automatically try to
				// recover from any errors encountered, the application
				// does not need to take action on them.
				fmt.Printf("KafkProducerError: %v\n", ev)
			default:
				fmt.Printf("KafkaProducerIgnored event: %s\n", ev)
			}
		}
	}
}

func (k *kafkaProducerModule) generateSerializer(modCtx gpipe.ModuleContext) func(interface{}) ([]byte, error) {
	type serializer func(interface{}) ([]byte, error)
	var serializerFunc serializer = func(data interface{}) ([]byte, error) {
		if msg, ok := data.([]byte); !ok {
			return nil, errors.New(fmt.Sprintf("input data type is %T, not []byte", data))
		} else {
			return msg, nil
		}
	}
	switch k.kafkaConfigMap.Serializer {
	case kafkaProducerSerializerJson:
		serializerFunc = func(data interface{}) ([]byte, error) {
			return json.Marshal(data)
		}
	case kafkaProducerSerializerYaml:
		serializerFunc = func(data interface{}) ([]byte, error) {
			return yaml.Marshal(data)
		}
	default:
		modCtx.Logger().Warn(modCtx, "unknown serializer %s", k.kafkaConfigMap.Serializer)
	}
	return serializerFunc
}

func (k *kafkaProducerModule) generateProducer(kafkaProducer *kafka.Producer, eventCh chan kafka.Event) func(*kafka.Message) error {
	prodCh := kafkaProducer.ProduceChannel()
	asyncP := func(msg *kafka.Message) error {
		prodCh <- msg
		return nil
	}
	syncP := func(msg *kafka.Message) error {
		return kafkaProducer.Produce(msg, eventCh)
	}
	producer := syncP
	if k.kafkaConfigMap.Async {
		producer = asyncP
	}
	return producer
}
