package fsm

import (
    "context"
    "time"
    "strconv"
)


func StatePending(ctx context.Context, args *StateArgs) (*StateArgs, StateFunc, string, error) {
    if args.Input == "/start" {
        return args, StatePending, "", nil
    } else if args.Input == "/createPromocode" {
        return args, StateWaitingForDiscount, "Enter a valid discount from 0 to 100 percent.", nil
    }
    return args, StatePending, "", nil
}

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
    Registry.Register("StatePending", StatePending)
    Registry.Register("WaitingForDiscount", StateWaitingForDiscount)
    Registry.Register("WaitingForPromocodeExpirationDate", StateWaitingForPromocodeExpirationDate)
    Registry.Register("CreatingPromocode", StateCreatingPromocode)
    // optional if final state
}