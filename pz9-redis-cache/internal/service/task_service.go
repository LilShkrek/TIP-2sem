package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"example.com/pz9-redis-cache/internal/cache"
	"example.com/pz9-redis-cache/internal/config"
	"example.com/pz9-redis-cache/internal/task"
	"github.com/redis/go-redis/v9"
)

type TaskService struct {
	repo  *task.Repo
	redis *redis.Client
	cfg   config.Config
}

func NewTaskService(repo *task.Repo, redisClient *redis.Client, cfg config.Config) *TaskService {
	return &TaskService{
		repo:  repo,
		redis: redisClient,
		cfg:   cfg,
	}
}

func (s *TaskService) GetTaskByID(ctx context.Context, id int64) (task.Task, error) {
	key := cache.TaskByIDKey(id)

	if s.redis != nil {
		cached, err := s.redis.Get(ctx, key).Result()
		switch {
		case err == nil:
			var t task.Task
			if err := json.Unmarshal([]byte(cached), &t); err == nil {
				log.Println("cache hit:", key)
				return t, nil
			}
			log.Println("redis cached value decode error:", key, err)
		case errors.Is(err, redis.Nil):
			log.Println("cache miss:", key)
		default:
			log.Println("redis read error:", key, err)
		}
	}

	t, err := s.repo.GetByID(id)
	if err != nil {
		return task.Task{}, err
	}

	s.setCache(ctx, key, t)
	return t, nil
}

func (s *TaskService) PatchTask(ctx context.Context, id int64, patch task.Patch) (task.Task, error) {
	t, err := s.repo.Patch(id, patch)
	if err != nil {
		return task.Task{}, err
	}

	s.invalidateCache(ctx, id)
	return t, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id int64) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	s.invalidateCache(ctx, id)
	return nil
}

func (s *TaskService) setCache(ctx context.Context, key string, t task.Task) {
	if s.redis == nil {
		return
	}

	payload, err := json.Marshal(t)
	if err != nil {
		log.Println("cache encode error:", key, err)
		return
	}

	ttl := cache.TTLWithJitter(s.cfg.CacheTTL, s.cfg.CacheTTLJitter)
	if err := s.redis.Set(ctx, key, payload, ttl).Err(); err != nil {
		log.Println("redis write error:", key, err)
		return
	}

	log.Println("cache set:", key, "ttl:", ttl)
}

func (s *TaskService) invalidateCache(ctx context.Context, id int64) {
	if s.redis == nil {
		return
	}

	key := cache.TaskByIDKey(id)
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		log.Println("redis delete error:", key, err)
		return
	}

	log.Println("cache invalidated:", key)
}
