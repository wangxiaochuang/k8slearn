package cli

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
)

func Run(cmd *cobra.Command) int {
	if logsInitialized, err := run(cmd); err != nil {
		if !logsInitialized {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		} else {
			klog.ErrorS(err, "command failed")
		}
		return 1
	}
	return 0
}

func run(cmd *cobra.Command) (logsInitialized bool, err error) {
	rand.Seed(time.Now().UnixNano())
	defer logs.FlushLogs()

	cmd.SetGlobalNormalizationFunc(cliflag.WordSepNormalizeFunc)

	if !cmd.SilenceUsage {
		cmd.SilenceUsage = true
		cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
			// Re-enable usage printing.
			c.SilenceUsage = false
			return err
		})
	}

	cmd.SilenceErrors = true

	// 添加日志相关的flag
	logs.AddFlags(cmd.PersistentFlags())

	switch {
	case cmd.PersistentPreRun != nil:
		pre := cmd.PersistentPreRun
		cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
			logs.InitLogs()
			logsInitialized = true
			pre(cmd, args)
		}
	case cmd.PersistentPreRunE != nil:
		pre := cmd.PersistentPreRunE
		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			logs.InitLogs()
			logsInitialized = true
			return pre(cmd, args)
		}
	default:
		cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
			logs.InitLogs()
			logsInitialized = true
		}
	}

	err = cmd.Execute()
	return
}
