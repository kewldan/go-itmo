package myitmo

import (
	"context"
	"time"
)

// EPBankModuleListItem is a module of the module bank list.
type EPBankModuleListItem struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	EducationLevel *EPRef `json:"education_level"`
	// Capacity is in credits.
	Capacity      float64   `json:"capacity"`
	CreatedAt     time.Time `json:"created_at"`
	RuleType      *EPRef    `json:"rule_type"`
	RuleValue     []float64 `json:"rule_value"`
	StartSemester *int      `json:"start_semester"`
	// Creator carries the name only.
	Creator EPPerson `json:"creator"`
	// Status is [EPStatusInWork], [EPStatusArchived] or [EPStatusSigned] (approved).
	Status EPRef `json:"status"`
}

// EPBankModuleList is a page of the module bank.
type EPBankModuleList struct {
	Modules   []EPBankModuleListItem `json:"modules"`
	Count     int                    `json:"count"`
	CanCreate bool                   `json:"can_create"`
}

// EPBankModule is a module bank card.
type EPBankModule struct {
	ID                  int64      `json:"id"`
	Name                string     `json:"name"`
	NameRU              string     `json:"name_ru"`
	NameEN              string     `json:"name_en"`
	Block               EPRef      `json:"block"`
	EducationalStandard EPStandard `json:"educational_standard"`
	RuleType            *EPRef     `json:"rule_type"`
	RuleValue           []float64  `json:"rule_value"`
	Description         string     `json:"description"`
	// Status is [EPStatusInWork] (editable), [EPStatusArchived] or [EPStatusSigned] (approved).
	Status EPRef `json:"status"`
	// CanEdit is true when the current user is an editor.
	CanEdit bool `json:"canEdit"`
	IsAdmin bool `json:"is_admin"`
	// RootModule is present on the wire; shape unknown.
	RootModule RawJSON    `json:"root_module,omitzero"`
	Editors    []EPPerson `json:"editors"`
	// AccessiblePrograms are the programmes allowed to use the module; empty means all.
	AccessiblePrograms []EPRef `json:"accessible_programs"`
	// BelongsTo are the programmes whose plans contain the module.
	BelongsTo []EPBankModuleUsage `json:"belongs_to"`
}

// EPBankModuleUsage is a programme whose plan contains a bank module.
type EPBankModuleUsage struct {
	// ID is the programme ID.
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	ModuleID int64  `json:"module_id"`
}

// EPBankModuleForm creates or edits a bank module's main information.
type EPBankModuleForm struct {
	NameRU string `json:"name_ru"`
	// NameEN must be Latin.
	NameEN     string `json:"name_en"`
	BlockID    int64  `json:"block_id"`
	StandardID int64  `json:"standard_id"`
}

// EPBankModuleListParams filters [ConstructorEPService.BankModules]. Zero values are omitted.
type EPBankModuleListParams struct {
	Query            string
	EducationLevelID int64
	StartYear        int
	// StatusID is [EPStatusSigned] when picking approved modules for a plan.
	StatusID int64
	// BlockID and ProgramID are used when picking modules for a programme plan.
	BlockID   int64
	ProgramID int64
	// Limit is the page size (typically 20).
	Limit  int
	Offset int
}

func epBank(moduleID int64, parts ...string) string {
	return epPath(append([]string{"bank", "modules", id(moduleID)}, parts...)...)
}

// CreateBankModule creates a module bank module and returns its ID.
// Staff only.
// POST /api/constructor-ep/bank/modules/create
func (s *ConstructorEPService) CreateBankModule(ctx context.Context, form EPBankModuleForm) (int64, error) {
	return call[int64](ctx, s.c, post(epPath("bank", "modules", "create"), form))
}

// BankModules returns a page of the module bank.
// Staff only.
// GET /api/constructor-ep/bank/modules/list
func (s *ConstructorEPService) BankModules(ctx context.Context, p EPBankModuleListParams) (*EPBankModuleList, error) {
	v := q().set("query", p.Query)
	epSetID(v, "education_level_id", p.EducationLevelID)
	epSetID(v, "start_year", int64(p.StartYear))
	epSetID(v, "status_id", p.StatusID)
	epSetID(v, "block_id", p.BlockID)
	epSetID(v, "program_id", p.ProgramID)
	epSetID(v, "limit", int64(p.Limit))
	v.set("offset", p.Offset)
	return call[*EPBankModuleList](ctx, s.c, get(epPath("bank", "modules", "list"), v))
}

// BankModule returns a module bank card.
// Staff only.
// GET /api/constructor-ep/bank/modules/{module_id}/info
func (s *ConstructorEPService) BankModule(ctx context.Context, moduleID int64) (*EPBankModule, error) {
	return call[*EPBankModule](ctx, s.c, get(epBank(moduleID, "info"), nil))
}

// BankModuleTree returns the content tree of a bank module (its root node).
// Staff only.
// GET /api/constructor-ep/bank/modules/{module_id}/tree
func (s *ConstructorEPService) BankModuleTree(ctx context.Context, moduleID int64) (*EPPlanNode, error) {
	return call[*EPPlanNode](ctx, s.c, get(epBank(moduleID, "tree"), nil))
}

// UpdateBankModule edits a bank module's main information.
// Staff only.
// PATCH /api/constructor-ep/bank/modules/{module_id}
func (s *ConstructorEPService) UpdateBankModule(ctx context.Context, moduleID int64, form EPBankModuleForm) error {
	return exec(ctx, s.c, patch(epBank(moduleID), form))
}

// SetBankModuleDescription sets a bank module's description (status [EPStatusInWork] only).
// Staff only.
// PATCH /api/constructor-ep/bank/modules/{module_id}/description
func (s *ConstructorEPService) SetBankModuleDescription(ctx context.Context, moduleID int64, description string) error {
	body := struct {
		Description string `json:"description"`
	}{description}
	return exec(ctx, s.c, patch(epBank(moduleID, "description"), body))
}

// SetBankModulePrograms restricts which programmes may use a bank module;
// nil allows every programme (sent as null).
// Staff only.
// PATCH /api/constructor-ep/bank/modules/{module_id}/accessible_programs
func (s *ConstructorEPService) SetBankModulePrograms(ctx context.Context, moduleID int64, programIDs []int64) error {
	body := struct {
		AccessiblePrograms *[]int64 `json:"accessible_programs"`
	}{}
	if programIDs != nil {
		body.AccessiblePrograms = &programIDs
	}
	return exec(ctx, s.c, patch(epBank(moduleID, "accessible_programs"), body))
}

// ApproveBankModule approves a bank module. Administrators only.
// Staff only.
// POST /api/constructor-ep/bank/modules/{module_id}/approve
func (s *ConstructorEPService) ApproveBankModule(ctx context.Context, moduleID int64) error {
	return exec(ctx, s.c, post(epBank(moduleID, "approve"), nil))
}

// RevertBankModule returns an approved bank module to work. Administrators only.
// Staff only.
// POST /api/constructor-ep/bank/modules/{module_id}/revert
func (s *ConstructorEPService) RevertBankModule(ctx context.Context, moduleID int64) error {
	return exec(ctx, s.c, post(epBank(moduleID, "revert"), nil))
}

// ArchiveBankModule archives a bank module.
// Staff only.
// POST /api/constructor-ep/bank/modules/{module_id}/archive
func (s *ConstructorEPService) ArchiveBankModule(ctx context.Context, moduleID int64) error {
	return exec(ctx, s.c, post(epBank(moduleID, "archive"), nil))
}

// AddBankModuleEditor adds an editor to a bank module.
// Staff only.
// POST /api/constructor-ep/bank/modules/{module_id}/{isu}
func (s *ConstructorEPService) AddBankModuleEditor(ctx context.Context, moduleID, isu int64) error {
	return exec(ctx, s.c, post(epBank(moduleID, id(isu)), nil))
}

// DeleteBankModuleEditor removes an editor from a bank module.
// Staff only.
// DELETE /api/constructor-ep/bank/modules/{module_id}/{isu}
func (s *ConstructorEPService) DeleteBankModuleEditor(ctx context.Context, moduleID, isu int64) error {
	return exec(ctx, s.c, del(epBank(moduleID, id(isu)), nil))
}

// CreateBankSubmodule creates a submodule in a bank module tree and returns its ID.
// Staff only.
// POST /api/constructor-ep/bank/modules/{module_id}/tree/module
func (s *ConstructorEPService) CreateBankSubmodule(ctx context.Context, moduleID int64, form EPModuleForm) (int64, error) {
	return call[int64](ctx, s.c, post(epBank(moduleID, "tree", "module"), form))
}

// UpdateBankSubmodule edits a submodule of a bank module tree and returns its ID.
// Staff only.
// PATCH /api/constructor-ep/bank/modules/{module_id}/tree/module/{submodule_id}
func (s *ConstructorEPService) UpdateBankSubmodule(ctx context.Context, moduleID, submoduleID int64, form EPModuleForm) (int64, error) {
	return call[int64](ctx, s.c, patch(epBank(moduleID, "tree", "module", id(submoduleID)), form))
}

// DeleteBankSubmodule deletes a submodule from a bank module tree.
// Staff only.
// DELETE /api/constructor-ep/bank/modules/{module_id}/tree/module/{submodule_id}
func (s *ConstructorEPService) DeleteBankSubmodule(ctx context.Context, moduleID, submoduleID int64) error {
	return exec(ctx, s.c, del(epBank(moduleID, "tree", "module", id(submoduleID)), nil))
}

// AddBankDisciplines adds disciplines to a module of a bank module tree.
// Staff only.
// POST /api/constructor-ep/bank/modules/{module_id}/tree/module/{submodule_id}/discipline
func (s *ConstructorEPService) AddBankDisciplines(ctx context.Context, moduleID, submoduleID int64, disciplineIDs []int64) error {
	return exec(ctx, s.c, post(epBank(moduleID, "tree", "module", id(submoduleID), "discipline"), disciplineIDs))
}

// SetBankDisciplineSemesters sets the semesters of a discipline in a bank
// module tree; submoduleID is the discipline's EPPlanNode.ModuleID. For a
// multi-semester discipline pass only the start semester.
// Staff only.
// PATCH /api/constructor-ep/bank/modules/{module_id}/tree/module/{submodule_id}/discipline/{discipline_id}
func (s *ConstructorEPService) SetBankDisciplineSemesters(ctx context.Context, moduleID, submoduleID, disciplineID int64, semesters []int) error {
	return exec(ctx, s.c, patch(epBank(moduleID, "tree", "module", id(submoduleID), "discipline", id(disciplineID)), semesters))
}

// DeleteBankDiscipline removes a discipline from a bank module tree.
// Staff only.
// DELETE /api/constructor-ep/bank/modules/{module_id}/tree/module/{submodule_id}/discipline/{discipline_id}
func (s *ConstructorEPService) DeleteBankDiscipline(ctx context.Context, moduleID, submoduleID, disciplineID int64) error {
	return exec(ctx, s.c, del(epBank(moduleID, "tree", "module", id(submoduleID), "discipline", id(disciplineID)), nil))
}
