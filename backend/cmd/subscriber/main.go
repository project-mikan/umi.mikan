package main

import (
	"context"
	"fmt"
	"log"

	"github.com/project-mikan/umi.mikan/backend/constants"
	"github.com/redis/rueidis"
)

func main() {
	log.Print("=== umi.mikan subscriber started ===")

	// Redis設定の読み込み
	redisConfig, err := constants.LoadRedisConfig()
	if err != nil {
		log.Fatalf("Failed to load Redis config: %v", err)
	}

	// Redisクライアント作成
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port)},
	})
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Redis接続確認
	pingCmd := client.B().Ping().Build()
	if err := client.Do(ctx, pingCmd).Error(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Print("Connected to Redis successfully")

	log.Print("Subscriber is listening for messages...")

	// SUBSCRIBE コマンドでチャンネル購読
	err = client.Receive(ctx, client.B().Subscribe().Channel("diary_events").Build(), func(msg rueidis.PubSubMessage) {
		log.Printf("Received message: %s from channel: %s", msg.Message, msg.Channel)

		// TODO: ここでLLMの要約処理などを実装
		err := processMessage(msg.Message)
		if err != nil {
			log.Printf("Failed to process message: %v", err)
		}
	})

	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	log.Print("Subscriber ended")
}

func processMessage(payload string) error {
	// TODO: LLMの要約生成処理を実装
	log.Printf("Processing message: %s", payload)
	return nil
}