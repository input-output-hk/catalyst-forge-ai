package schemas

import "time"

#Project: {
    // Project metadata
    projectName: string & =~"^.+$"  // Non-empty string
    template: {
        source: string & =~"^(ghcr|docker|quay)\\.io/[a-z0-9-/]+$"  // Valid OCI registry URL
        version: string & =~"^v\\d+\\.\\d+\\.\\d+$"  // Semantic version (e.g., v1.0.0)
    }
    status: "active" | "maintenance" | "archived"
    activeTask?: string  // Must reference existing task ID when validated

    // Task registry
    tasks: [ID=string]: {
        path: string & =~"^tasks/[a-z0-9-]+$"  // Relative path to task directory
        phase: "planning" | "implementation" | "validation"
        status: "active" | "blocked" | "completed" | "abandoned"
        template_version: string & =~"^v\\d+$"  // e.g., "v1", "v2"
        created_at: time.Time
        completed_at?: time.Time

        // Business rule: completed_at required if status is completed/abandoned
        if status == "completed" || status == "abandoned" {
            completed_at: time.Time
        }
    }

    // Ensure activeTask references an existing task
    if activeTask != _|_ {
        activeTask: or([ for id, _ in tasks {id}])
    }
}