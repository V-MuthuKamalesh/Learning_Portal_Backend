package database

import (
	"time"

	"github.com/collegeassess/backend/configs"
	"github.com/collegeassess/backend/internal/models"
	"github.com/collegeassess/backend/pkg/hash"
	"github.com/collegeassess/backend/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// permissionCatalog enumerates every resource:action in the system.
var permissionCatalog = []models.Permission{
	{Resource: "*", Action: "*", Slug: "*", Label: "Super wildcard"},
	{Resource: "student", Action: "create", Slug: "student:create", Label: "Create students"},
	{Resource: "student", Action: "read", Slug: "student:read", Label: "View students"},
	{Resource: "student", Action: "update", Slug: "student:update", Label: "Edit students"},
	{Resource: "student", Action: "delete", Slug: "student:delete", Label: "Delete students"},
	{Resource: "assessment", Action: "create", Slug: "assessment:create", Label: "Create assessments"},
	{Resource: "assessment", Action: "read", Slug: "assessment:read", Label: "View assessments"},
	{Resource: "assessment", Action: "update", Slug: "assessment:update", Label: "Edit assessments"},
	{Resource: "assessment", Action: "publish", Slug: "assessment:publish", Label: "Publish assessments"},
	{Resource: "assessment", Action: "delete", Slug: "assessment:delete", Label: "Delete assessments"},
	{Resource: "result", Action: "read", Slug: "result:read", Label: "View results"},
	{Resource: "result", Action: "export", Slug: "result:export", Label: "Export results"},
	{Resource: "analytics", Action: "read", Slug: "analytics:read", Label: "View analytics"},
	{Resource: "admin", Action: "create", Slug: "admin:create", Label: "Create admins"},
	{Resource: "admin", Action: "read", Slug: "admin:read", Label: "View admins"},
	{Resource: "admin", Action: "update", Slug: "admin:update", Label: "Edit admins"},
	{Resource: "admin", Action: "delete", Slug: "admin:delete", Label: "Delete admins"},
	{Resource: "role", Action: "read", Slug: "role:read", Label: "View roles"},
	{Resource: "role", Action: "manage", Slug: "role:manage", Label: "Manage roles"},
	{Resource: "question", Action: "create", Slug: "question:create", Label: "Create questions"},
	{Resource: "question", Action: "read", Slug: "question:read", Label: "View questions"},
	{Resource: "question", Action: "update", Slug: "question:update", Label: "Edit questions"},
	{Resource: "question", Action: "delete", Slug: "question:delete", Label: "Delete questions"},
	{Resource: "group", Action: "create", Slug: "group:create", Label: "Create groups"},
	{Resource: "group", Action: "read", Slug: "group:read", Label: "View groups"},
	{Resource: "group", Action: "update", Slug: "group:update", Label: "Edit groups"},
	{Resource: "group", Action: "delete", Slug: "group:delete", Label: "Delete groups"},
	{Resource: "college", Action: "read", Slug: "college:read", Label: "View college"},
	{Resource: "college", Action: "manage", Slug: "college:manage", Label: "Manage college"},
}

// roleTemplates maps a system role slug to the permission slugs it grants.
var roleTemplates = map[string]struct {
	name  string
	perms []string
}{
	"super-admin":      {"Super Admin", []string{"*"}},
	"department-admin": {"Department Admin", []string{"student:create", "student:read", "student:update", "student:delete", "result:read", "analytics:read", "group:create", "group:read", "group:update", "group:delete"}},
	"faculty-admin":    {"Faculty Admin", []string{"assessment:create", "assessment:read", "assessment:update", "assessment:publish", "assessment:delete", "question:create", "question:read", "question:update", "question:delete", "result:read", "result:export", "analytics:read", "student:read"}},
	"viewer-admin":     {"Viewer Admin", []string{"student:read", "assessment:read", "result:read", "analytics:read", "question:read", "group:read", "admin:read", "role:read", "college:read"}},
}

// Seed creates permissions, system roles, a demo college and a super admin (idempotent).
func Seed(db *gorm.DB, cfg *configs.Config) error {
	// 1. Permissions (upsert by slug).
	if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "slug"}}, DoNothing: true}).
		Create(&permissionCatalog).Error; err != nil {
		return err
	}

	var perms []models.Permission
	if err := db.Find(&perms).Error; err != nil {
		return err
	}
	bySlug := map[string]models.Permission{}
	for _, p := range perms {
		bySlug[p.Slug] = p
	}

	// 2. System roles (templates, college_id NULL).
	for slug, tpl := range roleTemplates {
		var role models.Role
		err := db.Where("slug = ? AND is_system = ?", slug, true).First(&role).Error
		if err == gorm.ErrRecordNotFound {
			role = models.Role{Name: tpl.name, Slug: slug, IsSystem: true}
			if err := db.Create(&role).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		var rolePerms []models.Permission
		for _, ps := range tpl.perms {
			if p, ok := bySlug[ps]; ok {
				rolePerms = append(rolePerms, p)
			}
		}
		if err := db.Model(&role).Association("Permissions").Replace(rolePerms); err != nil {
			return err
		}
	}

	// 3. Demo college.
	var college models.College
	err := db.Where("code = ?", cfg.Seed.CollegeCode).First(&college).Error
	if err == gorm.ErrRecordNotFound {
		college = models.College{
			Name:         cfg.Seed.CollegeName,
			Code:         cfg.Seed.CollegeCode,
			ContactEmail: cfg.Seed.SuperAdminEmail,
			IsActive:     true,
		}
		if err := db.Create(&college).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// 4. Super admin.
	var superRole models.Role
	if err := db.Where("slug = ? AND is_system = ?", "super-admin", true).First(&superRole).Error; err != nil {
		return err
	}
	var count int64
	db.Model(&models.Admin{}).Where("email = ?", cfg.Seed.SuperAdminEmail).Count(&count)
	if count == 0 {
		pw, err := hash.Password(cfg.Seed.SuperAdminPassword)
		if err != nil {
			return err
		}
		admin := models.Admin{
			CollegeID:     college.ID,
			RoleID:        superRole.ID,
			Name:          "Super Admin",
			Email:         cfg.Seed.SuperAdminEmail,
			PasswordHash:  pw,
			IsActive:      true,
			EmailVerified: true,
		}
		if err := db.Create(&admin).Error; err != nil {
			return err
		}
		logger.Info("seeded super admin", "email", cfg.Seed.SuperAdminEmail)
	}

	// 5. Demo department, batch and student for portal login testing.
	var dept models.Department
	err = db.Where("college_id = ? AND code = ?", college.ID, "CSE").First(&dept).Error
	if err == gorm.ErrRecordNotFound {
		dept = models.Department{CollegeID: college.ID, Name: "Computer Science", Code: "CSE"}
		if err := db.Create(&dept).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	var batch models.Batch
	err = db.Where("college_id = ? AND name = ?", college.ID, "2022-2026").First(&batch).Error
	if err == gorm.ErrRecordNotFound {
		batch = models.Batch{CollegeID: college.ID, Name: "2022-2026", StartYear: 2022, EndYear: 2026}
		if err := db.Create(&batch).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	db.Model(&models.Student{}).Where("college_id = ? AND register_number = ?", college.ID, "CSE22001").Count(&count)
	if count == 0 {
		studentPW, err := hash.Password("CSE22001")
		if err != nil {
			return err
		}
		student := models.Student{
			CollegeID:      college.ID,
			DepartmentID:   &dept.ID,
			BatchID:        &batch.ID,
			Name:           "Demo Student",
			RegisterNumber: "CSE22001",
			Email:          "student@demo.edu",
			PasswordHash:   studentPW,
			Year:           "3",
			Section:        "A",
			IsActive:       true,
			EmailVerified:  true,
		}
		if err := db.Create(&student).Error; err != nil {
			return err
		}
		logger.Info("seeded demo student", "register", "CSE22001", "email", "student@demo.edu")
	}

	if err := seedDemoContent(db, college.ID); err != nil {
		return err
	}

	return nil
}

func seedDemoContent(db *gorm.DB, collegeID uuid.UUID) error {
	var assessCount int64
	db.Model(&models.Assessment{}).Where("college_id = ? AND title = ?", collegeID, "Aptitude Round").Count(&assessCount)
	if assessCount > 0 {
		return nil
	}

	questions := []struct {
		topic string
		body  string
		opts  []string
		ans   int
	}{
		{"Sorting", "What is the average time complexity of merge sort?", []string{"O(n)", "O(n log n)", "O(n^2)", "O(1)"}, 1},
		{"Arrays", "Which index is valid for a zero-based array of length 5?", []string{"-1", "0", "5", "6"}, 1},
		{"SQL", "Which clause filters rows before grouping?", []string{"WHERE", "HAVING", "ORDER BY", "LIMIT"}, 0},
	}

	var qIDs []uuid.UUID
	var qCount int64
	db.Model(&models.Question{}).Where("college_id = ?", collegeID).Count(&qCount)
	if qCount == 0 {
		for _, item := range questions {
			q := models.Question{
				CollegeID:  collegeID,
				Type:       models.QuestionMCQ,
				Difficulty: models.DifficultyEasy,
				Topic:      item.topic,
				Marks:      10,
			}
			if err := db.Create(&q).Error; err != nil {
				return err
			}
			mcq := models.MCQQuestion{
				QuestionID:   q.ID,
				Body:         item.body,
				Options:      models.StringSlice(item.opts),
				CorrectIndex: item.ans,
			}
			if err := db.Create(&mcq).Error; err != nil {
				return err
			}
			qIDs = append(qIDs, q.ID)
		}

		mod := models.PracticeModule{
			CollegeID:   collegeID,
			Name:        "Arrays",
			Category:    "mcq",
			Description: "Practice array fundamentals",
			Ord:         1,
		}
		if err := db.Create(&mod).Error; err != nil {
			return err
		}
		for i, qid := range qIDs {
			link := models.ModuleQuestion{ModuleID: mod.ID, QuestionID: qid, Ord: i}
			if err := db.Create(&link).Error; err != nil {
				return err
			}
		}
	} else {
		db.Model(&models.Question{}).Where("college_id = ?", collegeID).Pluck("id", &qIDs)
	}

	if len(qIDs) == 0 {
		return nil
	}

	now := time.Now()
	end := now.Add(30 * 24 * time.Hour)
	assess := models.Assessment{
		CollegeID:       collegeID,
		Title:           "Aptitude Round",
		Description:     "Demo MCQ assessment for portal testing",
		Type:            models.AssessmentMCQ,
		DurationMinutes: 30,
		TotalMarks:      30,
		PassingMarks:    15,
		Status:          models.StatusPublished,
		StartTime:       &now,
		EndTime:         &end,
		MCQCount:        len(qIDs),
	}
	if err := db.Create(&assess).Error; err != nil {
		return err
	}
	for i, qid := range qIDs {
		aq := models.AssessmentQuestion{AssessmentID: assess.ID, QuestionID: qid, Ord: i}
		if err := db.Create(&aq).Error; err != nil {
			return err
		}
	}
	assign := models.AssessmentAssignment{
		AssessmentID: assess.ID,
		TargetType:   models.TargetCollege,
	}
	if err := db.Create(&assign).Error; err != nil {
		return err
	}

	logger.Info("seeded demo questions, practice module and published assessment")
	return nil
}
