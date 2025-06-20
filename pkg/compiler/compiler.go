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
					"protocol":  rule.Protocol,
					"from_port": rule.FromPort,
					"to_port":   rule.ToPort,
				}

				if len(rule.CIDRBlocks) > 0 {
					ruleConfig["cidr_blocks"] = rule.CIDRBlocks
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
				azureResources["azurerm_subnet"].(map[string]interface{})[subnetName] = map[string]interface{}{
					"name":                 subnetName,
					"resource_group_name":  "rg-" + vnetName,
					"virtual_network_name": vnetName,
					"address_prefixes":     []string{subnet.CIDR},
				}
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
			azureResources["azurerm_network_interface"].(map[string]interface{})[vmName+"-nic"] = map[string]interface{}{
				"name":                vmName + "-nic",
				"resource_group_name": "rg-" + compute.VPC,
				"location":            getProviderRegion(provider),
				"subnet_id":           fmt.Sprintf("${azurerm_subnet.%s.id}", compute.Subnet),
			}
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

	for _, compute := range service.Spec.Infrastructure.Computes {
		if compute.Provider == provider.Name && compute.Type == "google_compute_instance" {
			vmName := compute.Name
			vm := map[string]interface{}{
				"name":         vmName,
				"machine_type": "e2-medium",
				"zone":         getProviderZone(provider),
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

			vm["network_interface"] = []map[string]interface{}{{
				"subnetwork": fmt.Sprintf("${google_compute_subnetwork.%s.self_link}", compute.Subnet),
			}}

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
	config := map[string]interface{}{}

	if region, ok := provider.Spec["region"].(string); ok {
		config["region"] = region
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
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}
	return os.WriteFile(path, bytes, 0644)
}
