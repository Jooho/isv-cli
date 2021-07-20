package testharness

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/otiai10/copy"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

var (
	testHarnessCreateLong = templates.LongDesc(`
		Create a Test Harness Standard repository for ISV Addon product.

		This command will create a test harness repo and manifests repo. 
		Test Harness repo will take care of the following:
		- CRD exist
		- All pods are running
		- Manifest Job control

		Manifests repo will take care of the following:
		- ISV product specific objects existences. 
		- jupyter notebook integration test with ISV product.

		Experimental: This command is under active development and may change without notice.
	`)

	testHarnessCreateExample = templates.Examples(`
		# Create odh test harness operator and odh manifests  
		# Sample config ini file: https://github.com/Jooho/isv-cli/tree/main/templates/test-harness/example-config.ini

		isv-cli test-harness create --config-path=./config.ini --dest-dir=/tmp
								
	`)
)

func NewCmdCreate(f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewTestHarnessOptions(streams)

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a Test Harness",
		Long:    testHarnessLong,
		Example: testHarnessCreateExample,
		Run: func(cmd *cobra.Command, args []string) {
			kcmdutil.CheckErr(o.Complete(f, cmd, args))
			kcmdutil.CheckErr(o.Validate())
			kcmdutil.CheckErr(o.Run(f))
		},
	}

	cmd.Flags().StringVar(&o.ConfigPath, "config-path", o.ConfigPath, "Specify config.yaml path.")

	cmd.Flags().StringVar(&o.DestDir, "dest-dir", o.DestDir, "The destination directory where the Test Harness repositories created.")

	return cmd
}

func NewTestHarnessOptions(streams genericclioptions.IOStreams) *TestHarnessOptions {
	return &TestHarnessOptions{
		ConfigPath: "./config.ini",
		DestDir:    "./TestHarness",
	}
}

func (o *TestHarnessOptions) Complete(f kcmdutil.Factory, cmd *cobra.Command, args []string) error {
	o.RESTClientGetter = f
	var err error
	if o.Config, err = f.ToRESTConfig(); err != nil {
		return err
	}
	if o.Client, err = kubernetes.NewForConfig(o.Config); err != nil {
		return err
	}

	if _, err := os.Stat(o.ConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("Config file does not exist")
	}

	if o.ConfigFile, err = ini.Load(o.ConfigPath); err != nil {
		fmt.Printf("Fail to read file: %v", err)
		return err
	}

	if testHarnessName := o.getConfValue("Customize", "TEST_HARNESS_NAME"); testHarnessName == "" {
		return fmt.Errorf("TEST_HARNESS_NAME is not found from config file")
	} else {
		o.TestHarnessPath = o.DestDir + "/" + o.getConfValue("Customize", "TEST_HARNESS_NAME")
	}

	if manifestsName := o.getConfValue("Customize", "MANIFESTS_NAME"); manifestsName == "" {
		return fmt.Errorf("MANIFESTS_NAME is not found from config file")
	} else {
		o.ManifestsPath = o.DestDir + "/" + o.getConfValue("Customize", "MANIFESTS_NAME")
	}
	return nil
}

func (o *TestHarnessOptions) Validate() error {
	var err error

	if _, err := os.Stat(o.DestDir); os.IsNotExist(err) {
		if err = os.Mkdir(o.DestDir, os.FileMode(0777)); err != nil {
			return fmt.Errorf("Failed to create the destination directory(%s): %s", o.DestDir, err)
		}
	}

	if err = os.Mkdir(o.TestHarnessPath, os.FileMode(0777)); err != nil {
		return fmt.Errorf("Failed to create the test harness repository(%s): %s", o.TestHarnessPath, err)
	}

	if err = os.Mkdir(o.ManifestsPath, os.FileMode(0777)); err != nil {
		return fmt.Errorf("Failed to create the manifests repository(%s): %s", o.ManifestsPath, err)
	}

	if err = o.cloneTemplateRepos(); err != nil {
		return fmt.Errorf("Failed to clone template repositories: %s", err)
	}
	return nil
}

func (o *TestHarnessOptions) Run(f kcmdutil.Factory) error {

	var err error

	fmt.Println("** Create Test Harness Repositories **")

	err = copy.Copy("/tmp/test-harness/operator-test-harness", o.TestHarnessPath)
	if err != nil {
		return fmt.Errorf("Failed to copy test harness repository(%s): %s", o.TestHarnessPath, err)
	}

	err = copy.Copy("/tmp/test-harness/manifests-test", o.ManifestsPath)
	if err != nil {
		return fmt.Errorf("Failed to copy manifests repository(%s): %s", o.ManifestsPath, err)
	}

	// Update parameters
	fmt.Println("** Update Variable in Test Harness Repositories **")
	var srcFiles []string

	err = filepath.Walk(o.DestDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if !(strings.Contains(path, "git") || strings.Contains(path, "vendor") || strings.Contains(path, "Makefile")) {
				if strings.Contains(path, "env.sh") {

					input, err := ioutil.ReadFile(o.ConfigPath)
					var output []byte
					if err != nil {
						return fmt.Errorf("failed to read file(%s): %s",path,err)
					}
					groupRe, _ := regexp.Compile("\\[.*]")
					output = groupRe.ReplaceAll(input, []byte(""))

					envShFile := o.ConfigPath + ".sh"
					if err = ioutil.WriteFile(envShFile, output, 0666); err != nil {
						return fmt.Errorf("Failed to write a file(%s): %s", envShFile, err)
					}

					if err = copy.Copy(envShFile, o.TestHarnessPath+"/env.sh"); err != nil {
						return fmt.Errorf("Copy env.sh file to test harness repo: %s", err)
					}

					if err = copy.Copy(envShFile, o.ManifestsPath+"/env.sh"); err != nil {
						return fmt.Errorf("Copy env.sh file to Manifests repo: %s", err)
					}
				} else {
					srcFiles = append(srcFiles, path)
				}
			}

		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	o.updateSrcValue(srcFiles)

	fmt.Printf("\n** Test Harness Repositories are Ready: %s **\n", o.DestDir)

	return nil
}

func (o *TestHarnessOptions) cloneTemplateRepos() error {

	if _, err := git.PlainClone("/tmp/test-harness/operator-test-harness", false, &git.CloneOptions{
		URL:          "https://github.com/Jooho/operator-test-harness.git",
		Progress:      nil,
		ReferenceName: plumbing.ReferenceName("refs/heads/template"),
	}); err != nil {
		return err
	}

	if _, err := git.PlainClone("/tmp/test-harness/manifests-test", false, &git.CloneOptions{
		URL:           "https://github.com/Jooho/manifests-test.git",
		Progress:      nil,
		ReferenceName: plumbing.ReferenceName("refs/heads/template"),
	}); err != nil {
		return err
	}

	os.RemoveAll("/tmp/test-harness/operator-test-harness/.git")
	os.RemoveAll("/tmp/test-harness/manifests-test/.git")
	return nil
}

func (o *TestHarnessOptions) getConfValue(group, param string) string {
	var value string
	value = o.ConfigFile.Section(group).Key(param).String()

	if strings.Contains(value, "${") {
		value = o.replaceEnvValues(value)
	}

	return value
}

func (o *TestHarnessOptions) replaceEnvValues(nestedValues string) string {
	nestevRe := regexp.MustCompile("\\${(.*?)}")
	split := nestevRe.FindAllString(nestedValues, -1)
	envRe := regexp.MustCompile("[\\${|}|' +']")

	for i := range split {
		envString := envRe.Split(split[i], -1)
		nestedValues = strings.Replace(nestedValues, split[i], o.getEnvValue(envString[2 : len(envString)-1][0]), 1)
	}
	if strings.Contains(nestedValues, "$") {
		nestedValues = o.replaceEnvValues(nestedValues)
	}

	return nestedValues
}

func (o *TestHarnessOptions) getEnvValue(key string) string {
	groups := []string{"Customize", "DoNotChange"}
	for i := range groups {

		if envValue := o.ConfigFile.Section(groups[i]).Key(key).String(); envValue != "" {
			return envValue
		}

	}
	return ""
}

func (o *TestHarnessOptions) updateSrcValue(srcFiles []string) error {

	groups := []string{"Customize", "DoNotChange"}
	for group := range groups {

		confKeys := o.ConfigFile.Section(groups[group]).KeyStrings()

		for src := range srcFiles {
			input, err := ioutil.ReadFile(srcFiles[src])
			var output []byte
			if err != nil {
				return fmt.Errorf("Failed to read a file(%s): %s ", srcFiles[src], err)
			}

			for confKey := range confKeys {
				tempEnvParam := confKeys[confKey]
				output = bytes.Replace(input, []byte("%"+tempEnvParam+"%"), []byte(o.getConfValue(groups[group], tempEnvParam)), -1)
				input = output

			}
			if err = ioutil.WriteFile(srcFiles[src], output, 0666); err != nil {
				return fmt.Errorf("Failed to write a file(%s): %s ", srcFiles[src], err)
			}

		}
	}

	return nil
}
