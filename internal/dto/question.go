package dto

// CreateMCQQuestionRequest adds an MCQ to the question bank.
type CreateMCQQuestionRequest struct {
	Topic        string   `json:"topic"`
	Difficulty   string   `json:"difficulty"`
	Marks        int      `json:"marks"`
	Body         string   `json:"body" binding:"required"`
	Options      []string `json:"options" binding:"required,len=4"`
	CorrectIndex int      `json:"correct_index" binding:"min=0,max=3"`
	Explanation  string   `json:"explanation"`
}

// TestCaseRequest defines a programming test case.
type TestCaseRequest struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output" binding:"required"`
	IsHidden       bool   `json:"is_hidden"`
	Weight         int    `json:"weight"`
}

// CreateProgrammingQuestionRequest adds a coding problem to the bank.
type CreateProgrammingQuestionRequest struct {
	Topic         string            `json:"topic"`
	Difficulty    string            `json:"difficulty"`
	Marks         int               `json:"marks"`
	Title         string            `json:"title" binding:"required"`
	Description   string            `json:"description" binding:"required"`
	InputFormat   string            `json:"input_format"`
	OutputFormat  string            `json:"output_format"`
	Constraints   string            `json:"constraints"`
	SampleInput   string            `json:"sample_input"`
	SampleOutput  string            `json:"sample_output"`
	Explanation   string            `json:"explanation"`
	TimeLimitMS   int               `json:"time_limit_ms"`
	MemoryLimitMB int               `json:"memory_limit_mb"`
	TestCases     []TestCaseRequest `json:"test_cases" binding:"required,min=1"`
}

// RunCodeRequest executes code against stdin (practice or dry-run).
type RunCodeRequest struct {
	Language    string `json:"language" binding:"required"`
	Source      string `json:"source" binding:"required"`
	Stdin       string `json:"stdin"`
	TimeLimitMS int    `json:"time_limit_ms"`
}

// SubmitCodingRequest submits code for an assessment question.
type SubmitCodingRequest struct {
	Language   string `json:"language" binding:"required"`
	SourceCode string `json:"source_code" binding:"required"`
}
