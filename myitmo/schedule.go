package myitmo

import "context"

// ScheduleService is the personal timetable (/api/schedule).
type ScheduleService struct{ c *Client }

// ScheduleDay is the timetable of one calendar day.
type ScheduleDay struct {
	// DayNumber is the day index within the requested range.
	DayNumber int `json:"day_number"`
	// WeekNumber is the academic week number.
	WeekNumber int      `json:"week_number"`
	Date       Date     `json:"date"`
	Note       string   `json:"note"`
	Lessons    []Lesson `json:"lessons"`
	// Type is present on the wire; its values are not confirmed.
	Type RawJSON `json:"type,omitzero"`
	// Intersections are groups of Lesson.PairID values whose times overlap.
	Intersections [][]int64 `json:"intersections"`
}

// Lesson is one class in the personal timetable.
type Lesson struct {
	PairID    int64  `json:"pair_id"`
	Subject   string `json:"subject"`
	SubjectID int64  `json:"subject_id"`
	Note      string `json:"note"`
	Type      string `json:"type"`
	// TimeStart and TimeEnd are "HH:MM" in Moscow time.
	TimeStart   string `json:"time_start"`
	TimeEnd     string `json:"time_end"`
	TeacherID   *int64 `json:"teacher_id"`
	TeacherName string `json:"teacher_name"`
	Room        string `json:"room"`
	Building    string `json:"building"`
	// Format is the display name of FormatID.
	Format string `json:"format"`
	// WorkType is the display name of WorkTypeID.
	WorkType string `json:"work_type"`
	// WorkTypeID: see the WorkType constants; sport lessons may send null (decoded as 0).
	WorkTypeID int    `json:"work_type_id"`
	Group      string `json:"group"`
	// FlowTypeID: 2 classes, 3 sport, 5 room booking.
	FlowTypeID   int    `json:"flow_type_id"`
	FlowID       int64  `json:"flow_id"`
	ZoomURL      string `json:"zoom_url"`
	ZoomPassword string `json:"zoom_password"`
	ZoomInfo     string `json:"zoom_info"`
	// BuildingID is the actual building; nil for online classes.
	BuildingID *int64 `json:"bld_id"`
	// FormatID: 1 on-site, 2 hybrid, 3 online.
	FormatID int `json:"format_id"`
	// MainBuildingID: 13 Kronverksky, 273 Lomonosova, 5 Vyazemsky, 319 virtual rooms (common values).
	MainBuildingID *int64 `json:"main_bld_id"`
}

// Lesson work types of Lesson.WorkTypeID.
const (
	WorkTypeLecture       = 1
	WorkTypeLab           = 2
	WorkTypePractice      = 3
	WorkTypeExam          = 5
	WorkTypeCredit        = 6
	WorkTypeCourseProject = 8
	WorkTypeGradedCredit  = 9
	WorkTypeConsultation  = 10
	WorkTypeSport         = 11
	WorkTypeBooking       = 131
)

// TimeSlot is a class period.
type TimeSlot struct {
	ID        int64  `json:"id"`
	TimeStart string `json:"time_start"`
	TimeEnd   string `json:"time_end"`
	// Order is the period number within the day (schedule slots only).
	Order int `json:"order,omitzero"`
}

// Personal returns the user's timetable for the inclusive range [from, to],
// optionally only for the given subjects (Lesson.SubjectID).
// GET /api/schedule/schedule/personal
func (s *ScheduleService) Personal(ctx context.Context, from, to Date, subjectIDs ...int64) ([]ScheduleDay, error) {
	qv := q().set("date_start", from).set("date_end", to).set("subject_id[]", subjectIDs)
	return callData[[]ScheduleDay](ctx, s.c, get("api/schedule/schedule/personal", qv))
}

// Subjects returns the user's subjects for the timetable filter. Shape not
// confirmed (likely a list of {id, name}).
// GET /api/schedule/subjects
func (s *ScheduleService) Subjects(ctx context.Context) (RawJSON, error) {
	return callData[RawJSON](ctx, s.c, get("api/schedule/subjects", nil))
}

// TimeSlots returns the standard class periods in display order.
// GET /api/schedule/meta/time_slots
func (s *ScheduleService) TimeSlots(ctx context.Context) ([]TimeSlot, error) {
	return callData[[]TimeSlot](ctx, s.c, get("api/schedule/meta/time_slots", nil))
}
