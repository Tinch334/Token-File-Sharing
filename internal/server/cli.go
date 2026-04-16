package server


import (
    "fmt"
    "bufio"
    "os"
    "time"
)


// cli starts an interactive cli.
func (s *Server) cli() {
    scanner := bufio.NewScanner(os.Stdin)

    for {
        fmt.Printf("> ")

        // Read input.
        scanner.Scan()
        text := scanner.Text()

        // No input.
        if len(text) == 0 {
            continue
        }

        switch text {
        // System.
        case "help":
            help()

        case "about":
            fmt.Printf("Token File Sharing (TFS) - A token based file sharing system\nMade by: Martín Goñi\nVersion: 0.4\n")

        // General info.
        case "tokens":
            fmt.Printf("Token length: %d\nToken count: %d\nGenerated tokens: %d\n",
                s.th.TokenLength(), s.th.TokenCount(), s.mtr.GetTokens())

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
func help () {
    helpInfo := map[string]string{
        "help" : "Show this help message",
        "about" : "Program information",
        "tokens" : "Show token information",
        "uptime" : "Shows server uptime",
        "conn" : "Shows connection information",
        "shutdown" : "Shuts down the server, instantly",
    }

    fmt.Printf("Available commands:\n")
    for c, d := range helpInfo {
        fmt.Printf("%8s - %s\n", c, d)
    }
}