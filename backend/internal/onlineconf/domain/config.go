package domain

import (
	"encoding/json"
	"errors"
	"time"
)

const (
	ValueTypeString  = "string"
	ValueTypeNumber  = "number"
	ValueTypeBoolean = "boolean"
	ValueTypeArray   = "array"
	ValueTypeJSON    = "json"
)

var (
	ErrConfigNotFound      = errors.New("конфиг не найден")
	ErrConfigAlreadyExists = errors.New("конфиг с таким ключом в данном scope уже существует")
	ErrConfigInvalid       = errors.New("некорректные данные конфига")
	ErrConfigValueType     = errors.New("значение конфига не соответствует указанному valueType")
)

type Config struct {
	ID          int64
	Key         string
	Scope       string
	ValueType   string
	Value       json.RawMessage
	Description string
	CreatedBy   int64
	UpdatedBy   int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateConfigInDTO struct {
	Key         string
	Scope       string
	ValueType   string
	Value       json.RawMessage
	Description string
	AdminUserID int64
}

type UpdateConfigInDTO struct {
	ID          int64
	ValueType   string
	Value       json.RawMessage
	Description string
	AdminUserID int64
}

type ConfigRepo interface {
	List(scope string, query string) ([]Config, error)
	GetByID(id int64) (Config, error)
	Create(cfg Config) (Config, error)
	Update(cfg Config) (Config, error)
	Delete(id int64) (Config, error)
}

type ConfigUsecase interface {
	List(scope string, query string) ([]Config, error)
	Create(dto CreateConfigInDTO) (Config, error)
	Update(dto UpdateConfigInDTO) (Config, error)
	Delete(id int64, adminUserID int64) error
}
