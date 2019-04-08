package States

import (
  "github.com/chimera-rpg/go-client/Client"
  "github.com/chimera-rpg/go-client/UI"
  "github.com/chimera-rpg/go-common/Net"
)

type Handshake struct {
  Client.State
  ServersWindow UI.Window
}

func (s *Handshake) Init(v interface{}) (state Client.StateI, nextArgs interface{}, err error) {
  //var cmd Net.Command
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
    s.Client.Log.Print("Bad server value passed to Handshake State")
    state = Client.StateI(&List{})
    return
  }

  err = s.Client.ConnectTo(server)
  if err != nil {
    s.Client.Log.Print(err)
    state = Client.StateI(&List{})
    return
  }

  cmd := <- s.Client.CmdChan
  switch cmd.(type) {
  case Net.CommandHandshake:
  default:
    s.Client.Log.Print("Server sent non-handshake")
    state = Client.StateI(&List{})
    return
  }

  s.Client.Send(Net.Command(Net.CommandHandshake{
    Version: Net.VERSION,
    Program: "Golang Client",
  }))

  cmd = <- s.Client.CmdChan
  switch t := cmd.(type) {
  case Net.CommandBasic:
    if t.Type == Net.NOK {
      s.Client.Log.Printf("Server rejected us: %s\n", t.String)
      state = Client.StateI(&List{})
      return
    }
  default:
    s.Client.Log.Print("Server sent non CommandBasic")
    state = Client.StateI(&List{})
    return
  }

  state = Client.StateI(&Login{})

  return
}
