package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mdryaan/kubewatch-cli/pkg/client"
	"github.com/mdryaan/kubewatch-cli/pkg/output"
	"github.com/mdryaan/kubewatch-cli/pkg/watcher"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch Kubernetes resources in real time",
	Long:  `Stream live updates for pods, deployments, services, nodes, or events.`,
}

var watchPodsCmd = &cobra.Command{
	Use:   "pods",
	Short: "Watch pods in real time",
	RunE:  runWatchPods,
}

var watchDeploymentsCmd = &cobra.Command{
	Use:   "deployments",
	Short: "Watch deployments in real time",
	RunE:  runWatchDeployments,
}

var watchServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Watch services in real time",
	RunE:  runWatchServices,
}

var watchNodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Watch nodes in real time",
	RunE:  runWatchNodes,
}

var watchEventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Stream live Kubernetes events",
	RunE:  runWatchEvents,
}

func init() {
	watchCmd.PersistentFlags().IntP("interval", "i", 0, "refresh interval in seconds (0 = stream)")
	viper.BindPFlag("interval", watchCmd.PersistentFlags().Lookup("interval"))

	watchCmd.AddCommand(watchPodsCmd)
	watchCmd.AddCommand(watchDeploymentsCmd)
	watchCmd.AddCommand(watchServicesCmd)
	watchCmd.AddCommand(watchNodesCmd)
	watchCmd.AddCommand(watchEventsCmd)
}

func runWatchPods(cmd *cobra.Command, args []string) error {
	return runWatcher("pods", func(kc *client.KubeClient, ctx context.Context, events chan<- watcher.WatchEvent) error {
		return watcher.NewPodWatcher(kc).Watch(ctx, namespace(), labelSelector(), events)
	})
}

func runWatchDeployments(cmd *cobra.Command, args []string) error {
	return runWatcher("deployments", func(kc *client.KubeClient, ctx context.Context, events chan<- watcher.WatchEvent) error {
		return watcher.NewDeploymentWatcher(kc).Watch(ctx, namespace(), labelSelector(), events)
	})
}

func runWatchServices(cmd *cobra.Command, args []string) error {
	return runWatcher("services", func(kc *client.KubeClient, ctx context.Context, events chan<- watcher.WatchEvent) error {
		return watcher.NewServiceWatcher(kc).Watch(ctx, namespace(), labelSelector(), events)
	})
}

func runWatchNodes(cmd *cobra.Command, args []string) error {
	return runWatcher("nodes", func(kc *client.KubeClient, ctx context.Context, events chan<- watcher.WatchEvent) error {
		return watcher.NewNodeWatcher(kc).Watch(ctx, "", labelSelector(), events)
	})
}

func runWatchEvents(cmd *cobra.Command, args []string) error {
	kc, err := client.New(kubeconfig())
	if err != nil {
		return fmt.Errorf("connecting to cluster: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	events := make(chan watcher.WatchEvent, 100)
	output.PrintHeader(fmt.Sprintf("Streaming Events — %s", namespace()))
	fmt.Fprintf(os.Stdout, "  %s  Press Ctrl+C to stop\n\n", output.Faint("watching..."))

	go func() {
		if err := watcher.NewEventWatcher(kc).Watch(ctx, namespace(), "", events); err != nil {
			output.PrintError(err.Error())
			cancel()
		}
		close(events)
	}()

	for ev := range events {
		icon := "·"
		colorFn := output.Info
		if ev.Status == "Warning" || ev.Type == watcher.EventType("Warning") {
			icon = "⚠"
			colorFn = output.Warning
		}
		ts := ev.Timestamp.Format("15:04:05")
		fmt.Fprintf(os.Stdout, "  %s  %s  %s  %s/%s  %s\n",
			output.Faint(ts),
			colorFn(icon),
			output.Info(string(ev.Kind)),
			ev.Namespace,
			ev.Name,
			ev.Message,
		)
	}
	return nil
}

type watchFn func(kc *client.KubeClient, ctx context.Context, events chan<- watcher.WatchEvent) error

func runWatcher(resource string, fn watchFn) error {
	kc, err := client.New(kubeconfig())
	if err != nil {
		return fmt.Errorf("connecting to cluster: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	events := make(chan watcher.WatchEvent, 100)
	output.PrintHeader(fmt.Sprintf("Watching %s — %s", resource, namespace()))
	fmt.Fprintf(os.Stdout, "  %s  Press Ctrl+C to stop\n\n", output.Faint("streaming..."))

	t := output.NewTable([]string{"TIME", "EVENT", "KIND", "NAME", "NAMESPACE", "STATUS", "MESSAGE"})

	go func() {
		if err := fn(kc, ctx, events); err != nil {
			output.PrintError(err.Error())
			cancel()
		}
		close(events)
	}()

	for ev := range events {
		ts := ev.Timestamp.Format("15:04:05")
		eventColor := output.Info(string(ev.Type))
		switch ev.Type {
		case watcher.EventAdded:
			eventColor = output.Healthy(string(ev.Type))
		case watcher.EventDeleted:
			eventColor = output.Critical(string(ev.Type))
		}
		t.AddRow([]string{
			output.Faint(ts),
			eventColor,
			output.Info(ev.Kind),
			ev.Name,
			ev.Namespace,
			output.StatusColor(ev.Status),
			ev.Message,
		})
		t.Render()
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}
