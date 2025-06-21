package graph

import (
	"bold/pkg/parser"
	"fmt"
	"strings"
)

type DependencyNode struct {
	ID        string
	Type      string
	Name      string
	Provider  string
	DependsOn []string
}

type DependencyGraph struct {
	Nodes []DependencyNode
	Edges map[string][]string
}

func GenerateDependencyGraph(service *parser.Service) *DependencyGraph {
	graph := &DependencyGraph{
		Nodes: []DependencyNode{},
		Edges: make(map[string][]string),
	}

	for _, network := range service.Spec.Infrastructure.Networks {
		nodeID := fmt.Sprintf("%s_%s", network.Provider, network.Name)
		node := DependencyNode{
			ID:       nodeID,
			Type:     "network",
			Name:     network.Name,
			Provider: network.Provider,
		}
		graph.Nodes = append(graph.Nodes, node)

		for _, subnet := range network.Subnets {
			subnetID := fmt.Sprintf("%s_%s", network.Provider, subnet.Name)
			subnetNode := DependencyNode{
				ID:        subnetID,
				Type:      "subnet",
				Name:      subnet.Name,
				Provider:  network.Provider,
				DependsOn: []string{nodeID},
			}
			graph.Nodes = append(graph.Nodes, subnetNode)
			graph.Edges[nodeID] = append(graph.Edges[nodeID], subnetID)
		}
	}

	for _, sg := range service.Spec.Infrastructure.SecurityGroups {
		sgID := fmt.Sprintf("%s_%s", sg.Provider, sg.Name)
		vpcID := fmt.Sprintf("%s_%s", sg.Provider, sg.VPC)

		node := DependencyNode{
			ID:        sgID,
			Type:      "security_group",
			Name:      sg.Name,
			Provider:  sg.Provider,
			DependsOn: []string{vpcID},
		}
		graph.Nodes = append(graph.Nodes, node)
		graph.Edges[vpcID] = append(graph.Edges[vpcID], sgID)
	}

	for _, cluster := range service.Spec.Infrastructure.KubernetesClusters {
		clusterID := fmt.Sprintf("%s_%s", cluster.Provider, cluster.Name)
		vpcID := fmt.Sprintf("%s_%s", cluster.Provider, cluster.VPC)

		node := DependencyNode{
			ID:        clusterID,
			Type:      "kubernetes",
			Name:      cluster.Name,
			Provider:  cluster.Provider,
			DependsOn: []string{vpcID},
		}
		graph.Nodes = append(graph.Nodes, node)
		graph.Edges[vpcID] = append(graph.Edges[vpcID], clusterID)
	}

	for _, compute := range service.Spec.Infrastructure.Computes {
		computeID := fmt.Sprintf("%s_%s", compute.Provider, compute.Name)
		subnetID := fmt.Sprintf("%s_%s", compute.Provider, compute.Subnet)

		dependencies := []string{subnetID}

		if compute.SecurityGroup != "" {
			sgID := fmt.Sprintf("%s_%s", compute.Provider, compute.SecurityGroup)
			dependencies = append(dependencies, sgID)
		}

		node := DependencyNode{
			ID:        computeID,
			Type:      "compute",
			Name:      compute.Name,
			Provider:  compute.Provider,
			DependsOn: dependencies,
		}
		graph.Nodes = append(graph.Nodes, node)

		for _, dep := range dependencies {
			graph.Edges[dep] = append(graph.Edges[dep], computeID)
		}
	}

	return graph
}

func GenerateMermaidDiagram(graph *DependencyGraph) string {
	var mermaid strings.Builder
	mermaid.WriteString("graph TD\n")

	for _, node := range graph.Nodes {
		var shape string
		switch node.Type {
		case "network":
			shape = "{{" + node.Name + "}}"
		case "subnet":
			shape = "[" + node.Name + "]"
		case "security_group":
			shape = "(" + node.Name + ")"
		case "kubernetes":
			shape = "[/" + node.Name + "/]"
		case "compute":
			shape = ">" + node.Name + "]"
		default:
			shape = "[" + node.Name + "]"
		}

		mermaid.WriteString(fmt.Sprintf("    %s%s\n", node.ID, shape))
	}

	for from, toList := range graph.Edges {
		for _, to := range toList {
			mermaid.WriteString(fmt.Sprintf("    %s --> %s\n", from, to))
		}
	}

	return mermaid.String()
}

func GenerateDotGraph(graph *DependencyGraph) string {
	var dot strings.Builder
	dot.WriteString("digraph G {\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    node [shape=box, style=filled];\n\n")

	dot.WriteString("    // Node definitions\n")
	for _, node := range graph.Nodes {
		var color string
		switch node.Type {
		case "network":
			color = "lightblue"
		case "subnet":
			color = "lightgreen"
		case "security_group":
			color = "lightyellow"
		case "kubernetes":
			color = "lightpink"
		case "compute":
			color = "lightcoral"
		default:
			color = "lightgray"
		}

		dot.WriteString(fmt.Sprintf("    %s [label=\"%s\\n(%s)\", fillcolor=\"%s\"];\n",
			node.ID, node.Name, node.Type, color))
	}

	dot.WriteString("\n    // Edges\n")
	for from, toList := range graph.Edges {
		for _, to := range toList {
			dot.WriteString(fmt.Sprintf("    %s -> %s;\n", from, to))
		}
	}

	dot.WriteString("}\n")
	return dot.String()
}

func PrintDependencyTree(graph *DependencyGraph) string {
	var tree strings.Builder
	tree.WriteString("Resource Dependency Tree:\n")
	tree.WriteString("========================\n\n")

	nodesByType := make(map[string][]DependencyNode)
	for _, node := range graph.Nodes {
		nodesByType[node.Type] = append(nodesByType[node.Type], node)
	}

	types := []string{"network", "subnet", "security_group", "kubernetes", "compute"}

	for _, nodeType := range types {
		if nodes, exists := nodesByType[nodeType]; exists {
			tree.WriteString(fmt.Sprintf("%s:\n", strings.Title(nodeType)))
			for _, node := range nodes {
				tree.WriteString(fmt.Sprintf("  - %s (%s)\n", node.Name, node.Provider))
				if len(node.DependsOn) > 0 {
					tree.WriteString("    Depends on: ")
					deps := make([]string, len(node.DependsOn))
					for i, dep := range node.DependsOn {
						parts := strings.Split(dep, "_")
						if len(parts) >= 2 {
							deps[i] = parts[1]
						} else {
							deps[i] = dep
						}
					}
					tree.WriteString(strings.Join(deps, ", "))
					tree.WriteString("\n")
				}
			}
			tree.WriteString("\n")
		}
	}

	return tree.String()
}
