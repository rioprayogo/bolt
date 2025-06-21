package cost

import (
	"bold/pkg/parser"
	"fmt"
	"strings"
)

type CostEstimate struct {
	ResourceType string
	ResourceName string
	Provider     string
	MonthlyCost  float64
	HourlyCost   float64
	Currency     string
	Details      map[string]interface{}
}

type CostReport struct {
	TotalMonthlyCost float64
	TotalHourlyCost  float64
	Currency         string
	Estimates        []CostEstimate
	Summary          map[string]float64
}

type PricingData struct {
	AWS   map[string]map[string]float64
	Azure map[string]map[string]float64
	GCP   map[string]map[string]float64
}

var defaultPricing = PricingData{
	AWS: map[string]map[string]float64{
		"ec2": map[string]float64{
			"t2.micro": 0.0116,
			"t3.micro": 0.0104,
			"t2.small": 0.023,
			"t3.small": 0.0208,
			"m5.large": 0.096,
			"c5.large": 0.085,
		},
		"storage": map[string]float64{
			"gp2": 0.10,
			"io1": 0.125,
		},
		"network": map[string]float64{
			"vpc":    0.0,
			"subnet": 0.0,
			"nat":    0.045,
		},
	},
	Azure: map[string]map[string]float64{
		"vm": map[string]float64{
			"Standard_B1s":    0.0104,
			"Standard_B2s":    0.0416,
			"Standard_D2s_v3": 0.096,
		},
		"storage": map[string]float64{
			"Standard_LRS": 0.0184,
			"Premium_LRS":  0.12288,
		},
		"network": map[string]float64{
			"vnet":   0.0,
			"subnet": 0.0,
		},
	},
	GCP: map[string]map[string]float64{
		"compute": map[string]float64{
			"e2-micro":      0.008474,
			"e2-small":      0.016948,
			"e2-medium":     0.033896,
			"n1-standard-1": 0.0475,
		},
		"storage": map[string]float64{
			"pd-standard": 0.04,
			"pd-ssd":      0.17,
		},
		"network": map[string]float64{
			"vpc":    0.0,
			"subnet": 0.0,
		},
	},
}

func EstimateCosts(service *parser.Service) *CostReport {
	report := &CostReport{
		Currency:  "USD",
		Estimates: []CostEstimate{},
		Summary:   make(map[string]float64),
	}

	for _, network := range service.Spec.Infrastructure.Networks {
		estimate := estimateNetworkCost(network)
		if estimate != nil {
			report.Estimates = append(report.Estimates, *estimate)
			report.TotalMonthlyCost += estimate.MonthlyCost
			report.TotalHourlyCost += estimate.HourlyCost
		}
	}

	for _, sg := range service.Spec.Infrastructure.SecurityGroups {
		estimate := estimateSecurityGroupCost(sg)
		if estimate != nil {
			report.Estimates = append(report.Estimates, *estimate)
			report.TotalMonthlyCost += estimate.MonthlyCost
			report.TotalHourlyCost += estimate.HourlyCost
		}
	}

	for _, compute := range service.Spec.Infrastructure.Computes {
		estimate := estimateComputeCost(compute)
		if estimate != nil {
			report.Estimates = append(report.Estimates, *estimate)
			report.TotalMonthlyCost += estimate.MonthlyCost
			report.TotalHourlyCost += estimate.HourlyCost
		}
	}

	for _, cluster := range service.Spec.Infrastructure.KubernetesClusters {
		estimate := estimateKubernetesCost(cluster)
		if estimate != nil {
			report.Estimates = append(report.Estimates, *estimate)
			report.TotalMonthlyCost += estimate.MonthlyCost
			report.TotalHourlyCost += estimate.HourlyCost
		}
	}

	for _, estimate := range report.Estimates {
		report.Summary[estimate.ResourceType] += estimate.MonthlyCost
	}

	return report
}

func estimateNetworkCost(network parser.Network) *CostEstimate {
	estimate := &CostEstimate{
		ResourceType: "network",
		ResourceName: network.Name,
		Provider:     network.Provider,
		MonthlyCost:  0.0,
		HourlyCost:   0.0,
		Currency:     "USD",
		Details: map[string]interface{}{
			"cidr":    network.CIDR,
			"subnets": len(network.Subnets),
		},
	}

	return estimate
}

func estimateSecurityGroupCost(sg parser.SecurityGroup) *CostEstimate {
	return &CostEstimate{
		ResourceType: "security_group",
		ResourceName: sg.Name,
		Provider:     sg.Provider,
		MonthlyCost:  0.0,
		HourlyCost:   0.0,
		Currency:     "USD",
		Details: map[string]interface{}{
			"vpc":   sg.VPC,
			"rules": len(sg.Rules),
		},
	}
}

func estimateComputeCost(compute parser.Compute) *CostEstimate {
	provider := strings.ToLower(compute.Provider)

	var hourlyCost float64
	var instanceType string

	if spec, ok := compute.Spec["instance_type"].(string); ok {
		instanceType = spec
	} else if spec, ok := compute.Spec["size"].(string); ok {
		instanceType = spec
	} else if spec, ok := compute.Spec["machine_type"].(string); ok {
		instanceType = spec
	} else {
		switch compute.Type {
		case "ec2":
			instanceType = "t3.micro"
		case "azurerm_linux_virtual_machine":
			instanceType = "Standard_B1s"
		case "google_compute_instance":
			instanceType = "e2-micro"
		}
	}

	hourlyCost = 0.0
	storageCost := 0.0

	if provider == "aws_local" || provider == "azurerm_local" || provider == "google_local" {
		hourlyCost = 0.0
		storageCost = 0.0
	} else {
		switch provider {
		case "aws":
			if pricing, exists := defaultPricing.AWS["ec2"][instanceType]; exists {
				hourlyCost = pricing
			}
		case "azurerm":
			if pricing, exists := defaultPricing.Azure["vm"][instanceType]; exists {
				hourlyCost = pricing
			}
		case "google":
			if pricing, exists := defaultPricing.GCP["compute"][instanceType]; exists {
				hourlyCost = pricing
			}
		}

		if rootDiskSize, ok := compute.Spec["root_disk_size_gb"].(int); ok {
			storageCost = float64(rootDiskSize) * 0.10
		} else {
			storageCost = 20 * 0.10
		}
	}

	monthlyCost := (hourlyCost * 730) + storageCost

	estimate := &CostEstimate{
		ResourceType: "compute",
		ResourceName: compute.Name,
		Provider:     compute.Provider,
		MonthlyCost:  monthlyCost,
		HourlyCost:   hourlyCost,
		Currency:     "USD",
		Details: map[string]interface{}{
			"instance_type": instanceType,
			"vpc":           compute.VPC,
			"subnet":        compute.Subnet,
			"storage_gb":    20,
			"environment":   getEnvironmentFromProvider(provider),
		},
	}

	return estimate
}

func estimateKubernetesCost(cluster parser.KubernetesCluster) *CostEstimate {
	provider := strings.ToLower(cluster.Provider)

	var monthlyCost float64
	var clusterType string
	var nodeCount int
	var nodeType string

	if spec, ok := cluster.Spec["node_count"].(int); ok {
		nodeCount = spec
	} else {
		nodeCount = 2
	}

	if spec, ok := cluster.Spec["node_type"].(string); ok {
		nodeType = spec
	} else if spec, ok := cluster.Spec["node_size"].(string); ok {
		nodeType = spec
	} else if spec, ok := cluster.Spec["machine_type"].(string); ok {
		nodeType = spec
	}

	if provider == "aws_local" || provider == "azurerm_local" || provider == "google_local" {
		monthlyCost = 0.0
	} else {
		switch provider {
		case "aws":
			clusterType = "eks"
			if nodeType == "" {
				nodeType = "t3.medium"
			}
			if pricing, exists := defaultPricing.AWS["ec2"][nodeType]; exists {
				monthlyCost = pricing * 730 * float64(nodeCount)
			}
			monthlyCost += 73.0 // EKS control plane cost per month
		case "azurerm":
			clusterType = "aks"
			if nodeType == "" {
				nodeType = "Standard_B2s"
			}
			if pricing, exists := defaultPricing.Azure["vm"][nodeType]; exists {
				monthlyCost = pricing * 730 * float64(nodeCount)
			}
			monthlyCost += 73.0 // AKS control plane cost per month
		case "google":
			clusterType = "gke"
			if nodeType == "" {
				nodeType = "e2-medium"
			}
			if pricing, exists := defaultPricing.GCP["compute"][nodeType]; exists {
				monthlyCost = pricing * 730 * float64(nodeCount)
			}
			monthlyCost += 73.0 // GKE control plane cost per month
		}
	}

	estimate := &CostEstimate{
		ResourceType: "kubernetes",
		ResourceName: cluster.Name,
		Provider:     cluster.Provider,
		MonthlyCost:  monthlyCost,
		HourlyCost:   monthlyCost / 730,
		Currency:     "USD",
		Details: map[string]interface{}{
			"cluster_type": clusterType,
			"vpc":          cluster.VPC,
			"node_count":   nodeCount,
			"node_type":    nodeType,
			"environment":  getEnvironmentFromProvider(provider),
		},
	}

	return estimate
}

func getEnvironmentFromProvider(provider string) string {
	if strings.Contains(provider, "local") {
		return "local"
	}
	return "production"
}

func FormatCostReport(report *CostReport) string {
	var output strings.Builder

	output.WriteString("💰 Cost Estimation Report\n")
	output.WriteString("========================\n\n")

	output.WriteString(fmt.Sprintf("Total Monthly Cost: $%.2f USD\n", report.TotalMonthlyCost))
	output.WriteString(fmt.Sprintf("Total Hourly Cost:  $%.4f USD\n\n", report.TotalHourlyCost))

	output.WriteString("📊 Cost Breakdown by Resource Type:\n")
	output.WriteString("-----------------------------------\n")
	for resourceType, cost := range report.Summary {
		output.WriteString(fmt.Sprintf("%s: $%.2f/month\n", strings.Title(resourceType), cost))
	}

	output.WriteString("\n📋 Detailed Resource Costs:\n")
	output.WriteString("---------------------------\n")
	for _, estimate := range report.Estimates {
		output.WriteString(fmt.Sprintf("• %s (%s): $%.2f/month\n",
			estimate.ResourceName, estimate.ResourceType, estimate.MonthlyCost))
	}

	output.WriteString("\n💡 Cost Optimization Tips:\n")
	output.WriteString("-------------------------\n")
	output.WriteString("• Use spot/preemptible instances for non-critical workloads\n")
	output.WriteString("• Consider reserved instances for predictable workloads\n")
	output.WriteString("• Monitor and right-size instances based on usage\n")
	output.WriteString("• Use appropriate storage classes for your use case\n")

	return output.String()
}

func GetPricingData() *PricingData {
	return &defaultPricing
}
