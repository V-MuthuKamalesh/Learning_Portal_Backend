package repositories

import (
	"time"

	"github.com/collegeassess/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AttemptRepository handles submissions and answers.
type AttemptRepository struct{ db *gorm.DB }

func NewAttemptRepository(db *gorm.DB) *AttemptRepository { return &AttemptRepository{db: db} }

func (r *AttemptRepository) FindSubmission(assessmentID, studentID uuid.UUID) (*models.Submission, error) {
	var s models.Submission
	err := r.db.Preload("Answers").Preload("CodingSubmissions").
		Where("assessment_id = ? AND student_id = ?", assessmentID, studentID).
		First(&s).Error
	return wrap(&s, err)
}

func (r *AttemptRepository) SubmissionByID(id uuid.UUID) (*models.Submission, error) {
	var s models.Submission
	err := r.db.Preload("Answers").Preload("CodingSubmissions").
		First(&s, "id = ?", id).Error
	return wrap(&s, err)
}

func (r *AttemptRepository) CreateSubmission(s *models.Submission) error {
	return r.db.Create(s).Error
}

func (r *AttemptRepository) UpdateSubmission(s *models.Submission) error {
	return r.db.Save(s).Error
}

func (r *AttemptRepository) UpsertAnswer(a *models.Answer) error {
	var existing models.Answer
	err := r.db.Where("submission_id = ? AND assessment_question_id = ?", a.SubmissionID, a.AssessmentQuestionID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(a).Error
	}
	if err != nil {
		return err
	}
	existing.SelectedIndex = a.SelectedIndex
	if a.IsCorrect != nil {
		existing.IsCorrect = a.IsCorrect
	}
	existing.MarksAwarded = a.MarksAwarded
	return r.db.Save(&existing).Error
}

// FindCodingSubmission returns the current submission record for a question attempt, or ErrNotFound.
func (r *AttemptRepository) FindCodingSubmission(submissionID, assessmentQuestionID uuid.UUID) (*models.CodingSubmission, error) {
	var cs models.CodingSubmission
	err := r.db.Where("submission_id = ? AND assessment_question_id = ?", submissionID, assessmentQuestionID).
		First(&cs).Error
	return wrap(&cs, err)
}

func (r *AttemptRepository) UpsertCodingSubmission(cs *models.CodingSubmission) error {
	var existing models.CodingSubmission
	q := r.db.Where("student_id = ? AND question_id = ?", cs.StudentID, cs.QuestionID)
	if cs.SubmissionID != nil {
		q = q.Where("submission_id = ?", *cs.SubmissionID)
	}
	if cs.AssessmentQuestionID != nil {
		q = q.Where("assessment_question_id = ?", *cs.AssessmentQuestionID)
	}
	err := q.First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		if cs.AttemptCount <= 0 {
			cs.AttemptCount = 1
		}
		return r.db.Create(cs).Error
	}
	if err != nil {
		return err
	}
	existing.Language = cs.Language
	existing.SourceCode = cs.SourceCode
	existing.Status = cs.Status
	existing.PassedCases = cs.PassedCases
	existing.TotalCases = cs.TotalCases
	existing.MarksAwarded = cs.MarksAwarded
	existing.RuntimeMS = cs.RuntimeMS
	existing.MemoryKB = cs.MemoryKB
	existing.Verdict = cs.Verdict
	existing.AttemptCount = cs.AttemptCount
	existing.FailedAttempts = cs.FailedAttempts
	return r.db.Save(&existing).Error
}

func (r *AttemptRepository) CreateResult(res *models.AssessmentResult) error {
	return r.db.Where(
		"assessment_id = ? AND student_id = ?",
		res.AssessmentID, res.StudentID,
	).Assign(*res).FirstOrCreate(res).Error
}

func (r *AttemptRepository) ListResults(collegeID uuid.UUID, assessmentID *uuid.UUID) ([]models.AssessmentResult, error) {
	q := r.db.Preload("Student").Joins(
		"JOIN assessments ON assessments.id = assessment_results.assessment_id",
	).Where("assessments.college_id = ?", collegeID)
	if assessmentID != nil {
		q = q.Where("assessment_results.assessment_id = ?", *assessmentID)
	}
	var rows []models.AssessmentResult
	err := q.Order("assessment_results.percentage DESC").Find(&rows).Error
	return rows, err
}

func (r *AttemptRepository) StudentResults(studentID uuid.UUID) ([]models.AssessmentResult, error) {
	var rows []models.AssessmentResult
	err := r.db.Preload("Student").
		Where("student_id = ? AND published = ?", studentID, true).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *AttemptRepository) Leaderboard(collegeID uuid.UUID, limit int) ([]models.AssessmentResult, error) {
	if limit <= 0 {
		limit = 20
	}
	var rows []models.AssessmentResult
	err := r.db.Preload("Student").
		Joins("JOIN assessments ON assessments.id = assessment_results.assessment_id").
		Where("assessments.college_id = ? AND assessment_results.published = ?", collegeID, true).
		Order("assessment_results.percentage DESC, assessment_results.marks_scored DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *AttemptRepository) RecalculateRanks(assessmentID uuid.UUID) error {
	var rows []models.AssessmentResult
	if err := r.db.Where("assessment_id = ?", assessmentID).
		Order("marks_scored DESC").Find(&rows).Error; err != nil {
		return err
	}
	for i := range rows {
		rows[i].Rank = i + 1
		if err := r.db.Model(&rows[i]).Update("rank", rows[i].Rank).Error; err != nil {
			return err
		}
	}
	return nil
}

// NotificationRepository stores in-app notifications.
type NotificationRepository struct{ db *gorm.DB }

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(n *models.Notification) error { return r.db.Create(n).Error }

func (r *NotificationRepository) ListForUser(userID uuid.UUID, userType string) ([]models.Notification, error) {
	var rows []models.Notification
	err := r.db.Where("user_id = ? AND user_type = ?", userID, userType).
		Order("created_at DESC").Limit(50).Find(&rows).Error
	return rows, err
}

func (r *NotificationRepository) MarkRead(id, userID uuid.UUID) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

// PracticeRepository handles practice modules.
type PracticeRepository struct{ db *gorm.DB }

func NewPracticeRepository(db *gorm.DB) *PracticeRepository { return &PracticeRepository{db: db} }

func (r *PracticeRepository) ListModules(collegeID uuid.UUID) ([]models.PracticeModule, error) {
	var mods []models.PracticeModule
	err := r.db.Where("college_id = ?", collegeID).Order("ord, name").Find(&mods).Error
	return mods, err
}

func (r *PracticeRepository) ModuleByID(collegeID, id uuid.UUID) (*models.PracticeModule, error) {
	var m models.PracticeModule
	err := r.db.Preload("Questions").Preload("Questions.MCQ").Preload("Questions.Programming").
		Where("college_id = ? AND id = ?", collegeID, id).First(&m).Error
	return wrap(&m, err)
}

func (r *PracticeRepository) UpsertProgress(p *models.StudentProgress) error {
	return r.db.Where("student_id = ? AND module_id = ?", p.StudentID, p.ModuleID).
		Assign(models.StudentProgress{
			Completed:  p.Completed,
			Total:      p.Total,
			Percentage: p.Percentage,
			UpdatedAt:  time.Now(),
		}).FirstOrCreate(p).Error
}

func (r *PracticeRepository) ProgressForStudent(studentID uuid.UUID) ([]models.StudentProgress, error) {
	var rows []models.StudentProgress
	err := r.db.Where("student_id = ?", studentID).Find(&rows).Error
	return rows, err
}

func (r *PracticeRepository) CreateModule(m *models.PracticeModule) error { return r.db.Create(m).Error }

func (r *PracticeRepository) UpdateModule(m *models.PracticeModule) error {
	return r.db.Omit("Questions").Save(m).Error
}

func (r *PracticeRepository) DeleteModule(collegeID, id uuid.UUID) error {
	return r.db.Where("college_id = ? AND id = ?", collegeID, id).Delete(&models.PracticeModule{}).Error
}

func (r *PracticeRepository) LinkQuestion(moduleID, questionID uuid.UUID, ord, marks, maxAttempts int) error {
	link := models.ModuleQuestion{ModuleID: moduleID, QuestionID: questionID, Ord: ord, Marks: marks, MaxAttempts: maxAttempts}
	return r.db.Where("module_id = ? AND question_id = ?", moduleID, questionID).
		Assign(models.ModuleQuestion{Ord: ord, Marks: marks, MaxAttempts: maxAttempts}).FirstOrCreate(&link).Error
}

func (r *PracticeRepository) UnlinkQuestion(moduleID, questionID uuid.UUID) error {
	return r.db.Where("module_id = ? AND question_id = ?", moduleID, questionID).
		Delete(&models.ModuleQuestion{}).Error
}

func (r *PracticeRepository) UpdateModuleQuestion(mq *models.ModuleQuestion) error {
	return r.db.Model(mq).Updates(map[string]any{
		"marks":        mq.Marks,
		"max_attempts": mq.MaxAttempts,
		"ord":          mq.Ord,
	}).Error
}

func (r *PracticeRepository) ModuleQuestionSlot(moduleID, questionID uuid.UUID) (*models.ModuleQuestion, error) {
	var mq models.ModuleQuestion
	err := r.db.Where("module_id = ? AND question_id = ?", moduleID, questionID).First(&mq).Error
	return wrap(&mq, err)
}

// ListModuleQuestions returns ordered question slots with full question data.
func (r *PracticeRepository) ListModuleQuestions(collegeID, moduleID uuid.UUID) ([]models.ModuleQuestion, []models.Question, error) {
	var slots []models.ModuleQuestion
	if err := r.db.Where("module_id = ?", moduleID).Order("ord, id").Find(&slots).Error; err != nil {
		return nil, nil, err
	}
	if len(slots) == 0 {
		return slots, nil, nil
	}
	qids := make([]uuid.UUID, len(slots))
	for i, s := range slots {
		qids[i] = s.QuestionID
	}
	var questions []models.Question
	err := r.db.Preload("MCQ").Preload("Programming").
		Where("college_id = ? AND id IN ?", collegeID, qids).Find(&questions).Error
	return slots, questions, err
}
