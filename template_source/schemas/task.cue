package schemas

import (
    "time"
    "list"
)

#Task: {
    // Task identity
    id: string & =~"^[a-z0-9][a-z0-9-]*$"
    title: string & =~"^.+$"
    description?: string
    created_at: time.Time

    // Current status
    current_phase: "planning" | "implementation" | "validation"
    status: "active" | "blocked" | "completed" | "abandoned"
    blocked_reason?: string

    if status == "blocked" {
        blocked_reason: string & =~"^.+$"
    }

    // Unified phase→step model
    phases: {
        planning: #Phase & {modifiable: bool}
        implementation: #Phase & {modifiable: bool}
        validation: #Phase & {modifiable: bool}
    }

    #Phase: {
        status: "pending" | "active" | "completed" | "skipped"
        started_at?: time.Time
        completed_at?: time.Time
        notes?: string
        steps: [...#PhaseStep]
        modifiable: bool | *false

        if status == "completed" { 
            completed_at: time.Time
        }
        if status == "active" || status == "completed" { 
            started_at: time.Time
        }
    }

    #PhaseStep: {
        id: string & =~"^[a-z][a-z0-9-]*$"
        description?: string & =~"^.+$"
        ai_function: "DISCOVER" | "PLAN" | "ASSESS" | "EXECUTE" | "VALIDATE" | "REPORT"
        status: "pending" | "in-progress" | "completed" | "abandoned" | "blocked"
        started_at?: time.Time
        completed_at?: time.Time
        abandoned_at?: time.Time
        blocked_at?: time.Time
        blocked_reason?: string
        artifacts?: [...string]
        evidence?: [...string]
        success_criteria?: [...string]

        if status == "in-progress" { started_at: time.Time }
        if status == "completed" {
            started_at: time.Time
            completed_at: time.Time
            evidence: [...string] & list.MinItems(1)
        }
        if status == "abandoned" { abandoned_at: time.Time }
        if status == "blocked" { 
            blocked_at: time.Time
            blocked_reason: string & =~"^.+$"
        }
    }

}