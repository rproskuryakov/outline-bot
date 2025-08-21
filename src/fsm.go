package src

import (
	"fmt"
	"log"
)

// State represents a unique state name
type State string

// Event represents a named transition
type Event string

// Transition maps: current state + event => next state
type Transition struct {
	From  State
	Event Event
	To    State
}

type StateMachine struct {
	current     State
	transitions map[State]map[Event]State
}

type StateMachineFactory struct {
	transitions map[State]map[Event]State
}

func NewStateMachineFactory() *StateMachineFactory {
	return &StateMachineFactory{
		transitions: make(map[State]map[Event]State),
	}
}

// AddTransition dynamically adds a transition
func (f *StateMachineFactory) AddTransition(from State, event Event, to State) {
	if f.transitions[from] == nil {
		f.transitions[from] = make(map[Event]State)
	}
	f.transitions[from][event] = to
}

// Create builds a new FSM with the initial state
func (f *StateMachineFactory) Create(initial State) *StateMachine {
	return &StateMachine{
		current:     initial,
		transitions: f.transitions,
	}
}

func (sm *StateMachine) Current() State {
	return sm.current
}

func (sm *StateMachine) Trigger(event Event) error {
	next, ok := sm.transitions[sm.current][event]
	if !ok {
		return fmt.Errorf("no transition from %s on event %s", sm.current, event)
	}
	log.Printf("Transitioning: %s --(%s)--> %s", sm.current, event, next)
	sm.current = next
	return nil
}