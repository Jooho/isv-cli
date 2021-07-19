/*
	This ISV cli is based on oc cli and it overrides some specific commands.

	As of 2021.01.20, these commands are overrided:
	- must-gather

	Wrappered:
	- login
	- logout
*/

package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/openshift/oc/pkg/cli/options"
	cmdutil "github.com/openshift/oc/pkg/helpers/cmd"
	"github.com/openshift/oc/pkg/helpers/term"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	kubecmd "k8s.io/kubectl/pkg/cmd"
	"k8s.io/kubectl/pkg/cmd/plugin"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/jooho/isv-cli/pkg/cli/mustgather"
	"github.com/jooho/isv-cli/pkg/cli/ocwrappers"
	"github.com/jooho/isv-cli/pkg/cli/testharness"
)

const (
	productName = `ISV`
	cliVersion = "0.2.0"
)


var (
	cliLong = heredoc.Doc(`
	` + productName + ` Client
	This client helps you regarding OpenShift Managed Serviced Operator.`)
			
	cliExplain = heredoc.Doc(`
	This isv-cli is for ISV related features such as must-gather, test harness.
			 `)
			)
			
			// CommandFor returns the appropriate command for this base name,
			// or the ISV CLI command.
			func CommandFor(basename string) *cobra.Command {
				var cmd *cobra.Command
				
	in, out, errout := os.Stdin, os.Stdout, os.Stderr
	
	// Make case-insensitive and strip executable suffix if present
	if runtime.GOOS == "windows" {
		basename = strings.ToLower(basename)
		basename = strings.TrimSuffix(basename, ".exe")
	}
	
	cmd = NewDefaultIsvCommand(in, out, errout)

	// // Treat oc as a kubectl plugin (it is not being used now but for future plans, it inherited)
	// if strings.HasPrefix(basename, "kubectl-") {
		// 	args := strings.Split(strings.TrimPrefix(basename, "kubectl-"), "-")
		
		// 	// The plugin mechanism interprets "_" as dashes. Convert any "_" our basename
		// 	// might have in order to find the appropriate command in the `oc` tree.
		// 	for i := range args {
			// 		args[i] = strings.Replace(args[i], "_", "-", -1)
			// 	}

			// 	if targetCmd, _, err := cmd.Find(args); targetCmd != nil && err == nil {
				// 		// since cobra refuses to execute a child command, executing its root
	// 		// any time Execute() is called, we must create a completely new command
	// 		// and "deep copy" the targetCmd information to it.
	// 		newParent := &cobra.Command{
		// 			Use:     targetCmd.Use,
	// 			Short:   targetCmd.Short,
	// 			Long:    targetCmd.Long,
	// 			Example: targetCmd.Example,
	// 			Run:     targetCmd.Run,
	// 		}

	// 		// copy flags
	// 		newParent.Flags().AddFlagSet(cmd.Flags())
	// 		newParent.Flags().AddFlagSet(targetCmd.Flags())
	// 		newParent.PersistentFlags().AddFlagSet(targetCmd.PersistentFlags())
	
	// 		// copy subcommands
	// 		newParent.AddCommand(targetCmd.Commands()...)
	// 		cmd = newParent
	// 	}
	
	// }
	
	return cmd
}

func NewDefaultIsvCommand(in io.Reader, out, errout io.Writer) *cobra.Command {
	cmd := NewIsvCommand(in, out, errout)
	
	if len(os.Args) <= 1 {
		return cmd
	}
	
	cmdPathPieces := os.Args[1:]
	pluginHandler := kubecmd.NewDefaultPluginHandler(plugin.ValidPluginFilenamePrefixes)
	
	// only look for suitable extension executables if
	// the specified command does not already exist
	if _, _, err := cmd.Find(cmdPathPieces); err != nil {
		if err := kubecmd.HandlePluginCommand(pluginHandler, cmdPathPieces); err != nil {
			fmt.Fprintf(errout, "%v\n", err)
			os.Exit(1)
		}
	}
	return cmd
}

func NewIsvCommand(in io.Reader, out, errout io.Writer) *cobra.Command {
	// cliVersion := os.Getenv("CLI_VERSION")
	// Main command
	cmds := &cobra.Command{
		Use:     "isv-cli",
		Short:   "Command line tools for ISV Managed Service Operator",
		Long:    cliLong,
		Version: cliVersion,
		Run: func(c *cobra.Command, args []string) {
			explainOut := term.NewResponsiveWriter(out)
			c.SetOutput(explainOut)
			kcmdutil.RequireNoArguments(c, args)
			fmt.Fprintf(explainOut, "%s\n\n%s\n", cliLong, cliExplain)
		},
	}
	kubeConfigFlags := genericclioptions.NewConfigFlags(true)
	kubeConfigFlags.AddFlags(cmds.PersistentFlags())
	matchVersionKubeConfigFlags := kcmdutil.NewMatchVersionFlags(kubeConfigFlags)
	matchVersionKubeConfigFlags.AddFlags(cmds.PersistentFlags())
	cmds.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	f := kcmdutil.NewFactory(matchVersionKubeConfigFlags)
	
	ioStreams := genericclioptions.IOStreams{In: in, Out: out, ErrOut: errout}
	
	loginCmd := ocwrappers.NewCmdLogin(f, ioStreams)
	mustgatherCmd := mustgather.NewMustGatherCommand(f, ioStreams)
	testHarnessCmd := testharness.NewtestHarnessCommand(f, ioStreams)
	groups := ktemplates.CommandGroups{
		{
			Message: "Basic Commands:",
			Commands: []*cobra.Command{
				mustgatherCmd,
				loginCmd,
				testHarnessCmd,
			},
		},
		{
			Message: "Settings Commands:",
			Commands: []*cobra.Command{
				ocwrappers.NewCmdLogout(f, ioStreams),
			},
		},
	}
	filters := []string{
		"options",
		"deploy",
	}
	groups.Add(cmds)

	cmdutil.ActsAsRootCommand(cmds, filters, groups...).
		ExposeFlags(loginCmd, "certificate-authority", "insecure-skip-tls-verify", "token")

	cmds.AddCommand(options.NewCmdOptions(ioStreams))
	
	return cmds
}
