# Contributing to KubeWatch CLI

Thank you for taking the time to contribute. This document explains how to set up your development environment, add new features, and submit changes.

---

## Dev Environment Setup

### Prerequisites

- Go 1.21+
- A Kubernetes cluster (local via [kind](https://kind.sigs.k8s.io/) or [minikube](https://minikube.sigs.k8s.io/), or any remote cluster)
- `kubectl` configured with a working kubeconfig

### 1. Fork the repository

Click the **Fork** button on [github.com/mdryaan/kubewatch-cli](https://github.com/mdryaan/kubewatch-cli) to create your own copy under your GitHub account.

### 2. Clone your fork

```bash
git clone https://github.com/Your-username/kubewatch-cli.git
cd kubewatch-cli
```

### 3. Add the upstream remote

```bash
git remote add upstream https://github.com/mdryaan/kubewatch-cli.git
```

### 4. Install dependencies and build

```bash
go mod download
make build
./kubewatch version
```

### 5. Create a feature branch

```bash
git checkout -b feat/your-feature-name
```

### Run against a local cluster

```bash
kind create cluster --name kubewatch-dev
export KUBECONFIG=$(kind get kubeconfig-path --name kubewatch-dev)
./kubewatch health
```

### Verify your changes compile

```bash
make vet
make build
```

### Keep your fork in sync

```bash
git fetch upstream
git rebase upstream/main
```

---

## How to Add a New Command

1. Create a new file in `cmd/`, e.g. `cmd/mycommand.go`
2. Define a `*cobra.Command` variable and implement the `RunE` function
3. Register the command in `cmd/root.go` inside `init()` via `rootCmd.AddCommand(myCmd)`
4. Add any packages your command needs under `pkg/`

Example skeleton:

```go
package cmd

import (
    "github.com/spf13/cobra"
)

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Short description of what this command does",
    RunE:  runMyCommand,
}

func runMyCommand(cmd *cobra.Command, args []string) error {
    kc, err := client.New(kubeconfig())
    if err != nil {
        return err
    }
    // ... use kc to talk to the cluster
    return nil
}
```

---

## How to Add a New Watcher

1. Create a new file in `pkg/watcher/`, e.g. `pkg/watcher/statefulset_watcher.go`
2. Define a struct embedding `BaseWatcher`
3. Implement the `Watch(ctx, namespace, labelSelector, events chan<- WatchEvent) error` method using the appropriate clientset lister and watcher
4. Add a constructor `NewXxxWatcher(kc *client.KubeClient) *XxxWatcher`
5. Wire it up in the relevant `cmd/watch.go` subcommand

The event loop pattern is consistent across all watchers:

```go
func (w *MyWatcher) Watch(ctx context.Context, namespace string, labelSelector string, events chan<- WatchEvent) error {
    watcher, err := w.client.Clientset.XxxV1().Resources(namespace).Watch(ctx, metav1.ListOptions{
        LabelSelector: labelSelector,
    })
    if err != nil {
        return err
    }
    defer watcher.Stop()

    for {
        select {
        case <-ctx.Done():
            return nil
        case event, ok := <-watcher.ResultChan():
            if !ok {
                return nil
            }
            // convert and send to events channel
        }
    }
}
```

---

## PR Guidelines

- Keep PRs focused on a single concern — one feature, one fix, one refactor
- Title format: `feat(scope): description`, `fix(scope): description`, `chore: description`
- Rebase on `upstream/main` before opening a PR; do not merge-commit
- All code must compile: `make build` must succeed
- `make vet` must pass with zero errors
- If you add a new package, add a short note to this file explaining where it lives and what it does

---

## Code Style

- No comments, no docstrings, no inline explanations — well-named identifiers are the documentation
- Use strong Go types everywhere; avoid `interface{}` or `any` unless forced by an external API
- Error values must be wrapped with context using `fmt.Errorf("context: %w", err)`
- Prefer early returns over nested `if` blocks
- Keep function bodies short — if a function exceeds ~50 lines, consider splitting it
- Do not use `init()` functions outside of `cmd/root.go`
- All public types and functions in `pkg/` must have meaningful names that explain purpose without comments
- Use `context.Context` as the first argument for any function that performs I/O

### Package conventions

| Package | Responsibility |
|---|---|
| `cmd/` | CLI command definitions only — thin layer over pkg/ |
| `pkg/client/` | Kubernetes client construction and configuration |
| `pkg/watcher/` | Real-time resource watchers using the Watch API |
| `pkg/health/` | Health status assessment logic per resource type |
| `pkg/anomaly/` | Anomaly detection rules per failure mode |
| `pkg/graph/` | Dependency graph construction and ASCII rendering |
| `pkg/summary/` | Namespace-level aggregation and reporting |
| `pkg/output/` | All formatting and terminal color helpers |
| `internal/config/` | Viper config loading and defaults |
| `internal/utils/` | Shared stateless helpers (time, labels, strings) |
