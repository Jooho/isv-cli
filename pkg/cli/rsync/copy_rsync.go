package rsync

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
)

// DefaultRsyncRemoteShellToUse is customised for ISV. it set rshCmd to use oc as a root cmd. 
func DefaultRsyncRemoteShellToUse(cmd *cobra.Command) string {
	
	rshCmd := []string{"oc", "rsh"}
	// do not add local flags, unless also rsh flags to the command
	localFlags := sets.NewString()
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		localFlags.Insert(flag.Name)
	})
	// flag.Name represents what was present on the CLI, so the excluded list needs
	// to have both short and long versions of flags
	excludeFlags := localFlags.Difference(sets.NewString("container", "c", "no-tty", "T", "shell", "timeout", "tty", "t"))
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		if excludeFlags.Has(flag.Name) {
			return
		}
		if flag.Name == flag.Shorthand {
			rshCmd = append(rshCmd, fmt.Sprintf("-%s=%s", flag.Name, flag.Value.String()))
		} else {
			rshCmd = append(rshCmd, fmt.Sprintf("--%s=%s", flag.Name, flag.Value.String()))
		}
	})
	return strings.Join(rsyncEscapeCommand(rshCmd), " ")
}


func rsyncEscapeCommand(command []string) []string {
	var escapedCommand []string
	for _, val := range command {
		needsQuoted := strings.ContainsAny(val, `'" `)
		if needsQuoted {
			val = strings.Replace(val, `"`, `""`, -1)
			val = `"` + val + `"`
		}
		escapedCommand = append(escapedCommand, val)
	}
	return escapedCommand
}


