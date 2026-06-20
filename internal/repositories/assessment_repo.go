package repositories

import (
	"time"

	"github.com/collegeassess/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssessmentRepository handles assessment persistence.
type AssessmentRepository struct{ db *gorm.DB }

func NewAssessmentRepository(db *gorm.DB) *AssessmentRepository { return &AssessmentRepository{db: db} }

func (r *AssessmentRepository) Create(a *models.Assessment) error { return r.db.Create(a).Error }

func (r *AssessmentRepository) Update(a *models.Assessment) error { return r.db.Save(a).Error }

func (r *AssessmentRepository) ByID(collegeID, id uuid.UUID) (*models.Assessment, error) {
	var a models.Assessment
	err := r.db.Preload("Questions.Question.MCQ").
		Preload("Questions.Question.Programming.TestCases", func(db *gorm.DB) *gorm.DB {
			return db.Order("ord ASC")
		}).
		Preload("Assignments").
		Where("college_id = ? AND id = ?", collegeID, id).
		First(&a).Error
	return wrap(&a, err)
}

func (r *AssessmentRepository) Delete(collegeID, id uuid.UUID) error {
	return r.db.Where("college_id = ?", collegeID).Delete(&models.Assessment{}, "id = ?", id).Error
}

func (r *AssessmentRepository) List(collegeID uuid.UUID) ([]models.Assessment, error) {
	var items []models.Assessment
	err := r.db.Where("college_id = ?", collegeID).Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *AssessmentRepository) AttachQuestions(assessmentID uuid.UUID, questionIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("assessment_id = ?", assessmentID).Delete(&models.AssessmentQuestion{}).Error; err != nil {
			return err
		}
		mcqCount, codingCount := 0, 0
		for i, qid := range questionIDs {
			var q models.Question
			if err := tx.First(&q, "id = ?", qid).Error; err != nil {
				return err
			}
			if q.Type == models.QuestionProgramming {
				codingCount++
			} else {
				mcqCount++
			}
			aq := models.AssessmentQuestion{
				AssessmentID: assessmentID,
				QuestionID:   qid,
				Ord:          i,
			}
			if err := tx.Create(&aq).Error; err != nil {
				return err
			}
		}
		return tx.Model(&models.Assessment{}).Where("id = ?", assessmentID).Updates(map[string]any{
			"mcq_count":    mcqCount,
			"coding_count": codingCount,
		}).Error
	})
}

func (r *AssessmentRepository) Assign(assessmentID uuid.UUID, targetType string, targetID *uuid.UUID) error {
	a := models.AssessmentAssignment{
		AssessmentID: assessmentID,
		TargetType:   targetType,
		TargetID:     targetID,
	}
	return r.db.Create(&a).Error
}

func (r *AssessmentRepository) UpdateAssessmentQuestion(assessmentID, questionID uuid.UUID, updates map[string]any) error {
	return r.db.Model(&models.AssessmentQuestion{}).
		Where("assessment_id = ? AND question_id = ?", assessmentID, questionID).
		Updates(updates).Error
}

func (r *AssessmentRepository) Publish(id uuid.UUID, start, end *time.Time) error {
	updates := map[string]any{"status": models.StatusPublished}
	if start != nil {
		updates["start_time"] = start
	}
	if end != nil {
		updates["end_time"] = end
	}
	return r.db.Model(&models.Assessment{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AssessmentRepository) CountByCollege(collegeID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.Model(&models.Assessment{}).Where("college_id = ?", collegeID).Count(&n).Error
	return n, err
}

func (r *AssessmentRepository) CountPublished(collegeID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.Model(&models.Assessment{}).
		Where("college_id = ? AND status = ?", collegeID, models.StatusPublished).
		Count(&n).Error
	return n, err
}

// ListForStudent returns published assessments assigned to the student.
func (r *AssessmentRepository) ListForStudent(collegeID, studentID uuid.UUID, deptID, batchID *uuid.UUID, groupIDs []uuid.UUID) ([]models.Assessment, error) {
	subq := r.db.Model(&models.AssessmentAssignment{}).Select("assessment_id")
	var collegeTargets []uuid.UUID
	subq.Where("target_type = ?", models.TargetCollege).Pluck("assessment_id", &collegeTargets)

	studentTargets := []uuid.UUID{}
	r.db.Model(&models.AssessmentAssignment{}).
		Where("target_type = ? AND target_id = ?", models.TargetStudent, studentID).
		Pluck("assessment_id", &studentTargets)

	deptTargets := []uuid.UUID{}
	if deptID != nil {
		r.db.Model(&models.AssessmentAssignment{}).
			Where("target_type = ? AND target_id = ?", models.TargetDepartment, *deptID).
			Pluck("assessment_id", &deptTargets)
	}

	batchTargets := []uuid.UUID{}
	if batchID != nil {
		r.db.Model(&models.AssessmentAssignment{}).
			Where("target_type = ? AND target_id = ?", models.TargetBatch, *batchID).
			Pluck("assessment_id", &batchTargets)
	}

	groupTargets := []uuid.UUID{}
	if len(groupIDs) > 0 {
		r.db.Model(&models.AssessmentAssignment{}).
			Where("target_type = ? AND target_id IN ?", models.TargetGroup, groupIDs).
			Pluck("assessment_id", &groupTargets)
	}

	allIDs := append([]uuid.UUID{}, collegeTargets...)
	allIDs = append(allIDs, studentTargets...)
	allIDs = append(allIDs, deptTargets...)
	allIDs = append(allIDs, batchTargets...)
	allIDs = append(allIDs, groupTargets...)

	if len(allIDs) == 0 {
		// Fallback: all published assessments in college when nothing explicitly assigned.
		var items []models.Assessment
		err := r.db.Where("college_id = ? AND status = ?", collegeID, models.StatusPublished).
			Order("created_at DESC").Find(&items).Error
		return items, err
	}

	seen := map[uuid.UUID]bool{}
	unique := []uuid.UUID{}
	for _, id := range allIDs {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}

	var items []models.Assessment
	err := r.db.Where("college_id = ? AND status = ? AND id IN ?", collegeID, models.StatusPublished, unique).
		Order("created_at DESC").Find(&items).Error
	return items, err
}
