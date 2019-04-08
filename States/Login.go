package States

import (
  "github.com/chimera-rpg/go-client/Client"
)

type Login struct {
  Client.State
}

func (s *Login) Init(t interface{}) (next Client.StateI, nextArgs interface{}, err error) {
  s.Client.Log.Print("Login State")
  next = Client.StateI(&Game{})
  return
}
