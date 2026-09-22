package myitmo_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestRequestsV2Routes(t *testing.T) {
	type C = *myitmo.Client
	d := func(c C) *myitmo.RequestsV2Service { return c.RequestsV2 }
	access := myitmo.RequestV2AccessListInput{Code: "list-a", Name: "Список", Type: myitmo.RequestV2AccessListDynamic, ConfigJSON: `{"url":"https://example.test/users"}`}
	runRequestsRoutes(t, []requestsRoute{
		// Student side.
		{name: "Forms", result: `[]`, method: "GET", path: "/api/requests/v2/forms",
			run: func(c C) error { return requestsDiscard(d(c).Forms(ctx)) }},
		{name: "Form", result: `{}`, method: "GET", path: "/api/requests/v2/forms/a%2Fb",
			run: func(c C) error { return requestsDiscard(d(c).Form(ctx, "a/b")) }},
		{name: "Submit", result: `{"id":7}`, method: "POST", path: "/api/requests/v2/requests",
			body: `{"alias":"certificate","payloadJson":"{\"q1\":\"x\"}"}`,
			run:  func(c C) error { return requestsDiscard(d(c).Submit(ctx, "certificate", `{"q1":"x"}`)) }},
		{name: "Mine", result: `[]`, method: "GET", path: "/api/requests/v2/requests/mine",
			run: func(c C) error { return requestsDiscard(d(c).Mine(ctx)) }},
		{name: "Get", result: `{}`, method: "GET", path: "/api/requests/v2/requests/7",
			run: func(c C) error { return requestsDiscard(d(c).Get(ctx, 7)) }},
		{name: "ResultFile", result: `{}`, method: "GET", path: "/api/requests/v2/requests/7/files/order%20doc",
			run: func(c C) error { return requestsDiscard(d(c).ResultFile(ctx, 7, "order doc")) }},
		{name: "MyTaskRequestIDs", result: `[1,2]`, method: "GET", path: "/api/requests/v2/tasks/mine/request-ids",
			run: func(c C) error { return requestsDiscard(d(c).MyTaskRequestIDs(ctx)) }},
		{name: "MyTask", method: "GET", path: "/api/requests/v2/tasks/mine/by-request/7",
			run: func(c C) error { return requestsDiscard(d(c).MyTask(ctx, 7)) }},
		{name: "AssignedRequests", result: `[]`, method: "GET", path: "/api/requests/v2/requests/my-tasks",
			run: func(c C) error { return requestsDiscard(d(c).AssignedRequests(ctx)) }},
		{name: "TaskCounts", result: `{}`, method: "GET", path: "/api/requests/v2/requests/task-counts",
			run: func(c C) error { return requestsDiscard(d(c).TaskCounts(ctx)) }},
		{name: "CompleteTask", method: "POST", path: "/api/requests/v2/tasks/t-1/complete",
			body: `{"outcome":"approve","variables":{"reviewComment":"ok","requestDataPatch":{"f":1}}}`,
			run: func(c C) error {
				return d(c).CompleteTask(ctx, "t-1", "approve", myitmo.RequestV2TaskVariables{ReviewComment: "ok", RequestDataPatch: myitmo.RawJSON(`{"f":1}`)})
			}},
		{name: "CompleteTaskNoVars", method: "POST", path: "/api/requests/v2/tasks/t-1/complete",
			body: `{"outcome":"reject","variables":{}}`,
			run:  func(c C) error { return d(c).CompleteTask(ctx, "t-1", "reject", myitmo.RequestV2TaskVariables{}) }},
		{name: "ResolveDictionary", result: `{"items":[],"totalCount":0}`, method: "POST", path: "/api/requests/v2/dictionaries/resolve", query: "code=faculties",
			body: `{"field":"fac","questionName":"fac","filter":"ин","query":"ин","skip":0,"take":25,"data":{"a":1},"dependsOnQuestion":"lvl","dependsOnValue":"1,2"}`,
			run: func(c C) error {
				return requestsDiscard(d(c).ResolveDictionary(ctx, "faculties", myitmo.RequestV2DictionaryContext{
					Field: "fac", QuestionName: "fac", Filter: "ин", Query: "ин", Take: 25, Data: myitmo.RawJSON(`{"a":1}`),
					DependsOnQuestion: "lvl", DependsOnValue: "1,2",
				}))
			}},
		{name: "ResolvePrefill", result: `{"value":"x"}`, method: "POST", path: "/api/requests/v2/prefills/resolve", query: "code=group",
			body: `{"field":"g","questionName":"g","data":{},"dep":"5"}`,
			run: func(c C) error {
				return requestsDiscard(d(c).ResolvePrefill(ctx, "group", myitmo.RequestV2PrefillContext{Field: "g", QuestionName: "g", Data: myitmo.RawJSON(`{}`), Dep: "5"}))
			}},

		// Categories.
		{name: "Categories", result: `[]`, method: "GET", path: "/api/requests/v2/categories",
			run: func(c C) error { return requestsDiscard(d(c).Categories(ctx)) }},
		{name: "SaveCategory", result: `{}`, method: "PUT", path: "/api/requests/v2/categories",
			body: `{"code":"education","name":"Учёба","icon":"question","color":"secondary-blue","sortOrder":2}`,
			run: func(c C) error {
				return requestsDiscard(d(c).SaveCategory(ctx, myitmo.RequestV2CategoryInput{Code: "education", Name: "Учёба", Icon: "question", Color: "secondary-blue", SortOrder: 2}))
			}},
		{name: "DeleteCategory", method: "DELETE", path: "/api/requests/v2/categories/4",
			run: func(c C) error { return d(c).DeleteCategory(ctx, 4) }},

		// Access lists.
		{name: "CanManageAccessLists", result: `true`, method: "GET", path: "/api/requests/v2/access-definitions/admin-access",
			run: func(c C) error { return requestsDiscard(d(c).CanManageAccessLists(ctx)) }},
		{name: "AccessLists", result: `[]`, method: "GET", path: "/api/requests/v2/access-definitions",
			run: func(c C) error { return requestsDiscard(d(c).AccessLists(ctx)) }},
		{name: "SaveAccessList", method: "PUT", path: "/api/requests/v2/access-definitions",
			body: `{"code":"list-a","name":"Список","type":"DYNAMIC","configJson":"{\"url\":\"https://example.test/users\"}"}`,
			run:  func(c C) error { return d(c).SaveAccessList(ctx, access) }},
		{name: "AccessListUsers", result: `["sub-1"]`, method: "GET", path: "/api/requests/v2/access-definitions/list%20a/users",
			run: func(c C) error { return requestsDiscard(d(c).AccessListUsers(ctx, "list a")) }},
		{name: "SetAccessListUsers", method: "PUT", path: "/api/requests/v2/access-definitions/list-a/users", body: `["sub-1","sub-2"]`,
			run: func(c C) error { return d(c).SetAccessListUsers(ctx, "list-a", []string{"sub-1", "sub-2"}) }},
		{name: "SetAccessListUsersEmpty", method: "PUT", path: "/api/requests/v2/access-definitions/list-a/users", body: `[]`,
			run: func(c C) error { return d(c).SetAccessListUsers(ctx, "list-a", nil) }},
		{name: "PushAccessListExternalUsers", method: "PUT", path: "/api/requests/v2/access-definitions/list-a/external-users", body: `["sub-3"]`,
			run: func(c C) error { return d(c).PushAccessListExternalUsers(ctx, "list-a", []string{"sub-3"}) }},
		{name: "TestAccessList", result: `{}`, method: "POST", path: "/api/requests/v2/access-definitions/test",
			body: `{"code":"list-a","name":"Список","type":"DYNAMIC","configJson":"{\"url\":\"https://example.test/users\"}"}`,
			run:  func(c C) error { return requestsDiscard(d(c).TestAccessList(ctx, access)) }},
		{name: "RefreshAccessList", method: "POST", path: "/api/requests/v2/access-definitions/list-a/refresh",
			run: func(c C) error { return d(c).RefreshAccessList(ctx, "list-a") }},
		{name: "DeleteAccessList", method: "DELETE", path: "/api/requests/v2/access-definitions/list-a",
			run: func(c C) error { return d(c).DeleteAccessList(ctx, "list-a") }},
		{name: "AccessListOptions", result: `[]`, method: "GET", path: "/api/requests/v2/access-definitions/options",
			run: func(c C) error { return requestsDiscard(d(c).AccessListOptions(ctx)) }},

		// Dictionaries and catalogs.
		{name: "Dictionaries", result: `[]`, method: "GET", path: "/api/requests/v2/dictionaries",
			run: func(c C) error { return requestsDiscard(d(c).Dictionaries(ctx)) }},
		{name: "CanManageDictionaries", result: `false`, method: "GET", path: "/api/requests/v2/dictionaries/admin-access",
			run: func(c C) error { return requestsDiscard(d(c).CanManageDictionaries(ctx)) }},
		{name: "DictionaryConfigs", result: `[]`, method: "GET", path: "/api/requests/v2/dictionaries/admin",
			run: func(c C) error { return requestsDiscard(d(c).DictionaryConfigs(ctx)) }},
		{name: "SaveDictionary", method: "PUT", path: "/api/requests/v2/dictionaries/admin",
			body: `{"code":"fac","name":"Факультеты","configJson":"{}"}`,
			run: func(c C) error {
				return d(c).SaveDictionary(ctx, myitmo.RequestV2DictionaryInput{Code: "fac", Name: "Факультеты", ConfigJSON: "{}"})
			}},
		{name: "TestDictionary", result: `{}`, method: "POST", path: "/api/requests/v2/dictionaries/test",
			body: `{"configJson":"{}","context":{"faculty":{"id":"1"}}}`,
			run: func(c C) error {
				return requestsDiscard(d(c).TestDictionary(ctx, "{}", myitmo.RawJSON(`{"faculty":{"id":"1"}}`)))
			}},
		{name: "TestDictionaryNoContext", result: `{}`, method: "POST", path: "/api/requests/v2/dictionaries/test",
			body: `{"configJson":"{}","context":{}}`,
			run:  func(c C) error { return requestsDiscard(d(c).TestDictionary(ctx, "{}", nil)) }},
		{name: "DeleteDictionary", method: "DELETE", path: "/api/requests/v2/dictionaries/admin/fac",
			run: func(c C) error { return d(c).DeleteDictionary(ctx, "fac") }},
		{name: "Prefills", result: `[]`, method: "GET", path: "/api/requests/v2/prefills",
			run: func(c C) error { return requestsDiscard(d(c).Prefills(ctx)) }},
		{name: "ContextEnrichments", result: `[]`, method: "GET", path: "/api/requests/v2/context-enrichments",
			run: func(c C) error { return requestsDiscard(d(c).ContextEnrichments(ctx)) }},

		// Platform roles and users.
		{name: "MyPlatformRole", result: `{"role":"EDITOR"}`, method: "GET", path: "/api/requests/v2/platform-roles/me",
			run: func(c C) error { return requestsDiscard(d(c).MyPlatformRole(ctx)) }},
		{name: "PlatformRoles", result: `[]`, method: "GET", path: "/api/requests/v2/platform-roles",
			run: func(c C) error { return requestsDiscard(d(c).PlatformRoles(ctx)) }},
		{name: "SetPlatformRole", method: "PUT", path: "/api/requests/v2/platform-roles/sub%2F1", query: "role=ADMIN",
			run: func(c C) error { return d(c).SetPlatformRole(ctx, "sub/1", myitmo.RequestV2RoleAdmin) }},
		{name: "ResetPlatformRole", method: "DELETE", path: "/api/requests/v2/platform-roles/sub-1",
			run: func(c C) error { return d(c).ResetPlatformRole(ctx, "sub-1") }},
		{name: "Users", result: `[]`, method: "GET", path: "/api/requests/v2/users", query: "limit=20&query=%D0%B8%D0%B2",
			run: func(c C) error { return requestsDiscard(d(c).Users(ctx, "ив", 20)) }},

		// Processes and forms.
		{name: "Processes", result: `[]`, method: "GET", path: "/api/requests/v2/processes",
			run: func(c C) error { return requestsDiscard(d(c).Processes(ctx)) }},
		{name: "ProcessXML", result: `"<definitions/>"`, method: "GET", path: "/api/requests/v2/processes/proc_1",
			run: func(c C) error { return requestsDiscard(d(c).ProcessXML(ctx, "proc_1")) }},
		{name: "DeployProcess", result: `{"id":"def:2"}`, method: "POST", path: "/api/requests/v2/processes/deploy",
			body: `{"bpmnXml":"<definitions/>"}`,
			run:  func(c C) error { return requestsDiscard(d(c).DeployProcess(ctx, "<definitions/>")) }},
		{name: "CreateProcessDraft", result: `{"alias":"proc_2"}`, method: "POST", path: "/api/requests/v2/processes/draft",
			body: `{"name":"Новый процесс"}`,
			run:  func(c C) error { return requestsDiscard(d(c).CreateProcessDraft(ctx, "Новый процесс")) }},
		{name: "ProcessConfig", result: `{}`, method: "GET", path: "/api/requests/v2/processes/proc_1/config",
			run: func(c C) error { return requestsDiscard(d(c).ProcessConfig(ctx, "proc_1")) }},
		{name: "UpdateProcessConfig", result: `{}`, method: "PUT", path: "/api/requests/v2/processes/proc_1/config",
			body: `{"visibleInCatalog":true,"observerUserIds":["sub-1"]}`,
			run: func(c C) error {
				return requestsDiscard(d(c).UpdateProcessConfig(ctx, "proc_1", true, []string{"sub-1"}))
			}},
		{name: "PublishProcess", result: `{"status":"PUBLISHED"}`, method: "POST", path: "/api/requests/v2/processes/proc_1/publish",
			run: func(c C) error { return requestsDiscard(d(c).PublishProcess(ctx, "proc_1")) }},
		{name: "SuspendProcess", result: `{"status":"SUSPENDED"}`, method: "POST", path: "/api/requests/v2/processes/proc_1/suspend",
			run: func(c C) error { return requestsDiscard(d(c).SuspendProcess(ctx, "proc_1")) }},
		{name: "DeleteProcess", method: "DELETE", path: "/api/requests/v2/processes/proc_1",
			run: func(c C) error { return d(c).DeleteProcess(ctx, "proc_1") }},
		{name: "SaveForm", result: `{}`, method: "POST", path: "/api/requests/v2/forms",
			body: `{"alias":"proc_1","name":"Справка","description":null,"htmlNote":"<b>!</b>","categoryId":4,"version":1,"schemaJson":"{}","processVariableMappingsJson":"[]","accessCodes":[]}`,
			run: func(c C) error {
				return requestsDiscard(d(c).SaveForm(ctx, myitmo.RequestV2FormInput{
					Alias: "proc_1", Name: "Справка", HTMLNote: myitmo.Ptr("<b>!</b>"), CategoryID: myitmo.Ptr[int64](4),
					Version: 1, SchemaJSON: "{}", ProcessVariableMappingsJSON: "[]",
				}))
			}},
		{name: "SetFormAccess", method: "PUT", path: "/api/requests/v2/forms/proc_1/access", body: `{"accessCodes":["list-a"]}`,
			run: func(c C) error { return d(c).SetFormAccess(ctx, "proc_1", []string{"list-a"}) }},

		// Processing.
		{name: "ObservedRequests", result: `{}`, method: "GET", path: "/api/requests/v2/requests/observed",
			query: "formCode=&initiator=&page=0&requestId=0&size=30&status=",
			run:   func(c C) error { return requestsDiscard(d(c).ObservedRequests(ctx, myitmo.RequestV2ObservedParams{})) }},
		{name: "ObservedRequestsFiltered", result: `{}`, method: "GET", path: "/api/requests/v2/requests/observed",
			query: "formCode=proc_1&initiator=ivan&page=2&requestId=7&size=10&status=REJECTED",
			run: func(c C) error {
				return requestsDiscard(d(c).ObservedRequests(ctx, myitmo.RequestV2ObservedParams{
					Page: 2, Size: 10, FormCode: "proc_1", Status: myitmo.RequestV2Rejected, RequestID: 7, Initiator: "ivan",
				}))
			}},
		{name: "ObservedRequestTypes", result: `[]`, method: "GET", path: "/api/requests/v2/requests/observed/types",
			run: func(c C) error { return requestsDiscard(d(c).ObservedRequestTypes(ctx)) }},
		{name: "Tasks", result: `[]`, method: "GET", path: "/api/requests/v2/tasks",
			run: func(c C) error { return requestsDiscard(d(c).Tasks(ctx)) }},

		// Process instances.
		{name: "ProcessInstances", result: `{}`, method: "GET", path: "/api/requests/v2/admin/process-instances", query: "page=0&size=50",
			run: func(c C) error { return requestsDiscard(d(c).ProcessInstances(ctx, myitmo.RequestV2InstanceParams{})) }},
		{name: "ProcessInstancesFiltered", result: `{}`, method: "GET", path: "/api/requests/v2/admin/process-instances",
			query: "page=1&search=abc&size=20&status=FAILED",
			run: func(c C) error {
				return requestsDiscard(d(c).ProcessInstances(ctx, myitmo.RequestV2InstanceParams{Status: myitmo.RequestV2InstanceFailed, Search: "abc", Page: 1, Size: 20}))
			}},
		{name: "ProcessInstance", result: `{}`, method: "GET", path: "/api/requests/v2/admin/process-instances/pi:1",
			run: func(c C) error { return requestsDiscard(d(c).ProcessInstance(ctx, "pi:1")) }},
		{name: "RetryProcessJob", method: "POST", path: "/api/requests/v2/admin/process-instances/pi-1/jobs/job-1/retry",
			run: func(c C) error { return d(c).RetryProcessJob(ctx, "pi-1", "job-1") }},
		{name: "RestartProcessActivity", method: "POST", path: "/api/requests/v2/admin/process-instances/pi-1/activities/task_a/restart",
			run: func(c C) error { return d(c).RestartProcessActivity(ctx, "pi-1", "task_a") }},
		{name: "DeleteProcessInstance", method: "DELETE", path: "/api/requests/v2/admin/process-instances/pi-1",
			run: func(c C) error { return d(c).DeleteProcessInstance(ctx, "pi-1") }},
	})
}

func TestRequestsV2DecodeStudent(t *testing.T) {
	f, c := newFake(t)
	s := c.RequestsV2

	f.result(`[{"id":null,"code":"other","name":null,"color":"secondary-blue","icon":"question","sortOrder":null,"forms":[
		{"id":"p12","alias":"certificate","name":"Справка","description":null,"schemaJson":{"pages":[]},"htmlNote":null},
		{"id":13,"alias":"leave","name":"Отпуск","description":"d","schemaJson":"{\"pages\":[]}","htmlNote":"<b>n</b>"}]}]`)
	cats := must[[]myitmo.RequestV2FormCategory](t)(s.Forms(ctx))
	if cats[0].ID != "" || cats[0].SortOrder != nil || cats[0].Forms[0].ID != "p12" || cats[0].Forms[1].ID != "13" {
		t.Errorf("categories = %+v", cats)
	}
	if cats[0].Forms[0].SchemaJSON != `{"pages":[]}` || cats[0].Forms[1].SchemaJSON != `{"pages":[]}` || cats[0].Forms[1].HTMLNote != "<b>n</b>" {
		t.Errorf("schema = %q / %q", cats[0].Forms[0].SchemaJSON, cats[0].Forms[1].SchemaJSON)
	}

	f.result(`{"id":7,"formCode":"certificate","formName":"Справка","status":"PROCESSED","subStatus":"Готово","createdAt":"2026-04-01T09:00:00Z","updatedAt":null,
		"initiator":"sub-1","initiatorDisplayName":"Тестов Тест","reviewComment":null,"payloadJson":"{\"q1\":\"x\"}",
		"renderedFields":[{"name":"q1","title":"Вопрос","displayValue":"x"}],
		"result":{"tone":"success","title":"Готово","message":null,"parameters":[{"label":"Номер","value":"A-1"}],"files":[{"variableName":"doc","label":"Документ","filename":"doc.pdf"}]}}`)
	req := must[*myitmo.RequestV2Request](t)(s.Get(ctx, 7))
	if req.Status != myitmo.RequestV2Processed || !req.Status.Finished() || req.UpdatedAt != nil || req.CreatedAt.Hour() != 9 ||
		req.RenderedFields[0].DisplayValue != "x" || req.Result.Files[0].VariableName != "doc" || req.Result.Parameters[0].Value != "A-1" {
		t.Errorf("request = %+v", req)
	}

	f.result(`[{"id":8,"formCode":"leave","formName":null,"status":"IN_REVIEW","createdAt":"2026-04-02T10:00:00+03:00","updatedAt":"2026-04-03T10:00:00+03:00","initiator":"sub-2","initiatorDisplayName":null}]`)
	mine := must[[]myitmo.RequestV2Summary](t)(s.Mine(ctx))
	if mine[0].Status != myitmo.RequestV2InReview || mine[0].Status.Finished() || mine[0].UpdatedAt == nil {
		t.Errorf("mine = %+v", mine)
	}

	f.result(`{"base64Content":"aGVsbG8=","contentType":"text/plain","filename":"a.txt"}`)
	file := must[*myitmo.RequestV2FileContent](t)(s.ResultFile(ctx, 7, "doc"))
	if b := must[[]byte](t)(file.Bytes()); string(b) != "hello" || file.Filename != "a.txt" {
		t.Errorf("file = %q", b)
	}
	if _, err := (&myitmo.RequestV2FileContent{}).Bytes(); err == nil {
		t.Error("empty content must fail")
	}

	f.result(`{"id":"t-1","requestId":7,"name":"Подпись","actions":[{"code":"sign","label":"Подписать","kind":"primary","requiresComment":false}],
		"taskFormSchemaJson":null,"requestPayloadJson":"{}","signatureTask":true,"signatureTasks":[{"taskId":55,"signatureId":"s-1"}]}`)
	task := must[*myitmo.RequestV2Task](t)(s.MyTask(ctx, 7))
	if task.ID != "t-1" || task.RequestID != 7 || task.Actions[0].Code != "sign" || string(task.SignatureTask) != "true" ||
		task.SignatureTasks[0].TaskID != "55" || task.SignatureTasks[0].SignatureID != "s-1" {
		t.Errorf("task = %+v", task)
	}

	f.result(`null`)
	if task := must[*myitmo.RequestV2Task](t)(s.MyTask(ctx, 7)); task != nil {
		t.Errorf("task = %+v, want nil", task)
	}

	f.result(`{"ownRequestTasks":2,"assignedRequestTasks":5}`)
	if n := must[*myitmo.RequestV2TaskCounts](t)(s.TaskCounts(ctx)); n.OwnRequestTasks != 2 || n.AssignedRequestTasks != 5 {
		t.Errorf("counts = %+v", n)
	}

	f.result(`{"items":[{"value":1,"text":"Один"},{"value":"b","text":"Два"}],"totalCount":40}`)
	page := must[*myitmo.RequestV2CountPage[myitmo.RequestV2DictionaryItem]](t)(s.ResolveDictionary(ctx, "x", myitmo.RequestV2DictionaryContext{}))
	if page.TotalCount != 40 || len(page.Items) != 2 || string(page.Items[1].Value) != `"b"` {
		t.Errorf("page = %+v", page)
	}
	f.result(`[{"value":1,"text":"Один"}]`)
	page = must[*myitmo.RequestV2CountPage[myitmo.RequestV2DictionaryItem]](t)(s.ResolveDictionary(ctx, "x", myitmo.RequestV2DictionaryContext{}))
	if page.TotalCount != 1 || page.Items[0].Text != "Один" {
		t.Errorf("bare page = %+v", page)
	}

	f.result(`{"value":{"id":3}}`)
	if p := must[*myitmo.RequestV2Prefill](t)(s.ResolvePrefill(ctx, "x", myitmo.RequestV2PrefillContext{})); string(p.Value) != `{"id":3}` {
		t.Errorf("prefill = %s", p.Value)
	}
}

func TestRequestsV2DecodeStaff(t *testing.T) {
	f, c := newFake(t)
	s := c.RequestsV2

	f.result(`[{"id":4,"code":"education","name":"Учёба","icon":"question","color":"primary-blue","sortOrder":1}]`)
	if cats := must[[]myitmo.RequestV2Category](t)(s.Categories(ctx)); cats[0].ID != 4 || cats[0].SortOrder != 1 {
		t.Errorf("categories = %+v", cats)
	}

	f.result(`[{"code":"list-a","name":"Список","type":"DYNAMIC","configJson":"{\"url\":\"https://example.test\"}","userCount":12,
		"lastAttemptAt":"2026-05-01T12:00:00Z","lastUpdateStatus":"SUCCESS","lastUpdatedBy":"sub-1","lastUpdateMessage":null},
		{"code":"list-b","name":"Статика","type":"STATIC","configJson":"{}","userCount":0,"lastAttemptAt":null}]`)
	lists := must[[]myitmo.RequestV2AccessList](t)(s.AccessLists(ctx))
	if lists[0].Type != myitmo.RequestV2AccessListDynamic || lists[0].UserCount != 12 || lists[0].LastAttemptAt == nil || lists[1].LastAttemptAt != nil {
		t.Errorf("lists = %+v", lists)
	}

	f.result(`{"statusCode":200,"userCount":3,"sample":["sub-1","sub-2"]}`)
	if r := must[*myitmo.RequestV2AccessListTest](t)(s.TestAccessList(ctx, myitmo.RequestV2AccessListInput{})); r.StatusCode != http.StatusOK || len(r.Sample) != 2 {
		t.Errorf("test = %+v", r)
	}

	for _, tc := range []struct {
		result string
		want   bool
	}{{`true`, true}, {`false`, false}, {`null`, false}, {`{"allowed":true}`, true}, {`0`, false}} {
		f.result(tc.result)
		if got := must[bool](t)(s.CanManageAccessLists(ctx)); got != tc.want {
			t.Errorf("probe %s = %v", tc.result, got)
		}
	}
	f.reply(http.StatusForbidden, `{"error_code":403,"error_message":"Нет доступа","result":null}`)
	if ok, err := s.CanManageDictionaries(ctx); ok || err == nil {
		t.Errorf("probe on 403 = %v, %v", ok, err)
	}

	f.result(`{"statusCode":200,"totalCount":null,"items":[{"id":1},"x"]}`)
	if r := must[*myitmo.RequestV2DictionaryTest](t)(s.TestDictionary(ctx, "{}", nil)); r.TotalCount != nil || len(r.Items) != 2 {
		t.Errorf("dictionary test = %+v", r)
	}

	f.result(`[{"code":"fac","name":"Факультеты","configJson":"{\"url\":\"https://example.test/fac\",\"itemsPath\":\"data\"}"}]`)
	dicts := must[[]myitmo.RequestV2Dictionary](t)(s.DictionaryConfigs(ctx))
	if dicts[0].Code != "fac" || dicts[0].ConfigJSON == "" {
		t.Errorf("dictionaries = %+v", dicts)
	}

	f.result(`[{"userId":"sub-1","fullName":"Тестов Тест","personalId":100001,"email":"t@example.test","phone":null}]`)
	if u := must[[]myitmo.RequestV2User](t)(s.Users(ctx, "те", 20)); u[0].PersonalID != "100001" || u[0].Phone != "" {
		t.Errorf("users = %+v", u)
	}

	f.result(`[{"userId":"sub-1","role":"ADMIN"}]`)
	if r := must[[]myitmo.RequestV2PlatformRole](t)(s.PlatformRoles(ctx)); r[0].Role != myitmo.RequestV2RoleAdmin {
		t.Errorf("roles = %+v", r)
	}

	f.result(`{"status":"PUBLISHED","visibleInCatalog":true,"observers":[{"userId":"sub-1","fullName":"Тестов Тест"}]}`)
	if cfg := must[*myitmo.RequestV2ProcessConfig](t)(s.PublishProcess(ctx, "proc_1")); cfg.Status != myitmo.RequestV2ProcessPublished || cfg.Observers[0].UserID != "sub-1" {
		t.Errorf("config = %+v", cfg)
	}

	f.reply(http.StatusNotFound, `{"error_code":404,"error_message":"not found","result":null}`)
	_, err := s.ProcessConfig(ctx, "proc_9")
	var e *myitmo.Error
	if !errors.As(err, &e) || !e.IsNotFound() {
		t.Errorf("err = %v", err)
	}

	f.result(`"<?xml version=\"1.0\"?><definitions/>"`)
	if x := must[string](t)(s.ProcessXML(ctx, "proc_1")); x != `<?xml version="1.0"?><definitions/>` {
		t.Errorf("xml = %q", x)
	}

	f.result(`{"items":[{"id":9,"formCode":"leave","formName":"Отпуск","status":"APPROVED","initiator":"sub-2","initiatorDisplayName":"Тест","createdAt":"2026-04-02T10:00:00+03:00"}],
		"totalElements":31,"totalPages":2,"page":1,"size":30}`)
	obs := must[*myitmo.RequestV2Page[myitmo.RequestV2Summary]](t)(s.ObservedRequests(ctx, myitmo.RequestV2ObservedParams{Page: 1}))
	if obs.TotalElements != 31 || obs.TotalPages != 2 || obs.Page != 1 || obs.Items[0].Status != myitmo.RequestV2Approved {
		t.Errorf("observed = %+v", obs)
	}

	f.result(`[{"id":"t-2","name":null,"taskDefinitionKey":"review","requestId":9,"assignee":null,"actions":[{"code":"reject","label":"Отклонить","kind":"danger","requiresComment":true}],"taskFormSchemaJson":"{}"}]`)
	tasks := must[[]myitmo.RequestV2Task](t)(s.Tasks(ctx))
	if tasks[0].Assignee != nil || tasks[0].TaskDefinitionKey != "review" || !tasks[0].Actions[0].RequiresComment {
		t.Errorf("tasks = %+v", tasks)
	}

	f.result(`{"items":[{"id":"pi-1","status":"FAILED","hasErrors":true,"processDefinitionName":"Отпуск","alias":"leave","processDefinitionVersion":3,
		"currentActivities":[{"activityId":"call","name":null,"type":"async-job"}],"requestId":9,"requestStatus":"IN_PROGRESS","initiator":"sub-2",
		"startTime":"2026-04-02T10:00:00Z","durationInMillis":null}],"totalCount":1,"statusCounts":{"ALL":10,"FAILED":1}}`)
	inst := must[*myitmo.RequestV2InstancePage](t)(s.ProcessInstances(ctx, myitmo.RequestV2InstanceParams{}))
	if inst.TotalCount != 1 || inst.StatusCounts["ALL"] != 10 || inst.Items[0].Status != myitmo.RequestV2InstanceFailed ||
		*inst.Items[0].RequestID != 9 || inst.Items[0].CurrentActivities[0].Type != "async-job" || inst.Items[0].DurationInMillis != nil {
		t.Errorf("instances = %+v", inst)
	}

	f.result(`{"process":{"id":"pi-1","status":"TERMINATED","alias":"leave","processDefinitionVersion":3,"startTime":"2026-04-02T10:00:00Z","endTime":"2026-04-02T11:00:00Z","durationInMillis":3600000},
		"errors":[{"id":"job-1","status":"DEAD_LETTER","retries":0,"elementId":"call","elementName":"Вызов","message":"timeout","stackTrace":"..."}],
		"variables":[{"name":"approved","type":"boolean","value":true,"executionId":"ex-1","taskId":null,"updatedAt":"2026-04-02T10:30:00Z"}],
		"activities":[{"activityId":"start","name":null,"type":"startEvent","status":"COMPLETED","startTime":"2026-04-02T10:00:00Z","durationInMillis":5,"assignee":null}],
		"deleteReason":"manual"}`)
	det := must[*myitmo.RequestV2InstanceDetail](t)(s.ProcessInstance(ctx, "pi-1"))
	if det.Process.EndTime == nil || *det.Process.DurationInMillis != 3600000 || det.Errors[0].Status != myitmo.RequestV2DeadLetter ||
		string(det.Variables[0].Value) != "true" || det.Activities[0].Status != "COMPLETED" || det.DeleteReason != "manual" {
		t.Errorf("detail = %+v", det)
	}
}

func TestRequestsV2ErrorDetails(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusBadRequest, `{"error_code":400,"error_message":"Не удалось отправить заявку.","result":{"details":["Поле q1 обязательно"]}}`)
	_, err := c.RequestsV2.Submit(ctx, "certificate", "{}")
	var e *myitmo.Error
	if !errors.As(err, &e) || string(e.Details) != `["Поле q1 обязательно"]` {
		t.Errorf("err = %v", err)
	}
}
