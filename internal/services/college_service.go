package services

import (
	"io"

	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/models"
	"github.com/collegeassess/backend/internal/repositories"
	"github.com/collegeassess/backend/pkg/storage"
	"github.com/google/uuid"
)

// CollegeService manages tenants, branding, departments and batches.
type CollegeService struct {
	repo  *repositories.CollegeRepository
	store storage.Storage
}

func NewCollegeService(repo *repositories.CollegeRepository, store storage.Storage) *CollegeService {
	return &CollegeService{repo: repo, store: store}
}

func (s *CollegeService) Branding(code string) (*dto.CollegeBranding, error) {
	c, err := s.repo.ByCode(code)
	if err != nil {
		return nil, err
	}
	return &dto.CollegeBranding{
		Name: c.Name, Code: c.Code, LogoURL: c.LogoURL,
		PrimaryColor: c.PrimaryColor, SecondaryColor: c.SecondaryColor,
	}, nil
}

func (s *CollegeService) Get(id uuid.UUID) (*models.College, error) { return s.repo.ByID(id) }
func (s *CollegeService) List() ([]models.College, error)           { return s.repo.List() }

func (s *CollegeService) Create(req dto.CreateCollegeRequest) (*models.College, error) {
	c := &models.College{
		Name: req.Name, Code: req.Code,
		PrimaryColor: orDefault(req.PrimaryColor, "#4f46e5"),
		SecondaryColor: orDefault(req.SecondaryColor, "#0ea5e9"),
		ContactEmail: req.ContactEmail, ContactPhone: req.ContactPhone,
		Address: req.Address, IsActive: true,
	}
	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CollegeService) Update(id uuid.UUID, req dto.UpdateCollegeRequest) (*models.College, error) {
	c, err := s.repo.ByID(id)
	if err != nil {
		return nil, err
	}
	applyStr(&c.Name, req.Name)
	applyStr(&c.PrimaryColor, req.PrimaryColor)
	applyStr(&c.SecondaryColor, req.SecondaryColor)
	applyStr(&c.ContactEmail, req.ContactEmail)
	applyStr(&c.ContactPhone, req.ContactPhone)
	applyStr(&c.Address, req.Address)
	if req.IsActive != nil {
		c.IsActive = *req.IsActive
	}
	return c, s.repo.Update(c)
}

func (s *CollegeService) Delete(id uuid.UUID) error { return s.repo.Delete(id) }

// UploadLogo stores the logo and updates the college record.
func (s *CollegeService) UploadLogo(id uuid.UUID, filename string, r io.Reader) (string, error) {
	c, err := s.repo.ByID(id)
	if err != nil {
		return "", err
	}
	url, err := s.store.Save("logos", filename, r)
	if err != nil {
		return "", err
	}
	c.LogoURL = url
	if err := s.repo.Update(c); err != nil {
		return "", err
	}
	return url, nil
}

// ── Departments ──
func (s *CollegeService) ListDepartments(collegeID uuid.UUID) ([]models.Department, error) {
	return s.repo.ListDepartments(collegeID)
}
func (s *CollegeService) CreateDepartment(collegeID uuid.UUID, req dto.DepartmentRequest) (*models.Department, error) {
	d := &models.Department{CollegeID: collegeID, Name: req.Name, Code: req.Code}
	return d, s.repo.CreateDepartment(d)
}
func (s *CollegeService) UpdateDepartment(collegeID, id uuid.UUID, req dto.DepartmentRequest) (*models.Department, error) {
	d, err := s.repo.DepartmentByID(collegeID, id)
	if err != nil {
		return nil, err
	}
	d.Name, d.Code = req.Name, req.Code
	return d, s.repo.UpdateDepartment(d)
}
func (s *CollegeService) DeleteDepartment(collegeID, id uuid.UUID) error {
	return s.repo.DeleteDepartment(collegeID, id)
}

// ── Batches ──
func (s *CollegeService) ListBatches(collegeID uuid.UUID) ([]models.Batch, error) {
	return s.repo.ListBatches(collegeID)
}
func (s *CollegeService) CreateBatch(collegeID uuid.UUID, req dto.BatchRequest) (*models.Batch, error) {
	b := &models.Batch{CollegeID: collegeID, Name: req.Name, StartYear: req.StartYear, EndYear: req.EndYear}
	return b, s.repo.CreateBatch(b)
}
func (s *CollegeService) UpdateBatch(collegeID, id uuid.UUID, req dto.BatchRequest) (*models.Batch, error) {
	b, err := s.repo.BatchByID(collegeID, id)
	if err != nil {
		return nil, err
	}
	b.Name, b.StartYear, b.EndYear = req.Name, req.StartYear, req.EndYear
	return b, s.repo.UpdateBatch(b)
}
func (s *CollegeService) DeleteBatch(collegeID, id uuid.UUID) error {
	return s.repo.DeleteBatch(collegeID, id)
}

// ── small helpers ──
func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
func applyStr(dst *string, src *string) {
	if src != nil {
		*dst = *src
	}
}
