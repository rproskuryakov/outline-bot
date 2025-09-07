package handlers

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"log"
	"strconv"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

// 	"github.com/uptrace/bun"
    "github.com/redis/go-redis/v9"

	"github.com/rproskuryakov/outline-bot/internal/fsm"
	"github.com/rproskuryakov/outline-bot/internal/clients"
	"github.com/rproskuryakov/outline-bot/internal/model"
	"github.com/rproskuryakov/outline-bot/internal/repositories"
)

type Server struct {
    UserRepository *repositories.UserRepository
    RedisClient *redis.Client
    OutlineClients *clients.OutlineVPNClients
}

//
func (server *Server) DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    usernameTelegramID := strconv.FormatInt(update.Message.From.ID, 10)
    hasher := md5.New()
    hasher.Write([]byte(usernameTelegramID))
    usernameHashed := hex.EncodeToString(hasher.Sum(nil))

    userInput := update.Message.Text

    redisKey := fmt.Sprintf("user:%d:state", usernameHashed)

    args := &fsm.StateArgs{
        Input:     userInput,
        RedisKey:  redisKey,
        StateName: "start", // used only on first time
    }

    msg, done, err := fsm.Run(ctx, args, fsm.Registry)
    if !done || err != nil {
        b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: "error"})
        return
    }

    b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: msg})
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



func CheckAuthorized(server *Server, fn func(ctx context.Context, b *bot.Bot, update *models.Update)) func(ctx context.Context, b *bot.Bot, update *models.Update) {
    return func(ctx context.Context, b *bot.Bot, update *models.Update) {
        usernameTelegramID := update.Message.From.ID
        exists, err := server.UserRepository.CheckIfUserExists(ctx, usernameTelegramID)
        if err != nil {
            panic(err)
        }
        if exists {
            fn(ctx, b, update)
        } else {
            b.SendMessage(ctx, &bot.SendMessageParams{
		        ChatID:      update.Message.Chat.ID,
		        Text:        "User " + strconv.FormatInt(usernameTelegramID, 10) + " is not found. \n" +
		                     "Please, sign up: \n" +
    		                 "/signUp \n",
            })
        }
    }
}

func CheckAuthorizedAdmin(server *Server, fn func(ctx context.Context, b *bot.Bot, update *models.Update)) func(ctx context.Context, b *bot.Bot, update *models.Update) {
    return func(ctx context.Context, b *bot.Bot, update *models.Update) {
        username := update.Message.From.ID

        user, err := server.UserRepository.GetUserAttributes(ctx, username)
        if err != nil {
            log.Printf(err.Error())
            panic(err)
        }
        exists := false
        if *user == (model.User{}) {
            exists = false
            return
        }
        if !exists || !user.IsAdmin{
            b.SendMessage(ctx, &bot.SendMessageParams{
		        ChatID:      update.Message.Chat.ID,
		        Text:        "User " + strconv.FormatInt(username, 10) + " is not authorized as an admin. \n" +
		                     "Please, contact an administrator.",
            })
            return
        }
        fn(ctx, b, update)
    }
}

func (server *Server) StartHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    usernameTelegramID := update.Message.From.ID
    // check if user exists
    user, err := server.UserRepository.GetUserAttributes(ctx, usernameTelegramID)
    if err != nil {
        log.Printf(err.Error())
        panic(err)
    }
    exists := false
    if *user != (model.User{}) {
        exists = true
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
    }
    if exists && user.IsAdmin {
        b.SendMessage(ctx, &bot.SendMessageParams{
		    ChatID:      update.Message.Chat.ID,
		    Text:        "Admin " + strconv.FormatInt(usernameTelegramID, 10) + " is found. \n" +
		                 "You can do one of the following: \n" +
		                 "/createServer \n" +
		                 "/changeLimits \n" +
		                 "/viewOverallTrafficUsed",
        })
        return
    }
    if !exists {
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
    // check if user exists
    // check if user exists
    user, err := server.UserRepository.GetUserAttributes(ctx, usernameTelegramID)
    if err != nil {
        log.Printf(err.Error())
        panic(err)
    }
    exists := false
    if *user != (model.User{}) {
        exists = true
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
        server.UserRepository.InsertUser(ctx, usernameTelegramID)
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
    user, err := server.UserRepository.GetUserAttributes(ctx, usernameTelegramID)
    if err != nil {
        log.Printf(err.Error())
        panic(err)
    }
    exists := false
    if *user != (model.User{}) {
        exists = true
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
        usernameTelegramID := strconv.FormatInt(update.Message.From.ID, 10)
    hasher := md5.New()
    hasher.Write([]byte(usernameTelegramID))
    usernameHashed := hex.EncodeToString(hasher.Sum(nil))

    userInput := update.Message.Text

    redisKey := fmt.Sprintf("user:%d:state", usernameHashed)

    args := &fsm.StateArgs{
        Input:     userInput,
        RedisKey:  redisKey,
        StateName: "StateCreatingServer", // used only on first time
    }

    msg, done, err := fsm.Run(ctx, args, fsm.Registry)
    if !done || err != nil {
        b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: msg})
    }
}

