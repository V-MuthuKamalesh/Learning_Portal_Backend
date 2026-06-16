package repositories

import (
	"github.com/collegeassess/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GroupRepository handles group persistence with tenant scoping.
type GroupRepository struct{ db *gorm.DB }

func NewGroupRepository(db *gorm.DB) *GroupRepository { return &GroupRepository{db: db} }

func (r *GroupRepository) Create(g *models.Group) error { return r.db.Create(g).Error }
func (r *GroupRepository) Update(g *models.Group) error { return r.db.Save(g).Error }

func (r *GroupRepository) ByID(collegeID, id uuid.UUID) (*models.Group, error) {
	var g models.Group
	err := r.db.Where("college_id = ? AND id = ?", collegeID, id).First(&g).Error
	return wrap(&g, err)
}

func (r *GroupRepository) Delete(collegeID, id uuid.UUID) error {
	return r.db.Where("college_id = ?", collegeID).Delete(&models.Group{}, "id = ?", id).Error
}

func (r *GroupRepository) List(collegeID uuid.UUID) ([]models.Group, error) {
	var groups []models.Group
	err := r.db.Where("college_id = ?", collegeID).Order("created_at DESC").Find(&groups).Error
	if err != nil {
		return nil, err
	}
	for i := range groups {
		var count int64
		_ = r.db.Model(&models.StudentGroup{}).Where("group_id = ?", groups[i].ID).Count(&count).Error
		groups[i].MemberCount = int(count)
	}
	return groups, nil
}

func (r *GroupRepository) CountByCollege(collegeID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.Model(&models.Group{}).Where("college_id = ?", collegeID).Count(&n).Error
	return n, err
}

func (r *GroupRepository) ListMembers(collegeID, groupID uuid.UUID) ([]models.Student, error) {
	var students []models.Student
	err := r.db.Joins("JOIN student_groups ON student_groups.student_id = students.id").
		Where("students.college_id = ? AND student_groups.group_id = ?", collegeID, groupID).
		Order("students.name").
		Find(&students).Error
	return students, err
}

func (r *GroupRepository) AddMembers(groupID uuid.UUID, studentIDs []uuid.UUID) (int, error) {
	added := 0
	for _, sid := range studentIDs {
		sg := models.StudentGroup{StudentID: sid, GroupID: groupID}
		res := r.db.Where("student_id = ? AND group_id = ?", sid, groupID).FirstOrCreate(&sg)
		if res.Error != nil {
			return added, res.Error
		}
		if res.RowsAffected > 0 {
			added++
		}
	}
	return added, nil
}

func (r *GroupRepository) RemoveMember(groupID, studentID uuid.UUID) error {
	return r.db.Where("group_id = ? AND student_id = ?", groupID, studentID).
		Delete(&models.StudentGroup{}).Error
}
