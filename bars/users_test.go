package bars_test

import (
	"net/http"
	"testing"

	"github.com/kewldan/go-itmo/bars"
)

func TestConfigRoutes(t *testing.T) {
	f, c := session(t)
	f.reply(`{"id":1,"name":"daily_message","value":"hi"}`)
	ok(t, c.SetConfigValue(ctx, bars.SettingDailyMessage, "hi"))
	f.expect(http.MethodPost, "config").sameJSON(t, `{"name":"daily_message","value":"hi"}`)

	f.reply("")
	ok(t, c.SetSystemConfig(ctx, bars.SystemConfig{EditBars: true, CurrentYear: "2026/2027", CurrentTerm: "1"}))
	f.expect(http.MethodPost, "config/multiple").sameJSON(t, `{"edit_bars":true,"edit_regular":false,"edit_journal":false,
"edit_exams":false,"edit_personal":false,"current_year":"2026/2027","current_term":"1"}`)

	f.reply(`[{"id":3,"name":"current_year","value":"2026/2027"},{"id":4,"name":"x","value":null}]`)
	settings := must[[]bars.Setting](t)(c.PersonalSettings(ctx))
	f.expect(http.MethodGet, "config/personal")
	if len(settings) != 2 || *settings[0].ID != 3 || settings[1].Value != "" {
		t.Errorf("settings = %+v", settings)
	}

	f.reply("")
	ok(t, c.DeletePersonalSetting(ctx, "current_year"))
	f.expect(http.MethodDelete, "config/personal/current_year")
}

func TestUserSearches(t *testing.T) {
	f, c := session(t)
	f.reply(`[{"id":7,"login":"100001","first_name":"Имя","middle_name":"","last_name":"Фамилия"}]`)
	users := must[[]bars.UserSummary](t)(c.Users(ctx, "Фам & Ко"))
	f.expect(http.MethodGet, "users?filter=%D0%A4%D0%B0%D0%BC+%26+%D0%9A%D0%BE")
	if users[0].ID != 7 || users[0].LastName != "Фамилия" {
		t.Errorf("users = %+v", users)
	}

	f.reply(`[]`)
	must[[]bars.UserSummary](t)(c.Users(ctx, ""))
	f.expect(http.MethodGet, "users")

	f.reply(`[{"id":8,"first_name":"","middle_name":"","last_name":""}]`)
	must[[]bars.UserSummary](t)(c.Students(ctx, "ab"))
	f.expect(http.MethodGet, "users/students?filter=ab")

	f.reply(`[]`)
	must[[]bars.UserSummary](t)(c.UsersByRole(ctx, bars.RoleStudent, "x"))
	f.expect(http.MethodGet, "users/by_role/3?filter=x")

	f.reply(`[{"id":5,"type":"flow","name":"Поток","identifier":"a/b","term":1,"year":"2026/2027"}]`)
	entries := must[[]bars.WhiteListEntry](t)(c.WhiteListCandidates(ctx, bars.FlowTypeFlow, "По", "teacher1"))
	f.expect(http.MethodGet, "users/white_list/flow?filter=%D0%9F%D0%BE&realizerPeopleId=teacher1")
	if entries[0].Identifier != "a/b" || *entries[0].Term != 1 || string(entries[0].ID) != "5" {
		t.Errorf("entries = %+v", entries)
	}

	f.reply(`[]`)
	must[[]bars.WhiteListEntry](t)(c.WhiteListCandidates(ctx, bars.FlowTypeGroup, "", ""))
	f.expect(http.MethodGet, "users/white_list/group")
}

func TestUserRoles(t *testing.T) {
	f, c := session(t)
	f.reply(`[{"id":4,"name":"Родитель","locked":false,"selected":true,"allows_white_list":false,"allows_multiple":true,
"requires_main_user_user_role":3,"main_user":{"id":9,"login":"100002","first_name":"","middle_name":"","last_name":""}},
{"id":1,"name":"Преподаватель-Администратор","locked":false,"selected":false,"allows_white_list":true,"allows_multiple":false,
"white_list":[{"id":11,"type":"group","name":"G1","identifier":"G1"}]}]`)
	roles := must[[]bars.UserRole](t)(c.UserRoles(ctx, 12))
	f.expect(http.MethodGet, "users/12/roles")
	if roles[0].MainUser == nil || roles[0].MainUser.ID != 9 || *roles[0].RequiresMainUserUserRole != bars.RoleStudent {
		t.Errorf("role 0 = %+v", roles[0])
	}
	if len(roles[1].WhiteList) != 1 || roles[1].WhiteList[0].Name != "G1" {
		t.Errorf("role 1 = %+v", roles[1])
	}

	yes := true
	roles[1].Selected = &yes
	roles[1].WhiteList = append(roles[1].WhiteList, bars.WhiteListEntry{ID: bars.RawJSON(`"new_1"`), Type: "discipline", Name: "D", Identifier: "", DisciplineIdentifier: new(int64(77)), DisciplineName: "D"})
	f.reply("")
	ok(t, c.SetUserRoles(ctx, 12, roles[1:]))
	f.expect(http.MethodPost, "users/12/roles").sameJSON(t, `[{"id":1,"name":"Преподаватель-Администратор","locked":false,"selected":true,
"allows_white_list":true,"allows_multiple":false,"white_list":[{"id":11,"type":"group","name":"G1","identifier":"G1"},
{"id":"new_1","type":"discipline","name":"D","identifier":"","discipline_identifier":77,"discipline_name":"D"}]}]`)

	f.reply("")
	ok(t, c.SelectRole(ctx, bars.UserRole{ID: 3, Name: "Ученик"}))
	f.expect(http.MethodPost, "users/set_selected_role").sameJSON(t, `{"id":3,"name":"Ученик","locked":null,"selected":null,
"allows_white_list":null,"allows_multiple":null}`)
}

func TestCatalogRoutes(t *testing.T) {
	f, c := session(t)
	f.reply(`[{"id":1,"code":"09.03.01","name":"Информатика"}]`)
	programs := must[[]bars.EducationalProgram](t)(c.EducationalPrograms(ctx, 90, 2, 4))
	f.expect(http.MethodGet, "educational_programs/90?term=2%2C4")
	if programs[0].Code != "09.03.01" {
		t.Errorf("programs = %+v", programs)
	}
	f.reply(`[]`)
	must[[]bars.EducationalProgram](t)(c.EducationalPrograms(ctx, 90))
	f.expect(http.MethodGet, "educational_programs/90?term=null")

	f.reply(`[{"id":3,"name":"Тест 1","year":"2026/2027"}]`)
	tests := must[[]bars.Test](t)(c.Tests(ctx, "2026/2027"))
	f.expect(http.MethodGet, "tests?year=2026%2F2027")
	if tests[0].ID != 3 {
		t.Errorf("tests = %+v", tests)
	}
	f.reply("")
	ok(t, c.CreateTest(ctx, "Тест 2", "2026/2027"))
	f.expect(http.MethodPost, "tests").sameJSON(t, `{"name":"Тест 2","year":"2026/2027"}`)

	f.reply(`[{"id":21,"name":"Электронное тестирование в ЦДО"}]`)
	types := must[[]bars.CheckpointType](t)(c.CheckpointTypes(ctx, "", bars.CheckpointKindRegular))
	f.expect(http.MethodGet, "checkpoint_types?type=regular")
	if types[0].ID != 21 {
		t.Errorf("types = %+v", types)
	}
	f.reply(`[]`)
	must[[]bars.CheckpointType](t)(c.CheckpointTypes(ctx, "Эк", ""))
	f.expect(http.MethodGet, "checkpoint_types?name=%D0%AD%D0%BA")

	f.queue(step{status: http.StatusCreated})
	ok(t, c.CreateCheckpointType(ctx, bars.CheckpointKindFinal, "Экзамен"))
	f.expect(http.MethodPost, "checkpoint_types/final").sameJSON(t, `{"name":"Экзамен"}`)
}
