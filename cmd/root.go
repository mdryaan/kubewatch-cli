package cmd

import (
	"fmt"
	"os"

	"github.com/mdryaan/kubewatch-cli/internal/config"
	"github.com/mdryaan/kubewatch-cli/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "kubewatch",
	Short: "Real-time Kubernetes cluster monitoring and anomaly detection",
	Long: `KubeWatch CLI — watch Kubernetes resources in real time, detect anomalies,
report health status, and visualize resource dependencies.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		output.PrintError(err.Error())
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringP("namespace", "n", config.DefaultNamespace, "Kubernetes namespace")
	rootCmd.PersistentFlags().String("kubeconfig", "", "path to kubeconfig file (default: ~/.kube/config)")
	rootCmd.PersistentFlags().StringP("output", "o", config.DefaultOutputFormat, "output format: table, json, plain")
	rootCmd.PersistentFlags().StringP("selector", "l", "", "label selector (e.g. app=nginx)")
	rootCmd.PersistentFlags().Bool("all-namespaces", false, "list resources across all namespaces")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colorized output")

	viper.BindPFlag("namespace", rootCmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("selector", rootCmd.PersistentFlags().Lookup("selector"))
	viper.BindPFlag("all-namespaces", rootCmd.PersistentFlags().Lookup("all-namespaces"))
	viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color"))

	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(summaryCmd)
	rootCmd.AddCommand(anomaliesCmd)
	rootCmd.AddCommand(graphCmd)
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	config.SetDefaults()
	viper.AutomaticEnv()

	if viper.GetBool("no-color") {
		output.DisableColor()
	}
}

func kubeconfig() string {
	return viper.GetString("kubeconfig")
}

func namespace() string {
	return viper.GetString("namespace")
}

func outputFormat() output.Format {
	return output.ParseFormat(viper.GetString("output"))
}

func labelSelector() string {
	return viper.GetString("selector")
}

func allNamespaces() bool {
	return viper.GetBool("all-namespaces")
}

func exitWithError(err error) {
	output.PrintError(err.Error())
	fmt.Fprintln(os.Stderr, "")
	os.Exit(1)
}
