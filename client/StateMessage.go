package client

// StateMessage is a type that is passed to a Client to signal a change in
// its current state.
type StateMessage struct {
	State               StateI
	Args                interface{}
	Push, Pop, PopToTop bool
}
