package repositories

import (
	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/models"
	"github.com/collegeassess/backend/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StudentRepository handles student persistence with tenant-scoped queries.
type StudentRepository struct{ db *gorm.DB }

func NewStudentRepository(db *gorm.DB) *StudentRepository { return &StudentRepository{db: db} }

func (r *StudentRepository) Create(s *models.Student) error { return r.db.Create(s).Error }
func (r *StudentRepository) Update(s *models.Student) error { return r.db.Save(s).Error }

func (r *StudentRepository) ByID(collegeID, id uuid.UUID) (*models.Student, error) {
	var s models.Student
	err := r.db.Preload("Department").Preload("Batch").Preload("Groups").
		Where("college_id = ? AND id = ?", collegeID, id).First(&s).Error
	return wrap(&s, err)
}

// FindByID loads a student by primary key (for student-scoped services).
func (r *StudentRepository) FindByID(id uuid.UUID) (*models.Student, error) {
	var s models.Student
	err := r.db.Preload("Department").Preload("Batch").Preload("Groups").
		Where("id = ?", id).First(&s).Error
	return wrap(&s, err)
}

func (r *StudentRepository) Delete(collegeID, id uuid.UUID) error {
	return r.db.Where("college_id = ?", collegeID).Delete(&models.Student{}, "id = ?", id).Error
}

// scopedQuery builds a tenant + department-scoped base query.
func (r *StudentRepository) scopedQuery(collegeID uuid.UUID, deptScope []uuid.UUID) *gorm.DB {
	q := r.db.Model(&models.Student{}).Where("students.college_id = ?", collegeID)
	if len(deptScope) > 0 {
		q = q.Where("students.department_id IN ?", deptScope)
	}
	return q
}

// List returns a filtered, paginated page of students plus the total count.
// deptScope restricts results for department admins (nil/empty = all departments).
func (r *StudentRepository) List(collegeID uuid.UUID, deptScope []uuid.UUID, f dto.StudentFilter, p pagination.Params) ([]models.Student, int64, error) {
	q := r.scopedQuery(collegeID, deptScope)

	if f.DepartmentID != "" {
		q = q.Where("students.department_id = ?", f.DepartmentID)
	}
	if f.BatchID != "" {
		q = q.Where("students.batch_id = ?", f.BatchID)
	}
	if f.Year != "" {
		q = q.Where("students.year = ?", f.Year)
	}
	if f.Section != "" {
		q = q.Where("students.section = ?", f.Section)
	}
	switch f.Status {
	case "active":
		q = q.Where("students.is_active = ?", true)
	case "inactive":
		q = q.Where("students.is_active = ?", false)
	}
	if f.GroupID != "" {
		q = q.Joins("JOIN student_groups sg ON sg.student_id = students.id").
			Where("sg.group_id = ?", f.GroupID)
	}
	if p.Search != "" {
		like := "%" + p.Search + "%"
		q = q.Where("students.name ILIKE ? OR students.register_number ILIKE ? OR students.email ILIKE ?", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	allowed := map[string]bool{"name": true, "register_number": true, "created_at": true, "year": true}
	var students []models.Student
	err := q.Preload("Department").Preload("Batch").
		Order("students." + p.OrderClause(allowed, "created_at")).
		Limit(p.PageSize).Offset(p.Offset()).Find(&students).Error
	return students, total, err
}

// ExistsByRegister reports whether a register number is already used in the college.
func (r *StudentRepository) ExistsByRegister(collegeID uuid.UUID, register string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Student{}).
		Where("college_id = ? AND register_number = ?", collegeID, register).Count(&count).Error
	return count > 0, err
}

// CreateBatch inserts many students in a single transaction (bulk import).
func (r *StudentRepository) CreateBatch(students []models.Student) error {
	if len(students) == 0 {
		return nil
	}
	return r.db.CreateInBatches(students, 100).Error
}

// CountByCollege returns the active student total for dashboards.
func (r *StudentRepository) CountByCollege(collegeID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.Model(&models.Student{}).Where("college_id = ?", collegeID).Count(&n).Error
	return n, err
}
