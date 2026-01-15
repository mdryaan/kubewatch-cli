package cmd

import (
	"fmt"
	"runtime"

	"github.com/mdryaan/kubewatch-cli/pkg/output"
	v "github.com/mdryaan/kubewatch-cli/pkg/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		short, _ := cmd.Flags().GetBool("short")
		if short {
			fmt.Println(v.Version)
			return
		}

		output.PrintHeader("KubeWatch CLI")
		fmt.Printf("  %-15s %s\n", "Version:", output.Healthy(v.Version))
		fmt.Printf("  %-15s %s\n", "Git Commit:", output.Info(v.GitCommit))
		fmt.Printf("  %-15s %s\n", "Build Date:", v.BuildDate)
		fmt.Printf("  %-15s %s\n", "Go Version:", runtime.Version())
		fmt.Printf("  %-15s %s/%s\n", "OS/Arch:", runtime.GOOS, runtime.GOARCH)
		fmt.Println()
	},
}

func init() {
	versionCmd.Flags().Bool("short", false, "print version number only")
}
