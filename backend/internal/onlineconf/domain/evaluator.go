package domain

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"slices"
)

const (
	ReasonBlacklist = "blacklist"
	ReasonWhitelist = "whitelist"
	ReasonDisabled  = "disabled"
	ReasonRollout   = "rollout"
)

// ComputeRolloutBucket вычисляет детерминированный бакет 0..99 для пользователя.
func ComputeRolloutBucket(scope string, key string, userID int64) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(fmt.Sprintf("%s.%s:%d", scope, key, userID)))
	return int(h.Sum32() % 100)
}

// EvaluateFeatureFlag вычисляет итоговое состояние фича-флага для пользователя и причину.
func EvaluateFeatureFlag(flag FeatureFlag, userID int64) EvaluateFeatureFlagOutDTO {
	bucket := ComputeRolloutBucket(flag.Scope, flag.Key, userID)

	res := EvaluateFeatureFlagOutDTO{
		FlagID: flag.ID,
		Key:    flag.Key,
		Scope:  flag.Scope,
		UserID: userID,
		Bucket: bucket,
	}

	if slices.Contains(flag.Blacklist, userID) {
		res.Enabled = false
		res.Reason = ReasonBlacklist
		return res
	}

	if slices.Contains(flag.Whitelist, userID) {
		res.Enabled = true
		res.Reason = ReasonWhitelist
		return res
	}

	if !flag.Enabled {
		res.Enabled = false
		res.Reason = ReasonDisabled
		return res
	}

	res.Enabled = bucket < flag.RolloutPercent
	res.Reason = ReasonRollout
	return res
}

type RuntimeFrontendOutDTO struct {
	Configs map[string]json.RawMessage
	Flags   map[string]bool
}

type SnapshotConfigOutDTO struct {
	Scope     string
	Key       string
	ValueType string
	Value     json.RawMessage
}

type SnapshotFlagOutDTO struct {
	Scope          string
	Key            string
	Enabled        bool
	RolloutPercent int
	Whitelist      []int64
	Blacklist      []int64
}

type SnapshotOutDTO struct {
	Configs []SnapshotConfigOutDTO
	Flags   []SnapshotFlagOutDTO
}

type RuntimeUsecase interface {
	FrontendRuntime(userID int64) RuntimeFrontendOutDTO
	Snapshot() SnapshotOutDTO
	Invalidate()
}
