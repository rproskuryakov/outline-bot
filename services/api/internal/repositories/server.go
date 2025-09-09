package repositories

import (
    "context"
    "time"

    "github.com/uptrace/bun"

	"github.com/rproskuryakov/outline-bot/services/api/internal/model"
)


type ServerRepository struct {
    Db *bun.DB
}


func (repo *ServerRepository) InsertServerRecord(ctx context.Context, user *model.User) (error) {
    serverRecord := &model.ServerRecord{
        CreatedAt: time.Now(),
        Owner: *user,
        IsActive: true,
    }
    _, insertErr := repo.Db.NewInsert().Model(serverRecord).Exec(ctx)
    return insertErr
}