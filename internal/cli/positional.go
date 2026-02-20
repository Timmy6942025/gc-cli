package cli

import (
	"fmt"
	"strings"
)

func fillPositional(args []string, values ...*string) {
	for i := 0; i < len(args) && i < len(values); i++ {
		if strings.TrimSpace(*values[i]) != "" {
			continue
		}
		*values[i] = strings.TrimSpace(args[i])
	}
}

func requireValue(value string, flagName string, argName string) error {
	if strings.TrimSpace(value) != "" {
		return nil
	}
	return fmt.Errorf("missing %s: provide --%s or positional <%s>", argName, flagName, argName)
}
