package control

// Node is an interface with an address and a state
type Node interface {
	Address() (Address, error)
	State() (State, error)
}
