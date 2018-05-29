package kubectl

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func kubectl(args ...string) error {
	var cmd = exec.Command("kubectl", args...)

	log.Printf("kubectl %v...", strings.Join(args, " "))

	if out, err := cmd.Output(); err == nil {
		log.Printf("kubectl %v: %v", strings.Join(args, " "), string(out))

		return nil
	} else if exitErr, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("kubectl %v: %v: %v", strings.Join(args, " "), exitErr.String(), string(exitErr.Stderr))
	} else {
		return err
	}
}
