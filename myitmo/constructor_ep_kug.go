package myitmo

import (
	"context"
	"fmt"
	"time"
)

// Calendar study schedule (КУГ) activity type IDs, in the order of
// [ConstructorEPService.ActivityTypes]. Production
// calendar days use 1 and 8 as well.
const (
	EPActivityDayOff              = 1  // Выходной день
	EPActivityGIA                 = 2  // ГИА
	EPActivityVacation            = 3  // Каникулы
	EPActivityMobility            = 4  // Мобильность
	EPActivityResearch            = 5  // НИР
	EPActivityPreGraduatePractice = 6  // Преддипломная практика
	EPActivityIndustrialPractice  = 7  // Производственная практика
	EPActivityPublicHoliday       = 8  // Праздничный день
	EPActivityTheory              = 9  // Теоретическое обучение
	EPActivityEducationalPractice = 10 // Учебная практика
	EPActivityExamSession         = 11 // Экзаменационная сессия
	EPActivityNewPractice         = 12 // Новая практика
)

// Semester orders within an academic year.
const (
	EPSemesterAutumn = 1
	EPSemesterSpring = 2
)

// EPKUGSemesterID encodes a semester ID as the server does: year, year+1 and
// order concatenated (2024 autumn is 202420251).
func EPKUGSemesterID(year, order int) int64 {
	return int64(year)*100000 + int64(year+1)*10 + int64(order)
}

// EPKUGTemplateListItem is a KUG template of the list.
type EPKUGTemplateListItem struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	YearStart    int    `json:"year_start"`
	StandardName string `json:"standard_name"`
	UsedInEP     bool   `json:"used_in_ep"`
}

// EPKUGTemplateList is a page of KUG templates.
type EPKUGTemplateList struct {
	KUGTemplates []EPKUGTemplateListItem `json:"kug_templates"`
	Count        int                     `json:"count"`
}

// EPKUGTemplate is a KUG template's main information.
type EPKUGTemplate struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	StandardID int64  `json:"standard_id"`
	YearStart  int    `json:"year_start"`
	UsedInEP   bool   `json:"used_in_ep"`
}

// EPKUGTemplateForm creates a KUG template.
type EPKUGTemplateForm struct {
	Name       string `json:"name"`
	YearStart  int    `json:"year_start"`
	StandardID int64  `json:"standard_id"`
}

// EPKUGTemplateListParams filters [ConstructorEPService.KUGTemplates].
type EPKUGTemplateListParams struct {
	Query string
	// StandardID and YearStart are omitted when 0.
	StandardID int64
	YearStart  int
	UsedInEP   *bool
	// Personal lists only the user's templates.
	Personal bool
	// Valid lists only valid templates (used when picking one for a programme).
	Valid *bool
	// Limit is the page size (typically 20).
	Limit  int
	Offset int
}

// EPKUGTemplateUsage is a page of programmes using a KUG template.
type EPKUGTemplateUsage struct {
	Count    int                         `json:"count"`
	Programs []EPKUGTemplateUsageProgram `json:"programs"`
}

// EPKUGTemplateUsageProgram is a programme using a KUG template.
type EPKUGTemplateUsageProgram struct {
	ID              int64  `json:"id"`
	EPID            int64  `json:"ep_id"`
	EPName          string `json:"ep_name"`
	ImplementerName string `json:"implementer_name"`
}

// EPKUG is the KUG attached to a programme; its ID addresses the /kug/{kug_id} routes.
// ; other fields are unknown.
type EPKUG struct {
	ID        int64 `json:"id"`
	YearStart int   `json:"year_start"`
}

// EPKUGActivity is an activity period of a KUG or KUG template.
type EPKUGActivity struct {
	// Type carries the activity type ID (EPActivity* constants).
	Type      EPRef `json:"type"`
	DateStart Date  `json:"date_start"`
	DateEnd   Date  `json:"date_end"`
}

// EPPeriod is a date range.
type EPPeriod struct {
	DateStart Date `json:"date_start"`
	DateEnd   Date `json:"date_end"`
}

// EPKUGHoursByYear is the number of days per activity type in each semester of
// an academic year, keyed by activity type ID.
type EPKUGHoursByYear struct {
	AutumnSemHours map[string]float64 `json:"autumn_sem_hours"`
	SpringSemHours map[string]float64 `json:"spring_sem_hours"`
}

// EPKUGValidation lists the unfilled periods of a KUG, keyed by academic year.
type EPKUGValidation struct {
	YearsErrors map[string]EPKUGYearErrors `json:"years_errors"`
}

// EPKUGYearErrors are the errors of one academic year, keyed by semester order ("1" autumn, "2" spring).
type EPKUGYearErrors struct {
	SemestersErrors map[string]EPKUGSemesterErrors `json:"semesters_errors"`
}

// EPKUGSemesterErrors are the errors of one semester.
type EPKUGSemesterErrors struct {
	MissedActivitiesError *EPKUGMissedActivities `json:"missed_activities_error"`
}

// EPKUGMissedActivities lists periods without an activity.
type EPKUGMissedActivities struct {
	MissedPeriods []EPPeriod `json:"missed_periods"`
}

// EPKUGComment is an expert comment on a programme KUG.
type EPKUGComment struct {
	ID   int64     `json:"id"`
	Text string    `json:"text"`
	Date time.Time `json:"date"`
	User EPPerson  `json:"user"`
}

// EPKUGSemester is a configured semester.
type EPKUGSemester struct {
	// ID is encoded as by [EPKUGSemesterID]: year = ID/100000, order = ID%10.
	ID        int64 `json:"id"`
	DateStart Date  `json:"date_start"`
	DateEnd   Date  `json:"date_end"`
}

// EPProductionCalendarDay is a special day of the production calendar.
// ; other fields are unknown.
type EPProductionCalendarDay struct {
	Date Date `json:"date"`
	// TypeID is [EPActivityDayOff] (weekend) or [EPActivityPublicHoliday].
	TypeID int64 `json:"type_id"`
}

// EPKUGActivityForm sets an activity period. Dates are sent as DD-MM-YYYY.
// A period crossing the autumn/spring boundary must be split in two calls.
type EPKUGActivityForm struct {
	// TypeID is an EPActivity* constant.
	TypeID int64
	// SemesterID is encoded as by [EPKUGSemesterID].
	SemesterID int64
	DateStart  Date
	DateEnd    Date
}

// EPSemesterDatesForm sets the dates of one semester. Dates are sent as DD-MM-YYYY.
type EPSemesterDatesForm struct {
	// Year is the academic year start.
	Year int
	// Order is [EPSemesterAutumn] or [EPSemesterSpring].
	Order     int
	DateStart Date
	DateEnd   Date
}

func epDMY(d Date) string {
	return fmt.Sprintf("%02d-%02d-%04d", d.Day, d.Month, d.Year)
}

type epKUGActivityBody struct {
	TypeID        int64  `json:"type_id"`
	KUGEPID       int64  `json:"kug_ep_id,omitzero"`
	KUGTemplateID int64  `json:"kug_template_id,omitzero"`
	SemesterID    int64  `json:"semester_id"`
	DateStart     string `json:"date_start"`
	DateEnd       string `json:"date_end"`
}

func epActivityBody(a EPKUGActivityForm, kugID, templateID int64) epKUGActivityBody {
	return epKUGActivityBody{a.TypeID, kugID, templateID, a.SemesterID, epDMY(a.DateStart), epDMY(a.DateEnd)}
}

type epSemesterDatesBody struct {
	Year      int    `json:"year"`
	Order     int    `json:"order"`
	DateStart string `json:"date_start"`
	DateEnd   string `json:"date_end"`
}

func epSemesterBodies(sems []EPSemesterDatesForm) []epSemesterDatesBody {
	out := make([]epSemesterDatesBody, len(sems))
	for i, s := range sems {
		out[i] = epSemesterDatesBody{s.Year, s.Order, epDMY(s.DateStart), epDMY(s.DateEnd)}
	}
	return out
}

func epKUG(parts ...string) string { return epPath(append([]string{"kug"}, parts...)...) }

// KUGTemplates returns a page of KUG templates.
// Staff only.
// GET /api/constructor-ep/kug/templates/list
func (s *ConstructorEPService) KUGTemplates(ctx context.Context, p EPKUGTemplateListParams) (*EPKUGTemplateList, error) {
	v := q().set("query", p.Query)
	epSetID(v, "standard_id", p.StandardID)
	epSetID(v, "year_start", int64(p.YearStart))
	v.set("used_in_ep", p.UsedInEP)
	if p.Personal {
		v.set("personal", 1)
	}
	v.set("valid", p.Valid)
	epSetID(v, "limit", int64(p.Limit))
	v.set("offset", p.Offset)
	return call[*EPKUGTemplateList](ctx, s.c, get(epKUG("templates", "list"), v))
}

// CreateKUGTemplate creates a KUG template and returns its ID. Production
// calendars and semester dates must exist for every year it covers.
// Staff only.
// POST /api/constructor-ep/kug/templates/create
func (s *ConstructorEPService) CreateKUGTemplate(ctx context.Context, form EPKUGTemplateForm) (int64, error) {
	return call[int64](ctx, s.c, post(epKUG("templates", "create"), form))
}

// KUGTemplate returns a KUG template's main information.
// Staff only.
// GET /api/constructor-ep/kug/templates/{template_id}
func (s *ConstructorEPService) KUGTemplate(ctx context.Context, templateID int64) (*EPKUGTemplate, error) {
	return call[*EPKUGTemplate](ctx, s.c, get(epKUG("templates", id(templateID)), nil))
}

// RenameKUGTemplate renames a KUG template.
// Staff only.
// PATCH /api/constructor-ep/kug/templates/update_name/{template_id}
func (s *ConstructorEPService) RenameKUGTemplate(ctx context.Context, templateID int64, name string) error {
	body := struct {
		Name string `json:"name"`
	}{name}
	return exec(ctx, s.c, patch(epKUG("templates", "update_name", id(templateID)), body))
}

// KUGTemplateActivities returns the activity periods of a KUG template.
// Staff only.
// GET /api/constructor-ep/kug/templates/{template_id}/activities
func (s *ConstructorEPService) KUGTemplateActivities(ctx context.Context, templateID int64) ([]EPKUGActivity, error) {
	return call[[]EPKUGActivity](ctx, s.c, get(epKUG("templates", id(templateID), "activities"), nil))
}

// KUGTemplateHoursByYear returns the days per activity type of a KUG template
// for the academic year starting in year.
// Staff only.
// GET /api/constructor-ep/kug/templates/hours_by_year/{template_id}/{year}
func (s *ConstructorEPService) KUGTemplateHoursByYear(ctx context.Context, templateID int64, year int) (*EPKUGHoursByYear, error) {
	return call[*EPKUGHoursByYear](ctx, s.c, get(epKUG("templates", "hours_by_year", id(templateID), id(year)), nil))
}

// KUGTemplateSemesterDates returns the semester dates of a KUG template for a
// year. Shape not confirmed (likely a list of EPPeriod).
// Staff only.
// GET /api/constructor-ep/kug/templates/semesters_dates/{template_id}/{year}
func (s *ConstructorEPService) KUGTemplateSemesterDates(ctx context.Context, templateID int64, year int) (RawJSON, error) {
	return call[RawJSON](ctx, s.c, get(epKUG("templates", "semesters_dates", id(templateID), id(year)), nil))
}

// ValidateKUGTemplate checks a KUG template for unfilled periods; nil means valid.
// Staff only.
// GET /api/constructor-ep/kug/templates/validate/{template_id}
func (s *ConstructorEPService) ValidateKUGTemplate(ctx context.Context, templateID int64) (*EPKUGValidation, error) {
	return call[*EPKUGValidation](ctx, s.c, get(epKUG("templates", "validate", id(templateID)), nil))
}

// KUGTemplateUsage returns a page of the programmes using a KUG template.
// Staff only.
// GET /api/constructor-ep/kug/templates/using/{template_id}
func (s *ConstructorEPService) KUGTemplateUsage(ctx context.Context, templateID int64, query string, limit, offset int) (*EPKUGTemplateUsage, error) {
	v := q().set("query", query).set("limit", limit).set("offset", offset)
	return call[*EPKUGTemplateUsage](ctx, s.c, get(epKUG("templates", "using", id(templateID)), v))
}

// CreateKUGTemplateActivity adds an activity period to a KUG template, keeping existing ones.
// Staff only.
// POST /api/constructor-ep/kug/templates/activities/create
func (s *ConstructorEPService) CreateKUGTemplateActivity(ctx context.Context, templateID int64, a EPKUGActivityForm) error {
	return exec(ctx, s.c, post(epKUG("templates", "activities", "create"), epActivityBody(a, 0, templateID)))
}

// ReplaceKUGTemplateActivity sets an activity period on a KUG template, replacing what is there.
// Staff only.
// POST /api/constructor-ep/kug/templates/activities/replace
func (s *ConstructorEPService) ReplaceKUGTemplateActivity(ctx context.Context, templateID int64, a EPKUGActivityForm) error {
	return exec(ctx, s.c, post(epKUG("templates", "activities", "replace"), epActivityBody(a, 0, templateID)))
}

// ApplyKUGTemplate creates programme KUGs from a template.
// Staff only.
// POST /api/constructor-ep/kug/add_to_ep/{template_id}
func (s *ConstructorEPService) ApplyKUGTemplate(ctx context.Context, templateID int64, programIDs []int64) error {
	body := struct {
		EPIDs []int64 `json:"ep_ids"`
	}{programIDs}
	return exec(ctx, s.c, post(epKUG("add_to_ep", id(templateID)), body))
}

// KUGByProgram returns the KUG attached to a programme, or nil.
// Staff only.
// GET /api/constructor-ep/kug/by_ep/{ep_id}
func (s *ConstructorEPService) KUGByProgram(ctx context.Context, programID int64) (*EPKUG, error) {
	return call[*EPKUG](ctx, s.c, get(epKUG("by_ep", id(programID)), nil))
}

// KUGActivities returns the activity periods of a programme KUG.
// Staff only.
// GET /api/constructor-ep/kug/{kug_id}/activities
func (s *ConstructorEPService) KUGActivities(ctx context.Context, kugID int64) ([]EPKUGActivity, error) {
	return call[[]EPKUGActivity](ctx, s.c, get(epKUG(id(kugID), "activities"), nil))
}

// KUGHoursByYear returns the days per activity type of a programme KUG for
// the academic year starting in year.
// Staff only.
// GET /api/constructor-ep/kug/hours_by_year/{kug_id}/{year}
func (s *ConstructorEPService) KUGHoursByYear(ctx context.Context, kugID int64, year int) (*EPKUGHoursByYear, error) {
	return call[*EPKUGHoursByYear](ctx, s.c, get(epKUG("hours_by_year", id(kugID), id(year)), nil))
}

// KUGSemesterDates returns the semester dates of a programme KUG for a year.
// Shape not confirmed (likely a list of EPPeriod).
// Staff only.
// GET /api/constructor-ep/kug/activities/kug_semesters_dates/{kug_id}/{year}
func (s *ConstructorEPService) KUGSemesterDates(ctx context.Context, kugID int64, year int) (RawJSON, error) {
	return call[RawJSON](ctx, s.c, get(epKUG("activities", "kug_semesters_dates", id(kugID), id(year)), nil))
}

// ValidateKUG checks a programme KUG for unfilled periods; nil means valid.
// Staff only.
// GET /api/constructor-ep/kug/validate/{kug_id}
func (s *ConstructorEPService) ValidateKUG(ctx context.Context, kugID int64) (*EPKUGValidation, error) {
	return call[*EPKUGValidation](ctx, s.c, get(epKUG("validate", id(kugID)), nil))
}

// CreateKUGActivity adds an activity period to a programme KUG, keeping existing ones.
// Staff only.
// POST /api/constructor-ep/kug/activities/create
func (s *ConstructorEPService) CreateKUGActivity(ctx context.Context, kugID int64, a EPKUGActivityForm) error {
	return exec(ctx, s.c, post(epKUG("activities", "create"), epActivityBody(a, kugID, 0)))
}

// ReplaceKUGActivity sets an activity period on a programme KUG, replacing what is there.
// Staff only.
// POST /api/constructor-ep/kug/activities/replace
func (s *ConstructorEPService) ReplaceKUGActivity(ctx context.Context, kugID int64, a EPKUGActivityForm) error {
	return exec(ctx, s.c, post(epKUG("activities", "replace"), epActivityBody(a, kugID, 0)))
}

// KUGComments returns the expert comments on a programme KUG.
// Staff only.
// GET /api/constructor-ep/kug/{kug_id}/comments
func (s *ConstructorEPService) KUGComments(ctx context.Context, kugID int64) ([]EPKUGComment, error) {
	return call[[]EPKUGComment](ctx, s.c, get(epKUG(id(kugID), "comments"), nil))
}

type epText struct {
	Text string `json:"text"`
}

// AddKUGComment adds an expert comment to a programme KUG.
// Staff only.
// POST /api/constructor-ep/kug/{kug_id}/comments/create
func (s *ConstructorEPService) AddKUGComment(ctx context.Context, kugID int64, text string) error {
	return exec(ctx, s.c, post(epKUG(id(kugID), "comments", "create"), epText{text}))
}

// UpdateKUGComment edits a KUG comment.
// Staff only.
// PATCH /api/constructor-ep/kug/comments/{comment_id}/update
func (s *ConstructorEPService) UpdateKUGComment(ctx context.Context, commentID int64, text string) error {
	return exec(ctx, s.c, patch(epKUG("comments", id(commentID), "update"), epText{text}))
}

// DeleteKUGComment deletes a KUG comment.
// Staff only.
// DELETE /api/constructor-ep/kug/comments/{comment_id}/delete
func (s *ConstructorEPService) DeleteKUGComment(ctx context.Context, commentID int64) error {
	return exec(ctx, s.c, del(epKUG("comments", id(commentID), "delete"), nil))
}

// KUGToExpertise sends a programme KUG to expertise.
// Staff only.
// POST /api/constructor-ep/kug/{kug_id}/to_expertise
func (s *ConstructorEPService) KUGToExpertise(ctx context.Context, kugID int64) error {
	return exec(ctx, s.c, post(epKUG(id(kugID), "to_expertise"), nil))
}

// KUGToRevision returns a programme KUG for revision.
// Staff only.
// POST /api/constructor-ep/kug/{kug_id}/to_revision
func (s *ConstructorEPService) KUGToRevision(ctx context.Context, kugID int64) error {
	return exec(ctx, s.c, post(epKUG(id(kugID), "to_revision"), nil))
}

// KUGToSigning sends a programme KUG to signing.
// Staff only.
// POST /api/constructor-ep/kug/{kug_id}/to_signing
func (s *ConstructorEPService) KUGToSigning(ctx context.Context, kugID int64) error {
	return exec(ctx, s.c, post(epKUG(id(kugID), "to_signing"), nil))
}

// KUGDocument generates the KUG document of a programme (the programme ID,
// not the KUG ID); format is [EPFormatPDF] or [EPFormatXLSX].
// Staff only.
// POST /api/constructor-ep/kug/{ep_id}/{format}/generate
func (s *ConstructorEPService) KUGDocument(ctx context.Context, programID int64, format string) (*File, error) {
	return download(ctx, s.c, post(epKUG(id(programID), id(format), "generate"), nil))
}

// KUGOrder generates the KUG order (приказ) PDF for a calendar year.
// Staff only.
// POST /api/constructor-ep/kug/order/{year}/pdf/generate
func (s *ConstructorEPService) KUGOrder(ctx context.Context, year int) (*File, error) {
	return download(ctx, s.c, post(epKUG("order", id(year), "pdf", "generate"), nil))
}

// CreateProductionCalendar imports the production calendar of a year.
// Staff only.
// POST /api/constructor-ep/kug/production_calendar/create/for_year
func (s *ConstructorEPService) CreateProductionCalendar(ctx context.Context, year int) error {
	body := struct {
		Year int `json:"year"`
	}{year}
	return exec(ctx, s.c, post(epKUG("production_calendar", "create", "for_year"), body))
}

// ProductionCalendar returns the weekends and public holidays of a calendar year.
// Staff only.
// GET /api/constructor-ep/kug/production_calendar/list/by_year/{year}
func (s *ConstructorEPService) ProductionCalendar(ctx context.Context, year int) ([]EPProductionCalendarDay, error) {
	return call[[]EPProductionCalendarDay](ctx, s.c, get(epKUG("production_calendar", "list", "by_year", id(year)), nil))
}

// ProductionCalendarYears returns the years that have a production calendar.
// Staff only.
// GET /api/constructor-ep/kug/production_calendar/years/list
func (s *ConstructorEPService) ProductionCalendarYears(ctx context.Context) ([]int, error) {
	return call[[]int](ctx, s.c, get(epKUG("production_calendar", "years", "list"), nil))
}

// AddSemesters adds the autumn and spring semester dates of a new academic year.
// Staff only.
// POST /api/constructor-ep/kug/semesters/add
func (s *ConstructorEPService) AddSemesters(ctx context.Context, semesters []EPSemesterDatesForm) error {
	return exec(ctx, s.c, post(epKUG("semesters", "add"), epSemesterBodies(semesters)))
}

// UpdateSemesters updates the semester dates of an academic year.
// Staff only.
// PATCH /api/constructor-ep/kug/semesters/update
func (s *ConstructorEPService) UpdateSemesters(ctx context.Context, semesters []EPSemesterDatesForm) error {
	return exec(ctx, s.c, patch(epKUG("semesters", "update"), epSemesterBodies(semesters)))
}

// SemestersByYear returns the autumn ([0]) and spring ([1]) dates of the
// academic year starting in year.
// Staff only.
// GET /api/constructor-ep/kug/semesters/by_year/{year}
func (s *ConstructorEPService) SemestersByYear(ctx context.Context, year int) ([]EPPeriod, error) {
	return call[[]EPPeriod](ctx, s.c, get(epKUG("semesters", "by_year", id(year)), nil))
}

// Semesters returns every configured semester.
// Staff only.
// GET /api/constructor-ep/kug/semesters/list
func (s *ConstructorEPService) Semesters(ctx context.Context) ([]EPKUGSemester, error) {
	return call[[]EPKUGSemester](ctx, s.c, get(epKUG("semesters", "list"), nil))
}
