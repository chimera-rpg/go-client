package States

import (
  "github.com/chimera-rpg/go-client/Client"
  "github.com/chimera-rpg/go-common/Net"
)

type CharacterCreation struct {
  Client.State
}

func (s *CharacterCreation) Init(t interface{}) (next Client.StateI, nextArgs interface{}, err error) {
  s.Client.Log.Print("CharacterCreation State")

  for {
    cmd := <- s.Client.CmdChan
    switch t := cmd.(type) {
    case Net.CommandBasic:
      if t.Type == Net.REJECT {
        s.Client.Log.Printf("Server rejected us: %s\n", t.String)
      } else if t.Type == Net.OK {
        s.Client.Log.Printf("Server accepted us: %s\n", t.String)
        break
      }
    default:
      s.Client.Log.Print("Server sent non CommandBasic")
      next = Client.StateI(&List{})
      return
    }
  }

  next = Client.StateI(&Game{})
  return
}
