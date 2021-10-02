package states

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/chimera-rpg/go-common/network"
)

func (s *Game) isChatCommand(c string) bool {
	if strings.HasPrefix(c, s.Client.DataManager.Config.Game.CommandPrefix) {
		return true
	}
	return false
}

func (s *Game) processChatCommand(c string) {
	parts := strings.SplitAfterN(c[1:], " ", 2)
	parts[0] = strings.TrimSpace(strings.TrimPrefix(parts[0], s.Client.DataManager.Config.Game.CommandPrefix))
	s.handleChatCommand(parts[0], parts[1:]...)
}

func (s *Game) handleChatCommand(cmd string, args ...string) {
	switch cmd {
	case "quit":
		os.Exit(0)
	case "disconnect":
		s.inputChan <- DisconnectEvent{}
	case "say":
		s.Client.Send(network.CommandMessage{
			Type: network.PCMessage,
			Body: strings.Join(args, " "),
		})
	case "cmd":
		cmdMultiplier := regexp.MustCompile(`^([^*]*)[*]*\s*([0-9]*)`)
		if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
			s.Client.HandleMessageCommand(network.CommandMessage{
				Type: network.LocalMessage,
				Body: fmt.Sprintf("missing command for \"cmd\""),
			})
			s.UpdateMessagesWindow()
		} else if len(args) >= 2 {
			s.bindings.RunFunction(args[0], args[1:])
		} else if len(args) >= 1 {
			if strings.HasPrefix(args[0], "[") && strings.HasSuffix(args[0], "]") {
				a := strings.TrimSuffix(strings.TrimPrefix(args[0], "["), "]")
				parts := strings.Split(a, ",")
				for _, p := range parts {
					results := cmdMultiplier.FindStringSubmatch(p)
					cmdString := strings.TrimSpace(results[1])
					if results[2] == "" {
						s.bindings.RunFunction(cmdString)
					} else {
						i, err := strconv.Atoi(results[2])
						if err != nil {
							s.Client.HandleMessageCommand(network.CommandMessage{
								Type: network.LocalMessage,
								Body: fmt.Sprintf("couldn't parse number in %s", p),
							})
							s.UpdateMessagesWindow()
						} else {
							for j := 0; j < i; j++ {
								s.bindings.RunFunction(cmdString)
							}
						}
					}
				}
			} else {
				s.bindings.RunFunction(args[0])
			}
		}
	default:
		s.Client.HandleMessageCommand(network.CommandMessage{
			Type: network.LocalMessage,
			Body: fmt.Sprintf("unknown command \"%s\"", cmd),
		})
		s.UpdateMessagesWindow()
	}
}
