package ocwrappers

import (
	"bufio"
	"github.com/openshift/oc/pkg/cli/login"
	"github.com/openshift/oc/pkg/cli/logout"
	cmdutil "github.com/openshift/oc/pkg/helpers/cmd"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"
	"strings"
)

func adjustCmdExamples(cmd *cobra.Command, name string) {
	for _, subCmd := range cmd.Commands() {
		adjustCmdExamples(subCmd, cmd.Name())
	}
	cmd.Example = strings.Replace(cmd.Example, "oc", "isv-cli", -1)
	tabbing := "  "
	examples := []string{}
	scanner := bufio.NewScanner(strings.NewReader(cmd.Example))
	for scanner.Scan() {
		examples = append(examples, tabbing+strings.TrimSpace(scanner.Text()))
	}
	cmd.Example = strings.Join(examples, "\n")
}

func NewCmdLogout(f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	cmd := cmdutil.ReplaceCommandName("oc", "isv-cli", templates.Normalize(logout.NewCmdLogout(f, streams)))

	adjustCmdExamples(cmd, "logout")
	return cmd
}


func NewCmdLogin(f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	cmd := cmdutil.ReplaceCommandName("oc", "isv-cli", templates.Normalize(login.NewCmdLogin(f, streams)))

	adjustCmdExamples(cmd, "login")
	return cmd
}
