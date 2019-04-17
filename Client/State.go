package Client

import (
	"github.com/veandco/go-sdl2/sdl"
)

/*
~~~~ States In Brief ~~~~
Chimera's go client primarily works via a state machine. New states are passed into the state machine by either returning the next state during `Init()` or by sending a StateI through the Client's StateChannel.

The functions of a state are defined through the State interface known as StateI. Although a Client state may implement each function of the interface, it may be more convenient to simply embed the State type, which provides the base functions and methods that all states must implement.
*/

/*
StateI provides the base interface for all Client States.
*/
type StateI interface {
	Init(v interface{}) (state StateI, nextArgs interface{}, err error)
	Close()
	CommandLoop()
	SetClient(*Client)
	HandleRender()
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
	Client *Client
}

func (s *State) SetClient(c *Client) {
	s.Client = c
}

func (s *State) Init(t interface{}) (next StateI, nextArgs interface{}, err error) {
	return
}

func (s *State) Close() {
}

func (s *State) HandleRender() {
}

func (s *State) HandleNet() {
}

func (s *State) CommandLoop() {
}

func (s *State) HandleEvent(e *sdl.Event) {
}
