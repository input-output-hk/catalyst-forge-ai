# Template Distribution Architecture

## Overview

Forge AI templates are distributed as OCI (Open Container Initiative) images containing a single compressed archive of template files. This approach leverages existing container registry infrastructure for versioning, distribution, and authentication while keeping the implementation extremely simple.

## Core Design

### What Gets Distributed

Each Forge AI template is a small collection of plain text files:

```
.forge/ai/                        # Total size: ~50-100KB uncompressed
├── manifest.yml                  # Template metadata and version
├── templates/                    # Task templates
│   └── tasks/
│       ├── default.yaml
│       ├── bugfix.yaml
│       └── research.yaml
├── functions/                    # AI Function definitions (~30-50KB)
│   ├── planning/
│   │   ├── DISCOVER.ai.md
│   │   ├── PLAN.ai.md
│   │   └── ASSESS.ai.md
│   ├── implementation/
│   │   └── EXECUTE.ai.md
│   └── validation/
│       ├── VALIDATE.ai.md
│       └── REPORT.ai.md
├── schemas/                      # CUE schemas (~10-20KB)
│   ├── project.cue
│   ├── task.cue
│   └── memory_index.cue
└── seeds/                        # Optional starter content
    └── memory/                   # Initial memory entries
        └── 001-template-overview.md
```

### Distribution Method

Templates are packaged as a single tar.gz archive and attached to an OCI image using ORAS (OCI Registry As Storage). The entire template is shipped as one artifact - no complex layering or chunking.

## Manifest Specification

The `manifest.yml` file contains metadata about the template and declares its components:

```yaml
# Template identity
name: "forge-golang-cli"
version: "v1.2.0"              # Semantic version for entire template
description: "Forge AI template for Go CLI applications"

# Components included in this template
components:
  ai_functions:
    - planning/DISCOVER.ai.md
    - planning/PLAN.ai.md
    - planning/ASSESS.ai.md
    - implementation/EXECUTE.ai.md
    - implementation/CHECKPOINT.ai.md
    - validation/VALIDATE.ai.md
    - validation/REPORT.ai.md

  schemas:
    - project.cue
    - task.cue
    - memory_index.cue

# MCP tools shipped by template (specification)
tools:
  core:
    - next
    - artifact_save
    - step_start
    - step_complete
  planning:
    - step_add
  memory:
    - memory_add
    - memory_search
    - memory_relevant

# Task templates shipped by template (declarations)
templates:
  tasks:
    default: "default.yaml"
    available:
      - name: "default"
        file: "default.yaml"
        description: "Standard three-phase workflow"
      - name: "bugfix"
        file: "bugfix.yaml"
        description: "Streamlined bug fixing"
      - name: "research"
        file: "research.yaml"
        description: "Research and analysis"

# Optional metadata
author: "Catalyst Forge Team"
license: "MIT"
homepage: "https://github.com/catalyst-forge/golang-cli-template"
```

### Version Rules

- **Single Version**: The manifest version represents ALL template files
- **Semantic Versioning**: Major.Minor.Patch (e.g., v1.2.0)
- **Major Version Compatibility**: CLI embeds a single major version (e.g., v1)
- **Breaking Changes**: Occur when migrations cannot handle the changes

When ANY file in the template changes (AI Function, schema, etc.), the manifest version must be incremented.

Templates are versioned independently. Tasks record the `template_version` they were instantiated from; updating a template affects new tasks.

## Publishing Process

The `forge-ai` CLI has the ORAS Go library embedded directly in its binary. When publishing templates, the CLI handles all OCI operations internally without requiring any external tools:

```bash
forge-ai publish \
  --source=./template-source \
  --registry=ghcr.io/input-output-hk/catalyst-forge-ai:v1.2.0
forge-ai init --template=ghcr.io/input-output-hk/catalyst-forge-ai:v1.2.0 myproject
```

This single command:
1. Creates a tar.gz archive from the template source directory
2. Uses the embedded ORAS Go library to push the archive as an OCI artifact
3. Sets the artifact type as `application/vnd.forge.template`
4. Optionally tags the version as `latest` if specified

No external ORAS CLI or other tools are needed - everything is built into the `forge-ai` binary.

## Installation Process

When initializing a new project with a template:

```bash
forge-ai init myproject --template=ghcr.io/catalyst-forge/golang-cli:v1.2.0
```

The CLI performs these steps:

1. **Authentication**: Uses local Docker configuration (~/.docker/config.json)
2. **Fetch**: Downloads the template archive using ORAS
3. **Compatibility Check**: Compares manifest major version with CLI's embedded version
4. **Extract**: Unpacks the archive into the project's `.forge/ai` directory

### Compatibility Check

The CLI embeds a single major version:

```
CLI v1.5.0 embeds major version: v1
```

During installation:
- Template v1.0.0 → Compatible (major version v1)
- Template v1.9.5 → Compatible (major version v1)
- Template v2.0.0 → Incompatible (major version v2)

If major versions don't match, the CLI exits with an error:

```
Error: Template version v2.0.0 is incompatible with CLI v1.5.0
The CLI supports templates with major version v1.
Please upgrade your CLI or use a compatible template version.
```

## Upgrade Process

When upgrading an existing project to a new template version:

```bash
forge-ai template upgrade --to=ghcr.io/catalyst-forge/golang-cli:v1.3.0
```

The process is:

1. **Fetch New Template**: Download to temporary directory
2. **Check Compatibility**: Verify major versions match
3. **Backup Current State**: Create backup of existing `.forge/ai` directory
4. **Apply Migration**: Run migration strategy if needed
5. **Replace Files**: Copy new template files to `.forge/ai`
6. **Update project.yaml**: Record new template version

### Migration Strategy

For minor/patch version upgrades within the same major version:

**Simple Replacement** (default):
- Replace all template-provided files (functions/, schemas/)
- Preserve user state files (project.yaml, tasks/, memory/)

**Custom Migration** (when needed):
- Template can include migration instructions in manifest
- CLI applies specific transformations to state files
- Migrations must succeed or upgrade is rolled back

If migration fails, the backup is restored automatically.

## Registry Support

### Authentication

The CLI uses the local Docker configuration for all registry authentication:

```
~/.docker/config.json
```

No additional authentication methods are supported. Users must use `docker login` or equivalent to authenticate with registries before using Forge AI templates.

### Supported Registries

Any OCI-compliant registry that ORAS supports:
- GitHub Container Registry (ghcr.io)
- Docker Hub
- Google Artifact Registry
- Amazon ECR
- Azure Container Registry
- Harbor
- Local/private registries

## Caching

The CLI implements simple caching to avoid repeated downloads:

```
~/.forge/cache/
└── templates/
    └── ghcr.io/
        └── catalyst-forge/
            └── golang-cli/
                ├── v1.2.0.tar.gz         # Cached template archive
                ├── v1.2.0.manifest.yml   # Cached manifest for quick checks
                └── v1.2.0.timestamp      # Cache timestamp
```

### Cache Behavior

- **Cache Duration**: 24 hours by default
- **Cache Key**: Full template reference including version
- **Invalidation**: Manual via `forge-ai cache clean` or automatic after TTL
- **Size Limit**: 100MB total cache size (templates are tiny, this allows many)

When installing a template:
1. Check if cached and not expired
2. If valid cache exists, use it
3. Otherwise, fetch from registry and update cache

## Error Handling

### Common Errors

**Registry Authentication Failed**:
```
Error: Authentication failed for ghcr.io
Please run: docker login ghcr.io
```

**Template Not Found**:
```
Error: Template not found: ghcr.io/catalyst-forge/golang-cli:v1.2.0
Verify the template name and version are correct.
```

**Version Incompatibility**:
```
Error: Template requires CLI major version v2, but CLI supports v1
Please upgrade your CLI or use a compatible template.
```

**Network Issues**:
```
Error: Failed to fetch template: connection timeout
Check your network connection and try again.
```

## File Structure After Installation

After successful template installation:

```
myproject/
└── .forge/
    └── ai/
        ├── manifest.yml           # Template manifest (from template)
        ├── project.yaml           # Project state (created by CLI)
        ├── templates/             # Task templates (from template)
        │   └── tasks/
        │       ├── default.yaml
        │       └── ...
        ├── functions/            # AI Functions (from template)
        ├── schemas/              # CUE schemas (from template)
        ├── memory/               # Memory system (created by CLI)
        │   └── project/
        │       └── index.yaml
        └── tasks/                # Task storage (created by CLI)
```

Template provides:
- manifest.yml
- templates/
- functions/
- schemas/

CLI creates:
- project.yaml
- memory/ structure
- tasks/ structure (each `forge-ai task new` instantiates a template into `tasks/<task-id>/task.yaml`)

## Design Rationale

### Why OCI?

- **Existing Infrastructure**: Reuse container registries users already have
- **Standard Tooling**: ORAS provides robust OCI artifact support
- **Simple Security**: Leverage existing Docker authentication
- **Global Distribution**: Registry providers offer worldwide availability

### Why Single Archive?

- **Simplicity**: One artifact to manage, no complex layering
- **Size**: Templates are tiny (~50-100KB), no need for optimization
- **Atomicity**: All-or-nothing updates prevent partial states
- **Debugging**: Easy to inspect what's being distributed

### Why ORAS?

- **Purpose-Built**: Designed specifically for non-container OCI artifacts
- **Lightweight**: Minimal dependencies and simple CLI
- **Standard**: Follows OCI specifications properly
- **Maintained**: Active development and wide adoption

## Summary

The Forge AI distribution system is intentionally minimal:

1. **Single Archive**: Templates are distributed as one tar.gz file via OCI
2. **Simple Manifest**: Contains version, component list, and basic metadata
3. **ORAS Library**: Embedded Go library handles OCI artifact management
4. **Docker Auth**: Leverages existing Docker configuration
5. **Major Version Check**: Single compatibility check between CLI and template
6. **Basic Caching**: Simple time-based cache for efficiency
7. **Straightforward Upgrade**: Fetch, check, backup, replace

This design prioritizes simplicity and reliability over complex optimization, recognizing that templates are small text files that don't require elaborate distribution mechanisms. The entire system can be understood and debugged easily while still providing versioning, authentication, and global distribution through standard OCI registries.