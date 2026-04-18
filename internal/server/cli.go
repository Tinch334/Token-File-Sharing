package server


import (
    "os"
    "time"
    "strings"
    "text/tabwriter"
    "io"
    "fmt"
    "strconv"
    "context"

    "github.com/chzyer/readline"
    "github.com/rs/zerolog/log"

    "github.com/Tinch334/Token-File-Sharing/internal/credentials"
)


// A generic type for command handlers, takes a reference to the server and the commands arguments, if any.
type command struct {
    description string
    subcommands []string
}
// Map order is non deterministic.
var commandOrder = []string{
    "help", "about", "tokens", "user", "uptime", "conn", "shutdown",
}
// Help information for commands.
var commands = map[string]command{
    "help":  {description: "Show this help message"},
    "about": {description: "Program information"},
    "tokens": {description: "Show token information"},
    "user": {
        description: "Manage users",
        subcommands: []string{
            "new <username> <password> <permissions> - Creates a new user",
            "update (password|permissions) <new value> - Updates the corresponding user field",
            "delete <username> - Deletes the user",
            "list - Lists all users",
        },
    },
    "uptime":   {description: "Show server uptime"},
    "conn":     {description: "Show connection information"},
    "shutdown": {description: "Shut down the server immediately"},
}


// Completer for readline.
var completer = readline.NewPrefixCompleter(
    readline.PcItem("help"),
    readline.PcItem("about"),
    readline.PcItem("tokens"),
    readline.PcItem("user",
        readline.PcItem("new"),
        readline.PcItem("update",
            readline.PcItem("password"),
            readline.PcItem("permissions"),
        ),
        readline.PcItem("delete"),
        readline.PcItem("list"),
    ),
    readline.PcItem("uptime"),
    readline.PcItem("conn"),
    readline.PcItem("shutdown"),
)


// cli starts an interactive cli.
func (s *Server) cli(ctx context.Context) {
    // Start readline CLI
    rl, err := readline.NewEx(&readline.Config{
        Prompt:          "> ",
        HistoryFile:     "/tmp/tfs-history.tmp",
        AutoComplete:    completer,
        InterruptPrompt: "^C",
        EOFPrompt:       "exit",
    })
    if err != nil {
        log.Fatal().
                Err(err).
                Msg("Failed to initialize CLI")
        return
    }
    defer rl.Close()

    for {
        // Read input from readline.
        line, err := rl.Readline()
        if err == readline.ErrInterrupt || err == io.EOF {
            return
        }
        if err != nil {
            log.Fatal().
                Err(err).
                Msg("CLI input error")
            return
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
            fmt.Printf("Token File Sharing (TFS) - A token based file sharing system\nMade by: Martín Goñi\nVersion: 0.4\n")

        // General info.
        case "tokens":
            fmt.Printf("Token length: %d\nToken count: %d\nGenerated tokens: %d\n",
                s.th.TokenLength(), s.th.TokenCount(), s.mtr.GetTokens())

        // User handling.
        case "user":
            user(s, ctx, text[1:])

        // Metrics.
        case "uptime":
            sTime := s.mtr.GetStartTime()
            fmt.Printf("Server started on: %s\nUptime: %s\n",
                sTime.Format("2006-01-02 15:04:05"),
                time.Since(sTime).Truncate(time.Second).String(),
            )
        case "conn":
            fmt.Printf("Total connections: %d\n", s.mtr.GetConn())

        case "shutdown":
            return
        default:
            fmt.Printf("Unknown command\n\n")
            help()
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


// user modifies and displays user information.
func user(s *Server, ctx context.Context, subcommand []string) {
    formatErr := "ERROR: Invalid format for user command"

    if len(subcommand) == 0 {
        fmt.Println(formatErr)
        return
    }

    switch subcommand[0] {
    case "new":
        if len(subcommand) != 4 {
            fmt.Println(formatErr)
            return
        }

        // Check if given permission fits in a uint8.
        perm, err := strconv.ParseUint(subcommand[3], 10, 8)
        if err != nil {
            fmt.Println("ERROR: Invalid permission value")
            return
        }

        // Add user.
        err = s.credDB.AddUser(ctx, subcommand[1], subcommand[2], uint8(perm))
        if err != nil {
            if err == credentials.ErrUserAlreadyExists {
                fmt.Printf("ERROR: User `%s` already exists\n", subcommand[1])
            } else {
                fmt.Printf("ERROR: User could not be created: %v\n", err)
            }
            return
        }

        fmt.Printf("User `%s` created successfully\n", subcommand[1])

    case "update":
        if len(subcommand) != 4 {
            fmt.Println(formatErr)
            return
        }

        switch subcommand[2] {
        case "password":
            err := s.credDB.UpdateUserPassword(ctx, subcommand[1], subcommand[3])
            if err != nil {
                if err == credentials.ErrUserNotExists {
                    fmt.Printf("ERROR: User `%s` does not exist, password cannot be updated\n", subcommand[1])
                } else {
                    fmt.Printf("ERROR: User could not be updated: %v\n", err)
                }
                return
            }

        case "permissions":
            // Check if given permission fits in a uint8.
            perm, err := strconv.ParseUint(subcommand[3], 10, 8)
            if err != nil {
                fmt.Println("ERROR: Invalid permission value")
                return
            }

            err = s.credDB.UpdateUserPermissions(ctx, subcommand[1], uint8(perm))
            if err != nil {
                if err == credentials.ErrUserNotExists {
                    fmt.Printf("ERROR: User `%s` does not exist, permissions cannot be updated\n", subcommand[1])
                } else {
                    fmt.Printf("ERROR: User could not be updated: %v\n", err)
                }
                return
            }

        default:
            fmt.Println(formatErr)
        }

    case "delete":
        if len(subcommand) != 2 {
            fmt.Println(formatErr)
            return
        }

        // Delete user.
        err := s.credDB.DeleteUser(ctx, subcommand[1])
        if err != nil {
            if err == credentials.ErrUserNotExists {
                fmt.Printf("ERROR: User `%s` does not exist, cannot be deleted\n", subcommand[1])
            } else {
                fmt.Printf("ERROR: User could not be deleted: %v\n", err)
            }

            return
        }

        fmt.Printf("User `%s` deleted successfully\n", subcommand[1])

    case "list":
        if len(subcommand) != 1 {
            fmt.Println(formatErr)
            return
        }

        // Get users.
        userList, err := s.credDB.GetUsers(ctx)
        if err != nil {
            fmt.Printf("ERROR: Could not get user list: %v\n", err)
            return
        }

        // Start tabwriter to print users properly.
        w := tabwriter.NewWriter(os.Stdout, 4, 0, 2, ' ', 0)
        defer w.Flush()

        fmt.Fprintf(w, "Username\tPermissions\tLast log\n")
        for _, user := range userList {
            var timeStr string
            if user.LastLog.IsZero() {
                timeStr = "<No log>"
            } else {
                timeStr = user.LastLog.Format("2006-01-02 15:04:05")
            }

            fmt.Fprintf(w, "%s\t%d\t%s\n", user.Name, user.Perms, timeStr)
        } 

    default:
        fmt.Printf("ERROR: Unknown user command")
    }
}