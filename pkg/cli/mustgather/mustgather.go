/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package mustgather

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/exec"
	"k8s.io/kubectl/pkg/cmd/logs"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/util/templates"

	isvrsync "github.com/jooho/isv-cli/pkg/cli/rsync"
	imagereference "github.com/openshift/library-go/pkg/image/reference"
	"github.com/openshift/library-go/pkg/operator/resource/retry"
	"github.com/openshift/oc/pkg/cli/admin/inspect"
	"github.com/openshift/oc/pkg/cli/admin/mustgather"
	"github.com/openshift/oc/pkg/cli/rsync"
)

const (
	mustGatherServiceAccountName string = "must-gather-sa"
)

var (
	mustGatherLong = templates.LongDesc(`
		Launch a pod to gather debugging information

		This command will launch a pod in the namespace where you are in that gathers
		debugging information and then downloads the gathered information.

		Compared to 'oc adm must-gather', ISV must-gather only gather namespace level data including operands

		Experimental: This command is under active development and may change without notice.
	`)

	mustGatherExample = templates.Examples(`
		# gather information using operator must-gather plug-in image and downloaded it to ./must-gather.local.<rand>
		isv-cli must-gather --image=quay.io/kubevirt/must-gather
			
		# gather information with a specific local folder to copy to
		isv-cli must-gather --image=quay.io/kubevirt/must-gather --dest-dir=/local/directory
								
	`)
)

type MustGatherOptions struct {
	// Embed oc's MustGatherOptions directly.
	*mustgather.MustGatherOptions
	RestClient *rest.Config
	Clientset  kubernetes.Interface
	SourceDir  string
	Tar        bool
}

func NewMustGatherCommand(f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewMustGatherOptions(streams)

	cmd := &cobra.Command{
		Use:     "must-gather",
		Short:   "Launch a new instance of a pod for gathering debug information",
		Long:    mustGatherLong,
		Example: mustGatherExample,
		Run: func(cmd *cobra.Command, args []string) {
			kcmdutil.CheckErr(o.Complete(f, cmd, args))
			kcmdutil.CheckErr(o.Validate())
			kcmdutil.CheckErr(o.Run(f))
		},
	}

	cmd.Flags().StringSliceVar(&o.Images, "image", o.Images, "Requuired. Specify a operator must-gather plugin image to run.")
	cmd.Flags().StringVar(&o.DestDir, "dest-dir", o.DestDir, "Set a specific directory on the local machine to write gathered data to. Default ./must-gather.local.<rand>")
	cmd.Flags().Int64Var(&o.Timeout, "timeout", 600, "The length of time to gather data, in seconds. Defaults to 10 minutes.")
	cmd.Flags().BoolVar(&o.Tar, "notar", o.Tar, "Copy must-gather data without archive")
	cmd.Flags().BoolVar(&o.Keep, "keep", o.Keep, "Do not delete temporary resources when command completes.")
	cmd.Flags().MarkHidden("keep")

	cmd.MarkFlagRequired("image")
	return cmd
}

func NewMustGatherOptions(streams genericclioptions.IOStreams) *MustGatherOptions {
	return &MustGatherOptions{
		MustGatherOptions: mustgather.NewMustGatherOptions(streams),
		SourceDir:         "/opt/must-gather-root/must-gather",
	}
}

func (o *MustGatherOptions) Complete(f kcmdutil.Factory, cmd *cobra.Command, args []string) error {

	err := o.MustGatherOptions.Complete(f, cmd, args)
	if err != nil {
		return err
	}

	o.MustGatherOptions.RsyncRshCmd = isvrsync.DefaultRsyncRemoteShellToUse(cmd)

	// for execoptions
	restClient, err := f.ToRESTConfig()
	if err != nil {
		return err
	}
	o.RestClient = restClient

	clientset, err := f.KubernetesClientSet()
	if err != nil {
		return err
	}
	o.Clientset = clientset

	return nil
}

func (o *MustGatherOptions) Validate() error {
	return o.MustGatherOptions.Validate()
}

func (o *MustGatherOptions) Run(f kcmdutil.Factory) error {

	currNamespace, _, err := f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	// deploy must-gather pod
	var pods []*corev1.Pod
	for _, image := range o.Images {
		_, err := imagereference.Parse(image)
		if err != nil {
			o.log("unable to parse image reference %s: %v", image, err)
			return err
		}
		sa, err := o.Client.CoreV1().ServiceAccounts(currNamespace).Create(context.TODO(), o.newSA(), metav1.CreateOptions{})
		if err != nil {
			return err
		}
		roleBinding, err := o.Client.RbacV1().RoleBindings(currNamespace).Create(context.TODO(), o.newRoleBinding(), metav1.CreateOptions{})
		if err != nil {
			return err
		}

		if !o.Keep {
			defer func() {
				if err := o.Client.RbacV1().RoleBindings(currNamespace).Delete(context.TODO(), roleBinding.Name, metav1.DeleteOptions{}); err != nil {
					fmt.Printf("%v\n", err)
					return
				}
				if err := o.Client.CoreV1().ServiceAccounts(currNamespace).Delete(context.TODO(), sa.Name, metav1.DeleteOptions{}); err != nil {
					fmt.Printf("%v\n", err)
					return
				}

				o.PrinterDeleted.PrintObj(sa, o.LogOut)
			}()
		}

		pod, err := o.Client.CoreV1().Pods(currNamespace).Create(context.TODO(), o.newPod(o.NodeName, image), metav1.CreateOptions{})
		if err != nil {
			return err
		}
		o.log("pod for plug-in image %s created", image)
		pods = append(pods, pod)
	}

	// log timestamps...
	if err := os.MkdirAll(o.DestDir, os.ModePerm); err != nil {
		return err
	}
	if err := o.logTimestamp(); err != nil {
		return err
	}
	defer o.logTimestamp()

	var wg sync.WaitGroup
	wg.Add(len(pods))
	errCh := make(chan error, len(pods))
	for _, pod := range pods {

		go func(pod *corev1.Pod) {
			defer wg.Done()

			log := newPodOutLogger(o.Out, pod.Name)

			// wait for gather container to be running (gather is running)
			if err := o.waitForGatherContainerRunning(pod); err != nil {
				log("gather did not start: %s", err)
				errCh <- fmt.Errorf("gather did not start for pod %s: %s", pod.Name, err)
				return
			}
			// stream gather container logs
			if err := o.getGatherContainerLogs(pod); err != nil {
				log("gather logs unavailable: %v", err)
			}

			// wait for pod to be running (gather has completed)
			log("waiting for gather to complete")
			if err := o.waitForGatherToComplete(pod); err != nil {
				log("gather never finished: %v", err)
				errCh <- fmt.Errorf("gather never finished for pod %s: %s", pod.Name, err)
				return
			}
			
			pod, err = o.Client.CoreV1().Pods(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})

			// archive the gathered files into tarball format
			if !o.Tar {
				log("archieving the gathered data")
				if err := o.ExecCmdInPod(pod); err != nil {
					log("archieving failed: %v\n", err)
					errCh <- fmt.Errorf("unable to archive gathered output from pod %s: %s", pod.Name, err)
					return
				}
				o.SourceDir = "/opt/must-gather-root/tar"
			}

			// copy the gathered files or tarball into the local destination dir
			log("downloading gather output")
			if err != nil {
				log("gather output not downloaded: %v\n", err)
				errCh <- fmt.Errorf("unable to download output from pod %s: %s", pod.Name, err)
				return
			}

			if err := o.copyFilesFromPod(pod); err != nil {
				log("gather output not downloaded: %v\n", err)
				errCh <- fmt.Errorf("unable to download output from pod %s: %s", pod.Name, err)
				return
			}
		}(pod)
	}
	wg.Wait()
	close(errCh)
	var errs []error
	for i := range errCh {
		errs = append(errs, i)
	}

	// now gather all the events into a single file and produce a unified file
	if err := inspect.CreateEventFilterPage(o.DestDir); err != nil {
		errs = append(errs, err)
	}

	return errors.NewAggregate(errs)
}

// ExecCmdInPod is the same as `oc exec command` to archive must-gather data
func (o *MustGatherOptions) ExecCmdInPod(pod *corev1.Pod) error {

	options := &exec.ExecOptions{
		StreamOptions: exec.StreamOptions{
			IOStreams: genericclioptions.IOStreams{
				In:     o.In,
				Out:    o.Out,
				ErrOut: o.ErrOut,
			},

			Namespace:     pod.Namespace,
			PodName:       pod.Name,
			ContainerName: "copy",
		},
		Command:   []string{"tar", "cvf", "/opt/must-gather-root/tar/must-gather.tar", "./must-gather/"},
		Executor:  &exec.DefaultRemoteExecutor{},
		Config:    o.RestClient,
		PodClient: o.Clientset.CoreV1(),
	}

	err := o.execute(options)
	kcmdutil.CheckErr(err)

	return nil
}

func (o *MustGatherOptions) execute(options *exec.ExecOptions) error {
	if err := options.Validate(); err != nil {
		return err
	}

	if err := options.Run(); err != nil {
		return err
	}
	return nil
}

// oc mustgather util fuctinos are not exported so copied them from https://github.com/openshift/oc/blob/release-4.7/pkg/cli/admin/mustgather/mustgather.go

func newPodOutLogger(out io.Writer, podName string) func(string, ...interface{}) {
	writer := newPrefixWriter(out, fmt.Sprintf("[%s] OUT", podName))
	return func(format string, a ...interface{}) {
		fmt.Fprintf(writer, format+"\n", a...)
	}
}

func (o *MustGatherOptions) logTimestamp() error {
	f, err := os.OpenFile(path.Join(o.DestDir, "timestamp"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	_, err = f.WriteString(fmt.Sprintf("%v\n", time.Now()))
	return err
}

func (o *MustGatherOptions) copyFilesFromPod(pod *corev1.Pod) error {
	streams := o.IOStreams
	streams.Out = newPrefixWriter(streams.Out, fmt.Sprintf("[%s] OUT", pod.Name))
	destDir := path.Join(o.DestDir, regexp.MustCompile("[^A-Za-z0-9]+").ReplaceAllString(pod.Status.ContainerStatuses[0].ImageID, "-"))
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}
	rsyncOptions := &rsync.RsyncOptions{
		Namespace:     pod.Namespace,
		Source:        &rsync.PathSpec{PodName: pod.Name, Path: path.Clean(o.SourceDir) + "/"},
		ContainerName: "copy",
		Destination:   &rsync.PathSpec{PodName: "", Path: destDir},
		Client:        o.Client,
		Config:        o.Config,
		RshCmd:        fmt.Sprintf("%s --namespace=%s -c copy", o.RsyncRshCmd, pod.Namespace),
		IOStreams:     streams,
	}
	rsyncOptions.Strategy = rsync.NewDefaultCopyStrategy(rsyncOptions)

	return rsyncOptions.RunRsync()
}

func (o *MustGatherOptions) tarFilesInPod(pod *corev1.Pod) error {
	streams := o.IOStreams
	streams.Out = newPrefixWriter(streams.Out, fmt.Sprintf("[%s] OUT", pod.Name))
	destDir := path.Join(o.DestDir, regexp.MustCompile("[^A-Za-z0-9]+").ReplaceAllString(pod.Status.ContainerStatuses[0].ImageID, "-"))
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}

	rsyncOptions := &rsync.RsyncOptions{
		Namespace:     pod.Namespace,
		Source:        &rsync.PathSpec{PodName: pod.Name, Path: path.Clean(o.SourceDir) + "/"},
		ContainerName: "copy",
		Destination:   &rsync.PathSpec{PodName: "", Path: destDir},
		Client:        o.Client,
		Config:        o.Config,
		RshCmd:        fmt.Sprintf("%s --namespace=%s -c copy", o.RsyncRshCmd, pod.Namespace),
		IOStreams:     streams,
	}
	rsyncOptions.Strategy = rsync.NewDefaultCopyStrategy(rsyncOptions)

	// rsyncOptions.Strategy = rsync.NewTarStrategy(rsyncOptions)

	return rsyncOptions.RunRsync()
}

func (o *MustGatherOptions) getGatherContainerLogs(pod *corev1.Pod) error {
	return (&logs.LogsOptions{
		Namespace:   pod.Namespace,
		ResourceArg: pod.Name,
		Options: &corev1.PodLogOptions{
			Follow:    true,
			Container: pod.Spec.Containers[0].Name,
		},
		RESTClientGetter: o.RESTClientGetter,
		Object:           pod,
		ConsumeRequestFn: logs.DefaultConsumeRequest,
		LogsForObject:    polymorphichelpers.LogsForObjectFn,
		IOStreams:        genericclioptions.IOStreams{Out: newPrefixWriter(o.Out, fmt.Sprintf("[%s] POD", pod.Name))},
	}).RunLogs()
}

func newPrefixWriter(out io.Writer, prefix string) io.Writer {
	reader, writer := io.Pipe()
	scanner := bufio.NewScanner(reader)
	go func() {
		for scanner.Scan() {
			fmt.Fprintf(out, "%s %s\n", prefix, scanner.Text())
		}
	}()
	return writer
}

func (o *MustGatherOptions) waitForGatherToComplete(pod *corev1.Pod) error {
	err := wait.PollImmediate(10*time.Second, time.Duration(o.Timeout)*time.Second, func() (bool, error) {
		var err error
		if pod, err = o.Client.CoreV1().Pods(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{}); err != nil {
			// at this stage pod should exist, we've been gathering container logs, so error if not found
			if kerrors.IsNotFound(err) {
				return true, err
			}
			return false, nil
		}
		var state *corev1.ContainerState
		for _, cstate := range pod.Status.ContainerStatuses {
			if cstate.Name == "gather" {
				state = &cstate.State
				break
			}
		}

		// missing status for gather container => timeout in the worst case
		if state == nil {
			return false, nil
		}

		if state.Terminated != nil {
			if state.Terminated.ExitCode == 0 {
				return true, nil
			}
			return true, fmt.Errorf("%s/%s unexpectedly terminated: exit code: %v, reason: %s, message: %s", pod.Namespace, pod.Name, state.Terminated.ExitCode, state.Terminated.Reason, state.Terminated.Message)
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (o *MustGatherOptions) waitForGatherContainerRunning(pod *corev1.Pod) error {
	return wait.PollImmediate(10*time.Second, time.Duration(o.Timeout)*time.Second, func() (bool, error) {
		var err error
		if pod, err = o.Client.CoreV1().Pods(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{}); err == nil {
			if len(pod.Status.ContainerStatuses) == 0 {
				return false, nil
			}
			state := pod.Status.ContainerStatuses[0].State
			if state.Waiting != nil {
				switch state.Waiting.Reason {
				case "ErrImagePull", "ImagePullBackOff", "InvalidImageName":
					return true, fmt.Errorf("unable to pull image: %v: %v", state.Waiting.Reason, state.Waiting.Message)
				}
			}
			running := state.Running != nil
			terminated := state.Terminated != nil
			return running || terminated, nil
		}
		if retry.IsHTTPClientError(err) {
			return false, nil
		}
		return false, err
	})
}

func (o *MustGatherOptions) newPod(node, image string) *corev1.Pod {
	zero := int64(0)

	nodeSelector := map[string]string{
		corev1.LabelOSStable: "linux",
	}
	if node == "" {
		nodeSelector["node-role.kubernetes.io/worker"] = ""
	}

	ret := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "must-gather-",
			Labels: map[string]string{
				"app": "must-gather",
			},
		},
		Spec: corev1.PodSpec{
			NodeName:           node,
			RestartPolicy:      corev1.RestartPolicyNever,
			ServiceAccountName: mustGatherServiceAccountName,
			Volumes: []corev1.Volume{
				{
					Name: "must-gather-output",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
			Containers: []corev1.Container{
				{
					Name:            "gather",
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					// always force disk flush to ensure that all data gathered is accessible in the copy container
					Command: []string{"/bin/bash", "-c", "gather; sync"},
					Env: []corev1.EnvVar{
						{
							Name: "NAMESPACE",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.namespace",
								},
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "must-gather-output",
							MountPath: path.Clean(o.SourceDir),
							ReadOnly:  false,
						},
					},
				},
				{
					Name:            "copy",
					Image:           image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         []string{"/bin/bash", "-c", "trap : TERM INT; sleep infinity & wait"},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "must-gather-output",
							MountPath: path.Clean(o.SourceDir),
							ReadOnly:  false,
						},
					},
				},
			},
			NodeSelector:                  nodeSelector,
			TerminationGracePeriodSeconds: &zero,
			Tolerations: []corev1.Toleration{
				{
					Operator: "Exists",
				},
			},
		},
	}
	if len(o.Command) > 0 {
		// always force disk flush to ensure that all data gathered is accessible in the copy container
		ret.Spec.Containers[0].Command = []string{"/bin/bash", "-c", fmt.Sprintf("%s; sync", strings.Join(o.Command, " "))}
	}

	return ret
}

func (o *MustGatherOptions) log(format string, a ...interface{}) {
	fmt.Fprintf(o.LogOut, format+"\n", a...)
}

func (o *MustGatherOptions) newSA() *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{

		ObjectMeta: metav1.ObjectMeta{
			Name: mustGatherServiceAccountName,
			Labels: map[string]string{
				"app": "must-gather",
			},
		},
	}

	return sa
}

func (o *MustGatherOptions) newRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "must-gather-",
			Annotations: map[string]string{
				"oc.openshift.io/command": "isv-cli must-gather",
			},
			Labels: map[string]string{
				"app": "must-gather",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "admin",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: mustGatherServiceAccountName,
			},
		},
	}
}
