package internal

import (
    "strconv"
)


type StateHandler interface {
    Handle(username string, state UserState, msg string) (UserState, Event, error)
}


// func getStateForHandler(fsm *fsm.RedisFSM, userID int64, formDef struct) {
//     val, getErr := fsm.GetState(strconv.FormatInt(userID, 10))
//     if getErr != nil {
//         panic(getErr)
//     }
//     var form formDef
//     unmarshalErr := json.Unmarshal(val, &form)
//     if unmarshalErr != nil {
//         panic(unmarshalErr)
//     }
//     return form
// }


func StateWaitingForDiscountCallback(username string, state UserState, msg string) (UserState, Event, string) {
    // check if discount is correct
    discount, err := strconv.Atoi(msg)
    if err != nil {
        panic(err)
    }
    // check if discount is within the boundaries
    if discount <= 0 || discount > 100 {
        return state, EventInvalidDiscountEntered, "Invalid discount value entered. Please type in discount from 1 to 100."
    }
    state.stateData["discount"] = strconv.Itoa(discount)
    return state, EventValidDiscountEntered, "Please enter the promocode expiration date."
}


func StateWaitingForPromocodeExpirationDateCallback(username string, state UserState, msg string) (UserState, Event, string) {
    return state, EventValidExpirationDateEntered, "Creating promocode..."
//     value, err := fsm.GetState(strconv.FormatInt(userID, 10))
//     if err != nil {
//         panic(err)
//     }
//     currentState := string(value)
//     discount := 5
//     expirationDate := "05-03-2027"
//     currentDate, err := time.Parse("31-12-2026", expirationDate)
//     if err != nil {
//         panic(err)
//     }
//     if currentDate.Before(time.Now()) {
//         return EventInvalidExpirationDateEntered, "Please type in valid expiration date in format DD-MM-YYYY"
//     } else {
//         json, err := json.Marshal(Author{Name: "Elliot", Age: 25})
//         return EventValidExpirationDateEntered, "Creating promocode..."
//     }
}