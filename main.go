package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"

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
    dbConnInstance := &dbConn{db: db}
//     res, err := db.NewCreateTable().Model((*User)(nil)).Exec(ctx)
//     if err != nil {
//         panic(err)
//     }
    log.Printf("Table Users created")
	opts := []bot.Option{
		bot.WithDefaultHandler(dbConnInstance.defaultHandler),
		bot.WithCallbackQueryDataHandler("button", bot.MatchTypePrefix, dbConnInstance.callbackHandler),
		bot.WithMessageTextHandler("/start", bot.MatchTypeExact, dbConnInstance.defaultHandler),
	}
    b, err := bot.New(telegramToken, opts...)
	if err != nil {
		panic(err)
	}
    log.Printf("Starting bot...")
	b.Start(ctx)
	log.Printf("Bot shutdown...")
}

type dbConn struct {
    db *bun.DB
}

func (dbConnection *dbConn) callbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// answering callback query first to let Telegram know that we received the callback query,
	// and we're handling it. Otherwise, Telegram might retry sending the update repetitively
	// as it thinks the callback query doesn't reach to our application. learn more by
	// reading the footnote of the https://core.telegram.org/bots/api#callbackquery type.

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})
    var buttonName string = update.CallbackQuery.Data
    if buttonName == "get-vpn-key" {
        b.SendMessage(ctx, &bot.SendMessageParams{
		    ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		    Text:   "get vpn key: ",
	    })
    } else {
        b.SendMessage(ctx, &bot.SendMessageParams{
		    ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		    Text:   "You selected the button: " + buttonName,
	    })
    }

}


func (dbConnection *dbConn) defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "I am YetAnotherVPN Bot. \n"+
		             "I can sign up a new user, " +
		             "change limits on monthly traffic " +
		             "or regenerate an api key in case your's has stopped working. \n" + "\n" +
		             "I am built on open source outline vpn technology. \n"
		             "/start \n" +
		             "/signup \n" +
		             "/reissueApiKey \n" +
		             "/changeLimits \n",
    })
// 	kb := &models.InlineKeyboardMarkup{
// 		InlineKeyboard: [][]models.InlineKeyboardButton{
// 			{
// 				{Text: "Получить новый ключ", CallbackData: "get-vpn-key"},
// 				{Text: "Button 2", CallbackData: "button_2"},
// 			}, {
// 				{Text: "Button 3", CallbackData: "button_3"},
// 			},
// 		},
// 	}
//
// 	b.SendMessage(ctx, &bot.SendMessageParams{
// 		ChatID:      update.Message.Chat.ID,
// 		Text:        "Click by button",
// 		ReplyMarkup: kb, })
}
