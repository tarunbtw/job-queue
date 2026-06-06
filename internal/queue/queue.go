package queue

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

const JobQueueKey = "jobs:queue"

type Queue struct {
	client *redis.Client
	ctx    context.Context
}

func New() *Queue {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL not set")
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("invalid REDIS_URL:", err)
	}

	client := redis.NewClient(opts)

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal("failed to connect to redis:", err)
	}

	log.Println("redis connected")
	return &Queue{client: client, ctx: ctx}
}

// Push adds a job ID to the queue
func (q *Queue) Push(jobID string) error {
	return q.client.LPush(q.ctx, JobQueueKey, jobID).Err()
}

// Pop blocks until a job ID is available, then returns it
func (q *Queue) Pop() (string, error) {
	result, err := q.client.BRPop(q.ctx, 0, JobQueueKey).Result()
	if err != nil {
		return "", err
	}
	// BRPop returns [key, value]
	return result[1], nil
}