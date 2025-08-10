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
	"github.com/redis/go-redis/v9"
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
    OwnerID int64
	Owner User `bun:"rel:belongs-to,join:owner_id=id"`
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

type ChangeEvent struct {
    bun.BaseModel `bun:"table:cloud-providers,alias:e"`

    ID        int64     `bun:",pk,autoincrement"`
    UserID    int64     `bun:",notnull"` // FK to users
    Action    string    `bun:",notnull"` // e.g., "CREATE_KEY", "DELETE_KEY"
    Timestamp time.Time
}


type Server struct {
    db *bun.DB
    redisDb *redis.Client
}

func (server *Server) defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    usernameTelegramID := update.Message.From.ID
    hasher := md5.New()
    hasher.Write([]byte(strconv.FormatInt(usernameTelegramID, 10)))
    usernameHashed := hex.EncodeToString(hasher.Sum(nil))

    err := server.redisDb.Set(ctx, usernameHashed, "/default", 200).Err()
    if err != nil {
        panic(err)
    }
    val, err := server.redisDb.Get(ctx, usernameHashed).Result()
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

func (server *Server) addAdminHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
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
    if !exists {
        b.SendMessage(ctx, &bot.SendMessageParams{
		    ChatID:      update.Message.Chat.ID,
		    Text:        "User does not exists.",
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

func (server *Server) createServerHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    usernameTelegramID := update.Message.From.ID
    // check if user exists
    exists, existsError := checkIfUserExists(ctx, usernameTelegramID, server.db)
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
    user, getAttrsError := getUserAttributes(ctx, usernameTelegramID, server.db)
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
    _, insertErr := server.db.NewInsert().Model(serverRecord).Exec(ctx)

    if insertErr != nil {
        log.Printf(insertErr.Error())
        panic(insertErr)
    }
    log.Printf("Server record is created.")
}


func main() {
    var telegramToken string = os.Getenv("TELEGRAM_API_TOKEN")
    var postgresDsn string = os.Getenv("POSTGRES_DSN")
    var redisPassword string = os.Getenv("REDIS_PASSWORD")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

    // dsn := "unix://user:pass@dbname/var/run/postgresql/.s.PGSQL.5432"
    sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(postgresDsn)))
    db := bun.NewDB(sqldb, pgdialect.New())
    err := db.ResetModel(ctx, (*User)(nil))
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
    server := &Server{db: db, redisDb: redisDB}
	opts := []bot.Option{
		bot.WithDefaultHandler(server.defaultHandler),
		bot.WithMessageTextHandler("/start", bot.MatchTypeExact, server.startHandler),
		bot.WithMessageTextHandler("/signUp", bot.MatchTypeExact, server.signUpHandler),
		bot.WithMessageTextHandler("/issueApiKey", bot.MatchTypeExact, server.issueApiKeyHandler),
		bot.WithMessageTextHandler("/reissueApiKey", bot.MatchTypeExact, server.reissueApiKeyHandler),
		bot.WithMessageTextHandler("/createServer", bot.MatchTypeExact, server.createServerHandler),
		bot.WithMessageTextHandler("/changeLimits", bot.MatchTypeExact, server.changeLimitsHandler),
		bot.WithMessageTextHandler("/viewTrafficUsed", bot.MatchTypeExact, server.viewTrafficUsedHandler),
		bot.WithMessageTextHandler("/addAdmin", bot.MatchTypeExact, server.addAdminHandler),
	}
    b, err := bot.New(telegramToken, opts...)
	if err != nil {
		panic(err)
	}
    log.Printf("Starting bot...")
	b.Start(ctx)
	log.Printf("Bot shutdown...")
}
