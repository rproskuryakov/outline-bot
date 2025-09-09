package repositories

import (
//     "context"
//     "time"
//
//     "github.com/uptrace/bun"

	"github.com/rproskuryakov/outline-bot/services/api/internal/model"
)

type APIKeyStore interface {
    Add(name string, server model.AccessKey) error
    Get(name string) (model.AccessKey, error)
    Update(name string, server model.AccessKey) error
    List() (map[string]model.AccessKey, error)
    Remove(name string) error
}