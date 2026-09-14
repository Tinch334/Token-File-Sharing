package client_cli

import (
	"fmt"

	"github.com/fatih/color"
)


var green = color.New(color.FgGreen).SprintFunc()

// Success prints the given message preceeded by a green "[OK]".
func success(format string, args ...any) {
	fmt.Println(green("[OK]"), fmt.Sprintf(format, args...))
}