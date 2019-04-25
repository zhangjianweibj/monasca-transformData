package main

/*
author:zhangjianweibj
date:2019-3-22
*/
import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spf13/viper"
	"github.com/zhangjianweibj/monasca-transformData/models"
	"os"
	"runtime"
	"strings"
	"time"
)

var (
	config = initConfig()
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	//log.SetOutput(os.Stdout)
	//log.SetLevel(log.InfoLevel)
	logfile := config.GetString("logging.file")
	if logfile != "" {
		f, err := os.Create(logfile)
		if err != nil {
			log.SetOutput(os.Stdout)
			log.Fatalf("Failed to create log file: %v", err)
		}
		log.SetOutput(f)
	} else {
		log.SetOutput(os.Stdout)
	}

	loglevel := config.GetString("logging.level")
	switch strings.ToUpper(loglevel) {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}
}

func initConfig() *viper.Viper {
	config := viper.New()
	config.SetDefault("consumerTopic", "custom-metrics")
	config.SetDefault("producerTopic", "metrics")
	config.SetDefault("kafka.bootstrap.servers", "localhost:9092")
	config.SetDefault("kafka.group.id", "monasca-transformData")
	config.SetDefault("tenantId", "1231245")
	config.SetDefault("kafka.max.threads", 1)
	config.SetDefault("poolFactor", 100)
	config.SetConfigName("config")
	config.AddConfigPath(".")
	err := config.ReadInConfig()

	if err != nil {
		log.Fatalf("Fatal error reading config file: %s", err)
	}
	return config
}

func initConsumer(consumerTopic, groupID, bootstrapServers string) *kafka.Consumer {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               bootstrapServers,
		"group.id":                        groupID,
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"enable.auto.commit":              false,
		"default.topic.config":            kafka.ConfigMap{"auto.offset.reset": "earliest"},
	})

	if err != nil {
		log.Warnf("Failed to create consumer: %s", err)
		return nil
	}

	log.Infof("Created kafka consumer %v", c)

	err = c.Subscribe(consumerTopic, nil)

	if err != nil {
		log.Warnf("Failed to subscribe to topics %c", err)
		return nil
	}
	log.Infof("Subscribed to topic %s as group %s", consumerTopic, groupID)

	return c
}

func initProducer(bootstrapServers string) *kafka.Producer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":            bootstrapServers,
		"go.batch.producer":            true,
		"queue.buffering.max.messages": 100000,
		"message.timeout.ms":           5000,
	})

	if err != nil {
		log.Fatalf("Failed to create producer: %s", err)
	}

	log.Infof("Created kafka producer %v", p)

	return p
}

//transform input message tenantId property to admin
func processMessage(msg *kafka.Message, bound chan *models.MetricEnvelope, tenant string) {
	//unmarshal kafka message to MetricEnvelope
	metricEnvelope := models.MetricEnvelope{}
	err := json.Unmarshal([]byte(msg.Value), &metricEnvelope)
	if err != nil {
		log.Warnf("%% Invalid metric envelope on %s:%s", msg.TopicPartition, string(msg.Value))
		return
	}
	log.Debugf("before transform++ %#v", metricEnvelope)
	if metricEnvelope.Meta != nil {
		metricEnvelope.Meta["tenantId"] = tenant
	}
	//transform accountname to keystone user
	getKeystoneUserByAccountName("")

	log.Debugf("after transform++ %#v", metricEnvelope)
	bound <- &metricEnvelope
	log.Debugf("after bound++ %#v", metricEnvelope)
}

func sendMessage(msg chan *models.MetricEnvelope, p *kafka.Producer, topic string) {
	//go routine must at loop status.
	for {
		log.Debugf("send message before ++")
		deliveryChan := make(chan kafka.Event)
		message := <-msg
		value, _ := json.Marshal(message)
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          value,
		}, deliveryChan)

		e := <-deliveryChan
		m := e.(*kafka.Message)

		if m.TopicPartition.Error != nil {
			log.Warnf("Delivery failed: %v\n", m.TopicPartition.Error)
		} else {
			log.Printf("Delivered message to topic %s [%d] at offset %v\n",
				*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
		}
		close(deliveryChan)
		log.Debugf("send message after ++")
	}
}

func kafkaMessageProcess(c *kafka.Consumer, message chan *models.MetricEnvelope, tenantId string) {
	for true {
		select {
		case ev := <-c.Events():
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				log.Debugf("AssignedPartitions: %v\n", e)
				c.Assign(e.Partitions)
			case kafka.RevokedPartitions:
				log.Debugf("RevokedPartitions: %% %v\n", e)
				c.Unassign()
			case *kafka.Message:
				processMessage(e, message, tenantId)
				//commit offset at most consume once
				c.Commit()
			case kafka.PartitionEOF:
				log.Warnf("%% Reached %v\n", e)
			case kafka.Error:
				log.Warnf("%% Error: %v\n", e)
			}
		}
	}
}

func getKeystoneUserByAccountName(accountName string) (user string, err error) {
	//get user information from Gcache,LRU method evit cache key.
	//"github.com/bluele/gcache"

	return "", nil
}

func main() {

	consumerTopic := config.GetString("consumerTopic")
	producerTopic := config.GetString("producerTopic")
	threads := config.GetInt("kafka.max.threads")
	poolFactor := config.GetInt("poolFactor")
	if threads <= 0 {
		threads = 1
	}
	//set metrics pool factor default value.
	if poolFactor <= 0 {
		poolFactor = 100
	}

	bootstrapServers := config.GetString("kafka.bootstrap.servers")
	groupID := config.GetString("kafka.group.id")

	tenantId := config.GetString("tenantId")

Loop:
	c := initConsumer(consumerTopic, groupID, bootstrapServers)
	if c == nil {
		time.Sleep(time.Second * 5)
		goto Loop
	}
	defer c.Close()

	message := make(chan *models.MetricEnvelope, threads*poolFactor)
	p := initProducer(bootstrapServers)
	defer p.Close()

	log.Debugf("after go sendMessage::")
	kafkaMessageProcess(c, message, tenantId)

	for i := 0; i < threads; i++ {
		go sendMessage(message, p, producerTopic)
	}
}
