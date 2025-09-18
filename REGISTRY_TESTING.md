# Local Registry Testing with Zot

This setup provides a local OCI registry (Zot) for testing the `forge-ai publish` and `forge-ai init` commands.

**Note:** The registry runs on port 5050 (not 5000) to avoid conflicts with macOS AirPlay.

## Quick Start

### 1. Start the Registry

```bash
# Start Zot registry on port 5050
docker-compose up -d

# Check that it's running
docker-compose ps

# View logs if needed
docker-compose logs -f zot
```

### 2. Access the Web UI

Open http://localhost:5050 in your browser to access the Zot web UI.

### 3. Test Publishing

```bash
# Build the CLI if you haven't already
mkdir -p bin
cd cli && go build -o ../bin/forge-ai . && cd ..

# Create a sample template directory
mkdir -p /tmp/test-template
echo "# Test Template" > /tmp/test-template/README.md
echo "template content" > /tmp/test-template/template.yaml

# Publish to local registry (note port 5050)
./bin/forge-ai publish --source=/tmp/test-template --registry=localhost:5050/test/my-template:v1.0.0
```

### 4. Test Pulling (Init)

```bash
# Create a test directory
mkdir -p /tmp/test-project
cd /tmp/test-project

# Initialize a project from the published template
/path/to/bin/forge-ai init my-project --template=localhost:5050/test/my-template:v1.0.0
```

### 5. View Registry Contents

You can verify the published artifacts in the web UI at http://localhost:5050

### 6. Clean Up

```bash
# Stop and remove the registry
docker-compose down

# Remove registry data (optional)
rm -rf zot-data/
```

## Configuration Notes

The Zot registry is configured with:
- **No authentication** for easy local testing
- **Anonymous access** with full read/write permissions
- **Web UI** enabled for browsing artifacts
- **Search API** enabled for artifact discovery
- Data persisted in `./zot-data` directory

The CLI binary is built to `./bin/forge-ai` which is gitignored to prevent accidental commits.

## Troubleshooting

### Connection Refused
If you get connection refused errors, ensure:
1. Docker is running
2. Port 5050 is available (we use 5050 instead of 5000 to avoid macOS AirPlay conflicts)
3. The registry container is healthy: `docker-compose ps`

### Permission Denied
If you get permission errors when publishing:
- The local registry is configured for anonymous access
- Check docker-compose logs: `docker-compose logs zot`

### HTTP vs HTTPS
The local registry runs on HTTP. If the CLI requires HTTPS:
- You may need to configure the OCI client to allow insecure registries
- Or set up TLS certificates (more complex for local testing)