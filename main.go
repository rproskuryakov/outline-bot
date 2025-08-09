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
	"time"

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
	IsAdmin bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AccessKey struct {
    bun.BaseModel `bun:"table:vpn-access-keys,alias:k"`

    ID	 int64  `bun:",pk,autoincrement"`
    Name string
    URL string
    APIKey string
    UserID int64
	User User `bun:"rel:belongs-to,join:user_id=id"`
	ServerID int64
	Server ServerRecord `bun:"rel:belongs-to,join:server_id=id"`
	ByteLimit int64
}

type ServerRecord struct {
    bun.BaseModel `bun:"table:vpn-servers,alias:s"`

    ID	 int64  `bun:",pk,autoincrement"`
    CreatedAt time.Time
    APIKey string
    UserID int64
	User User `bun:"rel:belongs-to,join:user_id=id"`
	IsActive bool
	CloudProvider CloudProvider `bun:"rel:belongs-to,join:cloud_provider_id=id"`
	CloudProviderID int64
}

type CloudProvider struct {
    bun.BaseModel `bun:"table:cloud-providers,alias:c"`

    ID	 int64  `bun:",pk,autoincrement"`
    Name string
    CreatedAt time.Time

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

func checkIfUserExists(ctx context.Context, username int64, db *bun.DB) (f bool, err error) {
    hasher := md5.New()
    hasher.Write([]byte(strconv.FormatInt(username, 10)))
    usernameHashed := hex.EncodeToString(hasher.Sum(nil))

    user := new(User)
    exists, err := db.NewSelect().Model(user).Where("username = ?", usernameHashed).Exists(ctx)
    if err != nil {
        return false, err
    }
    return exists, nil
}

func getUserAttributes(ctx context.Context, username int64, db *bun.DB) (u *User, e error) {
    hasher := md5.New()
    hasher.Write([]byte(strconv.FormatInt(username, 10)))
    usernameHashed := hex.EncodeToString(hasher.Sum(nil))

    user := new(User)
    err := db.NewSelect().Model(user).Where("id = ?", usernameHashed).Scan(ctx)
    if err != nil {
        return user, err
    }
    return user, nil
}


func (server *Server) startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    usernameTelegramID := update.Message.From.ID
    // check if user exists
    exists, err := checkIfUserExists(ctx, usernameTelegramID, server.db)
    if err != nil {
        log.Printf(err.Error())
        panic(err)
    }

    user, err := getUserAttributes(ctx, usernameTelegramID, server.db)
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

func (server *Server) signUpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    usernameTelegramID := update.Message.From.ID
    exists, err := checkIfUserExists(ctx, usernameTelegramID, server.db)

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
        return
    } else {
        hasher := md5.New()
        hasher.Write([]byte(strconv.FormatInt(usernameTelegramID, 10)))
        usernameHashed := hex.EncodeToString(hasher.Sum(nil))

        user := &User{Name: usernameHashed, IsAdmin: false}
        _, err := server.db.NewInsert().Model(user).Exec(ctx)

        if err != nil {
            log.Printf(err.Error())
            panic(err)
        }
    }

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

func (server *Server) createServerHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "createServer",
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
		bot.WithMessageTextHandler("/createServer", bot.MatchTypeExact, server.createServerHandler),
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
