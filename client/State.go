package client

/*
~~~~ States In Brief ~~~~
Chimera's go client primarily works via a state machine. New states are passed into the state machine by either returning the next state during `Init()` or by sending a StateI through the Client's StateChannel.

The functions of a state are defined through the State interface known as StateI. Although a Client state may implement each function of the interface, it may be more convenient to simply embed the State type, which provides the base functions and methods that all states must implement.
*/

/*
StateI provides the base interface for all Client States.
*/
type StateI interface {
	Init(v interface{}) (state StateI, nextArgs interface{}, err error) // Init is when the state is created.
	CreateChannels()
	GetCloseChannel() chan bool
	Close() // Close is when the state is to be destroyed.
	SetClient(*Client)
	Enter(args ...interface{}) // Enter is when the state is entered via the next state being popped.
	Leave()                    // Leave is when the state is left via a new push, but is not replaced or popped.
	SetRunning(bool)
}

/*
State struct should be embedded in any Client State to provide the base struct methods and members. These can be overridden by the embedding state.

ex.:
  type MyState struct {
    Client.State
  }
  func (s *MyState Init(t interface{}) (next StateI, nextArgs interface{}, err error) {
    // .. custom init code
    return
  }
*/
type State struct {
	Client    *Client
	CloseChan chan bool
	Running   bool
}

// SetClient sets the state's Client pointer to the one provided.
func (s *State) SetClient(c *Client) {
	s.Client = c
}

func (s *State) SetRunning(b bool) {
	s.Running = b
}

// CreateChannels creates any channels needed by the State.
func (s *State) CreateChannels() {
	s.CloseChan = make(chan bool)
}

// GetCloseChannel returns the Close channel of the State.
func (s *State) GetCloseChannel() chan bool {
	return s.CloseChan
}

// Init is called to set up the State's initial... state.
func (s *State) Init(t interface{}) (next StateI, nextArgs interface{}, err error) {
	return
}

// Close cleans up the State.
func (s *State) Close() {
}

func (s *State) Enter(args ...interface{}) {
}

func (s *State) Leave() {
}
