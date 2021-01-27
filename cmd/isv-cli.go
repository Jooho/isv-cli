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
package main

import (
	goflag "flag"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jooho/isv-cli/pkg/cli"
	"github.com/spf13/pflag"
	utilflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func injectLoglevelFlag(flags *pflag.FlagSet) {
	from := goflag.CommandLine
	if flag := from.Lookup("v"); flag != nil {
		level := flag.Value.(*klog.Level)
		levelPtr := (*int32)(level)
		flags.Int32Var(levelPtr, "loglevel", 0, "Set the level of log output (0-10)")
		if flags.Lookup("v") == nil {
			flags.Int32Var(levelPtr, "v", 0, "Set the level of log output (0-10)")
		}
	}
}

func main() {

	logs.InitLogs()
	defer logs.FlushLogs()

	rand.Seed(time.Now().UTC().UnixNano())
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	// Prevents race condition present in vendored version of Docker.
	// See: https://github.com/moby/moby/issues/39859
	os.Setenv("MOBY_DISABLE_PIGZ", "true")

	pflag.CommandLine.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	injectLoglevelFlag(pflag.CommandLine)
	basename := filepath.Base(os.Args[0])

	command := cli.CommandFor(basename)
	if err := command.Execute(); err != nil {
		 os.Exit(1)
	}

}
