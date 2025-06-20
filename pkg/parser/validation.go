package parser

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

// ValidationError represents validation errors
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// ValidationResult contains all validation errors
type ValidationResult struct {
	Errors []ValidationError
}

func (r *ValidationResult) AddError(field, message string) {
	r.Errors = append(r.Errors, ValidationError{Field: field, Message: message})
}

func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

func (r *ValidationResult) Error() string {
	if !r.HasErrors() {
		return ""
	}

	var messages []string
	for _, err := range r.Errors {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// ValidateService validates the entire service configuration
func ValidateService(service *Service) error {
	result := &ValidationResult{}

	// Validate metadata
	validateMetadata(service.Metadata, result)

	// Validate providers
	for i, provider := range service.Providers {
		validateProvider(provider, fmt.Sprintf("providers[%d]", i), result)
	}

	// Validate spec
	validateSpec(service.Spec, result)

	if result.HasErrors() {
		return result
	}

	return nil
}

func validateMetadata(metadata Metadata, result *ValidationResult) {
	if metadata.Name == "" {
		result.AddError("metadata.name", "name is required")
	}

	if metadata.Owner == "" {
		result.AddError("metadata.owner", "owner is required")
	}

	// Validate name format (alphanumeric, hyphens, underscores only)
	if metadata.Name != "" {
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, metadata.Name)
		if !matched {
			result.AddError("metadata.name", "name must contain only alphanumeric characters, hyphens, and underscores")
		}
	}
}

func validateProvider(provider Provider, path string, result *ValidationResult) {
	if provider.Name == "" {
		result.AddError(path+".name", "provider name is required")
	}

	if provider.Type == "" {
		result.AddError(path+".type", "provider type is required")
	}

	// Validate provider type
	validTypes := map[string]bool{
		"aws":     true,
		"azurerm": true,
		"google":  true,
	}

	if provider.Type != "" && !validTypes[provider.Type] {
		result.AddError(path+".type", fmt.Sprintf("unsupported provider type: %s", provider.Type))
	}

	// Validate provider name format
	if provider.Name != "" {
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, provider.Name)
		if !matched {
			result.AddError(path+".name", "provider name must contain only alphanumeric characters, hyphens, and underscores")
		}
	}
}

func validateSpec(spec Spec, result *ValidationResult) {
	// Validate key pair
	validateKeyPair(spec.KeyPair, "spec.key_pair", result)

	// Validate infrastructure
	validateInfrastructure(spec.Infrastructure, "spec.infrastructure", result)
}

func validateKeyPair(keyPair KeyPair, path string, result *ValidationResult) {
	if keyPair.Name == "" {
		result.AddError(path+".name", "key pair name is required")
	}

	if keyPair.PublicKeyFile != "" {
		if _, err := os.Stat(keyPair.PublicKeyFile); os.IsNotExist(err) {
			result.AddError(path+".public_key_file", fmt.Sprintf("public key file not found: %s", keyPair.PublicKeyFile))
		}
	}
}

func validateInfrastructure(infra Infrastructure, path string, result *ValidationResult) {
	// Validate networks
	for i, network := range infra.Networks {
		validateNetwork(network, fmt.Sprintf("%s.networks[%d]", path, i), result)
	}

	// Validate security groups
	for i, sg := range infra.SecurityGroups {
		validateSecurityGroup(sg, fmt.Sprintf("%s.security_groups[%d]", path, i), result)
	}

	// Validate computes
	for i, compute := range infra.Computes {
		validateCompute(compute, fmt.Sprintf("%s.computes[%d]", path, i), result)
	}

	// Validate peerings
	for i, peering := range infra.Peerings {
		validatePeering(peering, fmt.Sprintf("%s.peerings[%d]", path, i), result)
	}
}

func validateNetwork(network Network, path string, result *ValidationResult) {
	if network.Name == "" {
		result.AddError(path+".name", "network name is required")
	}

	if network.Provider == "" {
		result.AddError(path+".provider", "network provider is required")
	}

	if !validateCIDR(network.CIDR) {
		result.AddError(path+".cidr", fmt.Sprintf("invalid CIDR: %s", network.CIDR))
	}

	// Validate subnets
	for i, subnet := range network.Subnets {
		validateSubnet(subnet, fmt.Sprintf("%s.subnets[%d]", path, i), result)
	}
}

func validateSubnet(subnet Subnet, path string, result *ValidationResult) {
	if subnet.Name == "" {
		result.AddError(path+".name", "subnet name is required")
	}

	if subnet.Zone == "" {
		result.AddError(path+".zone", "subnet zone is required")
	}

	if !validateCIDR(subnet.CIDR) {
		result.AddError(path+".cidr", fmt.Sprintf("invalid CIDR: %s", subnet.CIDR))
	}
}

func validateSecurityGroup(sg SecurityGroup, path string, result *ValidationResult) {
	if sg.Name == "" {
		result.AddError(path+".name", "security group name is required")
	}

	if sg.Provider == "" {
		result.AddError(path+".provider", "security group provider is required")
	}

	if sg.VPC == "" {
		result.AddError(path+".vpc", "security group VPC is required")
	}

	// Validate rules
	for i, rule := range sg.Rules {
		validateSecurityGroupRule(rule, fmt.Sprintf("%s.rules[%d]", path, i), result)
	}
}

func validateSecurityGroupRule(rule SecurityGroupRule, path string, result *ValidationResult) {
	if rule.Type == "" {
		result.AddError(path+".type", "rule type is required")
	}

	if rule.Type != "" && rule.Type != "ingress" && rule.Type != "egress" {
		result.AddError(path+".type", "rule type must be 'ingress' or 'egress'")
	}

	if rule.Protocol == "" {
		result.AddError(path+".protocol", "rule protocol is required")
	}

	// Validate port ranges
	if rule.FromPort < 0 || rule.FromPort > 65535 {
		result.AddError(path+".from_port", "from_port must be between 0 and 65535")
	}

	if rule.ToPort < 0 || rule.ToPort > 65535 {
		result.AddError(path+".to_port", "to_port must be between 0 and 65535")
	}

	if rule.FromPort > rule.ToPort {
		result.AddError(path+".from_port", "from_port cannot be greater than to_port")
	}

	// Validate CIDR blocks
	for i, cidr := range rule.CIDRBlocks {
		if !validateCIDR(cidr) {
			result.AddError(fmt.Sprintf("%s.cidr_blocks[%d]", path, i), fmt.Sprintf("invalid CIDR: %s", cidr))
		}
	}
}

func validateCompute(compute Compute, path string, result *ValidationResult) {
	if compute.Name == "" {
		result.AddError(path+".name", "compute name is required")
	}

	if compute.Type == "" {
		result.AddError(path+".type", "compute type is required")
	}

	if compute.Provider == "" {
		result.AddError(path+".provider", "compute provider is required")
	}

	if compute.VPC == "" {
		result.AddError(path+".vpc", "compute VPC is required")
	}

	if compute.Subnet == "" {
		result.AddError(path+".subnet", "compute subnet is required")
	}

	// Validate storage
	for i, storage := range compute.Storage {
		validateStorage(storage, fmt.Sprintf("%s.storage[%d]", path, i), result)
	}
}

func validateStorage(storage Storage, path string, result *ValidationResult) {
	if storage.Name == "" {
		result.AddError(path+".name", "storage name is required")
	}

	if storage.Size <= 0 {
		result.AddError(path+".size", "storage size must be greater than 0")
	}

	if storage.Type == "" {
		result.AddError(path+".type", "storage type is required")
	}
}

func validatePeering(peering Peering, path string, result *ValidationResult) {
	if peering.Name == "" {
		result.AddError(path+".name", "peering name is required")
	}

	if peering.Provider == "" {
		result.AddError(path+".provider", "peering provider is required")
	}

	if peering.VPCRequester == "" {
		result.AddError(path+".vpc_requester", "VPC requester is required")
	}

	if peering.VPCAccepter == "" {
		result.AddError(path+".vpc_accepter", "VPC accepter is required")
	}

	if peering.VPCRequester == peering.VPCAccepter {
		result.AddError(path+".vpc_accepter", "VPC requester and accepter cannot be the same")
	}
}

// validateCIDR validates CIDR notation
func validateCIDR(cidr string) bool {
	if cidr == "" {
		return false
	}

	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}
