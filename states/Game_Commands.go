package states

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/chimera-rpg/go-common/network"
)

// Print prints a local message.
func (s *Game) Print(str string) {
	s.HandleMessageCommand(network.CommandMessage{
		Type: network.LocalMessage,
		Body: str,
	})
}

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
	case "cmd":
		cmdMultiplier := regexp.MustCompile(`^([^*]*)[*]*\s*([0-9]*)`)
		if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
			s.HandleMessageCommand(network.CommandMessage{
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
							s.Print(fmt.Sprintf("couldn't parse number in %s", p))
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
		if s.bindings.HasFunction(cmd) {
			s.bindings.RunFunction(cmd, args)
		} else {
			s.Print(fmt.Sprintf("unknown command \"%s\"", cmd))
		}
	}
}
