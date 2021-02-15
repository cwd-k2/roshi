package confirm

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
)

var stdin = bufio.NewReader(os.Stdin)

func Confirm(msg string) bool {
	fmt.Fprintf(os.Stdout, "%s [y/n] ", msg)

	resp, err := stdin.ReadString('\n')
	if err != nil {
		log.Fatalf("%+v", errors.WithStack(err))
	}

	switch strings.TrimSpace(resp) {
	case "y", "yes", "Y", "Yes", "YES":
		return true
	case "n", "no", "N", "No", "NO":
		return false
	default:
		return Confirm(msg)
	}
}
