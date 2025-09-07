package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/redis/go-redis/v9"

	"github.com/rproskuryakov/outline-bot/internal/handlers"
	"github.com/rproskuryakov/outline-bot/internal/model"
)


func main() {
    var telegramToken string = os.Getenv("TELEGRAM_API_TOKEN")
    var postgresDsn string = os.Getenv("POSTGRES_DSN")
    var redisPassword string = os.Getenv("REDIS_PASSWORD")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

    // dsn := "unix://user:pass@dbname/var/run/postgresql/.s.PGSQL.5432"
    sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(postgresDsn)))
    db := bun.NewDB(sqldb, pgdialect.New())
    err := db.ResetModel(ctx, (*model.User)(nil))
    log.Printf("Table Users created")
    if err != nil {
        panic(err)
    }
    redisDB := redis.NewClient(&redis.Options{
        Addr:	  "cache:6379",
        Password: redisPassword, // No password set
        DB:		  0,  // Use default DB
        Protocol: 2,  // Connection protocol
    })

    server := &handlers.Server{Db: db, RedisClient: redisDB}
	opts := []bot.Option{
		bot.WithDefaultHandler(server.DefaultHandler),
		bot.WithMessageTextHandler("/start", bot.MatchTypeExact, handlers.CheckAuthorized(server, server.StartHandler)),
		bot.WithMessageTextHandler("/signUp", bot.MatchTypeExact, server.SignUpHandler),
		bot.WithMessageTextHandler("/issueApiKey", bot.MatchTypeExact, handlers.CheckAuthorized(server, server.IssueApiKeyHandler)),
		bot.WithMessageTextHandler("/reissueApiKey", bot.MatchTypeExact, handlers.CheckAuthorized(server, server.ReissueApiKeyHandler)),
		bot.WithMessageTextHandler("/createServer", bot.MatchTypeExact, handlers.CheckAuthorizedAdmin(server, server.CreateServerHandler)),
		bot.WithMessageTextHandler("/changeLimits", bot.MatchTypeExact, server.ChangeLimitsHandler),
		bot.WithMessageTextHandler("/viewTrafficUsed", bot.MatchTypeExact, handlers.CheckAuthorized(server, server.ViewTrafficUsedHandler)),
		bot.WithMessageTextHandler("/addAdmin", bot.MatchTypeExact, handlers.CheckAuthorizedAdmin(server, server.AddAdminHandler)),
	}
    b, err := bot.New(telegramToken, opts...)
	if err != nil {
		panic(err)
	}
    log.Printf("Starting bot...")
	b.Start(ctx)
	log.Printf("Bot shutdown...")
}
