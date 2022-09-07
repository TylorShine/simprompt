package simprompt

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type SimPromptCmd struct {
	Command  string
	Help     string
	Callback func([]string) bool // Callback(args) => isExit
}

type SimPrompt struct {
	Prompt string
	Cmds   map[string]SimPromptCmd
}

func NewSimPrompt() *SimPrompt {
	return &SimPrompt{
		Prompt: "> ",
		Cmds:   map[string]SimPromptCmd{},
	}
}

func (sp *SimPrompt) AppendCmd(cmd []string, help string, cb func([]string) bool) error {
	for _, v := range cmd {
		if _, ok := sp.Cmds[v]; ok {
			return errors.New("command \"" + v + "\" already defined")
		}
	}

	if cb == nil {
		return errors.New("no callback is not allowed")
	}

	for _, v := range cmd {
		sp.Cmds[v] = SimPromptCmd{
			Command:  v,
			Help:     help,
			Callback: cb,
		}
	}
	return nil
}

func (sp *SimPrompt) SetPrompt(s string) {
	sp.Prompt = s
}

func (sp *SimPrompt) SetCmds(cmds map[string]SimPromptCmd) {
	sp.Cmds = cmds
}

func (sp *SimPrompt) parseCommand(line string) (command string, args []string) {
	lineRune := ([]rune)(line)
	if len(lineRune) <= 0 {
		return "", nil
	}
	if lineRune[0] == '/' && len(lineRune) > 1 {
		lineRune = lineRune[1:]
	}
	commSplit := strings.Split(string(lineRune), " ")
	if len(commSplit) > 1 {
		return commSplit[0], commSplit[1:]
	}
	return commSplit[0], nil
}

func (sp *SimPrompt) Run(scan *os.File) chan bool {
	if scan == nil {
		scan = os.Stdin
	}
	scanner := bufio.NewScanner(scan)

	endChan := make(chan bool)

	textChan := make(chan string)

	go func() {
		for scanner.Scan() {
			textChan <- scanner.Text()
		}
	}()

	quitChan := make(chan os.Signal)

	signal.Notify(quitChan, os.Interrupt)
	signal.Notify(quitChan, syscall.SIGTERM)

	go func() {
		for {
			fmt.Print("\r" + sp.Prompt)
			select {
			case text, ok := <-textChan:
				if !ok {
					close(endChan)
					return
				}
				comm, commArgs := sp.parseCommand(text)
				commLower := strings.ToLower(comm)
				if v, ok := sp.Cmds[commLower]; ok {
					if len(commArgs) > 0 {
						lastArg := strings.TrimLeft(commArgs[len(commArgs)-1], "-/")
						switch lastArg {
						case "help", "Help", "h", "H":
							fmt.Println("Command", comm, "Help:")
							fmt.Println(sp.Cmds[commLower].Help)
							continue
						}
					}

					if !v.Callback(commArgs) {
						close(endChan)
						return
					}
				}
			case _, ok := <-endChan:
				if !ok {
					return
				}
			case <-quitChan:
				close(endChan)
				return
			}
		}
	}()

	return endChan
}