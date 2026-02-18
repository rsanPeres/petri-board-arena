package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	workerConfig "github.com/petri-board-arena/internal/infrastructure/config"
	kafkaconsumer "github.com/petri-board-arena/internal/infrastructure/messaging/kafka"
	infraredis "github.com/petri-board-arena/internal/infrastructure/persistence/redis"
	"github.com/petri-board-arena/internal/infrastructure/projector"
)

func main() {
	workerCfg, err := workerConfig.LoadConfig()
	if err != nil {
		log.Fatalf("config(worker): %v", err)
	}

	consumerCfg, err := kafkaconsumer.LoadConsumerConfig()
	if err != nil {
		log.Fatalf("config(consumer): %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rdb, err := infraredis.NewRedisClient(workerCfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rdb.Close()

	proj := projector.NewProjector(rdb, workerCfg)

	consumer := kafkaconsumer.NewConsumer(consumerCfg, proj)
	defer consumer.Close()

	if err := consumer.Run(ctx); err != nil {
		log.Fatalf("worker stopped with error: %v", err)
	}
}
