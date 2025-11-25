# Professional Backend Development Prompt - High-Performance & Secure Omni-Channel POS System

## Project Overview

Build a  **high-performance, security-hardened** , microservices-based backend for **OniChange** - an omni-channel Point of Sale (POS) system using Go-lang with enterprise-grade architecture patterns.

---

## Technical Stack Requirements

### **Core Framework - Performance Optimized**

* **Language:** Go 1.21+ (latest stable with performance improvements)
* **HTTP Framework:**
  * **Fiber v2** (Express-like, 15x faster than net/http) OR
  * **Chi** (lightweight, idiomatic Go) OR
  * **Gin** (high-performance with middleware support)
* **Architecture:** Microservices with optimized API Gateway
* **Concurrency:** Goroutines with worker pools, context-based cancellation

### **Database Layer - High Performance**

* **ORM:** GORM v2 with performance optimizations
* **Database:** PostgreSQL 15+ with performance tuning
* **Connection Pool:**
  * **pgxpool** (3x faster than database/sql)
  * Pool size optimization based on load
  * Connection reuse and prepared statements
* **Migration:** golang-migrate/migrate with versioning
* **Performance Optimizations:**
  * Composite indexes on frequent queries
  * Partial indexes for filtered queries
  * EXPLAIN ANALYZE for query optimization
  * Connection pooling (min: 10, max: 100)
  * Query timeout configuration
  * Batch insert/update operations
  * Database connection health checks
  * Read replicas with load balancing

### **Caching Layer - Performance Critical**

* **Primary Cache:** Redis 7+ with **go-redis/redis v9**
* **Local Cache:** **ristretto** or **bigcache** for ultra-fast in-memory caching
* **Strategy:**
  * Multi-tier caching (L1: in-memory, L2: Redis)
  * Cache stampede prevention (singleflight pattern)
  * TTL optimization per data type
  * Cache warming on startup
  * Redis pipelining for batch operations
  * Redis Cluster for horizontal scaling
* **Performance Features:**
  * Connection pooling (min: 20, max: 200)
  * Pipeline batching (reduce network RTT)
  * Compression for large values (snappy/gzip)
  * Cache hit rate monitoring

### **Real-Time Communication - Optimized**

* **WebSocket:** **gorilla/websocket** or **nhooyr/websocket**
* **Performance Features:**
  * Connection pooling and reuse
  * Message compression (permessage-deflate)
  * Binary protocol for efficiency
  * Backpressure handling
  * Automatic reconnection with exponential backoff
  * Horizontal scaling with Redis Pub/Sub
  * Connection rate limiting
  * Heartbeat/ping-pong mechanism (30s interval)
  * Maximum concurrent connections per node: 10,000+

### **Infrastructure & Deployment - Production Grade**

* **Containerization:**
  * Multi-stage Docker builds (alpine-based, <50MB final image)
  * Distroless images for security
  * Layer caching optimization
* **Orchestration:** Kubernetes with HPA (Horizontal Pod Autoscaler)
* **CI/CD Pipeline:**
  * **GitHub Actions** / **GitLab CI** / **Jenkins**
  * Automated testing (unit, integration, load tests)
  * **Security Scanning:**
    * **gosec** (Go security checker)
    * **Trivy** (vulnerability scanner)
    * **Snyk** (dependency scanning)
    * **SAST** tools (static analysis)
  * **Code Quality:**
    * **golangci-lint** (30+ linters)
    * **SonarQube** (code quality gate)
  * **Performance Testing:**
    * **k6** or **Apache JMeter** load tests
    * Benchmark tests in CI pipeline
  * Blue-green or canary deployment strategy

---

## Security Requirements - Enterprise Grade

### **Authentication & Authorization**

* **JWT:** HS256/RS256 with short expiry (15 min access, 7 days refresh)
* **Token Storage:**
  * HttpOnly, Secure, SameSite cookies
  * Redis whitelist for active tokens
  * Token rotation on refresh
* **OAuth2/OIDC:** Integration with external providers
* **MFA:** TOTP-based two-factor authentication
* **Session Management:**
  * Secure session invalidation
  * Concurrent session limits
  * Device fingerprinting

### **API Security**

* **Rate Limiting:**
  * Token bucket algorithm (100 req/min per IP)
  * Distributed rate limiting with Redis
  * Progressive rate limiting (increase on abuse)
* **Request Validation:**
  * **validator/v10** for input validation
  * Request size limits (max 10MB)
  * Content-Type verification
  * Schema validation (JSON Schema)
* **CORS:** Strict whitelist configuration
* **Headers Security:**
  * X-Content-Type-Options: nosniff
  * X-Frame-Options: DENY
  * X-XSS-Protection: 1; mode=block
  * Strict-Transport-Security (HSTS)
  * Content-Security-Policy (CSP)
* **API Versioning:** URL-based (/v1/, /v2/)

### **Data Security**

* **Encryption at Rest:**
  * PostgreSQL transparent data encryption
  * Sensitive field encryption (AES-256-GCM)
  * **argon2id** for password hashing (not bcrypt)
* **Encryption in Transit:**
  * TLS 1.3 only (disable 1.2 and below)
  * Certificate pinning
  * Perfect Forward Secrecy (PFS)
* **Secrets Management:**
  * **HashiCorp Vault** or **AWS Secrets Manager**
  * No secrets in code or environment variables
  * Automatic secret rotation
* **Data Sanitization:**
  * SQL injection prevention (parameterized queries)
  * XSS prevention (HTML escaping)
  * NoSQL injection prevention
  * Path traversal prevention

### **Application Security**

* **Dependency Security:**
  * **govulncheck** for vulnerability scanning
  * Automated dependency updates (Dependabot)
  * Minimal dependencies principle
* **Error Handling:**
  * Never expose stack traces to clients
  * Generic error messages externally
  * Detailed logging internally
* **Logging Security:**
  * No sensitive data in logs (PII, passwords, tokens)
  * Structured logging with **zerolog** or **zap**
  * Log aggregation with ELK/Loki
  * Audit trail for critical operations
* **Code Security:**
  * No hardcoded credentials
  * Secure random generation (crypto/rand)
  * Time-constant comparison for secrets
  * Input sanitization at boundaries

### **Infrastructure Security**

* **Container Security:**
  * Non-root user (USER 1000)
  * Read-only filesystem where possible
  * Capability dropping (--cap-drop=ALL)
  * Resource limits (CPU/Memory)
* **Network Security:**
  * Service mesh (Istio) for mTLS
  * Network policies in Kubernetes
  * Private subnets for databases
  * WAF (Web Application Firewall)
* **Compliance:**
  * PCI-DSS Level 1 for payment processing
  * GDPR data protection principles
  * SOC 2 Type II readiness

---

## Performance Optimization Strategies

### **Code-Level Optimizations**

* **Memory Management:**
  * Sync.Pool for object reuse
  * Minimize allocations (use pointers wisely)
  * Pre-allocate slices with known capacity
  * Avoid string concatenation (use strings.Builder)
* **Concurrency Patterns:**
  * Worker pool pattern (limit goroutines)
  * Fan-out/fan-in patterns
  * Context propagation for cancellation
  * Bounded channels to prevent memory leaks
* **Profiling:**
  * pprof integration for CPU/memory profiling
  * Continuous profiling in production (Pyroscope)
  * Benchmark tests (go test -bench)

### **Database Optimizations**

* **Query Optimization:**
  * Eager loading to prevent N+1 queries
  * Select specific columns (avoid SELECT *)
  * Limit/offset optimization
  * Batch operations (bulk insert/update)
* **Indexing Strategy:**
  * B-tree indexes for equality/range queries
  * GIN indexes for JSONB/array columns
  * Covering indexes where applicable
* **Connection Management:**
  * Pool size tuning based on load testing
  * Connection lifetime limits
  * Prepared statement caching

### **API Performance**

* **Response Optimization:**
  * Gzip/Brotli compression (5x reduction)
  * HTTP/2 server push
  * ETag/If-None-Match caching
  * Pagination for large datasets (cursor-based)
* **Request Processing:**
  * Request coalescing for duplicate requests
  * Timeouts on all external calls (3s default)
  * Circuit breaker pattern (gobreaker)
  * Bulkhead pattern for resource isolation

### **Monitoring & Optimization**

* **Metrics:** Prometheus with custom business metrics
* **APM:** Datadog/New Relic for performance monitoring
* **Alerting:**
  * P95 latency > 200ms
  * Error rate > 1%
  * Memory usage > 80%
  * Connection pool exhaustion

---

## Microservices Architecture - Performance Focused

### **Core Services**

#### 1. **Order Service** - High Throughput

* **Performance:** Handle 10,000+ orders/minute
* **Optimizations:**
  * Event sourcing for order history
  * CQRS pattern (read/write separation)
  * Async order processing with message queue
  * Idempotency keys for duplicate prevention
* **Caching:** Order status, recent orders (TTL: 5 min)
* **Security:**
  * PCI-DSS compliant data handling
  * Order encryption at rest
  * Audit logging for all modifications

#### 2. **User Service** - Security Critical

* **Performance:** 5,000+ authentications/second
* **Optimizations:**
  * JWT caching in Redis
  * Password hash result caching (1 hour)
  * Lazy loading of user profiles
* **Security:**
  * Argon2id password hashing (time: 2, memory: 64MB)
  * Account lockout after 5 failed attempts
  * IP-based brute force protection
  * GDPR-compliant data deletion
  * PII encryption (AES-256-GCM)

#### 3. **Store Service** - Low Latency

* **Performance:** < 10ms response time
* **Optimizations:**
  * Aggressive caching (TTL: 1 hour)
  * Geospatial indexing for location queries
  * Pre-computed aggregations
* **Security:** Role-based store access control

#### 4. **Payment Service** - Critical Security

* **Performance:** 99.99% uptime, < 500ms processing
* **Optimizations:**
  * Asynchronous payment processing
  * Retry logic with exponential backoff
  * Dead letter queue for failed payments
* **Security:**
  * **PCI-DSS Level 1 Compliance**
  * Tokenization (no raw card data storage)
  * End-to-end encryption
  * 3D Secure integration
  * Fraud detection hooks
  * Payment reconciliation audits
  * Chargeback tracking

#### 5. **Inventory Service** - High Consistency

* **Performance:** Real-time stock updates
* **Optimizations:**
  * Optimistic locking for concurrent updates
  * Event-driven stock notifications
  * Predictive caching based on trends
* **Security:** Audit trail for stock movements

#### 6. **Notification Service** - Async Processing

* **Performance:** 100,000+ notifications/minute
* **Optimizations:**
  * Message queue buffering (RabbitMQ/Kafka)
  * Batch sending
  * Template caching
* **Security:** User consent tracking, unsubscribe handling

---

## Service Communication - Optimized

### **Synchronous Communication**

* **gRPC with Protocol Buffers:**
  * 7x faster than JSON REST
  * Binary serialization
  * HTTP/2 multiplexing
  * Streaming support
  * Connection pooling
* **Load Balancing:** Client-side with health checks

### **Asynchronous Communication**

* **Message Queue:** RabbitMQ or Kafka
* **Event-Driven Architecture:**
  * Event sourcing for audit trail
  * CQRS for read/write optimization
  * Saga pattern for distributed transactions
* **Performance:**
  * Message batching
  * Compression (snappy)
  * Persistent connections

### **Service Discovery**

* **Consul** or **Kubernetes DNS**
* Health check endpoints (/health, /ready)
* Graceful shutdown (30s drain period)

---

## Observability - Production Ready

### **Logging - Structured & Secure**

* **Library:** zerolog (0 allocation, fastest)
* **Format:** JSON for machine parsing
* **Levels:** DEBUG, INFO, WARN, ERROR, FATAL
* **Features:**
  * Request ID propagation (trace across services)
  * Sensitive data masking (PII, tokens)
  * Sampling for high-volume logs
  * Log rotation and retention (30 days)
* **Aggregation:** ELK Stack or Grafana Loki

### **Metrics - Comprehensive**

* **Prometheus Metrics:**
  * RED metrics (Rate, Errors, Duration)
  * System metrics (CPU, memory, goroutines)
  * Business metrics (orders/min, revenue)
  * Custom histograms (p50, p95, p99 latencies)
* **Dashboards:** Grafana with alerting
* **SLIs/SLOs:**
  * Availability: 99.95%
  * Latency p95: < 200ms
  * Error rate: < 0.5%

### **Tracing - Distributed**

* **OpenTelemetry + Jaeger**
* **Features:**
  * End-to-end request tracing
  * Service dependency mapping
  * Performance bottleneck identification
  * Sampling strategy (1% in production)

### **Profiling - Continuous**

* **pprof endpoints:** /debug/pprof/*
* **Pyroscope** for continuous profiling
* **Profiles:** CPU, memory, goroutine, mutex

---

## Testing Strategy - Comprehensive

### **Unit Tests**

* Coverage target: > 80%
* Table-driven tests
* Mocking with testify/mock or gomock
* Parallel test execution

### **Integration Tests**

* Testcontainers for PostgreSQL/Redis
* HTTP client tests with httptest
* Database migration testing

### **Load Testing**

* **k6 scenarios:**
  * Ramp-up: 0 → 10,000 users over 5 min
  * Sustained: 10,000 users for 30 min
  * Spike: 50,000 users for 5 min
* **Success Criteria:**
  * P95 latency < 200ms
  * Error rate < 1%
  * No memory leaks

### **Security Testing**

* **OWASP ZAP** automated scanning
* Penetration testing checklist
* Dependency vulnerability scanning
* Secrets scanning (git-secrets)

---

## Deliverables - Production Ready

### **Phase 1: Secure Foundation (Week 1-2)**

1. Project structure (Clean Architecture + DDD)
2. Security-hardened Docker images
3. Database schema with encryption
4. CI/CD with security gates
5. API Gateway with rate limiting

### **Phase 2: High-Performance Services (Week 3-5)**

1. All microservices with performance optimization
2. gRPC service communication
3. Multi-tier caching implementation
4. WebSocket real-time features
5. Message queue integration

### **Phase 3: Security & Monitoring (Week 6-7)**

1. Complete security implementation (OAuth2, MFA, encryption)
2. Observability stack (Prometheus, Grafana, Jaeger)
3. Load testing and optimization
4. Security audit and penetration testing
5. Compliance documentation (PCI-DSS checklist)

### **Phase 4: Production Deployment (Week 8)**

1. Kubernetes deployment manifests
2. Auto-scaling configuration
3. Disaster recovery plan
4. Runbook and incident response
5. Performance benchmarks documentation

---

## Documentation Requirements

* **Architecture Diagrams:**
  * High-level system architecture
  * Service interaction flows
  * Security architecture
  * Data flow diagrams
* **API Documentation:**
  * OpenAPI 3.0 specs
  * Authentication flows
  * Rate limit policies
* **Security Documentation:**
  * Threat model
  * Security controls matrix
  * Incident response procedures
* **Operations Guide:**
  * Deployment procedures
  * Monitoring and alerting
  * Troubleshooting guide
  * Scaling strategies

---

## Performance Benchmarks - Expected

* **API Response Time:**
  * P50: < 50ms
  * P95: < 200ms
  * P99: < 500ms
* **Throughput:**
  * Orders: 10,000/minute
  * Authentication: 5,000/second
  * WebSocket connections: 10,000 concurrent
* **Resource Usage:**
  * Memory: < 512MB per service instance
  * CPU: < 50% under normal load
  * Database connections: < 50 per instance

---

## Security Compliance Checklist

* [ ] OWASP Top 10 mitigation
* [ ] PCI-DSS Level 1 requirements
* [ ] GDPR data protection (right to deletion, encryption)
* [ ] SOC 2 Type II controls
* [ ] Regular security audits (quarterly)
* [ ] Penetration testing (bi-annual)
* [ ] Vulnerability disclosure program
* [ ] Security training for developers

---

**Priority:** Security and performance are non-negotiable. Every feature must be secure by design and performance-tested before production deployment. Follow the principle of "secure by default, fast by design."
