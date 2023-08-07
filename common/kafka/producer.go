package kafka

import (
	"gfa/common/log"
	"gfa/core/global"
	"github.com/Shopify/sarama"
	"time"
)

var (
	producer sarama.AsyncProducer
	err      error
	C        Config
	Queue    = make(chan []byte, C.MaxQueue)
)

func Init() error {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Timeout = 5 * time.Second
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.MaxMessageBytes = 1024 * 1024 * 5
	kafkaConfig.Producer.Compression = sarama.CompressionSnappy
	producer, err = sarama.NewAsyncProducer(C.Brokers, kafkaConfig)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case msg := <-producer.Successes():
				//fmt.Printf("Send msg to kafka success, topic:%s, partition:%d, offset:%d\n", msg.Topic, msg.Partition, msg.Offset)
				_ = msg
			case err = <-producer.Errors():
				log.Errorf("Failed to send message: %s", err)
			case <-global.Ctx.Done():
				return
			}
		}
	}()
	for i := 0; i < C.Workers; i++ {
		go func() {
			for {
				select {
				case msg := <-Queue:
					send(msg)
				case <-global.Ctx.Done():
					return
				}
			}
		}()
	}
	return nil
}

func Close() {
	producer.AsyncClose()
}

func Q(msg []byte) {
	select {
	case Queue <- msg:
	default:
	}
}

func send(msg []byte) {
	producer.Input() <- &sarama.ProducerMessage{Topic: C.Topic, Value: sarama.StringEncoder(msg)}
}
