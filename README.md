# 🚀 Bolt - Multi-Cloud Infrastructure as Code

Bolt is a powerful Infrastructure as Code (IaC) tool that supports AWS, Azure, and GCP with a simple YAML configuration. Built with Go and OpenTofu, it provides a unified interface for managing multi-cloud infrastructure including Kubernetes clusters.

## ✨ Features

- **Multi-Cloud Support**: AWS, Azure, and GCP
- **Kubernetes Clusters**: EKS, AKS, and GKE support
- **Simple YAML Configuration**: Easy-to-read infrastructure definitions
- **Local Testing**: Test with LocalStack (AWS) and dev environments
- **Dependency Graph**: Visualize resource dependencies
- **Cost Estimation**: Get cost estimates before deployment
- **Flexible Key Management**: Use existing keys or generate new ones
- **Production Ready**: Built with Go for reliability and performance

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- OpenTofu
- LocalStack (for local AWS testing)
- Cloud provider credentials (for production)

### Installation

```bash
git clone <repository>
cd bold
go build -o bold .
```

### Basic Usage

```bash
# Analyze infrastructure
./bold analyze service.yaml

# Bootstrap infrastructure
./bold bootstrap service.yaml

# Destroy infrastructure
./bold destroy service.yaml
```

## 📋 Configuration Reference

### Service Manifest Structure

```yaml
apiVersion: bolt/v1
kind: Service
metadata:
  name: my-service
  owner: yourname
  tags:
    environment: "local"
    project: "my-project"

providers:
  - name: aws_local
    type: aws
    spec:
      region: us-east-1
      environment: local

spec:
  key_pair:
    name: "my-key"
    public_key_file: "~/.ssh/id_rsa.pub"
    use_existing: true
  
  infrastructure:
    networks: []
    security_groups: []
    kubernetes_clusters: []
    computes: []
```

## 🔑 Key Pair Configuration

Bolt supports three key pair configurations:

| Option | Description | Example |
|--------|-------------|---------|
| **No Key Pair** | Skip key pair entirely | Omit `key_pair` section |
| **Use Existing** | Use local SSH key | `use_existing: true` |
| **Generate New** | Create new key pair | `public_key_file: "./bolt-key.pub"` |

### Key Pair Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `name` | string | Yes | Key pair name | `"my-key"` |
| `public_key_file` | string | Yes* | Path to public key file | `"~/.ssh/id_rsa.pub"` |
| `use_existing` | boolean | No | Use existing local key | `true` |

*Required unless `use_existing: true`

### Key Pair Examples

#### 1. No Key Pair (Optional)
```yaml
spec:
  infrastructure:
    # ... resources without key_pair
```

#### 2. Use Existing Local SSH Key
```yaml
spec:
  key_pair:
    name: "local-ssh-key"
    public_key_file: "~/.ssh/id_rsa.pub"
    use_existing: true
```

#### 3. Generate New Key Pair
```yaml
spec:
  key_pair:
    name: "bolt-key"
    public_key_file: "./bolt-key.pub"
```

## 🌐 Network Configuration

### Network Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `name` | string | Yes | Network name | `"vpc-main"` |
| `provider` | string | Yes | Cloud provider | `"aws_local"` |
| `cidr` | string | Yes | Network CIDR | `"10.10.0.0/16"` |
| `subnets` | array | No | Subnet list | See below |

### Subnet Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `name` | string | Yes | Subnet name | `"subnet-public"` |
| `zone` | string | Yes | Availability zone | `"us-east-1a"` |
| `cidr` | string | Yes | Subnet CIDR | `"10.10.1.0/24"` |

### Network Examples

#### Multi-Zone Network
```yaml
networks:
  - name: vpc-main
    provider: aws_local
    cidr: 10.10.0.0/16
    subnets:
      - name: subnet-public-1a
        zone: us-east-1a
        cidr: 10.10.1.0/24
      - name: subnet-public-1b
        zone: us-east-1b
        cidr: 10.10.2.0/24
      - name: subnet-private-1a
        zone: us-east-1a
        cidr: 10.10.3.0/24
      - name: subnet-private-1b
        zone: us-east-1b
        cidr: 10.10.4.0/24
```

## 🔒 Security Group Configuration

### Security Group Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `name` | string | Yes | Security group name | `"web-sg"` |
| `provider` | string | Yes | Cloud provider | `"aws_local"` |
| `vpc` | string | Yes | VPC name | `"vpc-main"` |
| `rules` | array | Yes | Security rules | See below |

### Security Rule Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `type` | string | Yes | Rule type | `"ingress"` or `"egress"` |
| `protocol` | string | Yes | Protocol | `"tcp"`, `"udp"`, `"-1"` |
| `from_port` | integer | Yes | Start port | `22` |
| `to_port` | integer | Yes | End port | `22` |
| `cidr_blocks` | array | No* | CIDR blocks | `["0.0.0.0/0"]` |
| `source_vpc` | string | No* | Source VPC | `"vpc-main"` |

*Required one of `cidr_blocks` or `source_vpc`

### Security Group Examples

#### Web Server Security Group
```yaml
security_groups:
  - name: web-sg
    provider: aws_local
    vpc: vpc-main
    rules:
      - type: ingress
        protocol: tcp
        from_port: 22
        to_port: 22
        cidr_blocks: ["0.0.0.0/0"]
      - type: ingress
        protocol: tcp
        from_port: 80
        to_port: 80
        cidr_blocks: ["0.0.0.0/0"]
      - type: ingress
        protocol: tcp
        from_port: 443
        to_port: 443
        cidr_blocks: ["0.0.0.0/0"]
      - type: egress
        protocol: -1
        from_port: 0
        to_port: 0
        cidr_blocks: ["0.0.0.0/0"]
```

#### Application Security Group
```yaml
security_groups:
  - name: app-sg
    provider: aws_local
    vpc: vpc-main
    rules:
      - type: ingress
        protocol: tcp
        from_port: 22
        to_port: 22
        source_vpc: vpc-main
      - type: ingress
        protocol: tcp
        from_port: 8080
        to_port: 8080
        source_vpc: vpc-main
      - type: egress
        protocol: -1
        from_port: 0
        to_port: 0
        cidr_blocks: ["0.0.0.0/0"]
```

## 🐳 Kubernetes Configuration

### Kubernetes Cluster Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `name` | string | Yes | Cluster name | `"eks-cluster"` |
| `provider` | string | Yes | Cloud provider | `"aws_local"` |
| `vpc` | string | Yes | VPC name | `"vpc-main"` |
| `spec` | object | Yes | Cluster specification | See below |

### Kubernetes Spec Parameters

| Parameter | Type | Required | Description | AWS | Azure | GCP |
|-----------|------|----------|-------------|-----|-------|-----|
| `version` | string | No | K8s version | `"1.28"` | `"1.28"` | `"1.28"` |
| `node_type` | string | No | Instance type | `"t3.medium"` | - | - |
| `node_size` | string | No | VM size | - | `"Standard_B2s"` | - |
| `machine_type` | string | No | Machine type | - | - | `"e2-medium"` |
| `node_count` | integer | No | Worker nodes | `3` | `3` | `3` |
| `node_disk_size_gb` | integer | No | Disk size | `50` | - | - |

### Kubernetes Examples

#### EKS Cluster (AWS)
```yaml
kubernetes_clusters:
  - name: eks-cluster
    provider: aws_local
    vpc: vpc-main
    spec:
      version: "1.28"
      node_type: "t3.medium"
      node_count: 3
      node_disk_size_gb: 50
```

#### AKS Cluster (Azure)
```yaml
kubernetes_clusters:
  - name: aks-cluster
    provider: azurerm_local
    vpc: vnet-main
    spec:
      version: "1.28"
      node_size: "Standard_B2s"
      node_count: 3
```

#### GKE Cluster (GCP)
```yaml
kubernetes_clusters:
  - name: gke-cluster
    provider: google_local
    vpc: vpc-main
    spec:
      version: "1.28"
      machine_type: "e2-medium"
      node_count: 3
```

## 💻 Compute Configuration

### Compute Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `name` | string | Yes | Instance name | `"web-server"` |
| `type` | string | Yes | Instance type | `"ec2"` |
| `provider` | string | Yes | Cloud provider | `"aws_local"` |
| `vpc` | string | Yes | VPC name | `"vpc-main"` |
| `subnet` | string | Yes | Subnet name | `"subnet-public"` |
| `security_group` | string | No | Security group | `"web-sg"` |
| `spec` | object | Yes | Instance specification | See below |

### Compute Spec Parameters

| Parameter | Type | Required | Description | AWS | Azure | GCP |
|-----------|------|----------|-------------|-----|-------|-----|
| `instance_type` | string | No | Instance type | `"t3.micro"` | - | - |
| `size` | string | No | VM size | - | `"Standard_B1s"` | - |
| `machine_type` | string | No | Machine type | - | - | `"e2-micro"` |
| `root_disk_size_gb` | integer | No | Disk size | `20` | `20` | `20` |

### Compute Examples

#### EC2 Instance (AWS)
```yaml
computes:
  - name: web-server
    type: ec2
    provider: aws_local
    vpc: vpc-main
    subnet: subnet-public
    security_group: web-sg
    spec:
      instance_type: t3.micro
      root_disk_size_gb: 20
```

#### Azure VM
```yaml
computes:
  - name: web-server
    type: azurerm_linux_virtual_machine
    provider: azurerm_local
    vpc: vnet-main
    subnet: subnet-public
    security_group: nsg-web
    spec:
      size: Standard_B1s
      root_disk_size_gb: 20
```

#### GCP Instance
```yaml
computes:
  - name: web-server
    type: google_compute_instance
    provider: google_local
    vpc: vpc-main
    subnet: subnet-public
    security_group: firewall-web
    spec:
      machine_type: e2-micro
      root_disk_size_gb: 20
```

## 🌍 Multi-Cloud Examples

### AWS Comprehensive Example
```yaml
apiVersion: bolt/v1
kind: Service
metadata:
  name: aws-comprehensive-example
  owner: yourname
  tags:
    environment: "local"
    project: "bolt-aws-demo"
    cost-center: "engineering"
    team: "platform"

providers:
  - name: aws_local
    type: aws
    spec:
      region: us-east-1
      environment: local

spec:
  key_pair:
    name: "aws-key"
    public_key_file: "~/.ssh/id_rsa.pub"
    use_existing: true
  
  infrastructure:
    networks:
      - name: vpc-main
        provider: aws_local
        cidr: 10.10.0.0/16
        subnets:
          - name: subnet-public-1a
            zone: us-east-1a
            cidr: 10.10.1.0/24
          - name: subnet-public-1b
            zone: us-east-1b
            cidr: 10.10.2.0/24
          - name: subnet-private-1a
            zone: us-east-1a
            cidr: 10.10.3.0/24
          - name: subnet-private-1b
            zone: us-east-1b
            cidr: 10.10.4.0/24
    
    security_groups:
      - name: web-sg
        provider: aws_local
        vpc: vpc-main
        rules:
          - type: ingress
            protocol: tcp
            from_port: 22
            to_port: 22
            cidr_blocks: ["0.0.0.0/0"]
          - type: ingress
            protocol: tcp
            from_port: 80
            to_port: 80
            cidr_blocks: ["0.0.0.0/0"]
          - type: ingress
            protocol: tcp
            from_port: 443
            to_port: 443
            cidr_blocks: ["0.0.0.0/0"]
          - type: egress
            protocol: -1
            from_port: 0
            to_port: 0
            cidr_blocks: ["0.0.0.0/0"]
      
      - name: app-sg
        provider: aws_local
        vpc: vpc-main
        rules:
          - type: ingress
            protocol: tcp
            from_port: 22
            to_port: 22
            source_vpc: vpc-main
          - type: ingress
            protocol: tcp
            from_port: 8080
            to_port: 8080
            source_vpc: vpc-main
          - type: egress
            protocol: -1
            from_port: 0
            to_port: 0
            cidr_blocks: ["0.0.0.0/0"]
    
    kubernetes_clusters:
      - name: eks-cluster
        provider: aws_local
        vpc: vpc-main
        spec:
          version: "1.28"
          node_type: "t3.medium"
          node_count: 3
          node_disk_size_gb: 50
    
    computes:
      - name: web-server-1
        type: ec2
        provider: aws_local
        vpc: vpc-main
        subnet: subnet-public-1a
        security_group: web-sg
        spec:
          instance_type: t3.micro
          root_disk_size_gb: 20
      
      - name: web-server-2
        type: ec2
        provider: aws_local
        vpc: vpc-main
        subnet: subnet-public-1b
        security_group: web-sg
        spec:
          instance_type: t3.micro
          root_disk_size_gb: 20
      
      - name: app-server-1
        type: ec2
        provider: aws_local
        vpc: vpc-main
        subnet: subnet-private-1a
        security_group: app-sg
        spec:
          instance_type: t3.small
          root_disk_size_gb: 40
      
      - name: app-server-2
        type: ec2
        provider: aws_local
        vpc: vpc-main
        subnet: subnet-private-1b
        security_group: app-sg
        spec:
          instance_type: t3.small
          root_disk_size_gb: 40
```

### Azure Comprehensive Example
```yaml
apiVersion: bolt/v1
kind: Service
metadata:
  name: azure-comprehensive-example
  owner: yourname
  tags:
    environment: "local"
    project: "bolt-azure-demo"
    cost-center: "engineering"
    team: "platform"

providers:
  - name: azurerm_local
    type: azurerm
    spec:
      region: eastus
      environment: local

spec:
  key_pair:
    name: "azure-key"
    public_key_file: "~/.ssh/id_rsa.pub"
    use_existing: true
  
  infrastructure:
    networks:
      - name: vnet-main
        provider: azurerm_local
        cidr: 10.10.0.0/16
        subnets:
          - name: subnet-public-1
            zone: eastus
            cidr: 10.10.1.0/24
          - name: subnet-public-2
            zone: eastus2
            cidr: 10.10.2.0/24
          - name: subnet-private-1
            zone: eastus
            cidr: 10.10.3.0/24
          - name: subnet-private-2
            zone: eastus2
            cidr: 10.10.4.0/24
    
    security_groups:
      - name: nsg-web
        provider: azurerm_local
        vpc: vnet-main
        rules:
          - type: ingress
            protocol: tcp
            from_port: 22
            to_port: 22
            cidr_blocks: ["0.0.0.0/0"]
          - type: ingress
            protocol: tcp
            from_port: 80
            to_port: 80
            cidr_blocks: ["0.0.0.0/0"]
          - type: ingress
            protocol: tcp
            from_port: 443
            to_port: 443
            cidr_blocks: ["0.0.0.0/0"]
          - type: egress
            protocol: -1
            from_port: 0
            to_port: 0
            cidr_blocks: ["0.0.0.0/0"]
      
      - name: nsg-app
        provider: azurerm_local
        vpc: vnet-main
        rules:
          - type: ingress
            protocol: tcp
            from_port: 22
            to_port: 22
            source_vpc: vnet-main
          - type: ingress
            protocol: tcp
            from_port: 8080
            to_port: 8080
            source_vpc: vnet-main
          - type: egress
            protocol: -1
            from_port: 0
            to_port: 0
            cidr_blocks: ["0.0.0.0/0"]
    
    kubernetes_clusters:
      - name: aks-cluster
        provider: azurerm_local
        vpc: vnet-main
        spec:
          version: "1.28"
          node_size: "Standard_B2s"
          node_count: 3
    
    computes:
      - name: web-server-1
        type: azurerm_linux_virtual_machine
        provider: azurerm_local
        vpc: vnet-main
        subnet: subnet-public-1
        security_group: nsg-web
        spec:
          size: Standard_B1s
          root_disk_size_gb: 20
      
      - name: web-server-2
        type: azurerm_linux_virtual_machine
        provider: azurerm_local
        vpc: vnet-main
        subnet: subnet-public-2
        security_group: nsg-web
        spec:
          size: Standard_B1s
          root_disk_size_gb: 20
      
      - name: app-server-1
        type: azurerm_linux_virtual_machine
        provider: azurerm_local
        vpc: vnet-main
        subnet: subnet-private-1
        security_group: nsg-app
        spec:
          size: Standard_B2s
          root_disk_size_gb: 40
      
      - name: app-server-2
        type: azurerm_linux_virtual_machine
        provider: azurerm_local
        vpc: vnet-main
        subnet: subnet-private-2
        security_group: nsg-app
        spec:
          size: Standard_B2s
          root_disk_size_gb: 40
```

### GCP Comprehensive Example
```yaml
apiVersion: bolt/v1
kind: Service
metadata:
  name: gcp-comprehensive-example
  owner: yourname
  tags:
    environment: "local"
    project: "bolt-gcp-demo"
    cost-center: "engineering"
    team: "platform"

providers:
  - name: google_local
    type: google
    spec:
      region: us-central1
      environment: local

spec:
  key_pair:
    name: "gcp-key"
    public_key_file: "~/.ssh/id_rsa.pub"
    use_existing: true
  
  infrastructure:
    networks:
      - name: vpc-main
        provider: google_local
        cidr: 10.10.0.0/16
        subnets:
          - name: subnet-public-1a
            zone: us-central1-a
            cidr: 10.10.1.0/24
          - name: subnet-public-1b
            zone: us-central1-b
            cidr: 10.10.2.0/24
          - name: subnet-private-1a
            zone: us-central1-a
            cidr: 10.10.3.0/24
          - name: subnet-private-1b
            zone: us-central1-b
            cidr: 10.10.4.0/24
    
    security_groups:
      - name: firewall-web
        provider: google_local
        vpc: vpc-main
        rules:
          - type: ingress
            protocol: tcp
            from_port: 22
            to_port: 22
            cidr_blocks: ["0.0.0.0/0"]
          - type: ingress
            protocol: tcp
            from_port: 80
            to_port: 80
            cidr_blocks: ["0.0.0.0/0"]
          - type: ingress
            protocol: tcp
            from_port: 443
            to_port: 443
            cidr_blocks: ["0.0.0.0/0"]
          - type: egress
            protocol: -1
            from_port: 0
            to_port: 0
            cidr_blocks: ["0.0.0.0/0"]
      
      - name: firewall-app
        provider: google_local
        vpc: vpc-main
        rules:
          - type: ingress
            protocol: tcp
            from_port: 22
            to_port: 22
            source_vpc: vpc-main
          - type: ingress
            protocol: tcp
            from_port: 8080
            to_port: 8080
            source_vpc: vpc-main
          - type: egress
            protocol: -1
            from_port: 0
            to_port: 0
            cidr_blocks: ["0.0.0.0/0"]
    
    kubernetes_clusters:
      - name: gke-cluster
        provider: google_local
        vpc: vpc-main
        spec:
          version: "1.28"
          machine_type: "e2-medium"
          node_count: 3
    
    computes:
      - name: web-server-1
        type: google_compute_instance
        provider: google_local
        vpc: vpc-main
        subnet: subnet-public-1a
        security_group: firewall-web
        spec:
          machine_type: e2-micro
          root_disk_size_gb: 20
      
      - name: web-server-2
        type: google_compute_instance
        provider: google_local
        vpc: vpc-main
        subnet: subnet-public-1b
        security_group: firewall-web
        spec:
          machine_type: e2-micro
          root_disk_size_gb: 20
      
      - name: app-server-1
        type: google_compute_instance
        provider: google_local
        vpc: vpc-main
        subnet: subnet-private-1a
        security_group: firewall-app
        spec:
          machine_type: e2-small
          root_disk_size_gb: 40
      
      - name: app-server-2
        type: google_compute_instance
        provider: google_local
        vpc: vpc-main
        subnet: subnet-private-1b
        security_group: firewall-app
        spec:
          machine_type: e2-small
          root_disk_size_gb: 40
```

## 🔄 Environment Switching

### Local Development
```yaml
providers:
  - name: aws_local
    type: aws
    spec:
      region: us-east-1
      environment: local
```

### Production
```yaml
providers:
  - name: aws
    type: aws
    spec:
      region: us-east-1
      environment: production
```

## 📊 Analysis Features

### Dependency Graph
```bash
./bold analyze service.yaml
```

Shows:
- Resource dependency tree
- Mermaid diagram
- DOT graph for Graphviz
- Cost estimation

### Cost Estimation
- Monthly and hourly cost estimates
- Breakdown by resource type
- Local environment detection (free)
- Cost optimization tips

## 🛠️ Commands

```bash
# Analyze infrastructure
./bold analyze <service.yaml>

# Bootstrap infrastructure
./bold bootstrap <service.yaml>

# Destroy infrastructure
./bold destroy <service.yaml>
```

## 🔧 Development

### Project Structure
```
bold/
├── cmd/           # CLI commands
├── pkg/           # Core packages
│   ├── compiler/  # OpenTofu code generation
│   ├── config/    # Configuration management
│   ├── cost/      # Cost estimation
│   ├── engine/    # Deployment engine
│   ├── errors/    # Error handling
│   ├── graph/     # Dependency graph
│   ├── logger/    # Logging
│   ├── parser/    # YAML parsing
│   └── workflow/  # Workflow management
├── service.yaml   # Example configuration
└── README.md
```

### Building
```bash
go build -o bold .
```

### Testing
```bash
go test ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📞 Support

For support and questions, please open an issue on GitHub. 