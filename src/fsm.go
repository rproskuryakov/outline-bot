package src

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

type State string
type Event string

const (
	StatePending   State = "pending"
	StateCompleted State = "completed"

    StateWaitingForPromocode State = "waiting-for-promocode"
    StateCreatingPromocode State = "creating-promocode"
    StateWaitingForDiscount State = "entering-discount"

    StateWaitingForUserID State = "waiting-for-user-id"
    StateCheckingIfUserIsAdmin State = "checking-if-user-is-admin"
    StateAuthorizingUserAsAdmin State = "adding-admin-rights"

    EventCreatePromocode Event = "request-promocode-creation"
    EventValidDiscountEntered Event = "entered-valid-discount"
    EventInvalidDiscountEntered Event = "entered-invalid-discount"
    EventPromocodeCreationSuccess Event = "promocode-creation-success"
    EventPromocodeCreationError Event = "promocode-creation-error"

    EventAddAdmin Event = "request-admin-creation"
    EventCorrectUserIDEntered Event = "correct-user-id-entered"
    EventIncorrectUserIDEntered Event = "incorrect-user-id-entered"
    EventUserAlreadyHasAdminRights Event = "user-is-already-an-admin"
    EventUserHasNoAdminRights Event = "user-has-no-admin-rights"
    EventAdminCreationSuccess Event = "user-added-as-admin"
    EventAdminCreationError Event = "user-addition-as-admin-failed"
)

var (
	Transitions = map[State]map[Event]State{
	    StatePending: {
	        EventCreatePromocode: StateWaitingForDiscount,
	        EventAddAdmin: StateWaitingForUserID,
	    },
        // promocode creation
        StateWaitingForDiscount: {
            EventValidDiscountEntered: StateCreatingPromocode,
            EventInvalidDiscountEntered: StateWaitingForDiscount,
        },
        StateCreatingPromocode: {
            EventPromocodeCreationSuccess: StateCompleted,
            EventPromocodeCreationError: StateCompleted,
        },
        // add admin
        StateWaitingForUserID: {
            EventCorrectUserIDEntered: StateCheckingIfUserIsAdmin,
            EventIncorrectUserIDEntered: StateCompleted,
        },
        StateCheckingIfUserIsAdmin: {
            EventUserAlreadyHasAdminRights: StateCompleted,
            EventUserHasNoAdminRights: StateAuthorizingUserAsAdmin,
        },
        StateAuthorizingUserAsAdmin: {
            EventAdminCreationSuccess: StateCompleted,
            EventAdminCreationError: StateCompleted,
        },
    }
)


var ctx = context.Background()

type GenericFSM interface {
    GetState(key string) (State, error)
    SetState(key string, state State) error
    Trigger(key string, event Event) error
}


type RedisFSM struct {
    redisClient *redis.Client
    transitions map[State]map[Event]State
    strategies  map[State]StateHandler
}

type StateHandler interface {
    Handle(event Event) (State, error)
}

func NewFSM(redisAddr string) *RedisFSM {
	return &RedisFSM{
		redisClient: redis.NewClient(&redis.Options{Addr: redisAddr}),
	}
}


func (fsm *RedisFSM) GetState(key string) (State, error) {
    value, err := fsm.redisClient.Get(ctx, key).Result()
    if err != nil {
        return State(""), err
    }
    return State(value), nil
}

func (fsm *RedisFSM) SetState(key string, state State) error {
	return fsm.redisClient.Set(ctx, key, state, 0).Err()
}

func (fsm *RedisFSM) Trigger(key string, event Event) error {
	value, err := fsm.redisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}
    currentState := State(value)

	handler, ok := fsm.strategies[currentState]
	if !ok {
		return fmt.Errorf("no strategy for state: %s", currentState)
	}

	nextState, err := handler.Handle(event)
	if err != nil {
		return err
	}

	if err := fsm.redisClient.Set(ctx, key, nextState, 0).Err(); err != nil {
		return err
	}

	fmt.Printf("FSM Transitioned '%s' from %s -> %s on event '%s'\n", key, currentState, nextState, event)
	return nil
}
