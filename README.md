# Assignment4

# AWS Infrastructure Concepts

## EC2 vs ECS: Key Differences

### EC2 (Elastic Compute Cloud)
- **Virtual machines** that you manage completely
- You control the OS, patches, Docker installation
- Better for: Custom environments, specific OS requirements, full control
- More management overhead

### ECS (Elastic Container Service)
- **Container orchestration** service
- AWS manages the underlying infrastructure
- You just provide Docker images
- Better for: Microservices, scalable applications, less operational overhead
- Can run on EC2 or Fargate (serverless)

**Choose EC2 when:** You need specific OS configurations, legacy applications, or full infrastructure control  
**Choose ECS when:** You want to focus on applications, not infrastructure, and need container orchestration

## VPC and Subnets

### VPC (Virtual Private Cloud)
- Your **isolated network** in AWS
- Like having your own data center in the cloud
- Controls IP ranges, routing, security

### Subnets
- **Subdivisions** of your VPC
- Can be public (internet access) or private
- Spread across Availability Zones for resilience

### Default VPC Access
- AWS creates a **default VPC** in each region automatically
- Pre-configured with:
  - Internet gateway for public access
  - Public subnets in each AZ
  - Route tables configured
- Your resources use it automatically unless you specify otherwise

## TCP vs UDP

### TCP (Transmission Control Protocol)
- **Reliable, ordered** delivery
- Establishes connection before sending
- Guarantees packet delivery and order
- Use for: Web (HTTP/HTTPS), email, file transfers
- Higher overhead but more reliable

### UDP (User Datagram Protocol)
- **Fast, no guarantees**
- No connection establishment
- Packets might arrive out of order or get lost
- Use for: Video streaming, gaming, DNS
- Lower overhead but less reliable

## Controlling Task Resources in ECS

### CPU and Memory Allocation
- Defined in Task Definition
- CPU: Measured in CPU units (256 = 0.25 vCPU)
- Memory: Measured in MB

### Resource Limits per Container
- **Soft limits** (reservation): Guaranteed resources
- **Hard limits** (maximum): Cannot exceed

### Common Configurations:
- **Minimal**: 0.25 vCPU, 0.5 GB RAM
- **Small**: 0.5 vCPU, 1 GB RAM  
- **Medium**: 1 vCPU, 2 GB RAM
- **Large**: 2 vCPU, 4 GB RAM

### Auto-scaling Options:
- Scale based on CPU/memory utilization
- Custom CloudWatch metrics
- Target tracking policies
