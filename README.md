# Go Backend for Rinha de Backend 2025

> [!WARNING]
> OBS: Unfortunately, the project was not finished in time for the competition,
> but it is still a good example of a high-performance backend.

This project implements a high-performance payment processing backend using fasthttp
designed for the [Rinha de Backend 2025](https://github.com/zanfranceschi/rinha-de-backend-2025/)

## Technologies Used

- **Go 1.24.5** - Programming language
- **fasthttp** - Ultra-fast HTTP framework (10x faster than net/http)
- **fastjson** - High-performance JSON parser
- **Nginx** - Load balancer and reverse proxy
- **Redis** - Optional caching layer (minimal usage to stay within memory limits)
- **Docker** - Containerization

## Architecture

The application follows a architecture with:

1. **Nginx Load Balancer** - Distributes traffic between 2 API instances
2. **Two API Instances** - Go services handling payment processing
3. **In-Memory Storage** - Ultra-fast storage with pre-computed summaries
4. **Payment Processor Integration** - Connects to default and fallback processors

## Performance Optimizations

- **Pre-computed summaries** for instant /payments-summary responses
- **Nginx optimizations** with keepalive and connection reuse

## Resource Allocation

Total: 1.5 CPU units, 350MB RAM

- **Nginx**: 0.2 CPU, 32MB RAM
- **API Instance 1**: 0.6 CPU, 128MB RAM
- **API Instance 2**: 0.6 CPU, 128MB RAM
- **Redis**: 0.1 CPU, 62MB RAM

## Endpoints

- `POST /payments` - Process payment requests (returns 202 Accepted for speed)
- `GET /payments-summary` - Get payment processing summary with optional time filters

## Performance Targets

- **Sub-10ms p99** response times for performance bonuses
- **Zero downtime** with load balancing

## Running Locally

1. Start the backend:

```bash
make docker-up
```

2. Start Payment Processors:

```bash
make test-preview
```
