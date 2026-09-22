package myitmo_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestDormitoryRoutes(t *testing.T) {
	tests := []struct {
		name   string
		call   func(c *myitmo.Client) error
		method string
		path   string
		query  string
		body   string
	}{
		{"Status", func(c *myitmo.Client) error { _, err := c.Dormitory.Status(ctx); return err }, "GET", "/api/dormitory/status/full", "", ""},
		{"Documents", func(c *myitmo.Client) error { _, err := c.Dormitory.Documents(ctx); return err }, "GET", "/api/dormitory/documents", "", ""},
		{"Apartments", func(c *myitmo.Client) error { _, err := c.Dormitory.Apartments(ctx); return err }, "GET", "/api/dormitory/apartments", "", ""},
		{"Slots", func(c *myitmo.Client) error { _, err := c.Dormitory.Slots(ctx); return err }, "GET", "/api/dormitory/settlement/slots", "", ""},
		{"Register", func(c *myitmo.Client) error { return c.Dormitory.Register(ctx, 77) }, "POST", "/api/dormitory/settlement/register", "", `{"timeSlotId":77}`},
		{"RemoveAppointment", func(c *myitmo.Client) error { return c.Dormitory.RemoveAppointment(ctx) }, "DELETE", "/api/dormitory/settlement/remove", "", ""},
		{"Cancel", func(c *myitmo.Client) error { return c.Dormitory.Cancel(ctx) }, "DELETE", "/api/dormitory/settlement/cancel", "", ""},
		{"ChangeCategory", func(c *myitmo.Client) error { return c.Dormitory.ChangeCategory(ctx) }, "POST", "/api/dormitory/settlement/change_category", "", ""},
		{"ContractPeriods", func(c *myitmo.Client) error { _, err := c.Dormitory.ContractPeriods(ctx); return err }, "GET", "/api/dormitory/payments/contracts/periods", "", ""},
		{"Contracts", func(c *myitmo.Client) error {
			_, err := c.Dormitory.Contracts(ctx, myitmo.NewDate(2026, 1, 1), myitmo.NewDate(2026, 12, 31))
			return err
		}, "GET", "/api/dormitory/payments/contracts", "from=2026-01-01&to=2026-12-31", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, c := newFake(t)
			if err := tt.call(c); err != nil {
				t.Fatal(err)
			}
			r := f.expect(tt.method, tt.path)
			if got := r.Query.Encode(); got != tt.query {
				t.Errorf("query = %q, want %q", got, tt.query)
			}
			if tt.body != "" {
				r.sameJSON(t, tt.body)
			} else if len(r.Body) != 0 {
				t.Errorf("unexpected body %s", r.Body)
			}
		})
	}
}

func TestDormitoryPay(t *testing.T) {
	f, c := newFake(t)
	f.result(`"https://pay.example.test/dorm"`)
	url := must[string](t)(c.Dormitory.Pay(ctx, myitmo.PayRequest{
		ContractID: "55",
		Sum:        4200,
		SuccessURL: "https://my.itmo.ru/services/dormitories?payment=success",
		FailureURL: "https://my.itmo.ru/services/dormitories?payment=failure",
	}))
	f.expect(http.MethodPost, "/api/dormitory/payments/pay").sameJSON(t,
		`{"contractId":55,"sum":4200,"successUrl":"https://my.itmo.ru/services/dormitories?payment=success","failureUrl":"https://my.itmo.ru/services/dormitories?payment=failure"}`)
	if url != "https://pay.example.test/dorm" {
		t.Errorf("url = %q", url)
	}
}

func TestDormitoryDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`{"statusId":3,"dormType":"dorm","queuePlace":null,
		"assignedDormitory":{"buildingId":43,"name":"Общежитие №1","address":"ул. Примерная, 1","settlementProcedureText":"<p>Возьмите паспорт</p>"},
		"appointment":{"timeSlot":"2026-08-28T10:30:00+03:00","canRemoveUntil":"2026-08-27T10:30:00"}}`)
	st := must[*myitmo.DormStatus](t)(c.Dormitory.Status(ctx))
	if st.StatusID != myitmo.DormStatusAppointed || st.DormType != myitmo.DormTypeDorm || st.QueuePlace != nil ||
		st.AssignedDormitory.BuildingID != 43 || st.Appointment.TimeSlot.Hour() != 10 ||
		!st.Appointment.CanRemoveUntil.Equal(time.Date(2026, 8, 27, 10, 30, 0, 0, myitmo.MSK)) {
		t.Errorf("status = %+v", st)
	}
	f.result(`null`)
	if st := must[*myitmo.DormStatus](t)(c.Dormitory.Status(ctx)); st != nil {
		t.Errorf("null status = %+v", st)
	}

	f.result(`[{"documentTypeName":"Флюорография","status":-1,"templateId":4001},{"documentTypeName":"Справка","status":2,"templateId":"4002"}]`)
	docs := must[[]myitmo.DormDocument](t)(c.Dormitory.Documents(ctx))
	if len(docs) != 2 || docs[0].Status != myitmo.DormDocumentRejected || docs[0].TemplateID != "4001" || docs[1].TemplateID != "4002" {
		t.Errorf("documents = %+v", docs)
	}

	f.result(`[{"id":9,"name":"ITMO.Aparts A","address":"пр. Тестовый, 2"}]`)
	ap := must[[]myitmo.DormApartment](t)(c.Dormitory.Apartments(ctx))
	if len(ap) != 1 || ap[0].ID != "9" {
		t.Errorf("apartments = %+v", ap)
	}

	f.result(`[{"timeSlotId":501,"timeSlot":"2026-08-28T09:00:00+03:00"}]`)
	slots := must[[]myitmo.DormTimeSlot](t)(c.Dormitory.Slots(ctx))
	if len(slots) != 1 || slots[0].TimeSlotID != 501 || slots[0].TimeSlot.IsZero() {
		t.Errorf("slots = %+v", slots)
	}

	f.result(`[{"year":2026,"dateFrom":"2026-01-01","dateTo":"2026-12-31"}]`)
	periods := must[[]myitmo.DormPeriod](t)(c.Dormitory.ContractPeriods(ctx))
	if len(periods) != 1 || periods[0].DateTo != myitmo.NewDate(2026, 12, 31) {
		t.Errorf("periods = %+v", periods)
	}

	f.result(`[{"contractId":"D-5","name":"Проживание","number":"5/26","dateStart":"2026-09-01","dateEnd":"2027-06-30","active":true,"balance":150.5,
		"payments":[{"period":"Сентябрь","sum":"4200","paid":"4200","payUntil":"2026-09-10"}]}]`)
	contracts := must[[]myitmo.PaymentContract](t)(c.Dormitory.Contracts(ctx, periods[0].DateFrom, periods[0].DateTo))
	if len(contracts) != 1 || contracts[0].ContractID != "D-5" || contracts[0].Balance != 150.5 || contracts[0].Payments[0].Paid != 4200 {
		t.Errorf("contracts = %+v", contracts)
	}
}
