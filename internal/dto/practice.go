package dto

// CreatePracticeModuleRequest creates a new practice folder.
type CreatePracticeModuleRequest struct {
	Name        string `json:"name" binding:"required"`
	Category    string `json:"category"`    // "mcq" | "coding" | "mixed"
	Description string `json:"description"`
	Tags        string `json:"tags"`        // comma-separated topic tags
	Ord         int    `json:"ord"`
}

// UpdatePracticeModuleRequest updates module metadata.
type UpdatePracticeModuleRequest struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	IsPublished *bool  `json:"is_published"`
	Ord         int    `json:"ord"`
}

// ModuleQuestionEntry is one question slot inside a module, with optional overrides.
type ModuleQuestionEntry struct {
	QuestionID  string `json:"question_id" binding:"required"`
	Ord         int    `json:"ord"`
	Marks       int    `json:"marks"`        // 0 = inherit question default
	MaxAttempts int    `json:"max_attempts"` // 0 = unlimited
}

// AddModuleQuestionsRequest bulk-adds (or updates) questions in a module.
type AddModuleQuestionsRequest struct {
	Questions []ModuleQuestionEntry `json:"questions" binding:"required,min=1"`
}

// UpdateModuleQuestionRequest patches a single slot's overrides.
type UpdateModuleQuestionRequest struct {
	Marks       int `json:"marks"`
	MaxAttempts int `json:"max_attempts"`
	Ord         int `json:"ord"`
}

// ReorderModuleQuestionsRequest sets the full ordering of a module's questions.
type ReorderModuleQuestionsRequest struct {
	QuestionIDs []string `json:"question_ids" binding:"required,min=1"`
}

// BulkMCQRequest imports many MCQ questions in one call.
type BulkMCQRequest struct {
	Questions []CreateMCQQuestionRequest `json:"questions" binding:"required,min=1"`
}

// BulkCodingRequest imports many programming questions in one call.
type BulkCodingRequest struct {
	Questions []CreateProgrammingQuestionRequest `json:"questions" binding:"required,min=1"`
}

// ModuleQuestionDetail is returned per question slot (with overrides visible).
type ModuleQuestionDetail struct {
	ID          string   `json:"id"`
	QuestionID  string   `json:"question_id"`
	Ord         int      `json:"ord"`
	Marks       int      `json:"marks"`
	MaxAttempts int      `json:"max_attempts"`
	Question    *QuestionSummary `json:"question,omitempty"`
}

// QuestionSummary is a lightweight question view for module listings.
type QuestionSummary struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Topic      string `json:"topic"`
	Difficulty string `json:"difficulty"`
	Marks      int    `json:"marks"`
	Title      string `json:"title,omitempty"` // coding
	Body       string `json:"body,omitempty"`  // mcq
}
