package schemas

// Task represents per-task state with phases and steps.
//
// go:generate cue exp gengotypes

#Step: {
	id:          string & !~"\\s" & !=""
	description: string & !=""
	ai_function: "DISCOVER" | "PLAN" | "ASSESS" | "EXECUTE" | "VALIDATE" | "REPORT"
	status:      "pending" | "in-progress" | "completed" | *"pending"
	success_criteria: [...string & !=""] | *[]
}

#Phase: {
	status:      "active" | "pending" | "completed" | *"pending"
	modifiable?: bool | *false
	steps: [...#Step] | *[]
}

// Exported alias for Go type generation
Task: #Task

#Task: {
	id:            string & !=""
	title:         string & !=""
	created_at?:   string
	current_phase: "planning" | "implementation" | "validation"
	status:        "active" | "blocked" | "completed" | *"active"
	phases: {
		planning:       #Phase
		implementation: #Phase
		validation:     #Phase
	}
}
