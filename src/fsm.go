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
	Data map[string]string
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

func GetFiniteStateMachine() *StateMachine {
    // init state machine
    Factory := NewStateMachineFactory()

    // Define states and events
    const (
        StateBeginning State = "Beginning"
        StateWaitingForTheCommand State = "WaitingForTheCommand"
	    StateCheckingIfUserExists    State = "CheckingIfUserExists"
	    StateCheckingIfUserHasAdminRights State = "CheckingIfUserHasAdminRights"
	    StateSigningUpUser  State = "SigningUpUser"
	    StateCheckingPromocodeActivity State = "CheckingPromocodeActivity"
	    StateRegisteringPromocode State = "RegisteringPromocode"
	    StateCreatingAPIKey State = "CreatingAPIKey"
	    StateEnd State = "End"

	    EventStart  Event = "Start"

	    EventSignUpRequest Event = "SignUpRequest"
	    EventUserSignedUpSuccessfully Event = "UserSignedUpSuccessfully"
	    EventUserCouldNotBeSigneUp Event = "UserCouldNotBeSignedUp"

	    EventIssueAPIKey Event = "IssueAPIKey"
	    EventReissueAPIKey Event = "ReissueAPIKey"
	    EventCreateServer Event = "CreateServer"
	    EventChangeLimits Event = "ChangeLimits"
	    EventAddAdmin Event = "AddAdmin"
	    EventViewTrafficUsed Event = "ViewTrafficUsed"
	    EventUserExists  Event = "UserExists"
	    EventUserDoesNotExists  Event = "UserDoesNotExists"
	    EventUserHasAdminRights  Event = "UserHasAdminRights"
	    EventUserDoesNotHaveAdminRights Event = "UserDoesNotHaveAdminRights"
	    EventResume Event = "RESUME"
	    EventStop   Event = "STOP"
    )

    // SignUpRequest
    Factory.AddTransition(StateBeginning, EventSignUpRequest, StateCheckingIfUserExists)
    Factory.AddTransition(StateCheckingIfUserExists, EventUserExists, StateEnd)
    Factory.AddTransition(StateCheckingIfUserExists, EventUserDoesNotExists, StateSigningUpUser)
    Factory.AddTransition(StateSigningUpUser, EventUserSignedUpSuccessfully, StateEnd)
    Factory.AddTransition(StateSigningUpUser, EventUserCouldNotBeSigneUp, StateEnd)

    // Create FSM instance with initial state
    fsm := Factory.Create(StateBeginning)

    return fsm
}



// Simulate events
// 	events := []Event{EventStart, EventPause, EventResume, EventStop}
// 	for _, e := range events {
// 		if err := fsm.Trigger(e); err != nil {
// 			log.Println("Error:", err)
// 			break
// 		}
// 		log.Println("Current State:", fsm.Current())
// 	}