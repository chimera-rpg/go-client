package states

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-common/network"
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

	// There should be some sort of way to detect secure vs. insecure hosts. Perhaps this would just be through the server list service? Otherwise, we could have an "info" port where server information is queried.
	err = s.Client.SecureConnectTo(server, &tls.Config{
		InsecureSkipVerify: true, // Skip verification for now. In the future this will be configurable.
	})
	if err != nil {
		s.Client.Log.Print(err)
		// For now, just fall back to attempting an insecure connection.
		s.Client.Log.Print("Falling back to insecure connection.")
		err = s.Client.ConnectTo(server)
		if err != nil {
			s.Client.Log.Print(err)
			state = client.StateI(&List{})
			nextArgs = err
			return
		}
	}

	select {
	case cmd := <-s.Client.CmdChan:
		switch cmd.(type) {
		case network.CommandHandshake:
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

	s.Client.Send(network.Command(network.CommandHandshake{
		Version: network.Version,
		Program: "Golang Client",
	}))

	cmd := <-s.Client.CmdChan
	switch t := cmd.(type) {
	case network.CommandBasic:
		if t.Type == network.Nokay {
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
