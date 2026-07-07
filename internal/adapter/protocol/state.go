package protocol

import "fmt"

type State uint8

const (
	StateHandshake State = iota
	StateStatus
	StateLogin
	StateConfiguration
	StatePlay
)

func StateFromNext(n int32) (State, error) {
	switch n {
	case 1:
		return StateStatus, nil
	case 2:
		return StateLogin, nil
	case 3:
		return State(0), fmt.Errorf("transfer not supported")
	default:
		return State(0), fmt.Errorf("invalid next state: %d", n)
	}
}
