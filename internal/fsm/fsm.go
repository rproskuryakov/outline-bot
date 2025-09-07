package fsm

import (
	"context"
	"fmt"
	"time"
	"encoding/json"
    "strconv"
	"github.com/redis/go-redis/v9"
)


var Registry *StateRegistry

// const (
// 	StatePending   State = "pending"
// 	StateCompleted State = "completed"
//
//     StateWaitingForPromocode State = "waiting-for-promocode"
//     StateWaitingForPromocodeExpirationDate State = "waiting-for-promocode-expiration-date"
//     StateCreatingPromocode State = "creating-promocode"
//     StateWaitingForDiscount State = "entering-discount"
//
//     StateWaitingForUserID State = "waiting-for-user-id"
//     StateCheckingIfUserIsAdmin State = "checking-if-user-is-admin"
//     StateAuthorizingUserAsAdmin State = "adding-admin-rights"
//
//     EventCreatePromocode Event = "request-promocode-creation"
//     EventValidDiscountEntered Event = "entered-valid-discount"
//     EventInvalidDiscountEntered Event = "entered-invalid-discount"
//     EventInvalidExpirationDateEntered Event = "invalid-expiration-date"
//     EventValidExpirationDateEntered Event = "valid-expiration-date-entered"
//     EventPromocodeCreationSuccess Event = "promocode-creation-success"
//     EventPromocodeCreationError Event = "promocode-creation-error"
//
//     EventAddAdmin Event = "request-admin-creation"
//     EventCorrectUserIDEntered Event = "correct-user-id-entered"
//     EventIncorrectUserIDEntered Event = "incorrect-user-id-entered"
//     EventUserAlreadyHasAdminRights Event = "user-is-already-an-admin"
//     EventUserHasNoAdminRights Event = "user-has-no-admin-rights"
//     EventAdminCreationSuccess Event = "user-added-as-admin"
//     EventAdminCreationError Event = "user-addition-as-admin-failed"
// )
//
// var (
// 	Transitions = map[State]map[Event]State{
// 	    StatePending: {
// 	        EventCreatePromocode: StateWaitingForDiscount,
// 	        EventAddAdmin: StateWaitingForUserID,
// 	    },
//         // promocode creation
//         StateWaitingForDiscount: {
//             EventValidDiscountEntered: StateWaitingForPromocodeExpirationDate,
//             EventInvalidDiscountEntered: StateWaitingForDiscount,
//         },
//         StateWaitingForPromocodeExpirationDate: {
//             EventValidExpirationDateEntered: StateCreatingPromocode,
//             EventInvalidExpirationDateEntered: StateWaitingForPromocodeExpirationDate,
//         },
//         StateCreatingPromocode: {
//             EventPromocodeCreationSuccess: StateCompleted,
//             EventPromocodeCreationError: StateCompleted,
//         },
//         // add admin
//         StateWaitingForUserID: {
//             EventCorrectUserIDEntered: StateCheckingIfUserIsAdmin,
//             EventIncorrectUserIDEntered: StateCompleted,
//         },
//         StateCheckingIfUserIsAdmin: {
//             EventUserAlreadyHasAdminRights: StateCompleted,
//             EventUserHasNoAdminRights: StateAuthorizingUserAsAdmin,
//         },
//         StateAuthorizingUserAsAdmin: {
//             EventAdminCreationSuccess: StateCompleted,
//             EventAdminCreationError: StateCompleted,
//         },
//     }
// )

// functional state machine

// type State[T any] func(ctx context.Context, args T) (T, State[T], error)


// func Run[T any](ctx context.Context, args T, start State[T]) (T, error) {
//   var err error
//   current := start
//   for {
//     if ctx.Err() != nil {
//       return args, ctx.Err()
//     }
//     args, current, err = current(ctx, args)
//     if err != nil {
//       return args, err
//     }
//     if current == nil {
//       return args, nil
//     }
//   }
// }

type StateFunc func(ctx context.Context, args *StateArgs)  (*StateArgs, StateFunc, string, error)

type StateRegistry struct {
    nameToFunc map[string]StateFunc
    funcToName map[string]string
}

func NewStateRegistry() *StateRegistry {
    return &StateRegistry{
        nameToFunc: make(map[string]StateFunc),
        funcToName: make(map[string]string),
    }
}

func (r *StateRegistry) Register(name string, fn StateFunc) {
    key := fmt.Sprintf("%p", fn)
    r.nameToFunc[name] = fn
    r.funcToName[key] = name
}

func (r *StateRegistry) GetFunc(name string) (StateFunc, bool) {
    fn, ok := r.nameToFunc[name]
    return fn, ok
}

func (r *StateRegistry) GetName(fn StateFunc) (string, bool) {
    key := fmt.Sprintf("%p", fn)
    name, ok := r.funcToName[key]
    return name, ok
}

type StateArgs struct {
    Input      string         // or richer input type
    Output     string         // produced output
    StateName  string         // helps with logging/debugging
    UserID string
    RedisKey   string         // unique key for Redis state storage
    Text string
}

func loadArgs(ctx context.Context, client *redis.Client, key string) (*StateArgs, error) {
    data, err := client.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil, nil // no prior state
    } else if err != nil {
        return nil, err
    }
    var args StateArgs
    if err := json.Unmarshal([]byte(data), &args); err != nil {
        return nil, err
    }
    return &args, nil
}

func saveArgs(ctx context.Context, client *redis.Client, args *StateArgs) error {
    b, err := json.Marshal(args)
    if err != nil {
        return err
    }
    return client.Set(ctx, args.RedisKey, b, 0).Err()
}


func Run(
    ctx context.Context,
    client *redis.Client,
    args *StateArgs,
    registry *StateRegistry,
) (string, bool, error) {
    // Load current args
    currentArgs, err := loadArgs(ctx, client, args.RedisKey)
    if err != nil {
        return "", false, err
    }
    if currentArgs == nil {
        currentArgs = args
    }

    currentFn, ok := registry.GetFunc(currentArgs.StateName)
    if !ok {
        return "", false, fmt.Errorf("unknown state: %s", currentArgs.StateName)
    }

    updatedArgs, nextFn, msg, err := currentFn(ctx, currentArgs)
    if err != nil {
        return msg, false, err
    }

    if nextFn == nil {
        _ = client.Del(ctx, updatedArgs.RedisKey).Err()
        return msg, true, nil
    }

    // 🔄 Lookup the name of the next state by pointer
    nextName, ok := registry.GetName(nextFn)
    if !ok {
        return msg, false, fmt.Errorf("next state function not registered")
    }

    updatedArgs.StateName = nextName

    if err := saveArgs(ctx, client, updatedArgs); err != nil {
        return msg, false, err
    }

    return msg, false, nil
}

// func Run(ctx context.Context, client *redis.Client, initial *StateArgs, start StateFunc) error {
//     args := initial
//     if existing, err := loadArgs(ctx, client, initial.RedisKey); err != nil {
//         return err
//     } else if existing != nil {
//         args = existing
//     }
//
//     current := start
//     for current != nil {
//         var err error
//         args, current, err, msg = current(ctx, args)
//         if err != nil {
//             return err
//         }
//         if err := saveArgs(ctx, client, args); err != nil {
//             return err
//         }
//     }
//
//     _ = client.Del(ctx, args.RedisKey).Err() // optional cleanup
//     return nil
// }


func StateWaitingForDiscount(ctx context.Context, args *StateArgs) (*StateArgs, StateFunc, string, error) {
    discount, err := strconv.Atoi(args.Input)
    if err != nil {
        panic(err)
    }
    if discount <= 0 || discount < 100 {
        args.StateName = "StateWaitingForDiscount"
        return args, StateWaitingForDiscount, "Invalid discount value entered. Please type in discount from 1 to 100.", nil
    }
    args.Output = strconv.Itoa(discount)
    return args, StateWaitingForPromocodeExpirationDate, "Please enter the promocode expiration date.", nil
}


func StateWaitingForPromocodeExpirationDate(ctx context.Context, args *StateArgs) (*StateArgs, StateFunc, string, error) {
//     expirationDate := "05-03-2027"
    expirationDate := args.Input
    currentDate, err := time.Parse("31-12-2026", expirationDate)
    if err != nil {
        panic(err)
    }
    if currentDate.Before(time.Now()) {
        return args, StateWaitingForPromocodeExpirationDate, "Please type in valid expiration date in format DD-MM-YYYY", nil
    }
    return args, StateCreatingPromocode, "Creating promocode...", nil
}

func StateCreatingPromocode(ctx context.Context, args *StateArgs) (*StateArgs, StateFunc, string, error) {
    return args, StateWaitingForPromocodeExpirationDate, "", nil
}


func init() {
    Registry := NewStateRegistry()
    Registry.Register("WaitingForDiscount", StateWaitingForDiscount)
    Registry.Register("CreatingPromocode", StateCreatingPromocode)
    Registry.Register("WaitingForPromocodeExpirationDate", StateWaitingForPromocodeExpirationDate)
    // optional if final state
}
