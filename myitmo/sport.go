package myitmo

import (
	"context"
	"time"

	"github.com/kewldan/go-itmo/internal/jsonx"
	"github.com/kewldan/go-itmo/internal/rest"
)

// SportService is physical education: schedule, sign-up, scores (/api/sport).
type SportService struct{ c *Client }

// Sport error codes (Error.Code). The first call of a sport page decides
// whether the section is usable at all.
const (
	// SportErrorChoiceClosed: section choice is closed; SportSemester.ChoiceStart tells when it opens.
	SportErrorChoiceClosed = 135
	// SportErrorNoAccess: the user has no access to sport sign-up.
	SportErrorNoAccess = 114
	// SportErrorCannotSignIn: sign-in refused; Message lists the reasons
	// ("Выбрано 1 занятие в этот день", "Есть запись на занятия в это время: ...", ...).
	SportErrorCannotSignIn = 137
	// SportErrorCannotSignOut: sign-out refused; Message lists the reasons
	// ("Вы не записаны на это занятие", ...).
	SportErrorCannotSignOut = 130
)

// Sport lesson levels of SportLesson.LessonLevel and SportLessonGroup.Level.
const (
	// SportLessonLevelOpen is an open or free-visit lesson; colour it by TypeID.
	SportLessonLevelOpen = 1
	// SportLessonLevelBasic is the basic (training) level of a section.
	SportLessonLevelBasic = 2
	// SportLessonLevelAdvanced is the advanced (intermediate) level of a section.
	SportLessonLevelAdvanced = 3
	// SportLessonLevelTeam is the university team.
	SportLessonLevelTeam = 4
)

// Sport lesson types of SportLesson.TypeID, SportOtherLesson.TypeID and ChosenSportLesson.TypeID.
const (
	// SportLessonTypeOpen is an open lesson; in a selective section (LessonLevel != 1)
	// signing in requires the open-class form, see [SportService.SignInGroup].
	SportLessonTypeOpen = 1
	// SportLessonTypeFreeVisit is a free-visit lesson.
	SportLessonTypeFreeVisit = 2
	// SportLessonTypeDebt is a lesson for students with a debt.
	SportLessonTypeDebt = 5
	// SportLessonTypeNormatives is a fitness test session.
	SportLessonTypeNormatives = 6
	// SportLessonTypeExternat is an external-study (externat) session.
	SportLessonTypeExternat = 7
	// SportLessonTypeExtra is an additional lesson.
	SportLessonTypeExtra = 8
)

// Sport section levels of SportLesson.SectionLevel.
const (
	// SportSectionLevelFreeVisit is a free-visit section; any other value is a selective section.
	SportSectionLevelFreeVisit = 1
	// SportSectionLevelSelective is a section with a selection.
	SportSectionLevelSelective = 2
)

// Sport building and room IDs with a special meaning.
const (
	// SportBuildingOnline is the "Online" entry of SportFilters.BuildingID.
	SportBuildingOnline = -1
	// SportBuildingOther is the "other venues" entry of SportFilters.BuildingID. It is a
	// filter value only: lessons at other venues carry their own positive building IDs.
	SportBuildingOther = 0
	// SportBuildingLomonosova is the default building.
	SportBuildingLomonosova = 273
	// SportRoomOnline is the SportLesson.RoomID of an online lesson.
	SportRoomOnline = -1
)

// Externat application statuses of SportExternat.ExternatStatusID.
const (
	SportExternatPending  = 1
	SportExternatDeclined = 2
	SportExternatApproved = 3
)

// Special project IDs with a known meaning (SportProject.ID).
const (
	// SportProjectKronbarsRunning needs no link when signing in.
	SportProjectKronbarsRunning = 1
	// SportProjectExternal is the external credit ("Экстернат"); signing in
	// sends a link to the supporting document.
	SportProjectExternal = 3
	// SportProjectTheory is the theoretical credit.
	SportProjectTheory = 4
)

// SportHealthLevelSpecial is the SportProject.HealthLevelID of projects for the special health group.
const SportHealthLevelSpecial = 1

// Known SportAttendance.Type values; the set may grow.
const (
	SportAttendanceLesson      = "lesson"
	SportAttendanceCompetition = "competition"
)

// SportTimeSlot is a row of the weekly sport grid.
type SportTimeSlot struct {
	ID int64 `json:"id"`
	// TimeStart and TimeEnd are "HH:MM" in Moscow time.
	TimeStart string `json:"time_start"`
	TimeEnd   string `json:"time_end"`
}

// SportSemesterOption is an entry of the sport semester list.
type SportSemesterOption struct {
	// ID is the semester_id of [SportService.Score].
	ID int64 `json:"id"`
	// Value is the label, e.g. "Весна 2025/2026".
	Value   string `json:"value"`
	Comment string `json:"comment"`
}

// SportSemester is the current sport semester and its key dates.
type SportSemester struct {
	ID int64 `json:"id"`
	// StudyYear is "YYYY/YYYY".
	StudyYear string `json:"study_year"`
	// Semester is the server's semester number; it may be 0.
	Semester  int       `json:"semester"`
	DateStart time.Time `json:"date_start"`
	// DateEnd is the soft end of the main period.
	DateEnd time.Time `json:"date_end"`
	// HardDateEnd is when the semester closes completely.
	HardDateEnd time.Time `json:"hard_date_end"`
	Current     bool      `json:"current"`
	// ChoiceStart is when section choice opens; may be a technical minimum date.
	ChoiceStart time.Time `json:"choice_start"`
	// BachelorBound may be a technical minimum date.
	BachelorBound time.Time `json:"bachelor_bound"`
	PPA1Start     time.Time `json:"ppa1_start"`
	PPA1End       time.Time `json:"ppa1_end"`
	PPA2Start     time.Time `json:"ppa2_start"`
	PPA2End       time.Time `json:"ppa2_end"`
	// SignDuration is the allowed sign-up duration in days.
	SignDuration int `json:"sign_duration"`
}

// SportFilters are the values of the sign-up schedule filters, not a full
// catalogue: BuildingID holds -1 (Online), 0 (other venues) and the main
// buildings, while lessons at other venues carry positive IDs missing here.
type SportFilters struct {
	BuildingID  []IDValue `json:"building_id"`
	SectionID   []IDValue `json:"section_id"`
	SportTypeID []IDValue `json:"sport_type_id"`
	// TeacherISU entries have the teacher's ISU number as ID.
	TeacherISU []IDValue `json:"teacher_isu"`
}

// SportScheduleParams selects lessons of [SportService.Schedule].
type SportScheduleParams struct {
	// DateStart and DateEnd bound the range, typically 7 days.
	DateStart Date
	DateEnd   Date
	// BuildingID is always sent; see SportFilters.BuildingID.
	BuildingID int64
	// SportTypeIDs and TeacherISUs are optional multi-value filters.
	SportTypeIDs []int64
	TeacherISUs  []int64
}

// SportScheduleDay is the sport lessons of one date.
type SportScheduleDay struct {
	Date Date `json:"date"`
	// Lessons may be null instead of empty.
	Lessons []SportLesson `json:"lessons"`
}

// SportLesson is a lesson open for sign-up or an entry of the personal calendar.
type SportLesson struct {
	ID          int64     `json:"id"`
	Date        time.Time `json:"date"`
	DateEnd     time.Time `json:"date_end"`
	SectionID   int64     `json:"section_id"`
	SectionName string    `json:"section_name"`
	// SectionLevel: see the SportSectionLevel constants.
	SectionLevel  int   `json:"section_level"`
	LessonGroupID int64 `json:"lesson_group_id"`
	// LessonLevel: see the SportLessonLevel constants.
	LessonLevel int `json:"lesson_level"`
	// TypeID: see the SportLessonType constants.
	TypeID int `json:"type_id"`
	// BuildingID is the real venue, not necessarily listed in SportFilters;
	// nil for online lessons.
	BuildingID *int64 `json:"building_id"`
	// RoomID is the real room; SportRoomOnline (-1) for online lessons.
	RoomID   int64  `json:"room_id"`
	RoomName string `json:"room_name"`
	// Limit and Available are the places; absent in the personal calendar.
	// Live values come from [SportService.Limits].
	Limit     *int64 `json:"limit"`
	Available *int64 `json:"available"`
	Comment   string `json:"comment"`
	// TimeSlotID, TimeSlotStart and TimeSlotEnd place the lesson in the grid of [SportService.TimeSlots].
	TimeSlotID    int64  `json:"time_slot_id"`
	TimeSlotStart string `json:"time_slot_start"`
	TimeSlotEnd   string `json:"time_slot_end"`
	// Intersection reports an overlap with another event of the user.
	Intersection bool            `json:"intersection"`
	CanSignIn    *SportCanSignIn `json:"can_sign_in"`
	// OtherLessons are all dates of the same lesson group in the semester.
	OtherLessons []SportOtherLesson `json:"other_lessons"`
	Signed       *bool              `json:"signed"`
	TeacherISU   int64              `json:"teacher_isu"`
	TeacherFIO   string             `json:"teacher_fio"`
	// LinkURL is the video-call link of an online lesson (personal calendar).
	LinkURL string `json:"link_url"`
}

// SportCanSignIn is the server's decision whether the user may sign in.
type SportCanSignIn struct {
	CanSignIn bool `json:"can_sign_in"`
	// UnavailableReasons are localised; empty when signing in is allowed.
	UnavailableReasons []string `json:"unavailable_reasons"`
}

// SportOtherLesson is another date of the same lesson group.
type SportOtherLesson struct {
	ID        int64     `json:"id"`
	DateStart time.Time `json:"date_start"`
	// Weekday is 0 = Sunday ... 6 = Saturday, like time.Weekday.
	Weekday        int    `json:"weekday"`
	RoomID         int64  `json:"room_id"`
	RoomName       string `json:"room_name"`
	EvaluationID   int64  `json:"evaluation_id"`
	EvaluationName string `json:"evaluation_name"`
	TimeSlotID     int64  `json:"time_slot_id"`
	TimeSlotStart  string `json:"time_slot_start"`
	TimeSlotEnd    string `json:"time_slot_end"`
	// TimeStart and TimeEnd are "HH:MM"; when empty use the time slot.
	TimeStart    string          `json:"time_start"`
	TimeEnd      string          `json:"time_end"`
	Repeatable   bool            `json:"repeatable"`
	TeacherISU   int64           `json:"teacher_isu"`
	TeacherFIO   string          `json:"teacher_fio"`
	TypeID       int             `json:"type_id"`
	Comment      string          `json:"comment"`
	Intersection bool            `json:"intersection"`
	CanSignIn    *SportCanSignIn `json:"can_sign_in"`
}

// SportSignLimit is the number of places of a lesson or competition.
type SportSignLimit struct {
	Limit int `json:"limit"`
	// Available may be negative; treat it as 0.
	Available int `json:"available"`
}

// SportOtherSelection is the answer of [SportService.OtherLessons]. The server
// sends a list of lesson IDs for free-visit lessons (lesson_level 1) and
// {"signed": bool} otherwise.
type SportOtherSelection struct {
	// LessonIDs are the lessons of the group the user is already signed in to.
	LessonIDs []int64
	// Signed is set for section lessons.
	Signed bool
}

// UnmarshalJSON accepts both shapes.
func (o *SportOtherSelection) UnmarshalJSON(b []byte) error {
	*o = SportOtherSelection{}
	if sportIsJSONArray(b) {
		return jsonx.Unmarshal(b, &o.LessonIDs)
	}
	var v struct {
		Signed bool `json:"signed"`
	}
	if err := jsonx.Unmarshal(b, &v); err != nil {
		return err
	}
	o.Signed = v.Signed
	return nil
}

// SportOpenFormSubmit is the questionnaire sent when signing in to an open
// class of a selective section. Nil fields are sent as null.
type SportOpenFormSubmit struct {
	SchoolName *string `json:"school_name"`
	// RankID is an ID of [SportService.OpenFormRanks].
	RankID       *int64  `json:"rank_id"`
	Comment      *string `json:"comment"`
	Achievements *string `json:"achievements"`
	SectionID    int64   `json:"section_id"`
	// ISU is the current user's ISU number.
	ISU int64 `json:"isu"`
}

// SportOpenForm is a previously filled open-class questionnaire.
type SportOpenForm struct {
	ID           int64  `json:"id"`
	SchoolName   string `json:"school_name"`
	RankID       *int64 `json:"rank_id"`
	Comment      string `json:"comment"`
	Achievements string `json:"achievements"`
}

// ChosenSportSection is a section the user has chosen.
type ChosenSportSection struct {
	ID          int64  `json:"id"`
	SectionName string `json:"section_name"`
	// Level is 1 for free visit, greater for selective sections.
	Level        int                `json:"level"`
	LessonGroups []SportLessonGroup `json:"lesson_groups"`
}

// SportLessonGroup is a group of regular lessons inside a chosen section.
type SportLessonGroup struct {
	ID int64 `json:"id"`
	// Level: see the SportLessonLevel constants.
	Level     int    `json:"level"`
	LevelName string `json:"level_name"`
	// HasFutureLessons reports whether dated lessons are still ahead.
	HasFutureLessons bool `json:"has_future_lessons"`
	// Lessons are the dated lessons; may be null.
	Lessons []ChosenSportLesson `json:"lessons"`
	// Weekdays are the regular lessons described by weekday.
	Weekdays []ChosenSportWeekday `json:"weekdays"`
}

// ChosenSportLesson is a dated lesson of a chosen lesson group.
type ChosenSportLesson struct {
	ID         int64     `json:"id"`
	DateStart  time.Time `json:"date_start"`
	DateEnd    time.Time `json:"date_end"`
	TimeSlotID int64     `json:"time_slot_id"`
	// TimeStart and TimeEnd are "HH:MM"; when empty use the time slot.
	TimeStart  string `json:"time_start"`
	TimeEnd    string `json:"time_end"`
	RoomID     int64  `json:"room_id"`
	RoomName   string `json:"room_name"`
	TeacherISU int64  `json:"teacher_isu"`
	TeacherFIO string `json:"teacher_fio"`
	// TypeID: see the SportLessonType constants.
	TypeID int `json:"type_id"`
	// LinkURL is the link of an online lesson; may be empty.
	LinkURL      string `json:"link_url"`
	Comment      string `json:"comment"`
	Intersection bool   `json:"intersection"`
}

// ChosenSportWeekday is a regular lesson given by weekday instead of a date.
type ChosenSportWeekday struct {
	ChosenSportLesson `json:",inline"`
	// Weekday is 0 = Sunday ... 6 = Saturday, like time.Weekday.
	Weekday int `json:"weekday"`
}

// SportScore is the sport points of a semester and their history.
type SportScore struct {
	Sum SportScoreSum `json:"sum"`
	// Attendances is null when nothing was credited yet.
	Attendances []SportAttendance `json:"attendances"`
}

// SportScoreSum splits the total by source.
type SportScoreSum struct {
	// Attendances are points for lessons; 60 or more means the credit is reached.
	Attendances int64 `json:"attendances"`
	// Other are bonus points.
	Other int64 `json:"other"`
}

// SportAttendance is one credit of sport points.
type SportAttendance struct {
	// Type: see the SportAttendance constants.
	Type           string    `json:"type"`
	Name           string    `json:"name"`
	EvaluationID   int64     `json:"evaluation_id"`
	EvaluationName string    `json:"evaluation_name"`
	SectionLevel   int       `json:"section_level"`
	Score          int       `json:"score"`
	Date           time.Time `json:"date"`
	IsCompetition  bool      `json:"is_competition"`
	// DisciplineName and CompetitionName are empty for ordinary lessons.
	DisciplineName  string `json:"discipline_name"`
	CompetitionName string `json:"competition_name"`
	// Place is the result at a competition, e.g. "Участие".
	Place string `json:"place"`
}

// SportAttempts counts sign-up attempts.
type SportAttempts struct {
	// TotalAttempts is shown as "N visits per week/semester".
	TotalAttempts int  `json:"total_attempts"`
	UsedAttempts  int  `json:"used_attempts"`
	FreeAttempts  int  `json:"free_attempts"`
	CanSignIn     bool `json:"can_sign_in"`
}

// SportDebt is the physical-education debt.
type SportDebt struct {
	IsHavingDebt bool `json:"is_having_debt"`
	// NeededScore is absent without a debt.
	NeededScore *float64 `json:"needed_score"`
	// FreeAttempts are the attempts for debt lessons; absent without a debt.
	FreeAttempts *int `json:"free_attempts"`
}

// SportExternat is the state of the externat application.
type SportExternat struct {
	// Signed reports an accepted externat sign-up; with it lessons of
	// SportLessonTypeNormatives become available.
	Signed bool `json:"signed"`
	// ExternatStatusID: see the SportExternat constants; nil when never applied.
	ExternatStatusID *int64 `json:"externat_status_id"`
	// DeclineReason is set for declined applications.
	DeclineReason string `json:"decline_reason"`
}

// SportHealthLevel is the medical health group.
type SportHealthLevel struct {
	ID  int64 `json:"id"`
	ISU int64 `json:"isu"`
	// Name is localised, e.g. "Основная группа здоровья".
	Name string `json:"name"`
}

// SportSelection is a sport with passed selections.
type SportSelection struct {
	ID         int64            `json:"id"`
	Name       string           `json:"name"`
	Requisites []SportRequisite `json:"requisites"`
}

// SportRequisite is a level or test of a sport selection.
type SportRequisite struct {
	ID int64 `json:"id"`
	// LevelName is e.g. "Сборная команда".
	LevelName string `json:"level_name"`
	// Name is the sport and level.
	Name string `json:"name"`
}

// SportBriefingList is the state of the safety briefing.
type SportBriefingList struct {
	Signed bool         `json:"signed"`
	Briefs []SportBrief `json:"briefs"`
}

// SportBrief is a briefing document; other fields are unknown.
type SportBrief struct {
	// ID is likely the briefingID of [SportService.DownloadBriefing].
	ID int64 `json:"id"`
	// URL links to the signed document.
	URL string `json:"url"`
}

// SportBriefingSignTask is the document created for a simple electronic
// signature; pass it to the signing service.
type SportBriefingSignTask struct {
	// SignatureID and SignTaskID are numbers or strings.
	SignatureID RawJSON `json:"signature_id"`
	SignTaskID  RawJSON `json:"sign_task_id"`
}

// SportProject is a special format of the sport credit (Kronbars Running, theory, ...).
type SportProject struct {
	// ID: see the SportProject constants.
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Signed      bool   `json:"signed"`
	Limit       int    `json:"limit"`
	Available   int    `json:"available"`
	// InstructionLink links to the rules.
	InstructionLink string `json:"instruction_link"`
	// InstructionDescription is the rules confirmation text.
	InstructionDescription string `json:"instruction_description"`
	// RequisiteAvailable reports whether the prerequisites are met.
	RequisiteAvailable bool `json:"requisite_available"`
	// Link is the user's link sent when signing in; may be empty.
	Link string `json:"link"`
	// HealthLevelID is SportHealthLevelSpecial for projects of the special
	// health group.
	HealthLevelID *int64 `json:"health_level_id"`
}

// SportCompetition is a sport competition.
type SportCompetition struct {
	ID            int64                        `json:"id"`
	Name          string                       `json:"name"`
	SportTypeName string                       `json:"sport_type_name"`
	DateStart     time.Time                    `json:"date_start"`
	BuildingName  string                       `json:"building_name"`
	Disciplines   []SportCompetitionDiscipline `json:"disciplines"`
	// RegistrationLink is an external registration page, if any.
	RegistrationLink string `json:"registration_link"`
	// FreeRegistration reports whether registration is open.
	FreeRegistration bool `json:"free_registration"`
	Available        int  `json:"available"`
}

// SportCompetitionDiscipline is an event of a competition.
type SportCompetitionDiscipline struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Signed bool   `json:"signed"`
}

// TimeSlots returns the rows of the weekly sport grid.
// GET /api/sport/time_slots
func (s *SportService) TimeSlots(ctx context.Context) ([]SportTimeSlot, error) {
	return call[[]SportTimeSlot](ctx, s.c, get("api/sport/time_slots", nil))
}

// SportTypes returns the kinds of sport (competition filter).
// GET /api/sport/sport_types
func (s *SportService) SportTypes(ctx context.Context) ([]IDValue, error) {
	return call[[]IDValue](ctx, s.c, get("api/sport/sport_types", nil))
}

// Semesters returns the sport semesters of the score history. Error code
// SportErrorChoiceClosed or SportErrorNoAccess means the section is unavailable.
// GET /api/sport/semesters/list
func (s *SportService) Semesters(ctx context.Context) ([]SportSemesterOption, error) {
	return call[[]SportSemesterOption](ctx, s.c, get("api/sport/semesters/list", nil))
}

// CurrentSemester returns the current sport semester and its key dates.
// GET /api/sport/semesters/current
func (s *SportService) CurrentSemester(ctx context.Context) (*SportSemester, error) {
	return call[*SportSemester](ctx, s.c, get("api/sport/semesters/current", nil))
}

// Filters returns the values of the schedule filters. Error code
// SportErrorChoiceClosed or SportErrorNoAccess means sign-up is unavailable.
// GET /api/sport/sign/schedule/filters
func (s *SportService) Filters(ctx context.Context) (*SportFilters, error) {
	return call[*SportFilters](ctx, s.c, get("api/sport/sign/schedule/filters", nil))
}

// Schedule returns the lessons open for sign-up in a date range.
// GET /api/sport/sign/schedule
func (s *SportService) Schedule(ctx context.Context, p SportScheduleParams) ([]SportScheduleDay, error) {
	qv := q().
		set("date_start", p.DateStart).
		set("date_end", p.DateEnd).
		set("building_id", p.BuildingID).
		set("sport_type_id", p.SportTypeIDs).
		set("teacher_isu", p.TeacherISUs)
	return call[[]SportScheduleDay](ctx, s.c, get("api/sport/sign/schedule", qv))
}

// Limits returns the live places, keyed by lesson group ID and then lesson ID.
// GET /api/sport/sign/schedule/limits
func (s *SportService) Limits(ctx context.Context) (map[int64]map[int64]SportSignLimit, error) {
	return sportCallMap[map[int64]SportSignLimit](ctx, s.c, get("api/sport/sign/schedule/limits", nil))
}

// OtherLessons returns what the user already chose in the lesson group of lessonID.
// GET /api/sport/sign/schedule/lessons/{lesson_id}/other
func (s *SportService) OtherLessons(ctx context.Context, lessonID int64) (*SportOtherSelection, error) {
	return call[*SportOtherSelection](ctx, s.c, get("api/sport/sign/schedule/lessons/"+id(lessonID)+"/other", nil))
}

// SignIn signs the user in to free-visit lessons and returns the IDs of the
// records added. Refusals come as error code SportErrorCannotSignIn.
// POST /api/sport/sign/schedule/lessons
func (s *SportService) SignIn(ctx context.Context, lessonIDs ...int64) ([]int64, error) {
	return call[[]int64](ctx, s.c, post("api/sport/sign/schedule/lessons", sportIDs(lessonIDs)))
}

// SignOut signs the user out of lessons and returns the IDs signed out.
// Refusals come as error code SportErrorCannotSignOut.
// DELETE /api/sport/sign/schedule/lessons
func (s *SportService) SignOut(ctx context.Context, lessonIDs ...int64) ([]int64, error) {
	return call[[]int64](ctx, s.c, del("api/sport/sign/schedule/lessons", sportIDs(lessonIDs)))
}

// SignInGroup signs the user in to a whole lesson group of a section. form is
// required for an open class of a selective section (SportLesson.TypeID ==
// SportLessonTypeOpen with LessonLevel != SportLessonLevelOpen) and nil otherwise.
// POST /api/sport/sign/schedule/lesson_groups/{lesson_group_id}
func (s *SportService) SignInGroup(ctx context.Context, lessonGroupID int64, form *SportOpenFormSubmit) error {
	var body any
	if form != nil {
		body = form
	}
	return exec(ctx, s.c, post("api/sport/sign/schedule/lesson_groups/"+id(lessonGroupID), body))
}

// SignOutGroup signs the user out of a lesson group.
// DELETE /api/sport/sign/schedule/lesson_groups/{lesson_group_id}
func (s *SportService) SignOutGroup(ctx context.Context, lessonGroupID int64) error {
	return exec(ctx, s.c, del("api/sport/sign/schedule/lesson_groups/"+id(lessonGroupID), nil))
}

// Chosen returns the sections, lesson groups and regular lessons the user chose.
// GET /api/sport/sign/chosen
func (s *SportService) Chosen(ctx context.Context) ([]ChosenSportSection, error) {
	return call[[]ChosenSportSection](ctx, s.c, get("api/sport/sign/chosen", nil))
}

// Score returns the points of a sport semester (SportSemesterOption.ID) and their history.
// GET /api/sport/personal/score
func (s *SportService) Score(ctx context.Context, semesterID int64) (*SportScore, error) {
	return call[*SportScore](ctx, s.c, get("api/sport/personal/score", q().set("semester_id", semesterID)))
}

// Calendar returns the user's sport lessons in a date range, past ones included.
// GET /api/sport/personal/calendar
func (s *SportService) Calendar(ctx context.Context, from, to Date) ([]SportScheduleDay, error) {
	return call[[]SportScheduleDay](ctx, s.c, get("api/sport/personal/calendar", q().set("date_start", from).set("date_end", to)))
}

// SignAttempts returns the older free sign-up counter; prefer [SportService.Attempts].
// GET /api/sport/personal/sign_attempts
func (s *SportService) SignAttempts(ctx context.Context) (int, error) {
	return call[int](ctx, s.c, get("api/sport/personal/sign_attempts", nil))
}

// Attempts returns the used and remaining sign-up attempts.
// GET /api/sport/personal/have_attempts
func (s *SportService) Attempts(ctx context.Context) (*SportAttempts, error) {
	return call[*SportAttempts](ctx, s.c, get("api/sport/personal/have_attempts", nil))
}

// Debt returns the physical-education debt.
// GET /api/sport/personal/debt
func (s *SportService) Debt(ctx context.Context) (*SportDebt, error) {
	return call[*SportDebt](ctx, s.c, get("api/sport/personal/debt", nil))
}

// Externat returns the state of the externat application.
// GET /api/sport/personal/externat
func (s *SportService) Externat(ctx context.Context) (*SportExternat, error) {
	return call[*SportExternat](ctx, s.c, get("api/sport/personal/externat", nil))
}

// HealthLevel returns the user's medical health group, or nil when none is assigned.
// GET /api/sport/personal/health_level
func (s *SportService) HealthLevel(ctx context.Context) (*SportHealthLevel, error) {
	r, err := call[struct {
		HealthLevel *SportHealthLevel `json:"health_level"`
	}](ctx, s.c, get("api/sport/personal/health_level", nil))
	return r.HealthLevel, err
}

// Selections returns the passed sport selections and their levels.
// GET /api/sport/personal/selections
func (s *SportService) Selections(ctx context.Context) ([]SportSelection, error) {
	return call[[]SportSelection](ctx, s.c, get("api/sport/personal/selections", nil))
}

// OpenForm returns the open-class questionnaire filled earlier for a section, or nil.
// GET /api/sport/personal/open_form
func (s *SportService) OpenForm(ctx context.Context, sectionID int64) (*SportOpenForm, error) {
	return call[*SportOpenForm](ctx, s.c, get("api/sport/personal/open_form", q().set("section", sectionID)))
}

// OpenFormRanks returns the sport ranks offered in the open-class questionnaire.
// GET /api/sport/personal/open_form/ranks
func (s *SportService) OpenFormRanks(ctx context.Context) ([]IDValue, error) {
	return call[[]IDValue](ctx, s.c, get("api/sport/personal/open_form/ranks", nil))
}

// Briefings returns whether the safety briefing is signed.
// GET /api/sport/personal/briefing/list
func (s *SportService) Briefings(ctx context.Context) (*SportBriefingList, error) {
	return call[*SportBriefingList](ctx, s.c, get("api/sport/personal/briefing/list", nil))
}

// CreateBriefing creates the safety briefing document for signing.
// POST /api/sport/personal/briefing
func (s *SportService) CreateBriefing(ctx context.Context) (*SportBriefingSignTask, error) {
	return call[*SportBriefingSignTask](ctx, s.c, post("api/sport/personal/briefing", nil))
}

// SignBriefing signs a briefing; the route may be obsolete.
// The result shape is unknown.
// POST /api/sport/briefing/{briefing_id}/sign
func (s *SportService) SignBriefing(ctx context.Context, briefingID int64) (RawJSON, error) {
	return call[RawJSON](ctx, s.c, post("api/sport/briefing/"+id(briefingID)+"/sign", nil))
}

// DownloadBriefing downloads the signed briefing (PDF). The caller closes the file.
// GET /api/sport/personal/briefing/{briefing_id}/signed
func (s *SportService) DownloadBriefing(ctx context.Context, briefingID int64) (*File, error) {
	return download(ctx, s.c, get("api/sport/personal/briefing/"+id(briefingID)+"/signed", nil))
}

// Projects returns the special projects (online formats of the credit).
// GET /api/sport/projects/list
func (s *SportService) Projects(ctx context.Context) ([]SportProject, error) {
	return call[[]SportProject](ctx, s.c, get("api/sport/projects/list", nil))
}

// SignInProject signs the user in to a special project. link is the user's
// http(s) URL; "" sends null (SportProjectKronbarsRunning needs none).
// POST /api/sport/sign/projects/{project_id}
func (s *SportService) SignInProject(ctx context.Context, projectID int64, link string) error {
	body := struct {
		Link *string `json:"link"`
	}{}
	if link != "" {
		body.Link = &link
	}
	return exec(ctx, s.c, post("api/sport/sign/projects/"+id(projectID), body))
}

// SignOutProject signs the user out of a special project.
// DELETE /api/sport/sign/projects/{project_id}
func (s *SportService) SignOutProject(ctx context.Context, projectID int64) error {
	return exec(ctx, s.c, del("api/sport/sign/projects/"+id(projectID), nil))
}

// Competitions returns the competitions, optionally of one sport type ([SportService.SportTypes]).
// GET /api/sport/competitions/list
func (s *SportService) Competitions(ctx context.Context, sportTypeID *int64) ([]SportCompetition, error) {
	return call[[]SportCompetition](ctx, s.c, get("api/sport/competitions/list", q().set("sport_type_id", sportTypeID)))
}

// CompetitionLimits returns the places of competitions keyed by competition ID.
// GET /api/sport/competitions/list/limits
func (s *SportService) CompetitionLimits(ctx context.Context) (map[int64]SportSignLimit, error) {
	return sportCallMap[SportSignLimit](ctx, s.c, get("api/sport/competitions/list/limits", nil))
}

// SignInCompetition signs the user in to the chosen disciplines of a competition.
// POST /api/sport/sign/competitions/{competition_id}
func (s *SportService) SignInCompetition(ctx context.Context, competitionID int64, disciplineIDs ...int64) error {
	return exec(ctx, s.c, post("api/sport/sign/competitions/"+id(competitionID), sportIDs(disciplineIDs)))
}

// SignOutCompetition signs the user out of a competition.
// DELETE /api/sport/sign/competitions/{competition_id}
func (s *SportService) SignOutCompetition(ctx context.Context, competitionID int64) error {
	return exec(ctx, s.c, del("api/sport/sign/competitions/"+id(competitionID), nil))
}

// sportIDs makes sure an empty list is sent as [] rather than null.
func sportIDs(v []int64) []int64 {
	if v == nil {
		return []int64{}
	}
	return v
}

// sportCallMap decodes a result keyed by numeric IDs; an empty map may come as [].
func sportCallMap[V any](ctx context.Context, c *Client, r *rest.Request) (map[int64]V, error) {
	raw, err := call[RawJSON](ctx, c, r)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || sportIsJSONArray(raw) {
		return map[int64]V{}, nil
	}
	var out map[int64]V
	if err := jsonx.Unmarshal(raw, &out); err != nil {
		return nil, &DecodeError{Method: r.Method, Path: r.Path, Err: err}
	}
	return out, nil
}

func sportIsJSONArray(b []byte) bool {
	for _, c := range b {
		switch c {
		case ' ', '\t', '\r', '\n':
			continue
		}
		return c == '['
	}
	return false
}
