package client_cli


import (
    "os"
    "text/tabwriter"
    "fmt"
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
    "conn": {
        description: "Handle connection",
        cr: false,
        subcommands: []string {
            "cn <address> <username> <password> - Connects to the server in the given address using the specified username and password",
            "dc - Disconnect from the currently connected server, will invalidate session token",
            "st - Shows connection status",
        },
    },
    "exit": {description: "Exit the client", cr: false},

    "where": {description: "Shows the address of the connected server", cr: true},
}


// helpHandler prints the help information.
func (_ *status) helpHandler(_ []string) {
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


// aboutHandler prints information about the client.
func (_ *status) aboutHandler(_ []string) {
    fmt.Printf("Client for Token File Sharing (TFS)\nMade by: Martín Goñi\nVersion: 0.5\n")
}