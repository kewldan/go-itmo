package bars

import "context"

// Setting names. The edit_* switches hold the
// strings "true" or "false".
const (
	SettingCurrentYear  = "current_year"  // "2025/2026"
	SettingCurrentTerm  = "current_term"  // "0" spring, "1" autumn
	SettingDailyMessage = "daily_message" // banner text, up to 252 characters
	SettingEditBars     = "edit_bars"     // editing of BARS tables
	SettingEditRegular  = "edit_regular"  // editing of current control marks
	SettingEditJournal  = "edit_journal"  // editing of the journal
	SettingEditExams    = "edit_exams"    // editing of intermediate certification
	SettingEditPersonal = "edit_personal" // editing of personal plans
)

// SetConfigValue stores one global setting, for example the system banner
// (SettingDailyMessage, "" clears it).
//
// POST /config
//
// Audience: Admin (DOD).
func (c *Client) SetConfigValue(ctx context.Context, name, value string) error {
	return c.exec(ctx, post("config", Setting{Name: name, Value: value}))
}

// SystemConfig is the body of [Client.SetSystemConfig]: a flat object, not a
// list of settings. Every field is sent.
type SystemConfig struct {
	EditBars     bool `json:"edit_bars"`
	EditRegular  bool `json:"edit_regular"`
	EditJournal  bool `json:"edit_journal"`
	EditExams    bool `json:"edit_exams"`
	EditPersonal bool `json:"edit_personal"`
	// CurrentYear is "2025/2026".
	CurrentYear string `json:"current_year"`
	// CurrentTerm is "0" for spring and "1" for autumn.
	CurrentTerm string `json:"current_term"`
}

// SetSystemConfig saves the global period and edit switches at once.
//
// POST /config/multiple
//
// Audience: Admin (DOD).
func (c *Client) SetSystemConfig(ctx context.Context, cfg SystemConfig) error {
	return c.exec(ctx, post("config/multiple", cfg))
}

// PersonalSettings returns the user's personal settings, the same list as
// User.PersonalConfig. The shape is assumed to be that of [Client.Config].
//
// GET /config/personal
//
// Audience: Student, Teacher, Admin, Parent.
func (c *Client) PersonalSettings(ctx context.Context) ([]Setting, error) {
	return fetch[[]Setting](ctx, c, get("config/personal", nil))
}

// DeletePersonalSetting resets a personal setting by name, for example
// SettingCurrentYear.
//
// DELETE /config/personal/{configType}
//
// Audience: Student, Teacher, Admin, Parent.
func (c *Client) DeletePersonalSetting(ctx context.Context, name string) error {
	return c.exec(ctx, del(route("config", "personal", name), nil))
}
