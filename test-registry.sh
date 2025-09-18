#!/bin/bash
# Quick test script for local registry operations

set -e

echo "🚀 Building forge-ai CLI..."
mkdir -p bin
cd cli && go build -o ../bin/forge-ai . && cd ..

echo "✅ CLI built successfully (bin/forge-ai)"
echo ""
echo "📦 Creating test template..."
mkdir -p /tmp/forge-test-template
cat > /tmp/forge-test-template/README.md << 'EOF'
# Test Template

This is a test template for Forge AI.
EOF

cat > /tmp/forge-test-template/template.yaml << 'EOF'
template_version: v1.0.0
name: test-template
description: A test template for local registry testing
EOF

echo "✅ Test template created"
echo ""
echo "📤 Publishing to local registry..."
./bin/forge-ai publish \
  --source=/tmp/forge-test-template \
  --registry=localhost:5050/test/forge-template:v1.0.0

echo ""
echo "🔍 Verifying in registry..."
echo "Repositories:"
curl -s http://localhost:5050/v2/_catalog | jq .
echo ""
echo "Tags for test/forge-template:"
curl -s http://localhost:5050/v2/test/forge-template/tags/list | jq .

echo ""
echo "📥 Testing pull with init..."
TEST_DIR="/tmp/forge-init-test-$(date +%s)"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"
/Users/josh/work/catalyst-forge-ai/bin/forge-ai init test-project \
  --template=localhost:5050/test/forge-template:v1.0.0

echo ""
echo "✅ Project initialized at: $TEST_DIR"
echo "Project structure:"
ls -la .forge/
echo ""
echo "Project config:"
cat .forge/project.yaml

echo ""
echo "🎉 All tests passed! Registry is working correctly."
echo ""
echo "🌐 View the Zot web UI at: http://localhost:5050"