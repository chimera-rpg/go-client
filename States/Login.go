package States

import (
  "client/Client"
)

type Login struct {
  Client.State
}

func (s *Login) Init(t interface{}) (next Client.StateI, nextArgs interface{}, err error) {
  s.Client.Log.Print("Login State")
  next = Client.StateI(&Game{})
  return
}
