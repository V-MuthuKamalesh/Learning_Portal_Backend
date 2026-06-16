package models

import "github.com/google/uuid"

// Question type & difficulty enums.
const (
	QuestionMCQ         = "mcq"
	QuestionProgramming = "programming"

	DifficultyEasy   = "easy"
	DifficultyMedium = "medium"
	DifficultyHard   = "hard"
)

// Question is the polymorphic parent; details live in MCQQuestion / ProgrammingQuestion.
type Question struct {
	Base
	CollegeID  uuid.UUID   `gorm:"type:uuid;not null;index" json:"college_id"`
	Type       string      `gorm:"size:20;not null" json:"type"`
	Difficulty string      `gorm:"size:20;not null;default:'easy'" json:"difficulty"`
	Topic      string      `gorm:"size:100" json:"topic"`
	Tags       StringSlice `gorm:"type:jsonb;default:'[]'" json:"tags"`
	Marks      int         `gorm:"not null;default:1" json:"marks"`
	CreatedBy  *uuid.UUID  `gorm:"type:uuid" json:"created_by"`

	MCQ         *MCQQuestion         `gorm:"foreignKey:QuestionID" json:"mcq,omitempty"`
	Programming *ProgrammingQuestion `gorm:"foreignKey:QuestionID" json:"programming,omitempty"`
}

// MCQQuestion holds multiple-choice detail.
type MCQQuestion struct {
	Base
	QuestionID   uuid.UUID   `gorm:"type:uuid;not null;uniqueIndex" json:"question_id"`
	Body         string      `gorm:"not null" json:"body"`
	Options      StringSlice `gorm:"type:jsonb;not null" json:"options"`
	CorrectIndex int         `gorm:"not null" json:"correct_index"`
	Explanation  string      `json:"explanation"`
}

// ProgrammingQuestion holds coding-problem detail and its test cases.
type ProgrammingQuestion struct {
	Base
	QuestionID    uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"question_id"`
	Title         string     `gorm:"size:200;not null" json:"title"`
	Description   string     `json:"description"`
	InputFormat   string     `json:"input_format"`
	OutputFormat  string     `json:"output_format"`
	Constraints   string     `json:"constraints"`
	SampleInput   string     `json:"sample_input"`
	SampleOutput  string     `json:"sample_output"`
	Explanation   string     `json:"explanation"`
	TimeLimitMS   int        `gorm:"default:2000" json:"time_limit_ms"`
	MemoryLimitMB int        `gorm:"default:256" json:"memory_limit_mb"`
	TestCases     []TestCase `gorm:"foreignKey:ProgrammingQuestionID" json:"test_cases,omitempty"`
}

// TestCase is a single input/expected pair; hidden cases are never sent to students.
type TestCase struct {
	Base
	ProgrammingQuestionID uuid.UUID `gorm:"type:uuid;not null;index" json:"programming_question_id"`
	Input                 string    `json:"input"`
	ExpectedOutput        string    `json:"expected_output"`
	IsHidden              bool      `gorm:"default:false" json:"is_hidden"`
	Weight                int       `gorm:"default:1" json:"weight"`
	Ord                   int       `gorm:"default:0" json:"ord"`
}
