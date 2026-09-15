package client_cli

import (
    "fmt"

    "github.com/fatih/color"
)


var green = color.New(color.FgGreen).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()
var grey = color.RGB(140, 140, 140).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()


// success prints the given message preceded by a green "[OK]".
func success(format string, args ...any) {
    fmt.Println(green("[OK]"), fmt.Sprintf(format, args...))
}

// warn prints the given message preceded by a green "[Warn]".
func warn(format string, args ...any) {
    fmt.Println(yellow("[Warn]"), fmt.Sprintf(format, args...))
}

// info prints the given message preceded by a grey "[Info]".
func info(format string, args ...any) {
    fmt.Println(grey("[Info]"), fmt.Sprintf(format, args...))
}

// err prints the given message preceded by a grey "[Error]".
func errorPrint(format string, args ...any) {
    fmt.Println(red("[Error]"), fmt.Sprintf(format, args...))
}