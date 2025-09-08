# Go State Machine

A simple and extensible state machine implementation in Go, designed to orchestrate complex workflows. This project provides a core engine for defining and executing multi-step processes with support for branching, parallel execution, error handling, and more. It is inspired by concepts from AWS Step Functions.

## Features

- **State-based Execution:** Define your workflow as a series of distinct states that handle specific tasks.
- **Multiple State Types:**
  - `Task`: Executes a Go function.
  - `Pass`: Passes data from one state to the next, with optional data modification.
  - `Choice`: Implements conditional branching based on the state context.
  - `Map`: Processes an array of items concurrently by running a sub-workflow for each item.
  - `Parallel`: Executes multiple independent branches concurrently.
  - `Wait`: Pauses the workflow for a specified duration.
  - `Fail`: Halts the workflow with a failure.
  - `End`: Terminates a workflow successfully.
- **Resilient Error Handling:** Use `Retry` and `Catch` rules on `Task` states to automatically handle transient failures or transition to a different state on specific errors.
- **Timeouts:** Prevent a single task from blocking the entire workflow indefinitely by specifying a `TimeoutSeconds` property.
- **Declarative & Programmatic Definitions:** Define your workflows either directly in Go code using a fluent builder or with a declarative JSON file.

## Getting Started

### Prerequisites

- Go (1.23.3 or newer)

### Running the Example Workflow

1.  Clone this repository:
    ```bash
    git clone git@github.com:Basillica/go-statemachine.git
    cd go-state-machine
    ```
2.  Run the programmatic example:
    ```bash
    go run main.go
    ```
3.  Run the JSON-based example:
    ```bash
    go run main.go --json
    ```

### Running Tests

Execute the unit tests to verify the functionality of all states:

```bash
go test
```
