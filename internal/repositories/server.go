package repositories

import (
    "context"
    "time"

    "github.com/uptrace/bun"

	"github.com/rproskuryakov/outline-bot/internal/model"
)


func InsertServerRecord(ctx context.Context, user *model.User, Db *bun.DB) (error) {
    serverRecord := &model.ServerRecord{
        CreatedAt: time.Now(),
        Owner: *user,
        IsActive: true,
    }
    _, insertErr := Db.NewInsert().Model(serverRecord).Exec(ctx)
    return insertErr
}