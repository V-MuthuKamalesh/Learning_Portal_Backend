package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/models"
	"github.com/collegeassess/backend/internal/repositories"
	"github.com/collegeassess/backend/pkg/hash"
	"github.com/collegeassess/backend/pkg/pagination"
	"github.com/google/uuid"
)

// StudentService implements student management and bulk CSV import.
type StudentService struct {
	repo        *repositories.StudentRepository
	collegeRepo *repositories.CollegeRepository
}

func NewStudentService(repo *repositories.StudentRepository, collegeRepo *repositories.CollegeRepository) *StudentService {
	return &StudentService{repo: repo, collegeRepo: collegeRepo}
}

func (s *StudentService) List(collegeID uuid.UUID, deptScope []uuid.UUID, f dto.StudentFilter, p pagination.Params) ([]models.Student, int64, error) {
	return s.repo.List(collegeID, deptScope, f, p)
}

func (s *StudentService) Get(collegeID, id uuid.UUID) (*models.Student, error) {
	return s.repo.ByID(collegeID, id)
}

func (s *StudentService) Create(collegeID uuid.UUID, req dto.CreateStudentRequest) (*models.Student, error) {
	exists, err := s.repo.ExistsByRegister(collegeID, req.RegisterNumber)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("register number already exists")
	}
	pw := req.Password
	if pw == "" {
		pw = req.RegisterNumber // default initial password
	}
	hashed, err := hash.Password(pw)
	if err != nil {
		return nil, err
	}
	st := &models.Student{
		CollegeID:      collegeID,
		Name:           req.Name,
		RegisterNumber: req.RegisterNumber,
		Email:          req.Email,
		PasswordHash:   hashed,
		Year:           req.Year,
		Section:        req.Section,
		Phone:          req.Phone,
		IsActive:       true,
		DepartmentID:   parseUUIDPtr(req.DepartmentID),
		BatchID:        parseUUIDPtr(req.BatchID),
	}
	if err := s.repo.Create(st); err != nil {
		return nil, err
	}
	return st, nil
}

func (s *StudentService) Update(collegeID, id uuid.UUID, req dto.UpdateStudentRequest) (*models.Student, error) {
	st, err := s.repo.ByID(collegeID, id)
	if err != nil {
		return nil, err
	}
	applyStr(&st.Name, req.Name)
	applyStr(&st.Email, req.Email)
	applyStr(&st.Year, req.Year)
	applyStr(&st.Section, req.Section)
	applyStr(&st.Phone, req.Phone)
	if req.DepartmentID != nil {
		st.DepartmentID = parseUUIDPtr(req.DepartmentID)
	}
	if req.BatchID != nil {
		st.BatchID = parseUUIDPtr(req.BatchID)
	}
	return st, s.repo.Update(st)
}

func (s *StudentService) SetStatus(collegeID, id uuid.UUID, active bool) error {
	st, err := s.repo.ByID(collegeID, id)
	if err != nil {
		return err
	}
	st.IsActive = active
	return s.repo.Update(st)
}

func (s *StudentService) Delete(collegeID, id uuid.UUID) error {
	return s.repo.Delete(collegeID, id)
}

// BulkImport parses a CSV and creates students. Header (case-insensitive):
// name,register_number,email,department_code,year,section,phone
func (s *StudentService) BulkImport(collegeID uuid.UUID, r io.Reader) (*dto.BulkImportResult, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("invalid CSV: %w", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV has no data rows")
	}

	header := indexHeader(records[0])
	deptByCode, err := s.departmentCodeMap(collegeID)
	if err != nil {
		return nil, err
	}

	res := &dto.BulkImportResult{}
	var toCreate []models.Student
	for i, row := range records[1:] {
		lineNo := i + 2
		get := func(key string) string {
			if idx, ok := header[key]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}
		name, reg, email := get("name"), get("register_number"), get("email")
		if name == "" || reg == "" || email == "" {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Sprintf("line %d: name, register_number and email are required", lineNo))
			continue
		}
		if exists, _ := s.repo.ExistsByRegister(collegeID, reg); exists {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Sprintf("line %d: register number %s already exists", lineNo, reg))
			continue
		}
		hashed, err := hash.Password(reg) // default password = register number
		if err != nil {
			res.Failed++
			continue
		}
		var deptID *uuid.UUID
		if code := get("department_code"); code != "" {
			if id, ok := deptByCode[strings.ToUpper(code)]; ok {
				deptID = &id
			}
		}
		toCreate = append(toCreate, models.Student{
			CollegeID:      collegeID,
			Name:           name,
			RegisterNumber: reg,
			Email:          email,
			PasswordHash:   hashed,
			Year:           get("year"),
			Section:        get("section"),
			Phone:          get("phone"),
			DepartmentID:   deptID,
			IsActive:       true,
		})
		res.Created++
	}

	if err := s.repo.CreateBatch(toCreate); err != nil {
		return nil, fmt.Errorf("failed to persist students: %w", err)
	}
	return res, nil
}

func (s *StudentService) departmentCodeMap(collegeID uuid.UUID) (map[string]uuid.UUID, error) {
	depts, err := s.collegeRepo.ListDepartments(collegeID)
	if err != nil {
		return nil, err
	}
	m := make(map[string]uuid.UUID, len(depts))
	for _, d := range depts {
		m[strings.ToUpper(d.Code)] = d.ID
	}
	return m, nil
}

// ── helpers ──
func indexHeader(row []string) map[string]int {
	m := map[string]int{}
	for i, h := range row {
		m[strings.ToLower(strings.TrimSpace(h))] = i
	}
	return m
}

func parseUUIDPtr(s *string) *uuid.UUID {
	if s == nil || *s == "" {
		return nil
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return nil
	}
	return &id
}
