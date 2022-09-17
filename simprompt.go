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
	Command  []string
	Help     string
	Callback func([]string) bool // Callback(args) => isExit
	CmdIndex int
}

type SimPrompt struct {
	Prompt          string
	Cmds            map[string]SimPromptCmd
	DefaultCallback func([]string) bool
	CmdNum          int
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
			Command:  cmd,
			Help:     help,
			Callback: cb,
			CmdIndex: sp.CmdNum,
		}
	}
	sp.CmdNum++
	return nil
}

func (sp *SimPrompt) SetDefaultCallback(cb func([]string) bool) {
	sp.DefaultCallback = cb
}

func (sp *SimPrompt) SetPrompt(s string) {
	sp.Prompt = s
}

func (sp *SimPrompt) SetCmds(cmds map[string]SimPromptCmd) {
	sp.Cmds = cmds
}

func (sp *SimPrompt) SetHelp(cmd, help string) error {
	if _, ok := sp.Cmds[cmd]; !ok {
		return errors.New("command \"" + cmd + "\" was not found")
	} else {
		c := sp.Cmds[cmd]
		c.Help = help
		sp.Cmds[cmd] = c
	}

	return nil
}

func (sp *SimPrompt) GetHelp(cmd string) string {
	if v, ok := sp.Cmds[cmd]; ok {
		return v.Help
	}
	return ""
}

func (sp *SimPrompt) GetHelpAll() (ret []string) {
	ret = make([]string, sp.CmdNum)
	idx := 0
	m := map[int]struct{}{}
	for _, v := range sp.Cmds {
		if _, ok := m[v.CmdIndex]; ok {
			continue
		}
		ret[idx] = fmt.Sprint(v.Command, ":", v.Help)
		m[v.CmdIndex] = struct{}{}
	}
	return
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

func (sp *SimPrompt) getPipe() (r, w *os.File) {
	r, w, _ = os.Pipe()
	return
}

func (sp *SimPrompt) Run(scan *os.File) chan bool {
	if scan == nil {
		scan = os.Stdin
	}
	// scanner := bufio.NewScanner(scan)
	scanner := bufio.NewReader(scan)

	endChan := make(chan bool)

	textChan := make(chan string)

	readfunc := func() {
		if s, err := scanner.ReadString(byte('\n')); err == nil {
			textChan <- string(strings.TrimRight(s, "\r\n"))
		}
	}

	// go func() {
	// 	for scanner.Scan() {
	// 		textChan <- scanner.Text()
	// 	}
	// }()

	quitChan := make(chan os.Signal)

	signal.Notify(quitChan, os.Interrupt)
	signal.Notify(quitChan, syscall.SIGTERM)

	go func() {
		for {
			fmt.Print("\r" + sp.Prompt)
			go readfunc()
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
				} else if sp.DefaultCallback != nil {
					var ret bool
					if commArgs != nil {
						ret = sp.DefaultCallback(append([]string{comm}, commArgs...))
					} else {
						ret = sp.DefaultCallback([]string{comm})
					}
					if !ret {
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
