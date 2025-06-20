package parser

import (
	"os"
	"testing"
)

func TestParseManifest(t *testing.T) {
	// Create a temporary public key file for testing
	tmpKeyFile, err := os.CreateTemp("", "test-key-*.pub")
	if err != nil {
		t.Fatalf("Failed to create temp key file: %v", err)
	}
	defer os.Remove(tmpKeyFile.Name())

	// Write dummy public key content
	dummyKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... dummy-key-for-testing"
	if _, err := tmpKeyFile.WriteString(dummyKey); err != nil {
		t.Fatalf("Failed to write dummy key: %v", err)
	}
	tmpKeyFile.Close()

	testManifest := `
apiVersion: bolt/v1
kind: Service
metadata:
  name: test-service
  owner: test-owner
  tags:
    environment: "test"
    project: "test-project"
providers:
  - name: "aws_test"
    type: "aws"
    spec:
      region: "us-east-1"
      environment: "test"
spec:
  key_pair:
    name: "bolt-key"
    public_key_file: "` + tmpKeyFile.Name() + `"
  infrastructure:
    networks:
      - name: "vpc-test"
        provider: "aws_test"
        cidr: "10.0.0.0/16"
        subnets:
          - name: "subnet-test"
            zone: "us-east-1a"
            cidr: "10.0.1.0/24"
`

	tmpFile, err := os.CreateTemp("", "test-manifest-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testManifest); err != nil {
		t.Fatalf("Failed to write test manifest: %v", err)
	}
	tmpFile.Close()

	service, err := ParseManifest(tmpFile.Name())
	if err != nil {
		t.Fatalf("ParseManifest failed: %v", err)
	}

	if service.Metadata.Name != "test-service" {
		t.Errorf("Expected service name 'test-service', got '%s'", service.Metadata.Name)
	}

	if service.Metadata.Owner != "test-owner" {
		t.Errorf("Expected owner 'test-owner', got '%s'", service.Metadata.Owner)
	}

	if len(service.Providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(service.Providers))
	}

	if service.Providers[0].Name != "aws_test" {
		t.Errorf("Expected provider name 'aws_test', got '%s'", service.Providers[0].Name)
	}

	if len(service.Spec.Infrastructure.Networks) != 1 {
		t.Errorf("Expected 1 network, got %d", len(service.Spec.Infrastructure.Networks))
	}
}

func TestValidateService(t *testing.T) {
	tests := []struct {
		name    string
		service *Service
		wantErr bool
	}{
		{
			name: "valid service",
			service: &Service{
				Metadata: Metadata{
					Name:  "test-service",
					Owner: "test-owner",
				},
				Providers: []Provider{
					{
						Name: "aws_test",
						Type: "aws",
					},
				},
				Spec: Spec{
					KeyPair: KeyPair{
						Name: "bolt-key",
					},
					Infrastructure: Infrastructure{},
				},
			},
			wantErr: false,
		},
		{
			name: "missing service name",
			service: &Service{
				Metadata: Metadata{
					Owner: "test-owner",
				},
				Providers: []Provider{
					{
						Name: "aws_test",
						Type: "aws",
					},
				},
				Spec: Spec{
					KeyPair: KeyPair{
						Name: "bolt-key",
					},
					Infrastructure: Infrastructure{},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid provider type",
			service: &Service{
				Metadata: Metadata{
					Name:  "test-service",
					Owner: "test-owner",
				},
				Providers: []Provider{
					{
						Name: "invalid_test",
						Type: "invalid",
					},
				},
				Spec: Spec{
					KeyPair: KeyPair{
						Name: "bolt-key",
					},
					Infrastructure: Infrastructure{},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateService(tt.service)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCIDR(t *testing.T) {
	tests := []struct {
		name string
		cidr string
		want bool
	}{
		{"valid CIDR", "10.0.0.0/16", true},
		{"valid CIDR with /24", "192.168.1.0/24", true},
		{"valid CIDR with /8", "10.0.0.0/8", true},
		{"invalid CIDR - missing mask", "10.0.0.0", false},
		{"invalid CIDR - invalid IP", "256.0.0.0/16", false},
		{"invalid CIDR - invalid mask", "10.0.0.0/33", false},
		{"empty CIDR", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateCIDR(tt.cidr); got != tt.want {
				t.Errorf("validateCIDR() = %v, want %v", got, tt.want)
			}
		})
	}
}
