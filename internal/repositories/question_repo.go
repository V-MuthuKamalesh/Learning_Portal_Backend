package repositories

import (
	"github.com/collegeassess/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QuestionRepository handles question bank persistence.
type QuestionRepository struct{ db *gorm.DB }

func NewQuestionRepository(db *gorm.DB) *QuestionRepository { return &QuestionRepository{db: db} }

func (r *QuestionRepository) Create(q *models.Question, mcq *models.MCQQuestion) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(q).Error; err != nil {
			return err
		}
		mcq.QuestionID = q.ID
		return tx.Create(mcq).Error
	})
}

func (r *QuestionRepository) CreateProgramming(q *models.Question, prog *models.ProgrammingQuestion, cases []models.TestCase) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(q).Error; err != nil {
			return err
		}
		prog.QuestionID = q.ID
		if err := tx.Create(prog).Error; err != nil {
			return err
		}
		for i := range cases {
			cases[i].ProgrammingQuestionID = prog.ID
			cases[i].Ord = i
			if cases[i].Weight <= 0 {
				cases[i].Weight = 1
			}
		}
		if len(cases) > 0 {
			if err := tx.Create(&cases).Error; err != nil {
				return err
			}
		}
		prog.TestCases = cases
		q.Programming = prog
		return nil
	})
}

func (r *QuestionRepository) List(collegeID uuid.UUID) ([]models.Question, error) {
	var questions []models.Question
	err := r.db.Preload("MCQ").
		Preload("Programming.TestCases", func(db *gorm.DB) *gorm.DB {
			return db.Order("ord ASC")
		}).
		Where("college_id = ?", collegeID).
		Order("created_at DESC").
		Find(&questions).Error
	return questions, err
}

func (r *QuestionRepository) ByID(collegeID, id uuid.UUID) (*models.Question, error) {
	var q models.Question
	err := r.db.Preload("MCQ").
		Preload("Programming.TestCases", func(db *gorm.DB) *gorm.DB {
			return db.Order("ord ASC")
		}).
		Where("college_id = ? AND id = ?", collegeID, id).
		First(&q).Error
	return wrap(&q, err)
}

func (r *QuestionRepository) CountByCollege(collegeID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.Model(&models.Question{}).Where("college_id = ?", collegeID).Count(&n).Error
	return n, err
}
