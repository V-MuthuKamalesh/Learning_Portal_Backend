package dto

// DashboardStats summarizes tenant metrics for the admin home page.
type DashboardStats struct {
	Students    int64 `json:"students"`
	Groups      int64 `json:"groups"`
	Assessments int64 `json:"assessments"`
	Questions   int64 `json:"questions"`
}
