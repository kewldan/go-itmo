package myitmo

import (
	"context"
	"time"
)

// RecordBookService is the grade book (/api/record_book).
type RecordBookService struct{ c *Client }

// Specialization is an educational programme with a grade book.
type Specialization struct {
	// MainPlan is the plan ID used as specializationID in [RecordBookService.Entries].
	MainPlan           int64                `json:"main_plan"`
	SpecializationName string               `json:"specialization_name"`
	Semesters          []RecordBookSemester `json:"semesters"`
}

// RecordBookSemester is a semester available in the grade book.
type RecordBookSemester struct {
	// StudyYear is "YYYY/YYYY".
	StudyYear string `json:"study_year"`
	// Semester is the 1-based semester number within the plan.
	Semester int `json:"semester"`
	// Course is the 1-based year of study.
	Course int  `json:"course"`
	Actual bool `json:"actual"`
}

// Teacher is a teacher's name as the grade book sends it; every part may be empty.
type Teacher struct {
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	Patronymic string `json:"patronymic"`
}

// RecordBookEntry is the final result of one discipline.
type RecordBookEntry struct {
	// Name may carry stray spaces.
	Name         string `json:"name"`
	DisciplineID int64  `json:"discipline_id"`
	// EstID loads the control tree via [RecordBookService.ControlEntries].
	EstID int64 `json:"est_id"`
	// CurrentScore is nil until grading starts.
	CurrentScore *float64 `json:"current_score"`
	// Rate is the final grade: "5/A", "4/B", "4/C", "3/D", "3/E", "2/FX",
	// "Зачёт", "Незачёт", or empty. Also count
	// "зачет", "осв", "5", "4", "3" as passed and "неявка", "незач", "2FX", "2" as failed.
	Rate          string     `json:"rate"`
	Attempt       int        `json:"attempt"`
	ControlType   string     `json:"control_type"`
	ControlTypeID int64      `json:"control_type_id"`
	ExamDate      *time.Time `json:"exam_date"`
	HaveTree      bool       `json:"have_tree"`
	LMSLink       string     `json:"lms_link"`
	Teacher       *Teacher   `json:"teacher"`
}

// ControlEntry is one assessment inside a discipline; entries form a tree via ParentID.
type ControlEntry struct {
	ID          int64    `json:"id"`
	ControlName string   `json:"control_name"`
	ParentID    *int64   `json:"parent_id"`
	LowerValue  *float64 `json:"lower_value"`
	MaxValue    *float64 `json:"max_value"`
	MinValue    *float64 `json:"min_value"`
	Required    bool     `json:"required"`
	// Rate is the score received; nil when there is no result yet.
	Rate    *float64   `json:"rate"`
	Date    *time.Time `json:"date"`
	Teacher *Teacher   `json:"teacher"`
}

// Specializations returns the user's programmes and their grade-book semesters.
// GET /api/record_book/specializations
func (s *RecordBookService) Specializations(ctx context.Context) ([]Specialization, error) {
	return call[[]Specialization](ctx, s.c, get("api/record_book/specializations", nil))
}

// Entries returns the disciplines of a semester. specializationID is
// Specialization.MainPlan. One discipline may come as several rows (e.g. exam
// and course work) sharing DisciplineID.
// GET /api/record_book/{specialization_id}/{semester}
func (s *RecordBookService) Entries(ctx context.Context, specializationID int64, semester int) ([]RecordBookEntry, error) {
	return call[[]RecordBookEntry](ctx, s.c, get("api/record_book/"+id(specializationID)+"/"+id(semester), nil))
}

// ControlEntries returns the assessment tree of a discipline. estID is RecordBookEntry.EstID.
// GET /api/record_book/{est_id}
func (s *RecordBookService) ControlEntries(ctx context.Context, estID int64) ([]ControlEntry, error) {
	return call[[]ControlEntry](ctx, s.c, get("api/record_book/"+id(estID), nil))
}
