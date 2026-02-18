package kafka

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type ConsumerConfig struct {
	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string
	KafkaDLQ     string

	MaxRetries     int
	RetryBackoff   time.Duration
	PollMaxBytes   int
	PollMinBytes   int
	ReadTimeout    time.Duration
	CommitInterval time.Duration
}

func LoadConsumerConfig() (ConsumerConfig, error) {
	var cfg ConsumerConfig

	// Required
	brokers := strings.TrimSpace(os.Getenv("KAFKA_BROKERS")) // ex: "localhost:9092,localhost:9093"
	topic := strings.TrimSpace(os.Getenv("KAFKA_TOPIC"))
	group := strings.TrimSpace(os.Getenv("KAFKA_GROUP_ID"))
	dlq := strings.TrimSpace(os.Getenv("KAFKA_DLQ_TOPIC"))

	if brokers == "" || topic == "" || group == "" || dlq == "" {
		return cfg, fmt.Errorf("missing required env vars: KAFKA_BROKERS, KAFKA_TOPIC, KAFKA_GROUP_ID, KAFKA_DLQ_TOPIC")
	}

	cfg.KafkaBrokers = splitCSV(brokers)
	cfg.KafkaTopic = topic
	cfg.KafkaGroupID = group
	cfg.KafkaDLQ = dlq

	// Optional with defaults
	cfg.MaxRetries = envInt("KAFKA_MAX_RETRIES", 5)
	cfg.RetryBackoff = envDuration("KAFKA_RETRY_BACKOFF", 200*time.Millisecond)

	cfg.PollMinBytes = envInt("KAFKA_POLL_MIN_BYTES", 1)
	cfg.PollMaxBytes = envInt("KAFKA_POLL_MAX_BYTES", 10_000_000)

	cfg.ReadTimeout = envDuration("KAFKA_READ_TIMEOUT", 2*time.Second)
	cfg.CommitInterval = envDuration("KAFKA_COMMIT_INTERVAL", 1*time.Second)

	return cfg, nil
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func envInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func envDuration(key string, def time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
