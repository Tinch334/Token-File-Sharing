package client_cli

import (
    "strings"
    "io"

    "github.com/chzyer/readline"
)


// Completer for readline.
var completer = readline.NewPrefixCompleter(
    readline.PcItem("help"),
    readline.PcItem("about"),
    readline.PcItem("connect"),
    readline.PcItem("exit"),
    readline.PcItem("disconnect"),
    readline.PcItem("where"),
)


// Uses method expressions to pass a reference to the server to handlers whilst keeping the server definition lean.
var handlerMap = map[string]TypeHandler{
    "help": (*status).helpHandler,
    "about": (*status).aboutHandler,
    "conn": (*status).connectionHandler,
    "exit": (*status).exitHandler,
}


func Client() {
    // Get initial client status.
    st := newStatus()

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

        // Dispatch to appropriate handler.
        handler, ok := handlerMap[text[0]]
        if !ok {
            warn("Unknown command: `%s`\n", text[0])
            continue
        }

        // Call handler with given arguments.
        handler(st, text[1:])
    }
}