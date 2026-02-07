package graph

import (
	"fmt"
	"os"
	"strings"
)

func Render(g *Graph) {
	if len(g.Roots) == 0 {
		fmt.Fprintln(os.Stdout, "  no resources found")
		return
	}
	for i, root := range g.Roots {
		isLast := i == len(g.Roots)-1
		renderNode(root, "", isLast)
	}
}

func renderNode(node *GraphNode, prefix string, isLast bool) {
	connector := "├── "
	childPrefix := prefix + "│   "
	if isLast {
		connector = "└── "
		childPrefix = prefix + "    "
	}

	icon := kindIcon(node.Kind)
	statusStr := colorStatus(node.Kind, node.Status)
	fmt.Fprintf(os.Stdout, "%s%s%s %s  %s\n", prefix, connector, icon, node.Name, statusStr)

	for i, child := range node.Children {
		renderNode(child, childPrefix, i == len(node.Children)-1)
	}
}

func kindIcon(kind NodeKind) string {
	switch kind {
	case KindDeployment:
		return "[Deploy]"
	case KindReplicaSet:
		return "[RS]    "
	case KindPod:
		return "[Pod]   "
	case KindService:
		return "[Svc]   "
	case KindStatefulSet:
		return "[STS]   "
	default:
		return "[Res]   "
	}
}

func colorStatus(kind NodeKind, status string) string {
	if kind == KindPod {
		switch strings.ToLower(status) {
		case "running":
			return fmt.Sprintf("\033[32m%s\033[0m", status)
		case "pending":
			return fmt.Sprintf("\033[33m%s\033[0m", status)
		case "failed", "crashloopbackoff":
			return fmt.Sprintf("\033[31m%s\033[0m", status)
		}
	}
	return status
}
