package repositories

import (
    "context"
    "strconv"
	"crypto/md5"
	"encoding/hex"
    "log"
	"github.com/uptrace/bun"

	"github.com/rproskuryakov/outline-bot/internal/model"
)

func GetUserAttributes(ctx context.Context, username int64, Db *bun.DB) (u *model.User, e error) {
    hasher := md5.New()
    hasher.Write([]byte(strconv.FormatInt(username, 10)))
    usernameHashed := hex.EncodeToString(hasher.Sum(nil))

    user := new(model.User)
    err := Db.NewSelect().Model(user).Where("id = ?", usernameHashed).Scan(ctx)
    if err != nil {
        return user, err
    }
    return user, nil
}


func CheckIfUserExists(ctx context.Context, username int64, Db *bun.DB) (bool, error) {
    hasher := md5.New()
    hasher.Write([]byte(strconv.FormatInt(username, 10)))
    usernameHashed := hex.EncodeToString(hasher.Sum(nil))

    user := new(model.User)
    err := Db.NewSelect().Model(user).Where("id = ?", usernameHashed).Scan(ctx)
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

func InsertUser(ctx context.Context, username int64, Db *bun.DB) (error) {
    hasher := md5.New()
    hasher.Write([]byte(strconv.FormatInt(username, 10)))
    usernameHashed := hex.EncodeToString(hasher.Sum(nil))

    user := &model.User{Name: usernameHashed, IsAdmin: false}
    _, err := Db.NewInsert().Model(user).Exec(ctx)
    return err
}