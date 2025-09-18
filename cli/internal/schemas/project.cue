package schemas

// Project represents project-level state managed by the CLI.
// This is the single source of truth for generation.
//
// go:generate cue exp gengotypes

// Exported alias for Go type generation
Project: #Project

#Project: {
	projectName: string | *""
	template: {
		source:  string & !=""
		version: string & !=""
	}
	status: "active" | "maintenance" | "archived" | *"active"

	// Currently active task context for convenience. Optional.
	activeTask?: string & !~"\\s"

	// Registry of known tasks
	tasks?: [string]: {
		path:             string & !~"\\s"
		phase:            "planning" | "implementation" | "validation"
		status:           "active" | "blocked" | "completed"
		template_version: string & !=""
		created_at?:      string
		completed_at?:    string
	}
}
