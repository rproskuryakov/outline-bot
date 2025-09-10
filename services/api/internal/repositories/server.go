package repositories

import (
    "time"
    "context"
    "github.com/uptrace/bun"

	"github.com/rproskuryakov/outline-bot/services/api/internal/model"
)


type ServerStore interface {
    Add(name string, user model.User) error
    Get(name string) (model.ServerRecord, error)
    Update(name string, server model.ServerRecord) error
    List() (map[int]model.ServerRecord, error)
    Remove(name string) error
}

func NewServerStore(DB *bun.DB) *ServerRepository {
    return &ServerRepository{DB: DB}
}

type ServerRepository struct {
    DB *bun.DB
}

func (repo *ServerRepository) Add(name string, user model.User) (error) {
    ctx := context.Background()
    serverRecord := &model.ServerRecord{
        CreatedAt: time.Now(),
        Owner: user,
        IsActive: true,
    }
    _, insertErr := repo.DB.NewInsert().Model(serverRecord).Exec(ctx)
    return insertErr
}

func (repo *ServerRepository) List() (map[int]model.ServerRecord, error) {
    ctx := context.Background()
    var serverRecords []model.ServerRecord
    finalOutput := new(map[int]model.ServerRecord)
    err := repo.DB.NewSelect().Model(&serverRecords).Scan(ctx)
    if err != nil {
        return *finalOutput, err
    }
    for _, v := range serverRecords {
        (*finalOutput)[int(v.ID)] = v
    }
    return *finalOutput, err
}

func (repo *ServerRepository) Get(name string) (model.ServerRecord, error) {
    ctx := context.Background()
    serverRecord := new(model.ServerRecord)
    err := repo.DB.NewSelect().Model(serverRecord).Where("id = ?", name).Scan(ctx)
    if err != nil {
        return *serverRecord, err
    }
    return *serverRecord, err
}

func (repo *ServerRepository) Remove(name string) (error) {
    ctx := context.Background()
    _, err := repo.DB.NewDelete().Where("id = ?", name).Exec(ctx)
    if err != nil {
        return err
    }
    return nil
}


func (repo *ServerRepository) Update(name string, server model.ServerRecord) error {
//     ctx := context.Background()
//     _, err := repo.DB.NewDelete().Where("id = ?", name).Exec(ctx)
//     if err != nil {
//         return err
//     }
    return nil
}
