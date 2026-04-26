package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisTicketStore struct {
	redisClient *redis.Client
}

func NewRedisTicketStore(redisClient *redis.Client) *redisTicketStore {
	return &redisTicketStore{
		redisClient: redisClient,
	}
}

func (s *redisTicketStore) GenerateTicket(ctx context.Context, userID uint) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	ticketString := hex.EncodeToString(bytes)

	redisKey := "auth_ticket:" + ticketString
	//userIDStr := string(rune(userID))
	userIDStr := strconv.FormatUint(uint64(userID), 10)

	err := s.redisClient.Set(ctx, redisKey, userIDStr, 15*time.Second).Err()
	if err != nil {
		return "", err
	}
	return ticketString, nil
}

func (s *redisTicketStore) ConsumeTicket(ctx context.Context, ticket string) (uint, error) {
	redisKey := "auth_ticket:" + ticket

	val, err := s.redisClient.GetDel(ctx, redisKey).Result()
	if err == redis.Nil || err != nil {
		return 0, err
	}

	parsedID, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0, errors.New("invalid user ID in ticket")
	}
	return uint(parsedID), nil
}
