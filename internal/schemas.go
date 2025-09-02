package internal

import (
	"time"

	"github.com/uptrace/bun"
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
    bun.BaseModel `bun:"table:vpn-access-keys,alias:ak"`

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
    bun.BaseModel `bun:"table:vpn-servers,alias:sr"`

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
    bun.BaseModel `bun:"table:cloud-providers,alias:cp"`

    ID	 int64  `bun:",pk,autoincrement"`
    Name string
    CreatedAt time.Time
}

type ChangeEvent struct {
    bun.BaseModel `bun:"table:cloud-providers,alias:ch"`

    ID        int64     `bun:",pk,autoincrement"`
    UserID    int64     `bun:",notnull"`
    User User `bun:"rel:belongs-to,join:user_id=id"` // FK to users
    Action    string    `bun:",notnull"`
    EventTimestamp time.Time `bun:",notnull"`
}

type Promocode struct {
    bun.BaseModel `bun:"table:promocodes,alias:p"`

    ID        int64 `bun:",pk,autoincrement"`
    Name    string  `bun:",notnull"`
    Discount int64  `bun:",notnull"`
    CreatedAtTimestamp time.Time    `bun:",notnull"`
    ValidFromTimestamp time.Time    `bun:",notnull"`
    ValidToTimestamp time.Time  `bun:",notnull"`
}

type PromocodeUsers struct {
    bun.BaseModel `bun:"table:promocode-users,alias:pu"`

    ID        int64     `bun:",pk,autoincrement"`
    UserID    int64     `bun:",notnull"`
    User User `bun:"rel:belongs-to,join:user_id=id"`
    PromocodeID int64   `bun:",notnull"`
    Promocode Promocode `bun:"rel:belongs-to,join:promocode_id=id"`
    ActivationTimestamp    time.Time    `bun:",notnull"`
}
