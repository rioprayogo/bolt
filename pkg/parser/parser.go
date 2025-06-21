package parser

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Service represents the Go structure of the service.yaml file
type Service struct {
	APIVersion string     `yaml:"apiVersion"`
	Kind       string     `yaml:"kind"`
	Metadata   Metadata   `yaml:"metadata"`
	Providers  []Provider `yaml:"providers"`
	Spec       Spec       `yaml:"spec"`
}

type Metadata struct {
	Name  string            `yaml:"name"`
	Owner string            `yaml:"owner"`
	Tags  map[string]string `yaml:"tags"`
}

type Provider struct {
	Name string                 `yaml:"name"`
	Type string                 `yaml:"type"`
	Spec map[string]interface{} `yaml:"spec"`
}

type KeyPair struct {
	Name          string `yaml:"name"`
	PublicKeyFile string `yaml:"public_key_file"`
	UseExisting   bool   `yaml:"use_existing"`
}

type Infrastructure struct {
	KeyPair            KeyPair             `yaml:"key_pair"`
	Networks           []Network           `yaml:"networks"`
	Peerings           []Peering           `yaml:"peerings"`
	SecurityGroups     []SecurityGroup     `yaml:"security_groups"`
	KubernetesClusters []KubernetesCluster `yaml:"kubernetes_clusters"`
	Computes           []Compute           `yaml:"computes"`
}

type Spec struct {
	Provider       Provider       `yaml:"provider"`
	KeyPair        KeyPair        `yaml:"key_pair"`
	Infrastructure Infrastructure `yaml:"infrastructure"`
}

type Network struct {
	Name     string   `yaml:"name"`
	Provider string   `yaml:"provider"`
	CIDR     string   `yaml:"cidr"`
	Subnets  []Subnet `yaml:"subnets"`
}

type Subnet struct {
	Name string `yaml:"name"`
	Zone string `yaml:"zone"`
	CIDR string `yaml:"cidr"`
}

type Peering struct {
	Name         string `yaml:"name"`
	Provider     string `yaml:"provider"`
	VPCRequester string `yaml:"vpc_requester"`
	VPCAccepter  string `yaml:"vpc_accepter"`
}

type SecurityGroup struct {
	Name     string              `yaml:"name"`
	Provider string              `yaml:"provider"`
	VPC      string              `yaml:"vpc"`
	Rules    []SecurityGroupRule `yaml:"rules"`
}

type SecurityGroupRule struct {
	Type       string   `yaml:"type"`
	Protocol   string   `yaml:"protocol"`
	FromPort   int      `yaml:"from_port"`
	ToPort     int      `yaml:"to_port"`
	CIDRBlocks []string `yaml:"cidr_blocks,omitempty"`
	SourceVPC  string   `yaml:"source_vpc,omitempty"`
}

type KubernetesCluster struct {
	Name     string                 `yaml:"name"`
	Provider string                 `yaml:"provider"`
	VPC      string                 `yaml:"vpc"`
	Spec     map[string]interface{} `yaml:"spec"`
}

type Compute struct {
	Name          string                 `yaml:"name"`
	Type          string                 `yaml:"type"`
	Provider      string                 `yaml:"provider"`
	VPC           string                 `yaml:"vpc"`
	Subnet        string                 `yaml:"subnet"`
	SecurityGroup string                 `yaml:"security_group"`
	Storage       []Storage              `yaml:"storage"`
	Spec          map[string]interface{} `yaml:"spec"`
}

type Storage struct {
	Name      string `yaml:"name"`
	Path      string `yaml:"path"`
	Size      int    `yaml:"size"`
	Type      string `yaml:"type"`
	Encrypted bool   `yaml:"encrypted"`
}

// ParseManifest reads and parses the service manifest file
func ParseManifest(path string) (*Service, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var service Service
	if err := yaml.Unmarshal(data, &service); err != nil {
		return nil, err
	}

	// Validate the parsed service
	if err := ValidateService(&service); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &service, nil
}
