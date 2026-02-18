package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type WorkerConfig struct {
	AppEnv   string
	RedisURL string

	Kafka struct {
		Brokers        []string
		Topic          string
		GroupID        string
		DLQTopic       string
		PollMinBytes   int
		PollMaxBytes   int
		ReadTimeout    time.Duration
		CommitInterval time.Duration
		MaxRetries     int
		RetryBackoff   time.Duration
	}

	ShutdownTimeout time.Duration
	IdempotencyTTL  time.Duration
}

func LoadConfig() (WorkerConfig, error) {
	get := func(k, def string) string {
		v := strings.TrimSpace(os.Getenv(k))
		if v == "" {
			return def
		}
		return v
	}

	parseInt := func(k string, def int) (int, error) {
		s := get(k, "")
		if s == "" {
			return def, nil
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return 0, err
		}
		return n, nil
	}

	parseDur := func(k string, def time.Duration) (time.Duration, error) {
		s := get(k, "")
		if s == "" {
			return def, nil
		}
		d, err := time.ParseDuration(s)
		if err != nil {
			return 0, err
		}
		return d, nil
	}

	brokers := strings.Split(get("KAFKA_BROKERS", "localhost:9092"), ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}

	maxRetries, err := parseInt("WORKER_MAX_RETRIES", 5)
	if err != nil {
		return WorkerConfig{}, err
	}
	pollMaxBytes, err := parseInt("KAFKA_MAX_BYTES", 10<<20) // 10MB
	if err != nil {
		return WorkerConfig{}, err
	}
	pollMinBytes, err := parseInt("KAFKA_MIN_BYTES", 1)
	if err != nil {
		return WorkerConfig{}, err
	}

	retryBackoff, err := parseDur("WORKER_RETRY_BACKOFF", 750*time.Millisecond)
	if err != nil {
		return WorkerConfig{}, err
	}
	readTimeout, err := parseDur("KAFKA_READ_TIMEOUT", 10*time.Second)
	if err != nil {
		return WorkerConfig{}, err
	}
	commitInterval, err := parseDur("KAFKA_COMMIT_INTERVAL", 1*time.Second)
	if err != nil {
		return WorkerConfig{}, err
	}

	cfg := WorkerConfig{
		AppEnv: get("APP_ENV", "dev"),

		Kafka: struct {
			Brokers        []string
			Topic          string
			GroupID        string
			DLQTopic       string
			PollMinBytes   int
			PollMaxBytes   int
			ReadTimeout    time.Duration
			CommitInterval time.Duration
			MaxRetries     int
			RetryBackoff   time.Duration
		}{
			Brokers: brokers,
			Topic:   get("KAFKA_TOPIC", "petri.arena.events.v1"),
			GroupID: get("KAFKA_GROUP_ID", "petri-arena-projector"),

			MaxRetries:     maxRetries,
			RetryBackoff:   retryBackoff,
			PollMaxBytes:   pollMaxBytes,
			PollMinBytes:   pollMinBytes,
			ReadTimeout:    readTimeout,
			CommitInterval: commitInterval,
		},
	}

	if len(cfg.Kafka.Brokers) == 0 || cfg.Kafka.Topic == "" || cfg.Kafka.GroupID == "" {
		return WorkerConfig{}, errors.New("missing kafka envs")
	}
	if cfg.RedisURL == "" {
		return WorkerConfig{}, errors.New("REDIS_URL not set")
	}
	return cfg, nil
}
