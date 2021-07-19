package testharness

import (
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	testHarnessLong = templates.LongDesc(`
		For managed service, ISV should provide Test Harness image for ISV product.
		This command will provide the standard structure for Test Harness Standard repositories. The repositories allows ISV
		to develop and test easily with their own cluster. At the end, ISV will provide manifests test image which will be executed from test harness image for OSD e2e test

		Test Harness image can check if kubernetes object exist and the ISV product can integrate with Jupyter notebook properly.

		Experimental: This command is under active development and may change without notice.
	`)

)

type TestHarnessOptions struct {
	genericclioptions.IOStreams

	Config           *rest.Config
	Client           kubernetes.Interface
	RESTClientGetter genericclioptions.RESTClientGetter

	ConfigPath string
	DestDir    string
	ConfigFile *ini.File

	TestHarnessPath string
	ManifestsPath string
}

func NewtestHarnessCommand(f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	
	cmd := &cobra.Command{
		Use:     "test-harness",
		Short:   "Test Harness for ISV Addon operator",
		Long:    testHarnessLong,
		Run: kcmdutil.DefaultSubCommandRun(streams.ErrOut),
	}

	cmd.AddCommand(NewCmdCreate(f, streams))
	
	return cmd
}
