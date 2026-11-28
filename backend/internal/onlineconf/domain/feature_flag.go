package domain

import (
	"errors"
	"time"
)

var (
	ErrFeatureFlagNotFound       = errors.New("фича-флаг не найден")
	ErrFeatureFlagAlreadyExists  = errors.New("фича-флаг с таким ключом в данном scope уже существует")
	ErrFeatureFlagInvalid        = errors.New("некорректные данные фича-флага")
	ErrFeatureFlagRolloutInvalid = errors.New("процент раскатки должен быть от 0 до 100")
)

type FeatureFlag struct {
	ID             int64
	Key            string
	Scope          string
	Enabled        bool
	RolloutPercent int
	Whitelist      []int64
	Blacklist      []int64
	Description    string
	CreatedBy      int64
	UpdatedBy      int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateFeatureFlagInDTO struct {
	Key            string
	Scope          string
	Enabled        bool
	RolloutPercent int
	Whitelist      []int64
	Blacklist      []int64
	Description    string
	AdminUserID    int64
}

type UpdateFeatureFlagInDTO struct {
	ID             int64
	Enabled        bool
	RolloutPercent int
	Whitelist      []int64
	Blacklist      []int64
	Description    string
	AdminUserID    int64
}

type EvaluateFeatureFlagOutDTO struct {
	FlagID  int64
	Key     string
	Scope   string
	UserID  int64
	Enabled bool
	Reason  string
	Bucket  int
}

type FeatureFlagRepo interface {
	List(scope string, query string) ([]FeatureFlag, error)
	GetByID(id int64) (FeatureFlag, error)
	Create(flag FeatureFlag) (FeatureFlag, error)
	Update(flag FeatureFlag) (FeatureFlag, error)
	Delete(id int64) (FeatureFlag, error)
}

type FeatureFlagUsecase interface {
	List(scope string, query string) ([]FeatureFlag, error)
	Create(dto CreateFeatureFlagInDTO) (FeatureFlag, error)
	Update(dto UpdateFeatureFlagInDTO) (FeatureFlag, error)
	Delete(id int64, adminUserID int64) error
	Evaluate(id int64, userID int64) (EvaluateFeatureFlagOutDTO, error)
}
