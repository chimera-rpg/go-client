package states

import (
	"fmt"
	"time"

	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-common/Net"
)

// Handshake is the state responsible for the initial handshake with a server,
// ensuring that versions match and the server likes us in general.
type Handshake struct {
	client.State
	ServersWindow ui.Window
}

// Init Handshake
func (s *Handshake) Init(v interface{}) (state client.StateI, nextArgs interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			s.Client.Log.Print("Communication problematic with server, d/cing")
			s.Client.Close()
			err = r.(error)
		}
		return
	}()
	server, ok := v.(string)
	if ok == false {
		msg := fmt.Sprintf("Bad server value.")
		s.Client.Log.Print(msg)
		state = client.StateI(&List{})
		nextArgs = msg
		return
	}

	err = s.Client.ConnectTo(server)
	if err != nil {
		s.Client.Log.Print(err)
		state = client.StateI(&List{})
		nextArgs = err
		return
	}

	select {
	case cmd := <-s.Client.CmdChan:
		switch cmd.(type) {
		case Net.CommandHandshake:
		default:
			msg := fmt.Sprintf("Server \"%s\" sent non-handshake..", server)
			s.Client.Log.Print(msg)
			state = client.StateI(&List{})
			nextArgs = msg
			return
		}
	case <-time.After(2 * time.Second):
		msg := fmt.Sprintf("Server \"%s\" took too long to respond.", server)
		s.Client.Log.Printf(msg)
		state = client.StateI(&List{})
		nextArgs = msg
		return
	}

	s.Client.Send(Net.Command(Net.CommandHandshake{
		Version: Net.VERSION,
		Program: "Golang Client",
	}))

	cmd := <-s.Client.CmdChan
	switch t := cmd.(type) {
	case Net.CommandBasic:
		if t.Type == Net.NOK {
			msg := fmt.Sprintf("Server \"%s\" rejected us: %s", server, t.String)
			s.Client.Log.Printf(msg)
			state = client.StateI(&List{})
			nextArgs = msg
			return
		}
	default:
		msg := fmt.Sprintf("Server \"%s\" sent non CommandBasic.", server)
		s.Client.Log.Print(msg)
		state = client.StateI(&List{})
		nextArgs = msg
		return
	}

	state = client.StateI(&Login{})

	return
}
