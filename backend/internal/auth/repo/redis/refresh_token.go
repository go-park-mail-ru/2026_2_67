package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RefreshToken struct {
	UserID     int64     `json:"user_id"`
	TokenHash  string    `json:"token_hash"`
	DeviceName string    `json:"device_name"`
	IsRevoked  bool      `json:"is_revoked"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// RedisRefreshTokenRepo — реализация RefreshTokenRepo на Redis.
type RedisRefreshTokenRepo struct {
	client *redis.Client
}

// NewRedisRefreshTokenRepo создаёт новый репозиторий.
func NewRedisRefreshTokenRepo(client *redis.Client) *RedisRefreshTokenRepo {
	return &RedisRefreshTokenRepo{client: client}
}

// ---------- Ключи Redis ----------
//
// refresh_token:{hash}:{deviceName}      -> Hash { data: <json> }   (сам токен)
// user_tokens:{userID}                   -> Set  { "hash:deviceName", ... }  (все токены пользователя)
// device_tokens:{userID}:{deviceName}    -> Set  { "hash", ... }    (токены пользователя для устройства)

func tokenKey(hash, deviceName string) string {
	return fmt.Sprintf("refresh_token:%s:%s", hash, deviceName)
}

func userTokensSetKey(userID int64) string {
	return fmt.Sprintf("user_tokens:%d", userID)
}

func deviceTokensSetKey(userID int64, deviceName string) string {
	return fmt.Sprintf("device_tokens:%d:%s", userID, deviceName)
}

// tokenID формирует идентификатор токена для Set пользователя.
// Используем \x00 как разделитель, чтобы deviceName мог содержать ':'.
func tokenID(hash, deviceName string) string {
	return hash + "\x00" + deviceName
}

// splitTokenID разбирает идентификатор обратно на hash и deviceName.
func splitTokenID(id string) (hash, deviceName string, ok bool) {
	i := strings.IndexByte(id, 0)
	if i <= 0 || i == len(id)-1 {
		return "", "", false
	}
	return id[:i], id[i+1:], true
}

// ---------- Insert ----------

// Insert сохраняет refresh-токен в Redis.
func (r *RedisRefreshTokenRepo) Insert(t RefreshToken) error {
	if t.TokenHash == "" {
		return errors.New("token hash is empty")
	}
	if t.DeviceName == "" {
		return errors.New("device name is empty")
	}
	if t.UserID <= 0 {
		return errors.New("invalid user id")
	}

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

// ---------- SelectByHashAndDeviceName ----------

// SelectByHashAndDeviceName возвращает токен, если он существует, не отозван и не истёк.
func (r *RedisRefreshTokenRepo) SelectByHashAndDeviceName(hash string, deviceName string) (RefreshToken, bool) {
	if hash == "" || deviceName == "" {
		return RefreshToken{}, false
	}

	ctx := context.Background()
	tKey := tokenKey(hash, deviceName)

	data, err := r.client.HGet(ctx, tKey, "data").Bytes()
	if err != nil {
		// redis.Nil — ключ не найден; прочие ошибки тоже возвращаем как false
		return RefreshToken{}, false
	}

	var token RefreshToken
	if err := json.Unmarshal(data, &token); err != nil {
		return RefreshToken{}, false
	}

	if token.IsRevoked {
		return RefreshToken{}, false
	}
	if time.Now().After(token.ExpiresAt) {
		return RefreshToken{}, false
	}
	// Защита от рассинхронизации данных.
	if token.TokenHash != hash || token.DeviceName != deviceName {
		return RefreshToken{}, false
	}

	return token, true
}

// ---------- Revoke ----------

// Revoke помечает все refresh-токены пользователя для указанного устройства как отозванные.
// Возвращает true, если хотя бы один токен был обновлён.
func (r *RedisRefreshTokenRepo) Revoke(userID int64, deviceName string) bool {
	if userID <= 0 || deviceName == "" {
		return false
	}

	ctx := context.Background()
	dSetKey := deviceTokensSetKey(userID, deviceName)

	hashes, err := r.client.SMembers(ctx, dSetKey).Result()
	if err != nil || len(hashes) == 0 {
		return false
	}

	// Читаем все токены устройства одним pipeline.
	readPipe := r.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(hashes))
	for i, h := range hashes {
		cmds[i] = readPipe.HGet(ctx, tokenKey(h, deviceName), "data")
	}
	if _, err := readPipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return false
	}

	// Обновляем IsRevoked = true.
	updatePipe := r.client.Pipeline()
	updated := 0
	for _, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			continue // токен мог истечь по TTL — пропускаем
		}
		var token RefreshToken
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
		return false
	}

	if _, err := updatePipe.Exec(ctx); err != nil {
		return false
	}
	return true
}
