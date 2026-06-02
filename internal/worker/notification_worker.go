package worker

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	notificationQueueKey = "queue:notifications"
	maxRetries           = 3
)

type notificationJob struct {
	TaskID     int64 `json:"task_id"`
	UserID     int64 `json:"user_id"`
	RetryCount int   `json:"retry_count"`
}

func EnqueueNotificationJob(ctx context.Context, redisClient *redis.Client, taskID int64, userID int64) {
	if redisClient == nil {
		return
	}

	job := notificationJob{
		TaskID:     taskID,
		UserID:     userID,
		RetryCount: 0,
	}
	payload, err := json.Marshal(job)
	if err != nil {
		log.Printf("notification job marshal failed task_id=%d user_id=%d err=%v", taskID, userID, err)
		return
	}

	if err := redisClient.LPush(ctx, notificationQueueKey, payload).Err(); err != nil {
		log.Printf("notification job publish failed task_id=%d user_id=%d err=%v", taskID, userID, err)
		return
	}
	log.Printf("notification job published payload=%s", payload)
}

func StartNotificationWorker(redisClient *redis.Client) {
	if redisClient == nil {
		log.Println("notification worker disabled: redis client is nil")
		return
	}

	ctx := context.Background()
	log.Println("notification worker started")

	for {
		result, err := redisClient.BRPop(ctx, 0, notificationQueueKey).Result()
		if err != nil {
			log.Printf("notification worker BRPOP error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		if len(result) < 2 {
			log.Printf("notification worker received invalid job result=%v", result)
			continue
		}

		payload := result[1]
		var job notificationJob
		if err := json.Unmarshal([]byte(payload), &job); err != nil {
			log.Printf("notification worker invalid job payload=%s err=%v", payload, err)
			continue
		}

		if rand.Float32() < 0.5 {
			if job.RetryCount < maxRetries {
				job.RetryCount++
				retryPayload, marshalErr := json.Marshal(job)
				if marshalErr != nil {
					log.Printf("[Worker] ERROR: Không thể marshal retry job Task %d cho User %d: %v", job.TaskID, job.UserID, marshalErr)
					continue
				}

				if pushErr := redisClient.LPush(ctx, notificationQueueKey, retryPayload).Err(); pushErr != nil {
					log.Printf("[Worker] ERROR: Không thể push lại retry job Task %d cho User %d: %v", job.TaskID, job.UserID, pushErr)
					continue
				}

				log.Printf("[Worker] Job Task %d cho User %d thất bại lần %d. Đang tiến hành retry...", job.TaskID, job.UserID, job.RetryCount)
				continue
			}

			log.Printf("[Worker] ERROR: Job Task %d cho User %d đã đạt giới hạn retry tối đa (%d lần). Hủy bỏ job vĩnh viễn!", job.TaskID, job.UserID, maxRetries)
			continue
		}

		log.Printf("notification worker sending notification task_id=%d user_id=%d", job.TaskID, job.UserID)
		time.Sleep(2 * time.Second)
		log.Printf("notification worker sent notification task_id=%d user_id=%d", job.TaskID, job.UserID)
	}
}
