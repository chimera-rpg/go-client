package States

import (
	"github.com/chimera-rpg/go-client/Client"
	"github.com/chimera-rpg/go-common/Net"
)

type Login struct {
	Client.State
}

func (s *Login) Init(t interface{}) (next Client.StateI, nextArgs interface{}, err error) {
	s.Client.Log.Print("Login State")
	// Show UI for Username/Password input:
	//   * Main: Server Info Panel, Username, Password, Login, Register
	//     * Register: Enter Password Again, E-Mail(optional field)
	//       * User exists! (go back to Main)
	//       * Registered! (go back to Main w/ Login prefilled)
	//     * Login
	//       * Success! (go to Character Selection/Creation State)
	//       * Bad password/username! (go back to Main)
	s.Client.Send(Net.Command(Net.CommandLogin{
		Type: Net.LOGIN,
		User: "nommak",
		Pass: "nommak",
	}))

	isWaiting := true

	for isWaiting {
		cmd := <-s.Client.CmdChan
		switch t := cmd.(type) {
		case Net.CommandBasic:
			s.Client.Log.Print("Got basic")
			if t.Type == Net.REJECT {
				s.Client.Log.Printf("Server rejected us: %s\n", t.String)
			} else if t.Type == Net.OK {
				s.Client.Log.Printf("Server accepted us: %s\n", t.String)
				isWaiting = false
			}
		default:
			s.Client.Log.Print("Server sent non CommandBasic")
			next = Client.StateI(&List{})
			return
		}
	}

	next = Client.StateI(&CharacterCreation{})

	return
}
