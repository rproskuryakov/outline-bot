package repositories

import (
    "context"
    "strconv"
	"crypto/md5"
	"encoding/hex"
    "log"
	"github.com/uptrace/bun"

	"github.com/rproskuryakov/outline-bot/services/api/internal/model"
)


type UserStore interface {
    Add(name string, server model.ServerRecord) error
    Get(name string) (model.ServerRecord, error)
    Update(name string, server model.ServerRecord) error
    List() (map[string]model.ServerRecord, error)
    Remove(name string) error
}

type UserRepository struct {
    Db *bun.DB
}

func (repo *UserRepository) GetUserAttributes(ctx context.Context, username int64) (u *model.User, e error) {
    hasher := md5.New()
    hasher.Write([]byte(strconv.FormatInt(username, 10)))
    usernameHashed := hex.EncodeToString(hasher.Sum(nil))

    user := new(model.User)
    err := repo.Db.NewSelect().Model(user).Where("id = ?", usernameHashed).Scan(ctx)
    if err != nil {
        return user, err
    }
    return user, nil
}


func (repo *UserRepository) CheckIfUserExists(ctx context.Context, username int64) (bool, error) {
    hasher := md5.New()
    hasher.Write([]byte(strconv.FormatInt(username, 10)))
    usernameHashed := hex.EncodeToString(hasher.Sum(nil))

    user := new(model.User)
    err := repo.Db.NewSelect().Model(user).Where("id = ?", usernameHashed).Scan(ctx)
    if err != nil {
        log.Printf(err.Error())
        return false, nil
    }
    exists := false
    if *user != (model.User{}) {
        exists = true
    }
    return exists, nil
}

func (repo *UserRepository) InsertUser(ctx context.Context, username int64) (error) {
    hasher := md5.New()
    hasher.Write([]byte(strconv.FormatInt(username, 10)))
    usernameHashed := hex.EncodeToString(hasher.Sum(nil))

    user := &model.User{Name: usernameHashed, IsAdmin: false}
    _, err := repo.Db.NewInsert().Model(user).Exec(ctx)
    return err
}