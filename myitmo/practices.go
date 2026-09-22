package myitmo

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"net/http"
	"strconv"
)

// PracticesService is internships and practice reports (/api/practices).
type PracticesService struct{ c *Client }

// PracticePlan is an educational plan whose practices are listed together.
type PracticePlan struct {
	// ID is sent as op_id by [PracticesService.List].
	ID            int64  `json:"id"`
	OpName        string `json:"opName"`
	IsCurrentPlan bool   `json:"isCurrentPlan"`
}

// PracticeYear is the practices of one year of study.
type PracticeYear struct {
	// Year is shown as is; the server may send a number or a string.
	Year      RawJSON        `json:"year"`
	Practices []PracticeCard `json:"practices"`
}

// PracticeCard is one practice in the list.
type PracticeCard struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	IsCurrent bool   `json:"isCurrent"`
	// Mark is the grade text; nil while the practice is not graded.
	Mark *string `json:"mark"`
	// BaigColor is the badge colour: "Yellow", "Green", "Red".
	BaigColor string `json:"baigColor"`
	HeadName  string `json:"headName"`
	DateFrom  Date   `json:"dateFrom"`
	DateTo    Date   `json:"dateTo"`
}

// Thresholds of PracticeStatus.StatusOrder that unlock the tabs of a practice.
const (
	// PracticeOrderTask and above open the individual assignment.
	PracticeOrderTask = 2
	// PracticeOrderReport and above open the report.
	PracticeOrderReport = 4
	// PracticeOrderReview and above open the supervisor feedback.
	PracticeOrderReview = 6
	// PracticeOrderDone means every step is complete.
	PracticeOrderDone = 7
)

// PracticeStatus is the progress of a practice.
type PracticeStatus struct {
	// StatusOrder is 1..7; see the PracticeOrder constants.
	StatusOrder int `json:"statusOrder"`
}

// Practice is the details card of a practice.
type Practice struct {
	ID       int64          `json:"id"`
	Name     string         `json:"name"`
	Status   PracticeStatus `json:"status"`
	DateFrom Date           `json:"dateFrom"`
	DateTo   Date           `json:"dateTo"`
	// PlaceID is nil, 0 or 1 when no application document is needed.
	PlaceID        *int64 `json:"placeId"`
	PlaceName      string `json:"placeName"`
	DepartmentName string `json:"departmentName"`
	Position       string `json:"position"`
	Format         string `json:"format"`
	// CuratorID is the curator's ISU number.
	CuratorID    int64  `json:"curatorId"`
	CuratorName  string `json:"curatorName"`
	CuratorEmail string `json:"curatorEmail"`
	// CuratorPhoto is an image URL; append "cover/90/90/" for a thumbnail.
	CuratorPhoto string `json:"curatorPhoto"`
	HeadItmoName string `json:"headItmoName"`
	// HeadItmoIsu is spelled with a capital H on the wire.
	HeadItmoIsu       int64  `json:"HeadItmoIsu"`
	HeadItmoEmail     string `json:"headItmoEmail"`
	HeadItmoPhoto     string `json:"headItmoPhoto"`
	HeadExternalName  string `json:"headExternalName"`
	HeadExternalEmail string `json:"headExternalEmail"`
}

// Values of PracticeTask.StatusID.
const (
	// PracticeTaskApproved is an approved assignment.
	PracticeTaskApproved = 105
	// PracticeTaskOnApproval is an assignment awaiting approval; it can be recalled.
	PracticeTaskOnApproval = 107
	// PracticeTaskDraft is the only editable state: topic and stages can change
	// and the assignment can be sent for approval.
	PracticeTaskDraft = 108
	// PracticeTaskRejected was rejected by the external supervisor; see PracticeTask.Comment.
	PracticeTaskRejected = 109
	// PracticeTaskOnApprovalAlt is another awaiting-approval state that can be
	// recalled; how it differs from 107 is unknown.
	PracticeTaskOnApprovalAlt = 316
)

// PracticeTask is the individual assignment of a practice.
type PracticeTask struct {
	Topic string `json:"topic"`
	// Status is the localised label of StatusID.
	Status string `json:"status"`
	// StatusID is one of the PracticeTask constants.
	StatusID int `json:"statusId"`
	// Comment is the rejection reason.
	Comment string              `json:"comment"`
	Stages  []PracticeTaskStage `json:"stages"`
}

// PracticeTaskStage is one stage of the assignment work plan. A stage has
// either Duration or DateFrom and DateTo.
type PracticeTaskStage struct {
	StageID     int64  `json:"stageId"`
	StageNumber int    `json:"stageNumber"`
	StageName   string `json:"stageName"`
	StageTask   string `json:"stageTask"`
	DateFrom    Date   `json:"dateFrom"`
	DateTo      Date   `json:"dateTo"`
	// Duration is the length in days.
	Duration *int `json:"duration"`
}

// PracticeStageInput describes a stage to add or edit. Set either Duration
// or DateFrom and DateTo.
type PracticeStageInput struct {
	PracticeID int64  `json:"practiceId"`
	StageName  string `json:"stageName"`
	StageTask  string `json:"stageTask"`
	// Duration is the length in days.
	Duration int  `json:"duration,omitzero"`
	DateFrom Date `json:"dateFrom,omitzero"`
	DateTo   Date `json:"dateTo,omitzero"`
}

// PracticeReportState is the report tab of a practice.
type PracticeReportState struct {
	// Report is nil until a report is uploaded.
	Report *PracticeReport `json:"report"`
	// ShowFeedBack means the upload must also carry EmploymentStatID and PlaceGradeID.
	ShowFeedBack bool `json:"showFeedBack"`
}

// PracticeReport is an uploaded practice report.
type PracticeReport struct {
	// FileID is a number or a string; pass it back as PracticeReportUpload.FileID.
	FileID      RawJSON                 `json:"fileId"`
	FileName    string                  `json:"fileName"`
	ReportTypes []IDValue               `json:"reportTypes"`
	FeedBack    *PracticeReportFeedback `json:"feedBack"`
}

// PracticeReportFeedback is the student's answers about the practice place.
type PracticeReportFeedback struct {
	PlaceGradeID     int64  `json:"placeGradeId"`
	EmploymentStatID int64  `json:"employmentStatId"`
	PlaceGrade       string `json:"placeGrade"`
	EmploymentStat   string `json:"employmentStat"`
}

// PracticeReportUpload is the form of [PracticesService.CreateReport] and
// [PracticesService.UpdateReport].
type PracticeReportUpload struct {
	PracticeID int64
	// File is the report (one archive for several files); its field name is set to "file".
	// On update it may be nil to keep the file identified by FileID.
	File *Upload
	// FileID keeps the uploaded file on update when File is nil.
	FileID string
	// FileName defaults to File.Name.
	FileName string
	// ReportTypes are chosen from [PracticesService.ReportTypes].
	ReportTypes []IDValue
	// EmploymentStatID and PlaceGradeID are required when
	// PracticeReportState.ShowFeedBack is true.
	EmploymentStatID *int64
	PlaceGradeID     *int64
}

func (u PracticeReportUpload) form() (map[string]string, []Upload, error) {
	types := u.ReportTypes
	if types == nil {
		types = []IDValue{}
	}
	rt, err := json.Marshal(types)
	if err != nil {
		return nil, nil, fmt.Errorf("myitmo: encode reportTypes: %w", err)
	}
	fields := map[string]string{
		"reportTypes": string(rt),
		"practiceId":  strconv.FormatInt(u.PracticeID, 10),
	}
	name := u.FileName
	var files []Upload
	if u.File != nil {
		f := *u.File
		f.Field = "file"
		files = append(files, f)
		if name == "" {
			name = f.Name
		}
	} else if u.FileID != "" {
		fields["fileId"] = u.FileID
	}
	if name != "" {
		fields["fileName"] = name
	}
	if u.EmploymentStatID != nil {
		fields["employmentStatId"] = strconv.FormatInt(*u.EmploymentStatID, 10)
	}
	if u.PlaceGradeID != nil {
		fields["placeGradeId"] = strconv.FormatInt(*u.PlaceGradeID, 10)
	}
	return fields, files, nil
}

// PracticeFeedback is the supervisors' feedback on a practice.
type PracticeFeedback struct {
	InternalHead *PracticeHeadFeedback `json:"internalHead"`
	ExternalHead *PracticeHeadFeedback `json:"externalHead"`
}

// PracticeHeadFeedback is the feedback of one supervisor.
type PracticeHeadFeedback struct {
	FeedBack string `json:"feedBack"`
	// FeedBackGrade is the recommended grade, a string or a number.
	FeedBackGrade RawJSON `json:"feedBackGrade"`
	// FeedBackFileName is set when the external supervisor attached a file;
	// download it with [PracticesService.FeedbackFile].
	FeedBackFileName string                      `json:"feedBackFileName"`
	Stages           []PracticeHeadFeedbackStage `json:"stages"`
}

// PracticeHeadFeedbackStage is the supervisor's verdict on one stage.
type PracticeHeadFeedbackStage struct {
	ID           int64  `json:"id"`
	StageNumber  int    `json:"stageNumber"`
	StageName    string `json:"stageName"`
	StageResult  string `json:"stageResult"`
	StageComment string `json:"stageComment"`
}

type practiceRef struct {
	PracticeID int64 `json:"practiceId"`
}

type practiceTopic struct {
	PracticeID int64  `json:"practiceId"`
	Topic      string `json:"topic"`
}

type practiceStage struct {
	StageID            int64 `json:"stageId"`
	PracticeStageInput `json:",inline"`
}

type practiceStageRef struct {
	PracticeID int64 `json:"practiceId"`
	StageID    int64 `json:"stageId"`
}

// Plans returns the student's educational plans that have practices.
// GET /api/practices/practices/op
func (s *PracticesService) Plans(ctx context.Context) ([]PracticePlan, error) {
	return call[[]PracticePlan](ctx, s.c, get("api/practices/practices/op", nil))
}

// List returns the practices of a plan grouped by year; nil plan lets the server choose.
// GET /api/practices/practices/all
func (s *PracticesService) List(ctx context.Context, plan *PracticePlan) ([]PracticeYear, error) {
	qv := q()
	if plan != nil {
		qv.set("op_id", plan.ID).set("current", plan.IsCurrentPlan)
	}
	return call[[]PracticeYear](ctx, s.c, get("api/practices/practices/all", qv))
}

// Get returns the details of a practice.
// GET /api/practices/practices?id={practiceId}
func (s *PracticesService) Get(ctx context.Context, practiceID int64) (*Practice, error) {
	return call[*Practice](ctx, s.c, get("api/practices/practices", q().set("id", practiceID)))
}

// Application generates the practice application document (.docx).
// POST /api/practices/practices/{practiceId}/application
func (s *PracticesService) Application(ctx context.Context, practiceID int64) (*File, error) {
	return download(ctx, s.c, post("api/practices/practices/"+id(practiceID)+"/application", nil))
}

// Task returns the individual assignment; nil when none has been created yet.
// GET /api/practices/ind_task?id={practiceId}
func (s *PracticesService) Task(ctx context.Context, practiceID int64) (*PracticeTask, error) {
	return call[*PracticeTask](ctx, s.c, get("api/practices/ind_task", q().set("id", practiceID)))
}

// CreateTask creates the individual assignment with a topic.
// POST /api/practices/ind_task/create
func (s *PracticesService) CreateTask(ctx context.Context, practiceID int64, topic string) error {
	return exec(ctx, s.c, post("api/practices/ind_task/create", practiceTopic{practiceID, topic}))
}

// UpdateTopic changes the topic of the individual assignment.
// PUT /api/practices/ind_task/create
func (s *PracticesService) UpdateTopic(ctx context.Context, practiceID int64, topic string) error {
	return exec(ctx, s.c, put("api/practices/ind_task/create", practiceTopic{practiceID, topic}))
}

// AddStage adds a stage to the assignment work plan.
// POST /api/practices/ind_task/stage/add
func (s *PracticesService) AddStage(ctx context.Context, stage PracticeStageInput) error {
	return exec(ctx, s.c, post("api/practices/ind_task/stage/add", stage))
}

// UpdateStage edits a stage of the work plan.
// PUT /api/practices/ind_task/stage/update
func (s *PracticesService) UpdateStage(ctx context.Context, stageID int64, stage PracticeStageInput) error {
	return exec(ctx, s.c, put("api/practices/ind_task/stage/update", practiceStage{stageID, stage}))
}

// DeleteStage removes a stage from the work plan.
// DELETE /api/practices/ind_task/stage/delete
func (s *PracticesService) DeleteStage(ctx context.Context, practiceID, stageID int64) error {
	return exec(ctx, s.c, del("api/practices/ind_task/stage/delete", practiceStageRef{practiceID, stageID}))
}

// SubmitTask sends the assignment for approval (only from PracticeTaskDraft).
// POST /api/practices/ind_task/approve
func (s *PracticesService) SubmitTask(ctx context.Context, practiceID int64) error {
	return exec(ctx, s.c, post("api/practices/ind_task/approve", practiceRef{practiceID}))
}

// CancelTask recalls the assignment from approval or withdraws it.
// DELETE /api/practices/ind_task/cancel
func (s *PracticesService) CancelTask(ctx context.Context, practiceID int64) error {
	return exec(ctx, s.c, del("api/practices/ind_task/cancel", practiceRef{practiceID}))
}

// Report returns the uploaded report and whether feedback answers are required.
// GET /api/practices/report?id={practiceId}
func (s *PracticesService) Report(ctx context.Context, practiceID int64) (*PracticeReportState, error) {
	return call[*PracticeReportState](ctx, s.c, get("api/practices/report", q().set("id", practiceID)))
}

// ReportTypes returns the forms a report can take.
// GET /api/practices/report/types
func (s *PracticesService) ReportTypes(ctx context.Context) ([]IDValue, error) {
	return call[[]IDValue](ctx, s.c, get("api/practices/report/types", nil))
}

// PlaceGrades returns the grades for rating the practice place.
// GET /api/practices/report/grades
func (s *PracticesService) PlaceGrades(ctx context.Context) ([]IDValue, error) {
	return call[[]IDValue](ctx, s.c, get("api/practices/report/grades", nil))
}

// EmploymentStats returns the employment statuses a student can report.
// GET /api/practices/report/emp_stats
func (s *PracticesService) EmploymentStats(ctx context.Context) ([]IDValue, error) {
	return call[[]IDValue](ctx, s.c, get("api/practices/report/emp_stats", nil))
}

// CreateReport uploads the practice report. Call [PracticesService.SubmitReport] afterwards.
// POST /api/practices/report/create
func (s *PracticesService) CreateReport(ctx context.Context, report PracticeReportUpload) error {
	fields, files, err := report.form()
	if err != nil {
		return err
	}
	return exec(ctx, s.c, multipart(http.MethodPost, "api/practices/report/create", fields, files...))
}

// UpdateReport replaces the report file or its details. Call [PracticesService.SubmitReport] afterwards.
// PUT /api/practices/report/update
func (s *PracticesService) UpdateReport(ctx context.Context, report PracticeReportUpload) error {
	fields, files, err := report.form()
	if err != nil {
		return err
	}
	return exec(ctx, s.c, multipart(http.MethodPut, "api/practices/report/update", fields, files...))
}

// SubmitReport sends the uploaded report for review.
// POST /api/practices/report/approve
func (s *PracticesService) SubmitReport(ctx context.Context, practiceID int64) error {
	return exec(ctx, s.c, post("api/practices/report/approve", practiceRef{practiceID}))
}

// ReportFile downloads the uploaded report.
// GET /api/practices/report/file?id={practiceId}
func (s *PracticesService) ReportFile(ctx context.Context, practiceID int64) (*File, error) {
	return download(ctx, s.c, get("api/practices/report/file", q().set("id", practiceID)))
}

// Feedback returns the feedback of the ITMO and external supervisors.
// GET /api/practices/feedback/head?id={practiceId}
func (s *PracticesService) Feedback(ctx context.Context, practiceID int64) (*PracticeFeedback, error) {
	return call[*PracticeFeedback](ctx, s.c, get("api/practices/feedback/head", q().set("id", practiceID)))
}

// FeedbackFile downloads the external supervisor's feedback file.
// GET /api/practices/feedback/file?id={practiceId}
func (s *PracticesService) FeedbackFile(ctx context.Context, practiceID int64) (*File, error) {
	return download(ctx, s.c, get("api/practices/feedback/file", q().set("id", practiceID)))
}
