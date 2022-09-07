# simprompt - A Simple Prompt
simprompt is a minimum of minimum prompt implement for me, written in Go.  
  
## Usage  
`go get github.com/TylorShine/simprompt`  
  
See `cmd/example.go`:  
```Go
package main

import (
	"fmt"
	"strings"

	"github.com/TylorShine/simprompt"
)

func main() {
	// make the simprompt
	sp := simprompt.NewSimPrompt()

	// append commands
	sp.AppendCmd([]string{"t", "test"}, "Test Command", func(s []string) bool {
		fmt.Println("Argn:", len(s), ",Argv:", s, "\n...was given.")
		return true
	})

	sp.AppendCmd([]string{"p", "prompt"}, "Set Prompt", func(s []string) bool {
		if len(s) > 0 {
			sp.Prompt = strings.Join(s, "/") + "> "
			// or, simprompt.SetPrompt()
			// sp.SetPrompt(strings.Join(s, "/") + "> ")
		}
		return true
	})

	sp.AppendCmd([]string{"q", "quit", "exit"}, "Quit Program", func(s []string) bool {
		fmt.Println("Quit, bye!")
		return false
	})

	// nil given, read from os.Stdin
	done := sp.Run(nil)
	<-done
}
```
