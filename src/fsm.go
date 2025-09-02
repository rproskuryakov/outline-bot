package src

import (
	"context"
	"fmt"
	"time"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type State string

type UserState struct {
    state State
    stateData map[string]string
}

type Event string

const (
	StatePending   State = "pending"
	StateCompleted State = "completed"

    StateWaitingForPromocode State = "waiting-for-promocode"
    StateWaitingForPromocodeExpirationDate State = "waiting-for-promocode-expiration-date"
    StateCreatingPromocode State = "creating-promocode"
    StateWaitingForDiscount State = "entering-discount"

    StateWaitingForUserID State = "waiting-for-user-id"
    StateCheckingIfUserIsAdmin State = "checking-if-user-is-admin"
    StateAuthorizingUserAsAdmin State = "adding-admin-rights"

    EventCreatePromocode Event = "request-promocode-creation"
    EventValidDiscountEntered Event = "entered-valid-discount"
    EventInvalidDiscountEntered Event = "entered-invalid-discount"
    EventInvalidExpirationDateEntered Event = "invalid-expiration-date"
    EventValidExpirationDateEntered Event = "valid-expiration-date-entered"
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
            EventValidDiscountEntered: StateWaitingForPromocodeExpirationDate,
            EventInvalidDiscountEntered: StateWaitingForDiscount,
        },
        StateWaitingForPromocodeExpirationDate: {
            EventValidExpirationDateEntered: StateCreatingPromocode,
            EventInvalidExpirationDateEntered: StateWaitingForPromocodeExpirationDate,
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

type PromocodeForm struct {
	Discount int64 `json:"discount"`
	ExpirationDate time.Time `json:"expirationDate"`
}


var ctx = context.Background()

type GenericFSM interface {
    GetState(key string) (State, error)
    SetState(state UserState) error
    Trigger(key string, event Event) error
}


type RedisFSM struct {
    redisClient *redis.Client
    transitions map[State]map[Event]State
    callbacks  map[State]StateHandler
    startState State
}

func NewFSM(redisClient *redis.Client) *RedisFSM {
	return &RedisFSM{
		redisClient: redisClient,
		startState: StatePending,
		transitions: Transitions,
	}
}


func (fsm *RedisFSM) GetState(key string) (UserState, error) {
    val, err := fsm.redisClient.Get(ctx, key).Result()
    if err != nil {
        return UserState{}, err
    }
    //   unmarshal
    var userData UserState
    unmarshalErr := json.Unmarshal([]byte(val), &userData)
    if unmarshalErr != nil {
        return UserState{}, err
    }
    return userData, nil
}


func (fsm *RedisFSM) SetState(key string, state UserState) error {
    json, err := json.Marshal(state)
    if err != nil {
        return err
    }
    err = fsm.redisClient.Set(ctx, key, json, 0).Err()
    if err != nil {
        return err
    }
    return nil
}


func (fsm *RedisFSM) Transition(key string, event Event) error {
	value, err := fsm.redisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}
    currentState := State(value)

    nextState, ok := fsm.transitions[currentState][event]
    if !ok {
        return fmt.Errorf("Event " + string(event) + " unavailable for state " + string(currentState))
    }

	if err := fsm.redisClient.Set(ctx, key, nextState, 0).Err(); err != nil {
		return err
	}

	fmt.Printf("FSM Transitioned '%s' from %s -> %s on event '%s'\n", key, currentState, nextState, event)
	return nil
}
