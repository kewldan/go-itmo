package myitmo

import "context"

// AdviserService is academic adviser tools (/api/services/adviser). Every
// route is staff only: a tutor or adviser looks up students in their scope.
type AdviserService struct{ c *Client }

// AdviserStudentsParams filters Students. Zero values are not sent.
type AdviserStudentsParams struct {
	// Query is free search text.
	Query  string
	Course *int
	// Op is a programme ID from OpFilter.
	Op    *int64
	Group string
	// Limit defaults to 20.
	Limit  int
	Offset int
}

// AdviserStudent is a search hit.
type AdviserStudent struct {
	ISU        int64              `json:"isu"`
	LastName   string             `json:"last_name"`
	FirstName  string             `json:"first_name"`
	Patronymic string             `json:"patronymic"`
	Photo      string             `json:"photo"`
	Education  []AdviserEducation `json:"education"`
	Contacts   []AdviserContact   `json:"contacts"`
}

// AdviserEducation is one study record of a student.
type AdviserEducation struct {
	Course      int    `json:"course"`
	Level       string `json:"level"`
	FacultyName string `json:"faculty_name"`
	Group       string `json:"group"`
	ProgramCode string `json:"program_code"`
	ProgramName string `json:"program_name"`
	DirName     string `json:"dir_name"`
	Status      string `json:"status"`
	IsDebt      bool   `json:"is_debt"`
}

// AdviserContact is a group of contacts of one kind.
type AdviserContact struct {
	// ContactAlias is "phone", "email" or another kind.
	ContactAlias string   `json:"contact_alias"`
	Contact      []string `json:"contact"`
}

// AdviserPerson is a student's personal card.
type AdviserPerson struct {
	FIO       string             `json:"fio"`
	EngFIO    string             `json:"eng_fio"`
	Photo     string             `json:"photo"`
	BirthDate Date               `json:"birth_date"`
	Country   string             `json:"country"`
	Address   string             `json:"address"`
	Contacts  []AdviserContact   `json:"contacts"`
	Rooms     []AdviserRoom      `json:"rooms"`
	Order     *AdviserOrder      `json:"order"`
	Education []AdviserEducation `json:"education"`
}

// AdviserRoom is a dormitory room of a student.
type AdviserRoom struct {
	RoomNumber string `json:"room_number"`
	BldName    string `json:"bld_name"`
}

// AdviserOrder is the enrolment order of a student.
type AdviserOrder struct {
	// Form is the study form.
	Form string `json:"form"`
}

// AdviserActivitiesParams filters Activities. Zero values are not sent.
type AdviserActivitiesParams struct {
	// Query is free search text.
	Query string
	// Type is "project", "event", "article" or "rid"; empty for all.
	Type string
	// Limit defaults to 20.
	Limit  int
	Offset int
}

// AdviserActivity is a student's project, event, article or registered IP
// item. Dates are display strings.
type AdviserActivity struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	TypeName  string `json:"type_name"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	DateBegin string `json:"date_begin"`
	DateEnd   string `json:"date_end"`
	DateDoc   string `json:"date_doc"`
	Year      *int   `json:"year"`
	Info      string `json:"info"`
	Authors   string `json:"authors"`
	Group     string `json:"group"`
}

// AdviserScholarship is a scholarship of a student. Other wire fields are
// not known.
type AdviserScholarship struct {
	// Name is the scholarship name; the wire key is "Type" with a capital T.
	Name string `json:"Type"`
	// Type is the scholarship key, a string or a number.
	Type RawJSON `json:"type"`
}

// AdviserCourseOption is a course filter value.
type AdviserCourseOption struct {
	// Value is a course number, sent as a number or a string.
	Value RawJSON `json:"value"`
}

// AdviserGroupOption is a group filter value.
type AdviserGroupOption struct {
	Value string `json:"value"`
}

// AdviserPrograms is a student's programmes.
type AdviserPrograms struct {
	// Programs is nil when the student has no plan.
	Programs []AdviserProgram `json:"programs"`
}

// AdviserProgram is a programme and its plan.
type AdviserProgram struct {
	MainPlanID int64  `json:"mainPlanId"`
	OpName     string `json:"opName"`
}

// AdviserPlan is a student's individual curriculum.
type AdviserPlan struct {
	SemestersCount  int               `json:"semestersCount"`
	CurrentSemester int               `json:"currentSemester"`
	Structure       []AdviserPlanNode `json:"structure"`
}

// AdviserPlanNode is a module or discipline of the curriculum.
type AdviserPlanNode struct {
	ID int64 `json:"id"`
	// Type is "module" or "discipline".
	Type     string            `json:"type"`
	Name     string            `json:"name"`
	Children []AdviserPlanNode `json:"children"`
	ModuleID *int64            `json:"moduleId"`
	// ChoiceParameterID is 1, 2, 21 or 41.
	ChoiceParameterID int64   `json:"choiceParameterId"`
	Rules             []int64 `json:"rules"`
	NeedChoice        bool    `json:"needChoice"`
	ChoiceAvailable   bool    `json:"choiceAvailable"`
	CreditPoints      float64 `json:"creditPoints"`
	RPDURL            string  `json:"rpdUrl"`
	// Contents maps a start semester (as a string) to the workload.
	Contents map[string][]AdviserPlanContent `json:"contents"`
}

// AdviserPlanContent is the workload of a discipline in one semester.
type AdviserPlanContent struct {
	Semester     int                   `json:"semester"`
	CreditPoints float64               `json:"creditPoints"`
	Activities   []AdviserPlanActivity `json:"activities"`
}

// AdviserPlanActivity is one kind of work in a workload.
type AdviserPlanActivity struct {
	WorkTypeID int64 `json:"workTypeId"`
	// Volume is in academic hours.
	Volume float64 `json:"volume"`
}

// AdviserChoice is a student's saved elective choice.
type AdviserChoice struct {
	ChosenContents []AdviserChosenContent `json:"chosenContents"`
}

// AdviserChosenContent is one chosen discipline.
type AdviserChosenContent struct {
	ModuleID     int64 `json:"moduleId"`
	DisciplineID int64 `json:"disciplineId"`
	Semester     int   `json:"semester"`
}

// AdviserReplacement is a pending discipline replacement request. Other wire
// fields are not known.
type AdviserReplacement struct {
	DisciplineIDFrom int64 `json:"disciplineIdFrom"`
	ModuleID         int64 `json:"moduleId"`
	// Name is the new discipline name (not confirmed).
	Name string `json:"name"`
}

func limitOr20(n int) int {
	if n <= 0 {
		return 20
	}
	return n
}

// Students searches students in the adviser's scope.
// GET /api/services/adviser/students
// Staff only.
func (s *AdviserService) Students(ctx context.Context, p AdviserStudentsParams) (*Page[AdviserStudent], error) {
	qv := q().set("query", p.Query).set("course", p.Course).set("op", p.Op).set("group", p.Group).
		set("limit", limitOr20(p.Limit)).set("offset", p.Offset)
	return call[*Page[AdviserStudent]](ctx, s.c, get("api/services/adviser/students", qv))
}

// Student returns a student's personal card.
// GET /api/services/adviser/students/{isu}
// Staff only.
func (s *AdviserService) Student(ctx context.Context, isu int64) (*AdviserPerson, error) {
	return call[*AdviserPerson](ctx, s.c, get("api/services/adviser/students/"+id(isu), nil))
}

// Activities returns a student's projects, events, articles and registered IP items.
// GET /api/services/adviser/students/{isu}/activities
// Staff only.
func (s *AdviserService) Activities(ctx context.Context, isu int64, p AdviserActivitiesParams) (*Page[AdviserActivity], error) {
	qv := q().set("q", p.Query).set("type", p.Type).set("limit", limitOr20(p.Limit)).set("offset", p.Offset)
	return call[*Page[AdviserActivity]](ctx, s.c, get("api/services/adviser/students/"+id(isu)+"/activities", qv))
}

// Scholarship returns a student's scholarships over the last six months; empty when none were assigned.
// GET /api/services/adviser/students/{isu}/scholarship
// Staff only.
func (s *AdviserService) Scholarship(ctx context.Context, isu int64) ([]AdviserScholarship, error) {
	r, err := call[struct {
		Data []AdviserScholarship `json:"data"`
	}](ctx, s.c, get("api/services/adviser/students/"+id(isu)+"/scholarship", nil))
	return r.Data, err
}

// CourseFilter returns the course filter values.
// GET /api/services/adviser/filters/course
// Staff only.
func (s *AdviserService) CourseFilter(ctx context.Context) ([]AdviserCourseOption, error) {
	return call[[]AdviserCourseOption](ctx, s.c, get("api/services/adviser/filters/course", nil))
}

// OpFilter returns the educational programme filter values.
// GET /api/services/adviser/filters/op
// Staff only.
func (s *AdviserService) OpFilter(ctx context.Context) ([]IDValue, error) {
	return call[[]IDValue](ctx, s.c, get("api/services/adviser/filters/op", nil))
}

// GroupFilter returns the group filter values.
// GET /api/services/adviser/filters/groups
// Staff only.
func (s *AdviserService) GroupFilter(ctx context.Context) ([]AdviserGroupOption, error) {
	return call[[]AdviserGroupOption](ctx, s.c, get("api/services/adviser/filters/groups", nil))
}

// Schedule returns a student's timetable for the inclusive range [from, to].
// GET /api/services/adviser/services/schedule
// Staff only.
func (s *AdviserService) Schedule(ctx context.Context, isu int64, from, to Date) ([]ScheduleDay, error) {
	qv := q().set("date_start", from).set("date_end", to).set("isu", isu)
	return callData[[]ScheduleDay](ctx, s.c, get("api/services/adviser/services/schedule", qv))
}

// RecordBookSpecializations returns a student's programmes and their grade-book semesters.
// GET /api/services/adviser/services/record_book/specializations
// Staff only.
func (s *AdviserService) RecordBookSpecializations(ctx context.Context, isu int64) ([]Specialization, error) {
	return call[[]Specialization](ctx, s.c, get("api/services/adviser/services/record_book/specializations", q().set("isu", isu)))
}

// RecordBookEntries returns a student's grades for a semester; there is one row per attempt.
// planID is Specialization.MainPlan.
// GET /api/services/adviser/services/record_book/{planId}/{semester}
// Staff only.
func (s *AdviserService) RecordBookEntries(ctx context.Context, isu, planID int64, semester int) ([]RecordBookEntry, error) {
	return call[[]RecordBookEntry](ctx, s.c, get("api/services/adviser/services/record_book/"+id(planID)+"/"+id(semester), q().set("isu", isu)))
}

// RecordBookControls returns the assessment tree of a student's discipline. estID is RecordBookEntry.EstID.
// GET /api/services/adviser/services/record_book/details
// Staff only.
func (s *AdviserService) RecordBookControls(ctx context.Context, isu, estID int64) ([]ControlEntry, error) {
	return call[[]ControlEntry](ctx, s.c, get("api/services/adviser/services/record_book/details", q().set("est_id", estID).set("isu", isu)))
}

// Programs returns a student's programmes and plans.
// GET /api/services/adviser/services/program/{isu}
// Staff only.
func (s *AdviserService) Programs(ctx context.Context, isu int64) (*AdviserPrograms, error) {
	return call[*AdviserPrograms](ctx, s.c, get("api/services/adviser/services/program/"+id(isu), nil))
}

// PersonPlan returns a student's individual curriculum. planID is AdviserProgram.MainPlanID.
// GET /api/services/adviser/services/person_plan/{isu}/{planId}
// Staff only.
func (s *AdviserService) PersonPlan(ctx context.Context, isu, planID int64) (*AdviserPlan, error) {
	return call[*AdviserPlan](ctx, s.c, get("api/services/adviser/services/person_plan/"+id(isu)+"/"+id(planID), nil))
}

// Choice returns a student's saved elective choice in a plan.
// GET /api/services/adviser/services/choice/{isu}/{planId}
// Staff only.
func (s *AdviserService) Choice(ctx context.Context, isu, planID int64) (*AdviserChoice, error) {
	return call[*AdviserChoice](ctx, s.c, get("api/services/adviser/services/choice/"+id(isu)+"/"+id(planID), nil))
}

// Replacements returns a student's pending discipline replacement requests in a plan.
// GET /api/services/adviser/services/replace/{isu}/{planId}
// Staff only.
func (s *AdviserService) Replacements(ctx context.Context, isu, planID int64) ([]AdviserReplacement, error) {
	return call[[]AdviserReplacement](ctx, s.c, get("api/services/adviser/services/replace/"+id(isu)+"/"+id(planID), nil))
}
