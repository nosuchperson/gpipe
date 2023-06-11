package kafka

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/nosuchperson/gpipe"
)

const kafkaConsumerModuleName = "kafka/consumer"

type kafkaConsumerConfig struct {
	Topics []string        `yaml:"topics"`
	PollMs int             `yaml:"pollMs"`
	Config kafka.ConfigMap `yaml:"config"`
}

type kafkaConsumerModule struct {
	name           string
	topics         []string
	pollMs         int
	kafkaConfigMap kafka.ConfigMap
}

func init() {
	gpipe.RegisterModule(NewKafkaConsumerModule())
}
func NewKafkaConsumerModule() gpipe.ModuleFactory {
	return gpipe.NewSimpleModule(kafkaConsumerModuleName, func(name string, config interface{}) (gpipe.ModuleInstance, error) {
		if configMap, err := gpipe.ConfigMapUnmarshal(config, &kafkaConsumerConfig{}); err != nil {
			return nil, err
		} else {
			return &kafkaConsumerModule{
				name:           name,
				topics:         configMap.Topics,
				pollMs:         configMap.PollMs,
				kafkaConfigMap: configMap.Config,
			}, nil
		}
	})
}

func (k *kafkaConsumerModule) ModuleName() string {
	return kafkaConsumerModuleName
}

func (k *kafkaConsumerModule) InstanceName() string {
	return k.name
}

func (k *kafkaConsumerModule) Core(ctx context.Context, modCtx gpipe.ModuleContext) error {
	for {
		select {
		case _ = <-ctx.Done():
			return nil
		default:
			k.consumer(ctx, modCtx)
		}
	}
}

func (k *kafkaConsumerModule) consumer(ctx context.Context, modCtx gpipe.ModuleContext) error {
	kafkaConsumer, err := kafka.NewConsumer(&k.kafkaConfigMap)
	if err != nil {
		return err
	}
	defer func() {
		kafkaConsumer.Unsubscribe()
		kafkaConsumer.Close()
	}()
	if err := kafkaConsumer.SubscribeTopics(k.topics, func(consumer *kafka.Consumer, event kafka.Event) error {
		modCtx.Logger().Warn(modCtx, "KafkaConsumer ReBalancing")
		return nil
	}); err != nil {
		return err
	}

	// 扫描输入
	for {
		select {
		case _ = <-ctx.Done():
			return nil
		default:
			ev := kafkaConsumer.Poll(k.pollMs)
			if ev == nil {
				continue
			}
			switch e := ev.(type) {
			case *kafka.Message:
				modCtx.Collect(e.Value)
			case kafka.Error:
				// Errors should generally be considered
				// informational, the client will try to
				// automatically recover.
				// But in this example we choose to terminate
				// the application if all brokers are down.
				fmt.Printf("%% Error: %v: %v\n", e.Code(), e)
				if e.Code() == kafka.ErrAllBrokersDown {
					return e
				}
			}
		}
	}
}
