package onlineconf_client

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	defaultOnlineConfURL = "http://localhost:8081"
	snapshotPath         = "/api/v1/onlineconf/snapshot"
	commonScope          = "common"
	pollInterval         = 10 * time.Second
	requestTimeout       = 3 * time.Second
)

type SnapshotConfig struct {
	Scope     string          `json:"scope"`
	Key       string          `json:"key"`
	ValueType string          `json:"valueType"`
	Value     json.RawMessage `json:"value"`
}

type SnapshotFlag struct {
	Scope          string  `json:"scope"`
	Key            string  `json:"key"`
	Enabled        bool    `json:"enabled"`
	RolloutPercent int     `json:"rolloutPercent"`
	Whitelist      []int64 `json:"whitelist"`
	Blacklist      []int64 `json:"blacklist"`
}

type SnapshotResponse struct {
	Configs []SnapshotConfig `json:"configs"`
	Flags   []SnapshotFlag   `json:"flags"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger

	mu      sync.RWMutex
	configs map[string]json.RawMessage
	flags   map[string]SnapshotFlag
}

func NewClient(baseURL string, logger *zap.Logger) *Client {
	if baseURL == "" {
		baseURL = os.Getenv("ONLINECONF_URL")
	}
	if baseURL == "" {
		baseURL = defaultOnlineConfURL
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
		logger:  logger,
		configs: make(map[string]json.RawMessage),
		flags:   make(map[string]SnapshotFlag),
	}
}

func (c *Client) Start(ctx context.Context) {
	c.refresh(ctx)

	go func() {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.refresh(ctx)
			}
		}
	}()
}

func (c *Client) refresh(ctx context.Context) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+snapshotPath, nil)
	if err != nil {
		c.logger.Warn("не удалось создать запрос snapshot onlineconf", zap.Error(err))
		return
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("не удалось получить snapshot onlineconf", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("неожиданный статус snapshot onlineconf", zap.Int("status", resp.StatusCode))
		return
	}

	var snapshot SnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		c.logger.Warn("не удалось декодировать snapshot onlineconf", zap.Error(err))
		return
	}

	configs := make(map[string]json.RawMessage, len(snapshot.Configs))
	for _, cfg := range snapshot.Configs {
		configs[compoundKey(cfg.Scope, cfg.Key)] = cfg.Value
	}

	flags := make(map[string]SnapshotFlag, len(snapshot.Flags))
	for _, flag := range snapshot.Flags {
		flags[compoundKey(flag.Scope, flag.Key)] = flag
	}

	c.mu.Lock()
	c.configs = configs
	c.flags = flags
	c.mu.Unlock()
}

func (c *Client) GetConfigJSON(scope, key string) (json.RawMessage, bool) {
	if c == nil {
		return nil, false
	}

	c.mu.RLock()
	val, ok := c.configs[compoundKey(scope, key)]
	if !ok && scope != commonScope {
		val, ok = c.configs[compoundKey(commonScope, key)]
	}
	c.mu.RUnlock()

	return val, ok
}

func (c *Client) IsFlagEnabled(scope, key string, userID int64) bool {
	if c == nil {
		return false
	}

	c.mu.RLock()
	flag, ok := c.flags[compoundKey(scope, key)]
	if !ok && scope != commonScope {
		flag, ok = c.flags[compoundKey(commonScope, key)]
	}
	c.mu.RUnlock()
	if !ok {
		return false
	}

	for _, id := range flag.Blacklist {
		if id == userID {
			return false
		}
	}
	for _, id := range flag.Whitelist {
		if id == userID {
			return true
		}
	}
	if !flag.Enabled {
		return false
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(fmt.Sprintf("%s.%s:%d", scope, key, userID)))
	bucket := int(h.Sum32() % 100)

	return bucket < flag.RolloutPercent
}

func compoundKey(scope, key string) string {
	return scope + "." + key
}
