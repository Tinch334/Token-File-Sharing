package main

import (
    "os"

    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"

    "github.com/Tinch334/Token-File-Sharing/server"
)


func main() {
    logFile := createLogger("tfs.log")
    defer logFile.Close()

    sv := server.NewServer(":8000")
    sv.StartServer()
}


// createLogger creates a zerolog logger for the program, returns the pointer to the log file.
func createLogger(logPath string) *os.File {
    // Open log file.
    logFile, err := os.OpenFile(logPath, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0664)

    if err != nil {
        panic(err)
    }

    // Configure pretty zerolog writer.
    prettyWriter := zerolog.ConsoleWriter{
        Out:        logFile,
        NoColor:    true,
        // Chosen for readability, not the fastest, could be replaced with "TimeFormatUnix" if necessary.
        TimeFormat: "2006-01-02 15:04:05",//zerolog.IntegerTimeFieldFormat,
    }

    log.Logger = zerolog.New(prettyWriter).With().Timestamp().Logger()
    log.Info().Msg("Server logging started")

    return logFile
}