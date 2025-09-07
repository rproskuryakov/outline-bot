package fsm

import (
    "context"
    "time"
    "strconv"
)


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


func StatePending(ctx context.Context, args *StateArgs) (*StateArgs, StateFunc, string, error) {
    if args.Input == "/start" {
        return args, StatePending, "", nil
    } else if args.Input == "/createPromocode" {
        return args, StateWaitingForDiscount, "Enter a valid discount from 0 to 100 percent.", nil
    }
    return args, StatePending, "", nil
}

// promocode addition
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
    return args, StateWaitingForDiscount, "", nil
}


func init() {
    Registry := NewStateRegistry()
    Registry.Register("StatePending", StatePending)
    // promocode creation
    Registry.Register("WaitingForDiscount", StateWaitingForDiscount)
    Registry.Register("WaitingForPromocodeExpirationDate", StateWaitingForPromocodeExpirationDate)
    Registry.Register("CreatingPromocode", StateCreatingPromocode)
    //

    // optional if final state
}