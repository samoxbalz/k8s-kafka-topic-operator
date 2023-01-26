package util

import (
	"github.com/Shopify/sarama"
	apiv1 "kafka.samoxbalz.io/api/v1"
)

func InitKafkaConnect(topic *apiv1.TopicSpec) (sarama.ClusterAdmin, error) {
	config := sarama.NewConfig()
	return sarama.NewClusterAdmin(topic.Brokers, config)
}

func CreateTopic(kafkaClient *sarama.ClusterAdmin, topic *apiv1.TopicSpec) error {

	return nil
}

func UpdateTopic(kafkaClient *sarama.ClusterAdmin, topic *apiv1.TopicSpec) error {

	return nil
}
