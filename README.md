# Assignment 4 - MapReduce Implementation on AWS ECS

## Part 1: AWS Infrastructure Concepts

### EC2 vs ECS: Key Differences

#### EC2 (Elastic Compute Cloud)
- **Virtual machines** that you manage completely
- You control the OS, patches, Docker installation
- Better for: Custom environments, specific OS requirements, full control
- More management overhead

#### ECS (Elastic Container Service)
- **Container orchestration** service
- AWS manages the underlying infrastructure
- You just provide Docker images
- Better for: Microservices, scalable applications, less operational overhead
- Can run on EC2 or Fargate (serverless)

> **Choose EC2 when:** You need specific OS configurations, legacy applications, or full infrastructure control  
> **Choose ECS when:** You want to focus on applications, not infrastructure, and need container orchestration

### VPC and Subnets

#### VPC (Virtual Private Cloud)
- Your **isolated network** in AWS
- Like having your own data center in the cloud
- Controls IP ranges, routing, security

#### Subnets
- **Subdivisions** of your VPC
- Can be public (internet access) or private
- Spread across Availability Zones for resilience

#### Default VPC Access
- AWS creates a **default VPC** in each region automatically
- Pre-configured with:
  - Internet gateway for public access
  - Public subnets in each AZ
  - Route tables configured
- Your resources use it automatically unless you specify otherwise

### TCP vs UDP

#### TCP (Transmission Control Protocol)
- **Reliable, ordered** delivery
- Establishes connection before sending
- Guarantees packet delivery and order
- Use for: Web (HTTP/HTTPS), email, file transfers
- Higher overhead but more reliable

#### UDP (User Datagram Protocol)
- **Fast, no guarantees**
- No connection establishment
- Packets might arrive out of order or get lost
- Use for: Video streaming, gaming, DNS
- Lower overhead but less reliable

### Controlling Task Resources in ECS

#### CPU and Memory Allocation
- Defined in Task Definition
- CPU: Measured in CPU units (256 = 0.25 vCPU)
- Memory: Measured in MB

#### Resource Limits per Container
- **Soft limits** (reservation): Guaranteed resources
- **Hard limits** (maximum): Cannot exceed

#### Common Configurations:
| Type | CPU | Memory |
|------|-----|--------|
| **Minimal** | 0.25 vCPU | 0.5 GB RAM |
| **Small** | 0.5 vCPU | 1 GB RAM |
| **Medium** | 1 vCPU | 2 GB RAM |
| **Large** | 2 vCPU | 4 GB RAM |

#### Auto-scaling Options:
- Scale based on CPU/memory utilization
- Custom CloudWatch metrics
- Target tracking policies

---

## Part 2: MapReduce Implementation Results

### Experiment: 2 Chunks
```
============================================================
Running experiment with 2 chunks
============================================================
Starting MapReduce word count on: s3://mapreduce-wordcount-730335606003/input/hamlet.txt
Number of chunks: 2
--------------------------------------------------
1. Splitting file into chunks...
   Split completed in 188.00ms
   Created 2 chunks

2. Mapping chunks in parallel...
   Mapping completed
   Average map time: 103.00ms

3. Reducing results...
   Reduce completed in 123.00ms

==================================================
RESULTS:
Total unique words: 4701
Total word count: 30271
Final results saved to: s3://mapreduce-wordcount-730335606003/results/final-wordcount-1759260200.json

Top 10 words:
   1. the             - 993 occurrences
   2. and             - 863 occurrences
   3. to              - 685 occurrences
   4. of              - 610 occurrences
   5. i               - 574 occurrences
   6. you             - 527 occurrences
   7. a               - 511 occurrences
   8. my              - 502 occurrences
   9. it              - 419 occurrences
  10. in              - 400 occurrences

==================================================
PERFORMANCE METRICS:
Split time:      188.00ms
Map time (avg):   103.00ms
Map time (max):   109.00ms
Reduce time:     123.00ms
Total time:      683.32ms
```

### Experiment: 3 Chunks
```
============================================================
Running experiment with 3 chunks
============================================================
Starting MapReduce word count on: s3://mapreduce-wordcount-730335606003/input/hamlet.txt
Number of chunks: 3
--------------------------------------------------
1. Splitting file into chunks...
   Split completed in 143.00ms
   Created 3 chunks

2. Mapping chunks in parallel...
   Mapping completed
   Average map time: 99.00ms

3. Reducing results...
   Reduce completed in 125.00ms

==================================================
RESULTS:
Total unique words: 4701
Total word count: 30271
Final results saved to: s3://mapreduce-wordcount-730335606003/results/final-wordcount-1759260200.json

Top 10 words:
   1. the             - 993 occurrences
   2. and             - 863 occurrences
   3. to              - 685 occurrences
   4. of              - 610 occurrences
   5. i               - 574 occurrences
   6. you             - 527 occurrences
   7. a               - 511 occurrences
   8. my              - 502 occurrences
   9. it              - 419 occurrences
  10. in              - 400 occurrences

==================================================
PERFORMANCE METRICS:
Split time:      143.00ms
Map time (avg):    99.00ms
Map time (max):   109.00ms
Reduce time:     125.00ms
Total time:      620.44ms
```

### Experiment: 4 Chunks
```
============================================================
Running experiment with 4 chunks
============================================================
Starting MapReduce word count on: s3://mapreduce-wordcount-730335606003/input/hamlet.txt
Number of chunks: 4
--------------------------------------------------
1. Splitting file into chunks...
   Split completed in 201.00ms
   Created 4 chunks

2. Mapping chunks in parallel...
   Mapping completed
   Average map time: 87.00ms

3. Reducing results...
   Reduce completed in 164.00ms

==================================================
RESULTS:
Total unique words: 4701
Total word count: 30271
Final results saved to: s3://mapreduce-wordcount-730335606003/results/final-wordcount-1759260201.json

Top 10 words:
   1. the             - 993 occurrences
   2. and             - 863 occurrences
   3. to              - 685 occurrences
   4. of              - 610 occurrences
   5. i               - 574 occurrences
   6. you             - 527 occurrences
   7. a               - 511 occurrences
   8. my              - 502 occurrences
   9. it              - 419 occurrences
  10. in              - 400 occurrences

==================================================
PERFORMANCE METRICS:
Split time:      201.00ms
Map time (avg):    87.00ms
Map time (max):   109.00ms
Reduce time:     164.00ms
Total time:      848.17ms
```

### Experiment: 5 Chunks
```
============================================================
Running experiment with 5 chunks
============================================================
Starting MapReduce word count on: s3://mapreduce-wordcount-730335606003/input/hamlet.txt
Number of chunks: 5
--------------------------------------------------
1. Splitting file into chunks...
   Split completed in 190.00ms
   Created 5 chunks

2. Mapping chunks in parallel...
   Mapping completed
   Average map time: 81.71ms

3. Reducing results...
   Reduce completed in 203.00ms

==================================================
RESULTS:
Total unique words: 4701
Total word count: 30271
Final results saved to: s3://mapreduce-wordcount-730335606003/results/final-wordcount-1759260202.json

Top 10 words:
   1. the             - 993 occurrences
   2. and             - 863 occurrences
   3. to              - 685 occurrences
   4. of              - 610 occurrences
   5. i               - 574 occurrences
   6. you             - 527 occurrences
   7. a               - 511 occurrences
   8. my              - 502 occurrences
   9. it              - 419 occurrences
  10. in              - 400 occurrences

==================================================
PERFORMANCE METRICS:
Split time:      190.00ms
Map time (avg):    81.71ms
Map time (max):   109.00ms
Reduce time:     203.00ms
Total time:      889.32ms
```

### Experiment: 6 Chunks
```
============================================================
Running experiment with 6 chunks
============================================================
Starting MapReduce word count on: s3://mapreduce-wordcount-730335606003/input/hamlet.txt
Number of chunks: 6
--------------------------------------------------
1. Splitting file into chunks...
   Split completed in 203.00ms
   Created 6 chunks

2. Mapping chunks in parallel...
   Mapping completed
   Average map time: 80.30ms

3. Reducing results...
   Reduce completed in 189.00ms

==================================================
RESULTS:
Total unique words: 4701
Total word count: 30271
Final results saved to: s3://mapreduce-wordcount-730335606003/results/final-wordcount-1759260203.json

Top 10 words:
   1. the             - 993 occurrences
   2. and             - 863 occurrences
   3. to              - 685 occurrences
   4. of              - 610 occurrences
   5. i               - 574 occurrences
   6. you             - 527 occurrences
   7. a               - 511 occurrences
   8. my              - 502 occurrences
   9. it              - 419 occurrences
  10. in              - 400 occurrences

==================================================
PERFORMANCE METRICS:
Split time:      203.00ms
Map time (avg):    80.30ms
Map time (max):   114.00ms
Reduce time:     189.00ms
Total time:      915.83ms
```

---

## Performance Summary

### Overall Results Comparison

| Chunks | Split Time (ms) | Avg Map Time (ms) | Max Map Time (ms) | Reduce Time (ms) | **Total Time (ms)** |
|:------:|:--------------:|:----------------:|:----------------:|:---------------:|:------------------:|
| 2      | 188           | 103              | 109              | 123             | **683.32**         |
| **3**  | **143**       | **99**           | **109**          | **125**         | **620.44** ⭐      |
| 4      | 201           | 87               | 109              | 164             | **848.17**         |
| 5      | 190           | 81.71            | 109              | 203             | **889.32**         |
| 6      | 203           | 80.30            | 114              | 189             | **915.83**         |

### Key Observations
- ✅ **Optimal Performance**: 3 chunks yielded the fastest total time (620.44ms)
- 📈 **Map Time Trend**: Average map time decreased with more chunks (parallelization benefit)
- 📉 **Reduce Time Trend**: Reduce time increased with more chunks (aggregation overhead)
- ⚠️ **Diminishing Returns**: Performance degraded beyond 3 chunks due to coordination overhead
