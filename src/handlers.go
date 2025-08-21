package src

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"log"
	"strconv"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/uptrace/bun"
	"github.com/redis/go-redis/v9"
)


type Server struct {
    Db *bun.DB
    RedisDb *redis.Client
    StateMachine *StateMachine
    OutlineClients []*OutlineVPN
}

func (server *Server) DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    usernameTelegramID := update.Message.From.ID
    hasher := md5.New()
    hasher.Write([]byte(strconv.FormatInt(usernameTelegramID, 10)))
    usernameHashed := hex.EncodeToString(hasher.Sum(nil))

    err := server.RedisDb.Set(ctx, usernameHashed, "/default", 200).Err()
    if err != nil {
        panic(err)
    }
    val, err := server.RedisDb.Get(ctx, usernameHashed).Result()
    log.Printf(val)
    if err != nil {
        panic(err)
    }
    b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "I am YetAnotherVPN Bot. \n"+
		             "I can sign up a new user, " +
		             "change limits on monthly traffic " +
		             "or regenerate an api key in case your's has stopped working. \n" + "\n" +
		             "I am built on open source outline vpn technology. \n" +
		             "/start \n",
    })
}

func checkIfUserExists(ctx context.Context, username int64, Db *bun.DB) (f bool, err error) {
    hasher := md5.New()
    hasher.Write([]byte(strconv.FormatInt(username, 10)))
    usernameHashed := hex.EncodeToString(hasher.Sum(nil))

    user := new(User)
    exists, err := Db.NewSelect().Model(user).Where("username = ?", usernameHashed).Exists(ctx)
    if err != nil {
        return false, err
    }
    return exists, nil
}

func getUserAttributes(ctx context.Context, username int64, Db *bun.DB) (u *User, e error) {
    hasher := md5.New()
    hasher.Write([]byte(strconv.FormatInt(username, 10)))
    usernameHashed := hex.EncodeToString(hasher.Sum(nil))

    user := new(User)
    err := Db.NewSelect().Model(user).Where("id = ?", usernameHashed).Scan(ctx)
    if err != nil {
        return user, err
    }
    return user, nil
}


func (server *Server) StartHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    usernameTelegramID := update.Message.From.ID
    // check if user exists
    exists, err := checkIfUserExists(ctx, usernameTelegramID, server.Db)
    if err != nil {
        log.Printf(err.Error())
        panic(err)
    }

    user, err := getUserAttributes(ctx, usernameTelegramID, server.Db)
    if err != nil {
        log.Printf(err.Error())
        panic(err)
    }
    if exists && user.IsAdmin {
        b.SendMessage(ctx, &bot.SendMessageParams{
		    ChatID:      update.Message.Chat.ID,
		    Text:        "User " + strconv.FormatInt(usernameTelegramID, 10) + " is found. \n" +
		                 "You can do one of the following: \n" +
		                 "/issueApiKey \n" +
		                 "/reissueApiKey \n" +
		                 "/viewTrafficUsed",
        })
        return
    } else if exists && user.IsAdmin {
        b.SendMessage(ctx, &bot.SendMessageParams{
		    ChatID:      update.Message.Chat.ID,
		    Text:        "Admin " + strconv.FormatInt(usernameTelegramID, 10) + " is found. \n" +
		                 "You can do one of the following: \n" +
		                 "/createServer \n" +
		                 "/changeLimits \n" +
		                 "/viewOverallTrafficUsed",
        })
        return
    } else if !exists {
        b.SendMessage(ctx, &bot.SendMessageParams{
		    ChatID:      update.Message.Chat.ID,
		    Text:        "User " + strconv.FormatInt(usernameTelegramID, 10) + " is not found. \n" +
		                 "Please, sign up: \n" +
		                 "/signUp \n",
        })
        return
    }

}

func (server *Server) SignUpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    usernameTelegramID := update.Message.From.ID
    exists, err := checkIfUserExists(ctx, usernameTelegramID, server.Db)

    if err != nil {
        log.Printf(err.Error())
        panic(err)
    }
    if exists {
        b.SendMessage(ctx, &bot.SendMessageParams{
		    ChatID:      update.Message.Chat.ID,
		    Text:        "User " + strconv.FormatInt(usernameTelegramID, 10) + " already exists. \n" +
		                 "You can do one of the following: \n" +
		                 "/issueApiKey \n" +
		                 "/reissueApiKey \n" +
		                 "/viewTrafficUsed",
        })
        return
    } else {
        hasher := md5.New()
        hasher.Write([]byte(strconv.FormatInt(usernameTelegramID, 10)))
        usernameHashed := hex.EncodeToString(hasher.Sum(nil))

        user := &User{Name: usernameHashed, IsAdmin: false}
        _, err := server.Db.NewInsert().Model(user).Exec(ctx)

        if err != nil {
            log.Printf(err.Error())
            panic(err)
        }
    }

}


func (server *Server) IssueApiKeyHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "issueApiKey",
    })
}

func (server *Server) ReissueApiKeyHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "reissueApiKey",
    })
}

func (server *Server) ViewTrafficUsedHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "viewTrafficUsed",
    })
}

func (server *Server) ChangeLimitsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "changeLimits",
    })
}

func (server *Server) AddAdminHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    usernameTelegramID := update.Message.From.ID
    // check if user exists
    exists, err := checkIfUserExists(ctx, usernameTelegramID, server.Db)
    if err != nil {
        log.Printf(err.Error())
        panic(err)
    }

    user, err := getUserAttributes(ctx, usernameTelegramID, server.Db)
    if err != nil {
        log.Printf(err.Error())
        panic(err)
    }
    if !exists {
        b.SendMessage(ctx, &bot.SendMessageParams{
		    ChatID:      update.Message.Chat.ID,
		    Text:        "Your user does not exist.",
        })
        return
    }
    if !user.IsAdmin {
        b.SendMessage(ctx, &bot.SendMessageParams{
		    ChatID:      update.Message.Chat.ID,
		    Text:        "You are not authorized to add new admins.",
        })
        return
    }
    // update user admin rights
//     _, err = db.NewUpdate().
//     Model((*User)(nil)).
//     Set("last_login = ?", time.Now()).
//     Where("status = ?", "active").
//     Exec(ctx)
//     b.SendMessage(ctx, &bot.SendMessageParams{
// 		ChatID:      update.Message.Chat.ID,
// 		Text:        "addAdmin",
//     })
}

func (server *Server) CreateServerHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    usernameTelegramID := update.Message.From.ID
    // check if user exists
    exists, existsError := checkIfUserExists(ctx, usernameTelegramID, server.Db)
    if existsError != nil {
        log.Printf(existsError.Error())
        panic(existsError)
    }
    if !exists {
        b.SendMessage(ctx, &bot.SendMessageParams{
		    ChatID:      update.Message.Chat.ID,
		    Text:        "User does not exist.",
        })
        log.Printf("User does not exist.")
        return
    }
    // check if user is admin
    user, getAttrsError := getUserAttributes(ctx, usernameTelegramID, server.Db)
    if getAttrsError != nil {
        log.Printf(getAttrsError.Error())
        panic(getAttrsError)
    }
    if !user.IsAdmin {
        b.SendMessage(ctx, &bot.SendMessageParams{
		    ChatID:      update.Message.Chat.ID,
		    Text:        "You are not authorized to create a server.",
        })
        return
    }

    serverRecord := &ServerRecord{
        CreatedAt: time.Now(),
        Owner: *user,
        IsActive: true,
    }
    _, insertErr := server.Db.NewInsert().Model(serverRecord).Exec(ctx)

    if insertErr != nil {
        log.Printf(insertErr.Error())
        panic(insertErr)
    }
    log.Printf("Server record is created.")
}

