package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"vibe_market/backend/internal/auth/domain"

	"github.com/redis/go-redis/v9"
)

type refreshToken struct {
	UserID     int64     `json:"user_id"`
	TokenHash  string    `json:"token_hash"`
	DeviceName string    `json:"device_name"`
	IsRevoked  bool      `json:"is_revoked"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type RedisRefreshTokenRepo struct {
	client *redis.Client
}

func NewRefreshTokenRepo(client *redis.Client) *RedisRefreshTokenRepo {
	return &RedisRefreshTokenRepo{client: client}
}

func tokenKey(hash, deviceName string) string {
	return fmt.Sprintf("refresh_token:%s:%s", hash, deviceName)
}

func userTokensSetKey(userID int64) string {
	return fmt.Sprintf("user_tokens:%d", userID)
}

func deviceTokensSetKey(userID int64, deviceName string) string {
	return fmt.Sprintf("device_tokens:%d:%s", userID, deviceName)
}

func tokenID(hash, deviceName string) string {
	return hash + "\x00" + deviceName
}

// Insert сохраняет refresh-токен в Redis.
func (r *RedisRefreshTokenRepo) Insert(t domain.RefreshToken) error {
	ctx := context.Background()

	tKey := tokenKey(t.TokenHash, t.DeviceName)
	uSetKey := userTokensSetKey(t.UserID)
	dSetKey := deviceTokensSetKey(t.UserID, t.DeviceName)

	data, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("marshal refresh token: %w", err)
	}

	pipe := r.client.TxPipeline()

	pipe.HSet(ctx, tKey, "data", data)
	pipe.SAdd(ctx, uSetKey, tokenID(t.TokenHash, t.DeviceName))
	pipe.SAdd(ctx, dSetKey, t.TokenHash)

	if ttl := time.Until(t.ExpiresAt); ttl > 0 {
		pipe.Expire(ctx, tKey, ttl)
		pipe.Expire(ctx, uSetKey, ttl)
		pipe.Expire(ctx, dSetKey, ttl)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("exec redis pipeline: %w", err)
	}
	return nil
}

// SelectByHashAndDeviceName возвращает токен, если он существует, не отозван и не истёк.
func (r *RedisRefreshTokenRepo) SelectByHashAndDeviceName(hash string, deviceName string) (domain.RefreshToken, error) {
	if hash == "" || deviceName == "" {
		return domain.RefreshToken{}, domain.ErrRefreshTokenNotFound
	}

	ctx := context.Background()
	tKey := tokenKey(hash, deviceName)

	data, err := r.client.HGet(ctx, tKey, "data").Bytes()
	if err != nil {
		return domain.RefreshToken{}, err
	}

	var token refreshToken
	if err := json.Unmarshal(data, &token); err != nil {
		return domain.RefreshToken{}, err
	}

	if token.IsRevoked {
		return domain.RefreshToken{}, domain.ErrRefreshTokenRevoked
	}
	if time.Now().After(token.ExpiresAt) {
		return domain.RefreshToken{}, domain.ErrRefreshTokenExpired
	}

	tokenOut := domain.RefreshToken{
		UserID:     token.UserID,
		TokenHash:  token.TokenHash,
		DeviceName: token.DeviceName,
		IsRevoked:  token.IsRevoked,
		CreatedAt:  token.CreatedAt,
		ExpiresAt:  token.ExpiresAt,
	}

	return tokenOut, nil
}

// Revoke помечает все refresh-токены пользователя для указанного устройства как отозванные.
// Возвращает true, если хотя бы один токен был обновлён.
func (r *RedisRefreshTokenRepo) Revoke(userID int64, deviceName string) error {
	ctx := context.Background()
	dSetKey := deviceTokensSetKey(userID, deviceName)

	hashes, err := r.client.SMembers(ctx, dSetKey).Result()
	if err != nil {
		return err
	}
	if len(hashes) == 0 {
		return domain.ErrRefreshTokenNotFound
	}

	// Читаем все токены устройства одним pipeline.
	readPipe := r.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(hashes))
	for i, h := range hashes {
		cmds[i] = readPipe.HGet(ctx, tokenKey(h, deviceName), "data")
	}
	if _, err := readPipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	// Обновляем IsRevoked = true.
	updatePipe := r.client.Pipeline()
	updated := 0
	for _, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			continue
		}
		var token refreshToken
		if err := json.Unmarshal(data, &token); err != nil {
			continue
		}
		if token.IsRevoked {
			continue
		}
		token.IsRevoked = true
		newData, err := json.Marshal(token)
		if err != nil {
			continue
		}
		updatePipe.HSet(ctx, tokenKey(token.TokenHash, token.DeviceName), "data", newData)
		updated++
	}

	if updated == 0 {
		return domain.ErrRefreshTokenNotFound
	}

	if _, err := updatePipe.Exec(ctx); err != nil {
		return err
	}
	return nil
}
