package client_cli

import (
    "os"
    "strings"
    "text/tabwriter"
    "io"
    "fmt"

    "github.com/chzyer/readline"


)


// A generic type for command descriptions.
type command struct {
    description string
    subcommands []string
    cr bool // Indicates whether the command requires being connected to the server
}
// Map order is non deterministic, a list is used to keep them ordered.
var commandOrder = []string{
    "help", "about", "connect", "exit", "disconnect", "where",
}
// Help information for commands.
var commands = map[string]command{
    "help":  {description: "Show this help message", cr: false},
    "about": {description: "Program information", cr: false},
    "connect": {description: "Connect to the specified server", cr: false},
    "exit": {description: "Exit the client", cr: false},

    "disconnect": {description: "Disconnect from the connected server", cr: true},
    "where": {description: "Shows the address of the connected server", cr: true},
}


// Completer for readline.
var completer = readline.NewPrefixCompleter(
    readline.PcItem("help"),
    readline.PcItem("about"),
    readline.PcItem("connect"),
    readline.PcItem("exit"),
    readline.PcItem("disconnect"),
    readline.PcItem("where"),
)


func Client() {
	// Start readline CLI
    rl, err := readline.NewEx(&readline.Config{
        Prompt:          "> ",
        HistoryFile:     "/tmp/tfs-client-history.tmp",
        AutoComplete:    completer,
        InterruptPrompt: "^C",
        EOFPrompt:       "exit",
    })
    if err != nil {
        panic(err)
    }
    defer rl.Close()

    for {
        // Read input from readline.
        line, err := rl.Readline()
        if err == readline.ErrInterrupt || err == io.EOF {
            return
        }
        if err != nil {
            panic(err)
        }

        text := strings.Fields(line)

        // No input.
        if len(text) == 0 {
            continue
        }

        switch text[0] {
        // System.
        case "help":
            help()

        case "about":
            fmt.Printf("Token File Sharing (TFS) - A token based file sharing system\nMade by: Martín Goñi\nVersion: 0.5\n")

        case "connect":
        	success("Connected successfully to server on: %d", 1234)

		case "exit":
			/*
				Check for connection before disconnecting.
			*/
			success("Bye")
			return

		case "disconnect":
		case "where":

       
        }
    }
}


// help prints the help information.
func help() {
    w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
    defer w.Flush()

    // Print all command help.
    fmt.Fprintln(w, "Available commands:")
    for _, name := range commandOrder {
        cmd := commands[name]
        fmt.Fprintf(w, "  %s\t%s\n", name, cmd.description)
        for _, sub := range cmd.subcommands {
            fmt.Fprintf(w, "  \t  %s\n", sub)
        }
    }
}