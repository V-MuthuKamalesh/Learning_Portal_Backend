package repositories

import (
	"github.com/collegeassess/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CollegeRepository handles colleges, departments and batches.
type CollegeRepository struct{ db *gorm.DB }

func NewCollegeRepository(db *gorm.DB) *CollegeRepository { return &CollegeRepository{db: db} }

// ── Colleges ──
func (r *CollegeRepository) Create(c *models.College) error { return r.db.Create(c).Error }
func (r *CollegeRepository) Update(c *models.College) error { return r.db.Save(c).Error }

func (r *CollegeRepository) ByID(id uuid.UUID) (*models.College, error) {
	var c models.College
	return wrap(&c, r.db.First(&c, "id = ?", id).Error)
}

func (r *CollegeRepository) ByCode(code string) (*models.College, error) {
	var c models.College
	return wrap(&c, r.db.Where("code = ?", code).First(&c).Error)
}

func (r *CollegeRepository) List() ([]models.College, error) {
	var cs []models.College
	return cs, r.db.Order("name asc").Find(&cs).Error
}

func (r *CollegeRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.College{}, "id = ?", id).Error
}

// ── Departments ──
func (r *CollegeRepository) CreateDepartment(d *models.Department) error { return r.db.Create(d).Error }
func (r *CollegeRepository) UpdateDepartment(d *models.Department) error { return r.db.Save(d).Error }

func (r *CollegeRepository) DepartmentByID(collegeID, id uuid.UUID) (*models.Department, error) {
	var d models.Department
	return wrap(&d, r.db.Where("college_id = ? AND id = ?", collegeID, id).First(&d).Error)
}

func (r *CollegeRepository) ListDepartments(collegeID uuid.UUID) ([]models.Department, error) {
	var ds []models.Department
	return ds, r.db.Where("college_id = ?", collegeID).Order("name asc").Find(&ds).Error
}

func (r *CollegeRepository) DeleteDepartment(collegeID, id uuid.UUID) error {
	return r.db.Where("college_id = ?", collegeID).Delete(&models.Department{}, "id = ?", id).Error
}

// ── Batches ──
func (r *CollegeRepository) CreateBatch(b *models.Batch) error { return r.db.Create(b).Error }
func (r *CollegeRepository) UpdateBatch(b *models.Batch) error { return r.db.Save(b).Error }

func (r *CollegeRepository) BatchByID(collegeID, id uuid.UUID) (*models.Batch, error) {
	var b models.Batch
	return wrap(&b, r.db.Where("college_id = ? AND id = ?", collegeID, id).First(&b).Error)
}

func (r *CollegeRepository) ListBatches(collegeID uuid.UUID) ([]models.Batch, error) {
	var bs []models.Batch
	return bs, r.db.Where("college_id = ?", collegeID).Order("start_year desc").Find(&bs).Error
}

func (r *CollegeRepository) DeleteBatch(collegeID, id uuid.UUID) error {
	return r.db.Where("college_id = ?", collegeID).Delete(&models.Batch{}, "id = ?", id).Error
}
