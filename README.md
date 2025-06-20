# Bolt - Multi-Cloud Infrastructure as Code Tool

Bolt adalah tool untuk merakit infrastruktur dari definisi layanan abstrak menggunakan OpenTofu. Mendukung AWS, Azure, dan GCP dengan struktur YAML yang konsisten dan sangat fleksibel.

## 🚀 Fitur Utama

- **Multi-Provider Support**: AWS, Azure, GCP
- **Abstraksi Layanan**: Definisikan infrastruktur dalam YAML sederhana
- **Fleksibel**: Bisa hanya VM, hanya network, atau full stack
- **Key Pair Existing**: Bisa pakai key pair yang sudah ada
- **LocalStack Integration**: Testing lokal untuk AWS
- **Production Ready**: Siap untuk deployment production
- **Struktur YAML Konsisten**: Migrasi antar provider sangat mudah

## 📋 Prerequisites

### Untuk AWS (LocalStack):
```bash
brew install awscli localstack opentofu
yarn global add aws-cdk-local
localstack start
aws configure # (isi dengan dummy credentials untuk local)
```

### Untuk Azure:
```bash
brew install azure-cli opentofu
az login
```

### Untuk GCP:
```bash
brew install google-cloud-sdk opentofu
# Login dan set project
gcloud auth application-default login
gcloud config set project YOUR_PROJECT_ID
```

## 🏗️ Instalasi

```bash
# Clone repository
git clone <repository-url>
cd bold

# Build binary
go build -o bolt main.go

# Test instalasi
./bolt --help
```

## 📝 Contoh YAML Infrastruktur

### 1. **AWS - Minimal, Key Pair Existing, Hanya VM (Local Testing)**
```yaml
apiVersion: bolt/v1
kind: Service
metadata:
  name: example-aws-service
  owner: yourname
  tags:
    environment: "local"  # Ganti ke "production" untuk AWS asli
    project: "bolt-demo"
providers:
  - name: aws_local  # Ganti ke "aws_prod" untuk production
    type: aws
    spec:
      region: us-east-1
      environment: local  # Ganti ke "production" untuk AWS asli
      # Untuk production, tambahkan credentials di environment variables:
      # export AWS_ACCESS_KEY_ID="your-access-key"
      # export AWS_SECRET_ACCESS_KEY="your-secret-key"

spec:
  infrastructure:
    key_pair:
      name: "my-existing-key" # Key pair sudah ada di AWS
    networks:
      - name: vpc-main
        provider: aws_local  # Ganti ke "aws_prod" untuk production
        cidr: 10.10.0.0/16
        subnets:
          - name: subnet-public
            zone: us-east-1a
            cidr: 10.10.1.0/24
    security_groups:
      - name: web-sg
        provider: aws_local  # Ganti ke "aws_prod" untuk production
        vpc: vpc-main
        rules:
          - type: ingress
            protocol: tcp
            from_port: 22
            to_port: 22
            cidr_blocks: ["0.0.0.0/0"]
    computes:
      - name: web-vm
        type: ec2
        provider: aws_local  # Ganti ke "aws_prod" untuk production
        vpc: vpc-main
        subnet: subnet-public
        security_group: web-sg
        spec:
          instance_type: t3.micro
          root_disk_size_gb: 20
```

### 2. **Azure - Minimal, Key Pair Existing, Hanya VM (Local Testing)**
```yaml
apiVersion: bolt/v1
kind: Service
metadata:
  name: example-azure-service
  owner: yourname
  tags:
    environment: "local"  # Ganti ke "production" untuk Azure asli
    project: "bolt-demo"
providers:
  - name: azure_local  # Ganti ke "azure_prod" untuk production
    type: azurerm
    spec:
      region: eastus
      environment: local  # Ganti ke "production" untuk Azure asli
      # Untuk production, login dengan: az login
      # Atau gunakan service principal: az login --service-principal

spec:
  infrastructure:
    key_pair:
      name: "my-existing-key" # SSH public key sudah di-upload manual
    networks:
      - name: vnet-main
        provider: azure_local  # Ganti ke "azure_prod" untuk production
        cidr: 10.20.0.0/16
        subnets:
          - name: subnet-main
            zone: eastus
            cidr: 10.20.1.0/24
    security_groups:
      - name: web-nsg
        provider: azure_local  # Ganti ke "azure_prod" untuk production
        vpc: vnet-main
        rules:
          - type: ingress
            protocol: Tcp
            from_port: 22
            to_port: 22
            cidr_blocks: ["0.0.0.0/0"]
    computes:
      - name: web-vm
        type: azurerm_linux_virtual_machine
        provider: azure_local  # Ganti ke "azure_prod" untuk production
        vpc: vnet-main
        subnet: subnet-main
        security_group: web-nsg
        spec:
          size: Standard_B1s
          username: "azureuser"
          image:
            publisher: "Canonical"
            offer: "0001-com-ubuntu-server-jammy"
            sku: "22_04-lts-gen2"
          os_disk_size_gb: 30
```

### 3. **GCP - Minimal, Key Pair Existing, Hanya VM (Local Testing)**
```yaml
apiVersion: bolt/v1
kind: Service
metadata:
  name: example-gcp-service
  owner: yourname
  tags:
    environment: "local"  # Ganti ke "production" untuk GCP asli
    project: "bolt-demo"
providers:
  - name: gcp_local  # Ganti ke "gcp_prod" untuk production
    type: google
    spec:
      project: "your-gcp-project-id"  # Ganti dengan project ID asli untuk production
      region: us-central1
      environment: local  # Ganti ke "production" untuk GCP asli
      # Untuk production, set project dan login:
      # gcloud config set project YOUR_PROJECT_ID
      # gcloud auth application-default login

spec:
  infrastructure:
    key_pair:
      name: "my-existing-key" # SSH key di metadata project/user
    networks:
      - name: vpc-main
        provider: gcp_local  # Ganti ke "gcp_prod" untuk production
        cidr: 10.30.0.0/16
        subnets:
          - name: subnet-main
            zone: us-central1-a
            cidr: 10.30.1.0/24
    security_groups:
      - name: web-fw
        provider: gcp_local  # Ganti ke "gcp_prod" untuk production
        vpc: vpc-main
        rules:
          - type: ingress
            protocol: tcp
            from_port: 22
            to_port: 22
            cidr_blocks: ["0.0.0.0/0"]
    computes:
      - name: web-vm
        type: google_compute_instance
        provider: gcp_local  # Ganti ke "gcp_prod" untuk production
        vpc: vpc-main
        subnet: subnet-main
        security_group: web-fw
        spec:
          machine_type: e2-medium
          zone: us-central1-a
          image: "debian-cloud/debian-11"
          disk_size_gb: 20
```

### 4. **Fleksibilitas YAML**
- Bisa hanya 1 VM, atau full stack (multi VPC, multi subnet, multi SG, cluster, dsb)
- Bisa pakai key pair yang sudah ada (cukup isi `name` saja)
- Semua resource opsional, hanya isi yang dibutuhkan
- Bisa multi-provider dalam satu file

### 5. **Local vs Production**
- **Local Testing**: Gunakan `environment: local` dan provider name dengan suffix `_local`
- **Production**: Ganti ke `environment: production` dan provider name dengan suffix `_prod`
- **AWS Local**: Menggunakan LocalStack (tidak ada biaya)
- **Azure/GCP Local**: Tetap menggunakan cloud dev/test (ada biaya minimal)

## 🏗️ Bootstrap & Destroy Infrastructure

```bash
# Bootstrap infrastruktur
./bolt bootstrap service.yaml

# Atau untuk Azure
./bolt bootstrap service-azure.yaml

# Atau untuk GCP
./bolt bootstrap service-gcp.yaml

# Destroy infrastructure
./bolt destroy service.yaml
```

## 🔧 Konfigurasi Production

### AWS Production Setup

1. **Set Environment Variables**:
```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_DEFAULT_REGION="us-east-1"
```

2. **Update Service YAML**:
```yaml
providers:
  - name: "aws_prod"
    type: "aws"
    spec:
      region: "us-east-1"
      environment: "production"  # Penting!
```

### Azure Production Setup

1. **Login dengan Service Principal**:
```bash
az login --service-principal \
  --username APP_ID \
  --password PASSWORD \
  --tenant TENANT_ID
```

2. **Update Service YAML**:
```yaml
providers:
  - name: azure_prod
    type: azurerm
    spec:
      region: "eastus"
      environment: "production"
```

### GCP Production Setup

1. **Set Project dan Authenticate**:
```bash
gcloud config set project YOUR_PROJECT_ID
gcloud auth application-default login
```

2. **Update Service YAML**:
```yaml
providers:
  - name: gcp_prod
    type: google
    spec:
      project: "your-actual-project-id"
      region: "us-central1"
      environment: "production"
```

## 🧪 Testing Lokal

### AWS (LocalStack)
```bash
localstack start
./bolt bootstrap service.yaml
```

### Azure (Dev/Test Subscription)
```bash
az login
./bolt bootstrap service-azure.yaml
```

### GCP (Project Dev/Test)
```bash
gcloud auth application-default login
./bolt bootstrap service-gcp.yaml
```

**Catatan:**
- Untuk AWS, resource akan dibuat di LocalStack (tidak ke AWS asli).
- Untuk Azure & GCP, resource tetap dibuat di cloud dev/test, gunakan subscription/project khusus testing.
- Untuk production, cukup ganti `environment: production` dan credential ke yang production.

## 🚦 Migrasi ke Production
1. **AWS Production**:
   ```bash
   # Set credentials
   export AWS_ACCESS_KEY_ID="your-access-key"
   export AWS_SECRET_ACCESS_KEY="your-secret-key"
   
   # Update YAML: ganti semua "aws_local" ke "aws_prod"
   # Update YAML: ganti "environment: local" ke "environment: production"
   ```

2. **Azure Production**:
   ```bash
   # Login
   az login
   
   # Update YAML: ganti semua "azure_local" ke "azure_prod"
   # Update YAML: ganti "environment: local" ke "environment: production"
   ```

3. **GCP Production**:
   ```bash
   # Set project dan login
   gcloud config set project YOUR_PROJECT_ID
   gcloud auth application-default login
   
   # Update YAML: ganti semua "gcp_local" ke "gcp_prod"
   # Update YAML: ganti "environment: local" ke "environment: production"
   ```

## ⚠️ Troubleshooting & Tips
- **Resource tidak muncul di cloud?**
  - Pastikan provider, credential, dan region sudah benar.
- **Error permission?**
  - Cek credential dan role IAM/Service Principal/Service Account.
- **LocalStack error?**
  - Pastikan LocalStack sudah running dan environment di YAML `local`.
- **Resource tidak didukung di local?**
  - Beberapa resource (EKS/AKS/GKE) hanya bisa dites di cloud dev/test.
- **Security group terlalu permisif?**
  - Perketat rule sebelum ke production.

## 🏗️ Arsitektur

```
Bolt CLI
├── Parser (YAML → Go Structs)
├── Compiler (Go Structs → OpenTofu JSON)
└── Engine (OpenTofu Commands)
    ├── AWS Provider
    ├── Azure Provider
    └── GCP Provider
```

## 📁 Struktur File

```
bold/
├── main.go                 # CLI entry point
├── cmd/
│   ├── bootstrap.go        # Bootstrap command
│   └── destroy.go          # Destroy command
├── pkg/
│   ├── parser/             # YAML parsing
│   ├── compiler/           # OpenTofu compilation
│   ├── engine/             # OpenTofu execution
│   └── workflow/           # Workflow orchestration
├── service.yaml            # AWS example
├── service-azure.yaml      # Azure example
├── service-gcp.yaml        # GCP example
└── README.md
```

## 🤝 Contributing

1. Fork repository
2. Create feature branch
3. Make changes
4. Add tests
5. Submit pull request

## 📄 License

MIT License 