package myitmo

import (
	"encoding/json/jsontext"
	"fmt"

	"github.com/kewldan/go-itmo/internal/jsonx"
)

// GIARole is a role key of the GIA staff section: the value of the show_as
// filter and the approver key of [GIAService.ApproveStage].
type GIARole = string

// GIA role keys, derived from the flags of [GIAUserStatus].
const (
	GIARoleAdmin                   GIARole = "admin"
	GIARoleOSOP                    GIARole = "osop"
	GIARoleGeneralStateCoordinator GIARole = "general_state_coordinator"
	GIARoleEPManager               GIARole = "ep_manager"
	GIARoleSecretary               GIARole = "secretary"
	GIARoleSupervisor              GIARole = "supervisor"
	GIARoleFacultyManager          GIARole = "faculty_manager"
	GIARoleDeputyFacultyManager    GIARole = "deputy_faculty_manager"
)

// GIAStage is a stage of the graduation thesis (ВКР) workflow.
type GIAStage = string

// Thesis workflow stages.
const (
	GIAStageApplication GIAStage = "application"
	GIAStageTask        GIAStage = "task"
	GIAStageAnnotation  GIAStage = "annotation"
	GIAStageFile        GIAStage = "file"
)

// GIAFormat is the document format of the defense-day documents.
type GIAFormat = string

// Document formats.
const (
	GIAFormatDOCX GIAFormat = "docx"
	GIAFormatPDF  GIAFormat = "pdf"
)

// GIADocType selects the diploma or its supplement in the print section.
type GIADocType = string

// Print document types.
const (
	GIADocTypeDiploma    GIADocType = "diploma"
	GIADocTypeAttachment GIADocType = "attachment"
)

// GIAUserStatus holds the role flags of the current user in the GIA staff
// section; they drive every staff permission.
type GIAUserStatus struct {
	IsAdmin bool `json:"is_admin"`
	IsRoot  bool `json:"is_root"`
	// IsOSOP is the graduation office (ОСОП); approves stage status 2.
	IsOSOP bool `json:"is_osop"`
	// IsGeneralStateCoordinator is an ОГЭК coordinator.
	IsGeneralStateCoordinator bool `json:"is_general_state_coordinator"`
	IsGeneralStateSecretary   bool `json:"is_general_state_secretary"`
	// IsSupervisor is a thesis supervisor; approves stage status 8.
	IsSupervisor bool `json:"is_supervisor"`
	// IsEPManager is the head of an educational programme; approves status 9.
	IsEPManager bool `json:"is_ep_manager"`
	// IsFacultyManager and IsDeputyFacultyManager approve status 11.
	IsFacultyManager       bool `json:"is_faculty_manager"`
	IsDeputyFacultyManager bool `json:"is_deputy_faculty_manager"`
	// IsSecretary is a ГЭК secretary; approves status 10.
	IsSecretary bool `json:"is_secretary"`
	IsReviewer  bool `json:"is_reviewer"`
	IsManager   bool `json:"is_manager"`
}

// GIAIDName is a dictionary entry labelled by name.
type GIAIDName struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// GIANamedOption is a filter option labelled by name or value, depending on the endpoint.
type GIANamedOption struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// GIARefItem is a reference-book entry.
type GIARefItem struct {
	ID     int64  `json:"id"`
	NameRU string `json:"name_ru"`
}

// GIAYearOption is an academic year of the year filter.
type GIAYearOption struct {
	ID   int64 `json:"id"`
	Year int   `json:"year"`
	// Value is usually "2024/2025"; some pages read it as a number.
	Value RawJSON `json:"value"`
	Name  string  `json:"name"`
}

// GIAYearFilter is a year of the list filters.
type GIAYearFilter struct {
	Year int `json:"year"`
}

// GIAStatus is a status as the lists and dictionaries return it.
type GIAStatus struct {
	StatusID   int    `json:"status_id"`
	StatusName string `json:"status_name"`
	// StatusColor and Color are hex colours of the badge; either may be empty.
	StatusColor string `json:"status_color"`
	Color       string `json:"color"`
	Comment     string `json:"comment"`
}

// GIAGroupOption is a study group of a filter.
type GIAGroupOption struct {
	GroupID string `json:"group_id"`
}

// GIARoleOption is a role of the show_as selector. Defense-day filters send
// the role in ID instead of Key.
type GIARoleOption struct {
	Key  GIARole `json:"key"`
	ID   RawJSON `json:"id,omitzero"`
	Name string  `json:"name"`
}

// GIADateOption is a defense date of a filter.
type GIADateOption struct {
	Value string `json:"value"`
	Name  string `json:"name"`
}

// GIAApprovedOption is an entry of the compressed status filter.
type GIAApprovedOption struct {
	Approved bool   `json:"approved"`
	Name     string `json:"name"`
}

// GIAGrade is a mark: a criterion grade or a final assessment grade.
type GIAGrade struct {
	ID     int64  `json:"id"`
	NameRU string `json:"name_ru"`
	Number int    `json:"number"`
}

// GIAFileRef is a stored thesis file, downloadable with [GIAService.DiplomaFile].
type GIAFileRef struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	Extension string `json:"extension"`
}

// GIAEduDirection is a field of study (направление подготовки).
type GIAEduDirection struct {
	DirID   int64  `json:"dir_id"`
	DirCode string `json:"dir_code"`
	DirName string `json:"dir_name"`
}

// GIAEduProgram is an educational programme.
type GIAEduProgram struct {
	EPID             int64  `json:"ep_id"`
	EPName           string `json:"ep_name"`
	EPEnrollmentYear int    `json:"ep_enrollment_year"`
}

// GIADirectionOption is a direction of the direction filter.
type GIADirectionOption struct {
	ID    int64  `json:"id"`
	Value string `json:"value"`
	// DirCode and DirName are read by the ОГЭК list; may be empty.
	DirCode string `json:"dir_code"`
	DirName string `json:"dir_name"`
}

// GIAPersonOption is a person of a filter or search; pages track it by ID or ISU.
type GIAPersonOption struct {
	ID    int64  `json:"id"`
	ISU   int64  `json:"isu"`
	Value string `json:"value"`
}

// GIAGroupValue is a study group of the group filter. The server sends
// {"value": "..."} objects; plain strings are accepted too.
type GIAGroupValue struct {
	Value string `json:"value"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *GIAGroupValue) UnmarshalJSON(b []byte) error {
	v := jsontext.Value(b)
	switch v.Kind() {
	case '"':
		return jsonx.Unmarshal(b, &g.Value)
	case 'n':
		*g = GIAGroupValue{}
		return nil
	case '{':
		var w struct {
			Value string `json:"value"`
		}
		if err := jsonx.Unmarshal(b, &w); err != nil {
			return err
		}
		g.Value = w.Value
		return nil
	}
	return fmt.Errorf("myitmo: unexpected group filter entry %s", b)
}

// GIAPerson is a person block of the thesis card (co-supervisor, consultant).
type GIAPerson struct {
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	SecondName string `json:"second_name"`
	JobTitle   string `json:"job_title"`
	WorkPlace  string `json:"work_place"`
}
