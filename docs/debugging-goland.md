# Debugging Services in GoLand

This guide explains how to debug microservices in this repository using JetBrains GoLand IDE. Two approaches are covered:

1. **Local Debugging** - Run services directly from GoLand with the debugger attached
2. **Remote Debugging** - Attach GoLand to a Delve debugger running in a Docker container

Choose the approach that best fits your workflow. Local debugging offers faster iteration, while remote debugging provides a production-like environment.

## Prerequisites

- **GoLand IDE** - JetBrains GoLand installed and configured
- **Docker & Docker Compose** - For running infrastructure dependencies
- **Service Dependencies** - MySQL 8.0 and RabbitMQ 3.9.11 (managed via docker-compose)

## Approach 1: Local Debugging (Direct Run)

Run the service directly from GoLand with the debugger attached. This approach provides fast iteration cycles and native IDE integration.

### 1. Start Infrastructure Dependencies

Start only the required infrastructure services (MySQL and RabbitMQ) without running the application service:

```bash
docker-compose up -d mysql rabbitmq
```

Verify services are running:

```bash
docker-compose ps mysql rabbitmq
```

### 2. Create GoLand Run Configuration

1. Open GoLand and navigate to **Run → Edit Configurations**
2. Click the **+** button and select **Go Build**
3. Configure the following settings:

   - **Name**: `symbol-service (debug)`
   - **Run kind**: `Package`
   - **Package path**: `./services/symbols/cmd/symbol-service`
   - **Working directory**: `$ProjectFileDir$/services/symbols`
   - **Program arguments**: `--conf=./configs/config.yaml`
   - **Environment variables** (optional): Add any `KRATOS_*` overrides if needed
     - Example: `KRATOS_DATA_DATABASE_DSN=root:root@tcp(localhost:3306)/symbols?parseTime=True`
   - **Before launch** (optional): Add "Run External tool" → `make generate` to ensure Wire dependencies are up-to-date

4. Click **Apply** and **OK**

### 3. Set Breakpoints and Debug

1. Open the source file where you want to debug (e.g., `internal/service/symbols.go`)
2. Click in the gutter (left margin) next to the line number to set a breakpoint
3. Select the `symbol-service (debug)` configuration from the run configurations dropdown
4. Click the **Debug** button or press **Shift+F9**

The service will start with the debugger attached. Access endpoints at:

- **HTTP API**: `http://localhost:8000`
- **gRPC**: `localhost:7000`

### 4. Benefits and Limitations

**Benefits:**
- Fast iteration cycles (no Docker rebuild required)
- Native IDE integration with instant breakpoint updates
- Easy variable inspection and expression evaluation
- Direct access to service logs in GoLand console

**Limitations:**
- Dependencies must be managed separately (MySQL, RabbitMQ)
- Environment may differ from production Docker container
- Requires manual configuration of environment variables
- Traefik routing not available (must use direct ports)

## Approach 2: Remote Debugging (Docker + Delve)

Attach GoLand to a Delve debugger running inside a Docker container. This approach provides a production-like environment with all dependencies managed automatically.

### 1. Start Services with Debug Configuration

The `symbol-service` in docker-compose is already configured to use `Dockerfile.debug`, which includes the Delve debugger.

Start the service:

```bash
docker-compose up symbols
```

This will:
- Build the service with debug flags: `-gcflags="all=-N -l"` (disables optimizations)
- Start Delve debugger on port **2345**
- Expose HTTP on port **8000** and gRPC on port **7000**

Verify the service is running:

```bash
docker-compose ps
```

### 2. Create GoLand Remote Debug Configuration

1. Open GoLand and navigate to **Run → Edit Configurations**
2. Click the **+** button and select **Go Remote**
3. Configure the following settings:

   - **Name**: `symbol-service (remote debug)`
   - **Host**: `localhost`
   - **Port**: `2345`
   - **On disconnect**: `Leave it running` (allows reconnection without restarting the container)

4. Click **Apply** and **OK**

### 3. Attach Debugger

1. Ensure the service is running in Docker: `docker-compose ps`
2. Set breakpoints in your code (they will sync to the running container)
3. Select the `symbol-service (remote debug)` configuration
4. Click the **Debug** button

GoLand will connect to Delve on port 2345. You should see "Connected" in the Debug console.

### 4. Access Service Endpoints

The service is accessible through Traefik or directly:

**Via Traefik:**
- HTTP API: `http://symbols.localhost`
- gRPC: `symbols-grpc.localhost:7000`

**Direct access:**
- HTTP API: `http://localhost:8000`
- gRPC: `localhost:7000`

### 5. Viewing Logs

View service logs in real-time:

```bash
docker-compose logs -f symbols
```

### 6. Making Code Changes

When you modify code:

1. Stop the debugger in GoLand (click the Stop button)
2. Rebuild and restart the container:

   ```bash
   docker-compose up --build symbols
   ```

3. Reattach the debugger from GoLand

### 7. Benefits and Limitations

**Benefits:**
- Production-like environment (exact Docker container)
- All dependencies automatically managed (MySQL, RabbitMQ, Traefik)
- Traefik routing enabled (test with production-like URLs)
- No manual dependency configuration required
- Multiple developers can use identical setup

**Limitations:**
- Slower iteration cycle (requires Docker rebuild for code changes)
- Slightly higher latency due to Docker networking
- Container must be rebuilt after dependency changes

## Tips and Troubleshooting

### "Cannot connect to Delve on port 2345"

**Symptoms:** GoLand shows "Connection refused" when trying to attach remote debugger.

**Solutions:**
1. Verify service is running:
   ```bash
   docker-compose ps
   ```

2. Check port mapping:
   ```bash
   docker-compose port symbols 2345
   ```

3. Rebuild the container:
   ```bash
   docker-compose up --build symbols
   ```

4. Check if port 2345 is blocked by firewall or another process:
   ```bash
   lsof -i :2345
   ```

### Breakpoints Not Hitting

**Symptoms:** Debugger is connected but breakpoints are skipped or grayed out.

**Solutions:**
1. **Remote debugging:** Ensure code version matches the running binary
   - Rebuild container after code changes: `docker-compose up --build symbols`

2. Verify breakpoints are in an executed code path
   - Add a breakpoint in the handler method that processes your request

3. Check if the code is optimized (local debugging)
   - Ensure build tags don't enable optimizations

### Database Connection Errors (Local Debugging)

**Symptoms:** Service fails to start with MySQL connection errors.

**Solutions:**
1. Ensure MySQL is running:
   ```bash
   docker-compose ps mysql
   ```

2. Check `configs/config.yaml` has correct connection string:
   ```yaml
   data:
     database:
       dsn: root:root@tcp(localhost:3306)/symbols?parseTime=True
   ```

3. Override via environment variable:
   ```bash
   KRATOS_DATA_DATABASE_DSN=root:root@tcp(localhost:3306)/symbols?parseTime=True
   ```

4. Verify MySQL port is accessible:
   ```bash
   docker-compose port mysql 3306
   ```

### Port Conflicts

**Symptoms:** Service fails to start with "address already in use" error.

**Solutions:**
1. Check if ports 7000, 8000, or 2345 are already in use:
   ```bash
   lsof -i :7000
   lsof -i :8000
   lsof -i :2345
   ```

2. Stop conflicting services or processes

3. Modify port mappings in `docker-compose.yml` if needed:
   ```yaml
   ports:
     - "8001:8000"  # Map to different host port
   ```

### Wire Dependency Injection Errors

**Symptoms:** Service fails to start with "provider not found" or similar Wire errors.

**Solutions:**
1. Regenerate Wire code:
   ```bash
   cd services/symbols
   make generate
   ```

2. Ensure `wire_gen.go` is up-to-date and committed

3. Check `wire.go` for any missing providers in ProviderSets

## Quick Reference

| Aspect | Local Debugging | Remote Debugging |
|--------|----------------|------------------|
| **Start command** | Run from GoLand | `docker-compose up symbols` |
| **Debugger** | GoLand native | Delve (port 2345) |
| **Dependencies** | Manual (docker-compose) | Automatic (docker-compose) |
| **Iteration speed** | Fast | Slower (rebuild needed) |
| **Environment** | Local machine | Docker container |
| **Traefik routing** | Not available | Available |
| **Use case** | Rapid development | Production-like testing |
| **Code changes** | Instant | Requires rebuild |
| **Setup complexity** | Medium | Low |

## Additional Resources

- **[CLAUDE.md](../CLAUDE.md)** - Complete development guide with architecture patterns
- **[services/symbols/Dockerfile.debug](../services/symbols/Dockerfile.debug)** - Debug container configuration
- **[docker-compose.yml](../docker-compose.yml)** - Service orchestration and port mappings
- **[Delve Documentation](https://github.com/go-delve/delve/tree/master/Documentation)** - Official Delve debugger docs
- **[GoLand Debugging Guide](https://www.jetbrains.com/help/go/debugging-code.html)** - JetBrains official debugging docs
