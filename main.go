package main

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)


type User struct {
    bun.BaseModel `bun:"table:vpn-users,alias:u"`

	ID	 int64  `bun:",pk,autoincrement"`
	Name string
	Password string
}


type Server struct {
    db *bun.DB
}

func (server *Server) defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
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

func (server *Server) startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    hasher := md5.New()
    hasher.Write([]byte(strconv.FormatInt(update.Message.From.ID, 10)))
    b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Your hashed telegram id " + hex.EncodeToString(hasher.Sum(nil)),
    })
//     check if user exists in database
//     if so, then they can reissueApiKey, viewTrafficUsed, changeLimits
//     if user doesnt exist in database then they can sign up
//     and then issueApiKey

}

func (server *Server) signUpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "signUp",
    })
}


func (server *Server) issueApiKeyHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "issueApiKey",
    })
}

func (server *Server) reissueApiKeyHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "reissueApiKey",
    })
}

func (server *Server) viewTrafficUsedHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "viewTrafficUsed",
    })
}

func (server *Server) changeLimitsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "changeLimits",
    })
}


func main() {
    var telegramToken string = os.Getenv("TELEGRAM_API_TOKEN")
    var postgresDsn string = os.Getenv("POSTGRES_DSN")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

    // dsn := "unix://user:pass@dbname/var/run/postgresql/.s.PGSQL.5432"
    sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(postgresDsn)))
    db := bun.NewDB(sqldb, pgdialect.New())
    err := db.ResetModel(ctx, (*User)(nil))
    if err != nil {
        panic(err)
    }
    server := &Server{db: db}
    log.Printf("Table Users created")
	opts := []bot.Option{
		bot.WithDefaultHandler(server.defaultHandler),
		bot.WithMessageTextHandler("/start", bot.MatchTypeExact, server.startHandler),
		bot.WithMessageTextHandler("/signUp", bot.MatchTypeExact, server.signUpHandler),
		bot.WithMessageTextHandler("/issueApiKey", bot.MatchTypeExact, server.issueApiKeyHandler),
		bot.WithMessageTextHandler("/reissueApiKey", bot.MatchTypeExact, server.reissueApiKeyHandler),
		bot.WithMessageTextHandler("/changeLimits", bot.MatchTypeExact, server.changeLimitsHandler),
		bot.WithMessageTextHandler("/viewTrafficUsed", bot.MatchTypeExact, server.viewTrafficUsedHandler),
	}
    b, err := bot.New(telegramToken, opts...)
	if err != nil {
		panic(err)
	}
    log.Printf("Starting bot...")
	b.Start(ctx)
	log.Printf("Bot shutdown...")
}
