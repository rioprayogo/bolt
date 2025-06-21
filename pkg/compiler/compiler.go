// =============================================================================
// Bolt v2: Multi-Provider Support (AWS, Azure, GCP)
// =============================================================================
package compiler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bold/pkg/logger"
	"bold/pkg/parser"
)

// CompileToTofu generates the OpenTofu JSON configuration from the service manifest
func CompileToTofu(service *parser.Service, boltBuildPath string) error {
	logger.Info("Starting OpenTofu compilation", logger.Fields{
		"service_name": service.Metadata.Name,
		"build_path":   boltBuildPath,
		"providers":    len(service.Providers),
	})

	if err := os.MkdirAll(boltBuildPath, 0755); err != nil {
		logger.LogError(err, "creating build directory", logger.Fields{
			"build_path": boltBuildPath,
		})
		return fmt.Errorf("failed to create build directory: %w", err)
	}

	resources := make(map[string]interface{})
	providers := make(map[string]interface{})

	for _, provider := range service.Providers {
		switch provider.Type {
		case "aws":
			providers["aws"] = generateAWSProviderConfig(provider)
			awsResources := processAWSResources(service, provider)
			for resourceType, resourceMap := range awsResources {
				if resources[resourceType] == nil {
					resources[resourceType] = make(map[string]interface{})
				}
				for resourceName, resourceConfig := range resourceMap.(map[string]interface{}) {
					resources[resourceType].(map[string]interface{})[resourceName] = resourceConfig
				}
			}
		case "azurerm":
			providers["azurerm"] = generateAzureProviderConfig(provider)
			azureResources := processAzureResources(service, provider)
			for resourceType, resourceMap := range azureResources {
				if resources[resourceType] == nil {
					resources[resourceType] = make(map[string]interface{})
				}
				for resourceName, resourceConfig := range resourceMap.(map[string]interface{}) {
					resources[resourceType].(map[string]interface{})[resourceName] = resourceConfig
				}
			}
		case "google":
			providers["google"] = generateGCPProviderConfig(provider)
			gcpResources := processGCPResources(service, provider)
			for resourceType, resourceMap := range gcpResources {
				if resources[resourceType] == nil {
					resources[resourceType] = make(map[string]interface{})
				}
				for resourceName, resourceConfig := range resourceMap.(map[string]interface{}) {
					resources[resourceType].(map[string]interface{})[resourceName] = resourceConfig
				}
			}
		}
	}

	kubernetesResources := processKubernetesResources(service)
	for resourceType, resourceMap := range kubernetesResources {
		if resources[resourceType] == nil {
			resources[resourceType] = make(map[string]interface{})
		}
		for resourceName, resourceConfig := range resourceMap.(map[string]interface{}) {
			resources[resourceType].(map[string]interface{})[resourceName] = resourceConfig
		}
	}

	config := map[string]interface{}{
		"terraform": map[string]interface{}{
			"required_providers": map[string]interface{}{
				"aws": map[string]interface{}{
					"source":  "hashicorp/aws",
					"version": "~> 5.0",
				},
				"azurerm": map[string]interface{}{
					"source":  "hashicorp/azurerm",
					"version": "~> 3.0",
				},
				"google": map[string]interface{}{
					"source":  "hashicorp/google",
					"version": "~> 5.0",
				},
			},
		},
		"provider": providers,
		"resource": resources,
	}

	outputPath := filepath.Join(boltBuildPath, "main.tf.json")
	if err := writeToFile(config, outputPath); err != nil {
		logger.LogError(err, "writing OpenTofu configuration", logger.Fields{
			"output_path": outputPath,
		})
		return fmt.Errorf("failed to write OpenTofu configuration: %w", err)
	}

	logger.Info("OpenTofu configuration generated successfully", logger.Fields{
		"output_path": outputPath,
	})

	return nil
}

func processAWSResources(service *parser.Service, provider parser.Provider) map[string]interface{} {
	awsResources := make(map[string]interface{})

	if service.Spec.KeyPair.Name != "" && service.Spec.KeyPair.PublicKeyFile != "" {
		publicKeyBytes, err := os.ReadFile(service.Spec.KeyPair.PublicKeyFile)
		if err == nil {
			publicKey := string(publicKeyBytes)
			keyPairResourceName := strings.ReplaceAll(service.Spec.KeyPair.Name, " ", "_")
			if awsResources["aws_key_pair"] == nil {
				awsResources["aws_key_pair"] = make(map[string]interface{})
			}
			awsResources["aws_key_pair"].(map[string]interface{})[keyPairResourceName] = map[string]interface{}{
				"key_name":   service.Spec.KeyPair.Name,
				"public_key": publicKey,
				"tags":       mergeTags(service.Metadata.Tags, map[string]string{"Name": service.Spec.KeyPair.Name}),
			}
		}
	}

	for _, network := range service.Spec.Infrastructure.Networks {
		if network.Provider == provider.Name {
			vpcName := network.Name
			if awsResources["aws_vpc"] == nil {
				awsResources["aws_vpc"] = make(map[string]interface{})
			}
			awsResources["aws_vpc"].(map[string]interface{})[vpcName] = map[string]interface{}{
				"cidr_block": network.CIDR,
				"tags":       mergeTags(service.Metadata.Tags, map[string]string{"Name": vpcName}),
			}

			for _, subnet := range network.Subnets {
				subnetName := subnet.Name
				if awsResources["aws_subnet"] == nil {
					awsResources["aws_subnet"] = make(map[string]interface{})
				}
				awsResources["aws_subnet"].(map[string]interface{})[subnetName] = map[string]interface{}{
					"vpc_id":            fmt.Sprintf("${aws_vpc.%s.id}", vpcName),
					"cidr_block":        subnet.CIDR,
					"availability_zone": subnet.Zone,
					"tags":              mergeTags(service.Metadata.Tags, map[string]string{"Name": subnetName}),
				}
			}
		}
	}

	for _, sg := range service.Spec.Infrastructure.SecurityGroups {
		if sg.Provider == provider.Name {
			sgName := sg.Name
			if awsResources["aws_security_group"] == nil {
				awsResources["aws_security_group"] = make(map[string]interface{})
			}

			sgConfig := map[string]interface{}{
				"name":        sgName,
				"description": fmt.Sprintf("Security group for %s", sgName),
				"vpc_id":      fmt.Sprintf("${aws_vpc.%s.id}", sg.VPC),
				"tags":        mergeTags(service.Metadata.Tags, map[string]string{"Name": sgName}),
			}

			var ingressRules []map[string]interface{}
			var egressRules []map[string]interface{}

			for _, rule := range sg.Rules {
				ruleConfig := map[string]interface{}{
					"protocol":         rule.Protocol,
					"from_port":        rule.FromPort,
					"to_port":          rule.ToPort,
					"description":      "",
					"ipv6_cidr_blocks": []string{},
					"prefix_list_ids":  []string{},
					"security_groups":  []string{},
					"self":             false,
				}

				if len(rule.CIDRBlocks) > 0 {
					ruleConfig["cidr_blocks"] = rule.CIDRBlocks
				} else {
					ruleConfig["cidr_blocks"] = []string{"0.0.0.0/0"}
				}

				if rule.Type == "ingress" {
					ingressRules = append(ingressRules, ruleConfig)
				} else if rule.Type == "egress" {
					egressRules = append(egressRules, ruleConfig)
				}
			}

			if len(ingressRules) > 0 {
				sgConfig["ingress"] = ingressRules
			}
			if len(egressRules) > 0 {
				sgConfig["egress"] = egressRules
			}

			awsResources["aws_security_group"].(map[string]interface{})[sgName] = sgConfig
		}
	}

	for _, compute := range service.Spec.Infrastructure.Computes {
		if compute.Provider == provider.Name && compute.Type == "ec2" {
			vmName := compute.Name
			instance := map[string]interface{}{
				"ami":           "ami-local",
				"instance_type": "t2.micro",
				"subnet_id":     fmt.Sprintf("${aws_subnet.%s.id}", compute.Subnet),
				"tags":          mergeTags(service.Metadata.Tags, map[string]string{"Name": vmName}),
			}

			if spec, ok := compute.Spec["instance_type"].(string); ok {
				instance["instance_type"] = spec
			}
			if rootDiskSize, ok := compute.Spec["root_disk_size_gb"].(int); ok {
				instance["root_block_device"] = map[string]interface{}{
					"volume_size": rootDiskSize,
				}
			}

			if service.Spec.KeyPair.Name != "" {
				instance["key_name"] = service.Spec.KeyPair.Name
			}

			if compute.SecurityGroup != "" {
				instance["vpc_security_group_ids"] = []string{fmt.Sprintf("${aws_security_group.%s.id}", compute.SecurityGroup)}
			}

			if awsResources["aws_instance"] == nil {
				awsResources["aws_instance"] = make(map[string]interface{})
			}
			awsResources["aws_instance"].(map[string]interface{})[vmName] = instance
		}
	}

	return awsResources
}

func processAzureResources(service *parser.Service, provider parser.Provider) map[string]interface{} {
	azureResources := make(map[string]interface{})

	for _, network := range service.Spec.Infrastructure.Networks {
		if network.Provider == provider.Name {
			vnetName := network.Name
			if azureResources["azurerm_virtual_network"] == nil {
				azureResources["azurerm_virtual_network"] = make(map[string]interface{})
			}
			azureResources["azurerm_virtual_network"].(map[string]interface{})[vnetName] = map[string]interface{}{
				"name":                vnetName,
				"resource_group_name": "rg-" + vnetName,
				"location":            getProviderRegion(provider),
				"address_space":       []string{network.CIDR},
				"tags":                mergeTags(service.Metadata.Tags, map[string]string{"Name": vnetName}),
			}

			for _, subnet := range network.Subnets {
				subnetName := subnet.Name
				if azureResources["azurerm_subnet"] == nil {
					azureResources["azurerm_subnet"] = make(map[string]interface{})
				}

				subnetConfig := map[string]interface{}{
					"name":                 subnetName,
					"resource_group_name":  "rg-" + vnetName,
					"virtual_network_name": vnetName,
					"address_prefixes":     []string{subnet.CIDR},
				}

				azureResources["azurerm_subnet"].(map[string]interface{})[subnetName] = subnetConfig
			}
		}
	}

	for _, sg := range service.Spec.Infrastructure.SecurityGroups {
		if sg.Provider == provider.Name {
			sgName := sg.Name
			if azureResources["azurerm_network_security_group"] == nil {
				azureResources["azurerm_network_security_group"] = make(map[string]interface{})
			}

			azureResources["azurerm_network_security_group"].(map[string]interface{})[sgName] = map[string]interface{}{
				"name":                sgName,
				"resource_group_name": "rg-" + sg.VPC,
				"location":            getProviderRegion(provider),
				"tags":                mergeTags(service.Metadata.Tags, map[string]string{"Name": sgName}),
			}

			if azureResources["azurerm_network_security_rule"] == nil {
				azureResources["azurerm_network_security_rule"] = make(map[string]interface{})
			}

			for i, rule := range sg.Rules {
				ruleName := fmt.Sprintf("%s-rule-%d", sgName, i)
				ruleConfig := map[string]interface{}{
					"name":                        ruleName,
					"resource_group_name":         "rg-" + sg.VPC,
					"network_security_group_name": sgName,
					"priority":                    100 + i,
					"access":                      "Allow",
					"source_port_range":           "*",
					"destination_port_range":      fmt.Sprintf("%d-%d", rule.FromPort, rule.ToPort),
				}

				if rule.Type == "ingress" {
					ruleConfig["direction"] = "Inbound"
				} else {
					ruleConfig["direction"] = "Outbound"
				}

				if rule.Protocol == "tcp" {
					ruleConfig["protocol"] = "Tcp"
				} else if rule.Protocol == "udp" {
					ruleConfig["protocol"] = "Udp"
				} else {
					ruleConfig["protocol"] = "*"
				}

				if len(rule.CIDRBlocks) > 0 {
					ruleConfig["source_address_prefix"] = rule.CIDRBlocks[0]
				} else {
					ruleConfig["source_address_prefix"] = "*"
				}

				ruleConfig["destination_address_prefix"] = "*"

				azureResources["azurerm_network_security_rule"].(map[string]interface{})[ruleName] = ruleConfig
			}
		}
	}

	for _, compute := range service.Spec.Infrastructure.Computes {
		if compute.Provider == provider.Name && compute.Type == "azurerm_linux_virtual_machine" {
			vmName := compute.Name
			vm := map[string]interface{}{
				"name":                vmName,
				"resource_group_name": "rg-" + compute.VPC,
				"location":            getProviderRegion(provider),
				"size":                "Standard_B1s",
				"admin_username":      "boltadmin",
				"tags":                mergeTags(service.Metadata.Tags, map[string]string{"Name": vmName}),
				"source_image_reference": map[string]interface{}{
					"publisher": "Canonical",
					"offer":     "UbuntuServer",
					"sku":       "18.04-LTS",
					"version":   "latest",
				},
			}

			if size, ok := compute.Spec["size"].(string); ok {
				vm["size"] = size
			}
			if username, ok := compute.Spec["username"].(string); ok {
				vm["admin_username"] = username
			}
			if password, ok := compute.Spec["password"].(string); ok {
				vm["admin_password"] = password
			}

			if image, ok := compute.Spec["image"].(map[string]interface{}); ok {
				vm["source_image_reference"] = map[string]interface{}{
					"publisher": image["publisher"],
					"offer":     image["offer"],
					"sku":       image["sku"],
					"version":   "latest",
				}
			}

			vm["os_disk"] = []map[string]interface{}{{
				"caching":              "ReadWrite",
				"storage_account_type": "Standard_LRS",
			}}

			vm["network_interface_ids"] = []string{fmt.Sprintf("${azurerm_network_interface.%s.id}", vmName+"-nic")}

			if azureResources["azurerm_linux_virtual_machine"] == nil {
				azureResources["azurerm_linux_virtual_machine"] = make(map[string]interface{})
			}
			azureResources["azurerm_linux_virtual_machine"].(map[string]interface{})[vmName] = vm

			if azureResources["azurerm_network_interface"] == nil {
				azureResources["azurerm_network_interface"] = make(map[string]interface{})
			}

			nicConfig := map[string]interface{}{
				"name":                vmName + "-nic",
				"resource_group_name": "rg-" + compute.VPC,
				"location":            getProviderRegion(provider),
				"ip_configuration": []map[string]interface{}{{
					"name":                          "internal",
					"subnet_id":                     fmt.Sprintf("${azurerm_subnet.%s.id}", compute.Subnet),
					"private_ip_address_allocation": "Dynamic",
				}},
			}

			azureResources["azurerm_network_interface"].(map[string]interface{})[vmName+"-nic"] = nicConfig
		}
	}

	return azureResources
}

func processGCPResources(service *parser.Service, provider parser.Provider) map[string]interface{} {
	gcpResources := make(map[string]interface{})

	for _, network := range service.Spec.Infrastructure.Networks {
		if network.Provider == provider.Name {
			vpcName := network.Name
			if gcpResources["google_compute_network"] == nil {
				gcpResources["google_compute_network"] = make(map[string]interface{})
			}
			gcpResources["google_compute_network"].(map[string]interface{})[vpcName] = map[string]interface{}{
				"name":                    vpcName,
				"auto_create_subnetworks": false,
			}

			for _, subnet := range network.Subnets {
				subnetName := subnet.Name
				if gcpResources["google_compute_subnetwork"] == nil {
					gcpResources["google_compute_subnetwork"] = make(map[string]interface{})
				}
				gcpResources["google_compute_subnetwork"].(map[string]interface{})[subnetName] = map[string]interface{}{
					"name":          subnetName,
					"ip_cidr_range": subnet.CIDR,
					"network":       fmt.Sprintf("${google_compute_network.%s.self_link}", vpcName),
					"region":        getProviderRegion(provider),
				}
			}
		}
	}

	for _, sg := range service.Spec.Infrastructure.SecurityGroups {
		if sg.Provider == provider.Name {
			sgName := sg.Name
			if gcpResources["google_compute_firewall"] == nil {
				gcpResources["google_compute_firewall"] = make(map[string]interface{})
			}

			for i, rule := range sg.Rules {
				ruleName := fmt.Sprintf("%s-rule-%d", sgName, i)
				ruleConfig := map[string]interface{}{
					"name":    ruleName,
					"network": fmt.Sprintf("${google_compute_network.%s.self_link}", sg.VPC),
				}

				if rule.Type == "ingress" {
					ruleConfig["direction"] = "INGRESS"
					ruleConfig["source_ranges"] = rule.CIDRBlocks
					ruleConfig["target_tags"] = []string{sgName}
				} else {
					ruleConfig["direction"] = "EGRESS"
					ruleConfig["destination_ranges"] = rule.CIDRBlocks
					ruleConfig["target_tags"] = []string{sgName}
				}

				if rule.Protocol == "tcp" || rule.Protocol == "udp" {
					ruleConfig["allow"] = []map[string]interface{}{{
						"protocol": rule.Protocol,
						"ports":    []string{fmt.Sprintf("%d-%d", rule.FromPort, rule.ToPort)},
					}}
				} else {
					ruleConfig["allow"] = []map[string]interface{}{{
						"protocol": rule.Protocol,
					}}
				}

				gcpResources["google_compute_firewall"].(map[string]interface{})[ruleName] = ruleConfig
			}
		}
	}

	for _, compute := range service.Spec.Infrastructure.Computes {
		if compute.Provider == provider.Name && compute.Type == "google_compute_instance" {
			vmName := compute.Name
			vm := map[string]interface{}{
				"name":         vmName,
				"machine_type": "e2-medium",
				"zone":         getProviderZone(provider),
				"boot_disk": []map[string]interface{}{{
					"initialize_params": []map[string]interface{}{{
						"image": "debian-cloud/debian-11",
						"size":  20,
					}},
				}},
			}

			if machineType, ok := compute.Spec["machine_type"].(string); ok {
				vm["machine_type"] = machineType
			}
			if zone, ok := compute.Spec["zone"].(string); ok {
				vm["zone"] = zone
			}

			if image, ok := compute.Spec["image"].(string); ok {
				vm["boot_disk"] = []map[string]interface{}{{
					"initialize_params": []map[string]interface{}{{
						"image": image,
					}},
				}}
			}

			networkInterface := map[string]interface{}{
				"subnetwork": fmt.Sprintf("${google_compute_subnetwork.%s.self_link}", compute.Subnet),
			}

			if compute.SecurityGroup != "" {
				networkInterface["access_config"] = []map[string]interface{}{{
					"network_tier": "STANDARD",
				}}
				vm["tags"] = []string{compute.SecurityGroup}
			}

			vm["network_interface"] = []map[string]interface{}{networkInterface}

			if gcpResources["google_compute_instance"] == nil {
				gcpResources["google_compute_instance"] = make(map[string]interface{})
			}
			gcpResources["google_compute_instance"].(map[string]interface{})[vmName] = vm
		}
	}

	return gcpResources
}

func generateAWSProviderConfig(provider parser.Provider) map[string]interface{} {
	config := map[string]interface{}{}

	if region, ok := provider.Spec["region"].(string); ok {
		config["region"] = region
	}

	if env, ok := provider.Spec["environment"].(string); ok && env == "local" {
		config["access_key"] = "test"
		config["secret_key"] = "test"
		config["s3_use_path_style"] = true
		config["skip_credentials_validation"] = true
		config["skip_requesting_account_id"] = true
		config["skip_metadata_api_check"] = true
		config["endpoints"] = map[string]string{
			"ec2": "http://localhost:4566",
			"eks": "http://localhost:4566",
			"iam": "http://localhost:4566",
			"sts": "http://localhost:4566",
			"s3":  "http://localhost:4566",
		}
	}

	return config
}

func generateAzureProviderConfig(provider parser.Provider) map[string]interface{} {
	config := map[string]interface{}{
		"features": map[string]interface{}{},
	}

	return config
}

func generateGCPProviderConfig(provider parser.Provider) map[string]interface{} {
	config := map[string]interface{}{}

	if project, ok := provider.Spec["project"].(string); ok {
		config["project"] = project
	} else {
		config["project"] = "your-gcp-project-id"
	}

	if region, ok := provider.Spec["region"].(string); ok {
		config["region"] = region
	}

	if zone, ok := provider.Spec["zone"].(string); ok {
		config["zone"] = zone
	}

	return config
}

func getProviderRegion(provider parser.Provider) string {
	if region, ok := provider.Spec["region"].(string); ok {
		return region
	}
	return "us-east-1"
}

func getProviderZone(provider parser.Provider) string {
	if zone, ok := provider.Spec["zone"].(string); ok {
		return zone
	}
	return "us-central1-a"
}

func mergeTags(globalTags, resourceTags map[string]string) map[string]string {
	merged := make(map[string]string)
	for k, v := range globalTags {
		merged[k] = v
	}
	for k, v := range resourceTags {
		merged[k] = v
	}
	return merged
}

func writeToFile(config map[string]interface{}, path string) error {
	bytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(path, bytes, 0644)
}

func processKubernetesResources(service *parser.Service) map[string]interface{} {
	kubernetesResources := make(map[string]interface{})

	for _, cluster := range service.Spec.Infrastructure.KubernetesClusters {
		switch cluster.Provider {
		case "aws", "aws_local":
			processEKSCluster(cluster, kubernetesResources)
		case "azurerm", "azurerm_local":
			processAKSCluster(cluster, kubernetesResources)
		case "google", "google_local":
			processGKECluster(cluster, kubernetesResources)
		}
	}

	return kubernetesResources
}

func processEKSCluster(cluster parser.KubernetesCluster, resources map[string]interface{}) {
	clusterName := cluster.Name
	vpcName := cluster.VPC

	version := getStringSpec(cluster.Spec, "version", "1.28")
	nodeType := getStringSpec(cluster.Spec, "node_type", "t3.medium")
	nodeCount := getIntSpec(cluster.Spec, "node_count", 2)
	nodeDiskSize := getIntSpec(cluster.Spec, "node_disk_size_gb", 20)

	if resources["aws_eks_cluster"] == nil {
		resources["aws_eks_cluster"] = make(map[string]interface{})
	}

	resources["aws_eks_cluster"].(map[string]interface{})[clusterName] = map[string]interface{}{
		"name":     clusterName,
		"role_arn": fmt.Sprintf("${aws_iam_role.%s_cluster_role.arn}", clusterName),
		"version":  version,
		"vpc_config": map[string]interface{}{
			"subnet_ids": []string{
				fmt.Sprintf("${aws_subnet.%s_public.id}", vpcName),
				fmt.Sprintf("${aws_subnet.%s_private.id}", vpcName),
			},
			"endpoint_private_access": true,
			"endpoint_public_access":  true,
		},
		"depends_on": []string{
			fmt.Sprintf("aws_iam_role_policy_attachment.%s_cluster_policy", clusterName),
			fmt.Sprintf("aws_iam_role_policy_attachment.%s_vpc_resource_controller", clusterName),
		},
		"tags": map[string]string{
			"Name": clusterName,
		},
	}

	if resources["aws_eks_node_group"] == nil {
		resources["aws_eks_node_group"] = make(map[string]interface{})
	}

	resources["aws_eks_node_group"].(map[string]interface{})[clusterName] = map[string]interface{}{
		"cluster_name":    fmt.Sprintf("${aws_eks_cluster.%s.name}", clusterName),
		"node_group_name": fmt.Sprintf("%s-nodes", clusterName),
		"node_role_arn":   fmt.Sprintf("${aws_iam_role.%s_node_role.arn}", clusterName),
		"subnet_ids":      []string{fmt.Sprintf("${aws_subnet.%s_private.id}", vpcName)},
		"instance_types":  []string{nodeType},
		"scaling_config": map[string]interface{}{
			"desired_size": nodeCount,
			"max_size":     nodeCount,
			"min_size":     1,
		},
		"disk_size": nodeDiskSize,
		"depends_on": []string{
			fmt.Sprintf("aws_iam_role_policy_attachment.%s_worker_node_policy", clusterName),
			fmt.Sprintf("aws_iam_role_policy_attachment.%s_cni_policy", clusterName),
			fmt.Sprintf("aws_iam_role_policy_attachment.%s_ecr_read_only", clusterName),
		},
		"tags": map[string]string{
			"Name": fmt.Sprintf("%s-nodes", clusterName),
		},
	}
}

func processAKSCluster(cluster parser.KubernetesCluster, resources map[string]interface{}) {
	clusterName := cluster.Name
	nodeCount := getIntSpec(cluster.Spec, "node_count", 2)
	nodeSize := getStringSpec(cluster.Spec, "node_size", "Standard_B2s")

	if resources["azurerm_kubernetes_cluster"] == nil {
		resources["azurerm_kubernetes_cluster"] = make(map[string]interface{})
	}

	resources["azurerm_kubernetes_cluster"].(map[string]interface{})[clusterName] = map[string]interface{}{
		"name":                clusterName,
		"location":            fmt.Sprintf("${azurerm_resource_group.%s.location}", clusterName),
		"resource_group_name": fmt.Sprintf("${azurerm_resource_group.%s.name}", clusterName),
		"dns_prefix":          clusterName,
		"default_node_pool": map[string]interface{}{
			"name":       "default",
			"node_count": nodeCount,
			"vm_size":    nodeSize,
		},
		"identity": map[string]interface{}{
			"type": "SystemAssigned",
		},
		"network_profile": map[string]interface{}{
			"network_plugin": "azure",
			"network_policy": "azure",
		},
		"tags": map[string]string{
			"Name": clusterName,
		},
	}

	if resources["azurerm_resource_group"] == nil {
		resources["azurerm_resource_group"] = make(map[string]interface{})
	}

	resources["azurerm_resource_group"].(map[string]interface{})[clusterName] = map[string]interface{}{
		"name":     fmt.Sprintf("%s-rg", clusterName),
		"location": "eastus",
	}
}

func processGKECluster(cluster parser.KubernetesCluster, resources map[string]interface{}) {
	clusterName := cluster.Name
	vpcName := cluster.VPC
	nodeCount := getIntSpec(cluster.Spec, "node_count", 2)
	machineType := getStringSpec(cluster.Spec, "machine_type", "e2-medium")

	if resources["google_container_cluster"] == nil {
		resources["google_container_cluster"] = make(map[string]interface{})
	}

	resources["google_container_cluster"].(map[string]interface{})[clusterName] = map[string]interface{}{
		"name":                     clusterName,
		"location":                 "us-central1",
		"remove_default_node_pool": true,
		"initial_node_count":       1,
		"network":                  fmt.Sprintf("${google_compute_network.%s.name}", vpcName),
		"subnetwork":               fmt.Sprintf("${google_compute_subnetwork.subnet-private-1a.name}", vpcName),
		"ip_allocation_policy": map[string]interface{}{
			"cluster_ipv4_cidr_block":  "/16",
			"services_ipv4_cidr_block": "/22",
		},
		"private_cluster_config": map[string]interface{}{
			"enable_private_nodes":    true,
			"enable_private_endpoint": false,
			"master_ipv4_cidr_block":  "172.16.0.0/28",
		},
	}

	if resources["google_container_node_pool"] == nil {
		resources["google_container_node_pool"] = make(map[string]interface{})
	}

	resources["google_container_node_pool"].(map[string]interface{})[clusterName] = map[string]interface{}{
		"name":       fmt.Sprintf("%s-node-pool", clusterName),
		"location":   fmt.Sprintf("${google_container_cluster.%s.location}", clusterName),
		"cluster":    fmt.Sprintf("${google_container_cluster.%s.name}", clusterName),
		"node_count": nodeCount,
		"node_config": map[string]interface{}{
			"machine_type": machineType,
			"disk_size_gb": 20,
			"oauth_scopes": []string{
				"https://www.googleapis.com/auth/logging.write",
				"https://www.googleapis.com/auth/monitoring",
			},
			"metadata": map[string]string{
				"disable-legacy-endpoints": "true",
			},
		},
	}
}

func getStringSpec(spec map[string]interface{}, key, defaultValue string) string {
	if value, ok := spec[key].(string); ok {
		return value
	}
	return defaultValue
}

func getIntSpec(spec map[string]interface{}, key string, defaultValue int) int {
	if value, ok := spec[key].(int); ok {
		return value
	}
	return defaultValue
}
