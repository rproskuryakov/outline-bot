package fsm

import (
	"context"
	"fmt"
	"encoding/json"
	"github.com/redis/go-redis/v9"

	"github.com/rproskuryakov/outline-bot/internal/clients"
    "github.com/rproskuryakov/outline-bot/internal/repositories"
// 	"github.com/rproskuryakov/outline-bot/internal/infra"
)


var Registry *StateRegistry

type StateFunc func(ctx context.Context, args *StateArgs, machine *StateMachine)  (*StateArgs, StateFunc, string, error)

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
    UserID int64
    RedisKey   string         // unique key for Redis state storage
    Text string
}

type StateMachine struct {
    OutlineClients clients.OutlineVPNClients
    UserRepository repositories.UserRepository
    ServerRepository repositories.ServerRepository
    StateStorage StateStorage
}

type StateStorage struct {
    RedisClient *redis.Client
}


func (storage *StateStorage) loadArgs(ctx context.Context, key string) (*StateArgs, error) {
    data, err := storage.RedisClient.Get(ctx, key).Result()
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

func (storage *StateStorage) saveArgs(ctx context.Context, args *StateArgs) error {
    b, err := json.Marshal(args)
    if err != nil {
        return err
    }
    return storage.RedisClient.Set(ctx, args.RedisKey, b, 0).Err()
}


func Run(
    ctx context.Context,
    args *StateArgs,
    registry *StateRegistry,
    machine *StateMachine,
) (string, bool, error) {
    // Load current args
    currentArgs, err := machine.StateStorage.loadArgs(ctx, args.RedisKey)
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

    updatedArgs, nextFn, msg, err := currentFn(ctx, currentArgs, machine)
    if err != nil {
        return msg, false, err
    }

    if nextFn == nil {
        _ = machine.StateStorage.RedisClient.Del(ctx, updatedArgs.RedisKey).Err()
        return msg, true, nil
    }

    // 🔄 Lookup the name of the next state by pointer
    nextName, ok := registry.GetName(nextFn)
    if !ok {
        return msg, false, fmt.Errorf("next state function not registered")
    }

    updatedArgs.StateName = nextName

    if err := machine.StateStorage.saveArgs(ctx, updatedArgs); err != nil {
        return msg, false, err
    }

    return msg, false, nil
}

