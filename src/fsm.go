package src

import (
// 	"fmt"
// 	"log"

	"github.com/looplab/fsm"
)


type UserFSM struct {
    UserID int64
    StateMachine *fsm.FSM
}


func NewStateMachine(UserID int64) *UserFSM {
    return &UserFSM{
        UserID: UserID,
        StateMachine: fsm.NewFSM(
            "Beginning",
            fsm.Events{
                // event name, src states, destination state
                {Name: "UserFound", Src: []string{"Beginning"}, Dst: "Ending"},
                {Name: "UserNotFound", Src: []string{"Beginning"}, Dst: "SignUpStart"},
                {Name: "UserCouldNotBeSignedUp", Src: []string{"SignUpStart"}, Dst: "Ending"},
                {Name: "UserSignedUpSuccessfully", Src: []string{"SignUpStart"}, Dst: "Ending"},
            },
            fsm.Callbacks{},
        ),
    }
}


//                 {Name: "UserCouldNotBeSignedUp"},
//                 {Name: "IssueAPIKey"},
//                 {Name: "ReissueAPIKey"},
//                 {Name: "CreateServer"},
//                 {Name: "ChangeLimits"},
//                 {Name: "AddAdmin"},
//                 {Name: "ViewTrafficUsed"},
//                 {Name: "UserExists"},
//                 {Name: "UserDoesNotExists"},
//                 {Name: "UserHasAdminRights"},
//                 {Name: "UserDoesNotHaveAdminRights"},
//                 {Name: "End", Src: []string{"Started"}, Dst: "closed"},