package main

import (
    "os"
    "path/filepath"

    "github.com/Tinch334/Token-File-Sharing/internal/log"
    "github.com/Tinch334/Token-File-Sharing/internal/server"
    "github.com/Tinch334/Token-File-Sharing/internal/constants"
)


func main() {
    // Create data directory if it doesn't exist.
    if err := os.MkdirAll(constants.DATA_FOLDER, os.ModePerm); err != nil {
        panic(err)
    }
    logFilepath := filepath.Join(constants.DATA_FOLDER, constants.LOG_FILE)
    credentialsFilepath := filepath.Join(constants.DATA_FOLDER, constants.CREDENTIALS_FILE)

    // Start logger.
    logFile := log.CreateLogger(logFilepath)
    defer logFile.Close()

    // Start server.
    sv := server.NewServer(":8000", credentialsFilepath)
    sv.StartServer()
}