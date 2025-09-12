# **Forge AI: MVP Implementation Guide**

## **Objective**

This guide provides a sequential set of instructions to implement the Minimum Viable Product (MVP) of the forge-ai CLI. The focus is exclusively on the components required to demonstrate the core workflow: publishing a template, creating a task, and executing the full task lifecycle via the Model Context Protocol (MCP).

## **1.0 Setup and Distribution**

**Goal:** Implement the commands to prepare and distribute the initial MVP template files.

### **1.1 Implement Template Filesystem Preparation**

The first step is to create a minimal set of template files on disk that the publish command will use as its source.

**Instructions:**

* Create a source directory containing the following minimal structure and content. These files should be lean and contain only the essential fields and instructions needed for the MVP demo flow.

template\_source/  
├── manifest.yml  
├── functions/  
│   ├── planning/  
│   │   ├── DISCOVER.ai.md  
│   │   ├── PLAN.ai.md  
│   │   └── ASSESS.ai.md  
│   ├── implementation/  
│   │   └── EXECUTE.ai.md  
│   └── validation/  
│       ├── VALIDATE.ai.md  
│       └── REPORT.ai.md  
├── schemas/  
│   ├── project.cue  
│   └── task.cue  
└── templates/  
    └── tasks/  
        └── default.yaml

* **manifest.yml**: Must contain name and version fields.  
* **AI Functions (\*.ai.md)**: Each should contain the standard preamble and a simple one-sentence instruction relevant to its purpose (e.g., PLAN.ai.md should instruct the agent to use the step\_add tool).  
* **CUE Schemas (\*.cue)**: Implement the minimal viable schemas for project.cue and task.cue to validate the state files.  
* **default.yaml**: This is the task template. It should define the three phases (planning, implementation, validation) and their initial steps as specified in the architecture.

*For detailed file contents and structure, refer to the DISTRIBUTION.md and ARCHITECTURE.md documents.*

### **1.2 Implement the publish Command**

This command will package the source directory from step 1.1 and push it to an OCI registry.

**Instructions:**

* Create the command forge-ai publish \--source=\<path\> \--registry=\<oci-ref\>.  
* Internally, this command must use the existing Go module located at lib/oci.  
* Instantiate a client from the ocibundle package.  
* Call the client.Push(ctx, sourceDir, reference) method, passing the source path and registry reference.

*For details on the distribution method, refer to DISTRIBUTION.md.*

## **2.0 Project and Task Initialization**

**Goal:** Implement the user-facing commands to initialize a project and create a new task from the published template.

### **2.1 Implement the init Command**

This command pulls the template from the OCI registry and sets up the local project structure.

**Instructions:**

* Create the command forge-ai init \<project-name\> \--template=\<oci-ref\>.  
* This command will also use the lib/oci Go module.  
* It should call the client.Pull(ctx, reference, targetDir) method to download and extract the template files into the .forge/ai directory.  
* After pulling the template, the command must create the initial project state file: .forge/ai/project.yaml.

*For details on the file structure after installation and the project.yaml specification, refer to DISTRIBUTION.md and STATE.md.*

### **2.2 Implement the task new Command**

This command creates a new task instance from the default.yaml template.

**Instructions:**

* Create the command forge-ai task new \--title="\<task-title\>".  
* The command must:  
  1. Generate a unique task ID.  
  2. Create a new task directory: tasks/\<task-id\>/.  
  3. Copy the templates/tasks/default.yaml into tasks/\<task-id\>/task.yaml.  
  4. Update tasks/\<task-id\>/task.yaml with the unique ID and title.  
  5. Register the new task in the .forge/ai/project.yaml file.  
  6. The initial current\_phase in task.yaml must be set to planning.

*For the task state file structure and state transition rules, refer to STATE.md.*

## **3.0 Executing the Task Lifecycle via MCP**

**Goal:** Implement the MCP server and the core logic to guide an AI agent through the three-phase task lifecycle.

### **3.1 Implement the MCP Server and Agent Interaction Loop**

The forge-ai mcp serve command is the entry point for AI agent interaction. The core of this interaction is the next tool.

**Instructions:**

* Create the command forge-ai mcp serve. This starts the STDIO-based JSON-RPC server.  
* Implement the next tool within the MCP server.  
* When an agent calls next, the server must:  
  1. Read the active task's task.yaml state file.  
  2. Identify the current\_phase and find the first step within that phase that is not yet completed.  
  3. Return the corresponding AI Function (e.g., DISCOVER.ai.md) as a response.

*For the MCP protocol and architectural details, refer to ARCHITECTURE.md.*

### **3.2 Implement Phase 1: Planning**

This phase involves executing three AI Functions sequentially and requires the implementation of the step\_add tool.

**Instructions:**

* **AI Function Execution:** The agent will call next three times, executing DISCOVER, PLAN, and ASSESS in order.  
* **Implement step\_add Tool:** This tool is critical for the PLAN function. It must accept a phase ID and step details, and append the new step to the steps list for that phase in task.yaml. The CLI must enforce that steps can only be added to a phase marked as modifiable: true.  
* **Implement Human Gate:** The forge-ai task phase next command must be implemented. When run, it should:  
  1. Validate that all steps in the planning phase are complete.  
  2. Update the current\_phase in task.yaml from planning to implementation.

*For function-specific behaviors and phase transition rules, refer to FUNCTIONS.md and STATE.md.*

### **3.3 Implement Phase 2: Implementation**

This phase involves executing the steps defined during planning.

**Instructions:**

* **AI Function Execution:** The agent will repeatedly call next to receive the EXECUTE.ai.md function, contextualized for each step that was added via step\_add.  
* **Implement artifact\_save and step\_complete Tools:**  
  * artifact\_save: This tool should register a file path as a deliverable for the current step. For the MVP, this can simply be recording the path in the task.yaml file.  
  * step\_complete: This tool marks the current step's status as completed in task.yaml.  
* **Implement Human Gate:** The forge-ai task phase next command logic should be extended. When the current phase is implementation, it must:  
  1. Validate that all steps in the implementation phase are complete.  
  2. Update the current\_phase in task.yaml from implementation to validation.

*For details, refer to FUNCTIONS.md and STATE.md.*

### **3.4 Implement Phase 3: Validation**

This final phase verifies the work and completes the task.

**Instructions:**

* **AI Function Execution:** The agent will call next twice to execute the VALIDATE and REPORT functions.  
* **Implement task complete Command:** Create the forge-ai task complete command. This command is the final human gate. It should:  
  1. Validate that all steps in the validation phase are complete.  
  2. Update the task's status to completed in task.yaml.  
  3. Update the task's status to completed in the main .forge/ai/project.yaml registry.

*For details on the final state transitions, refer to STATE.md.*