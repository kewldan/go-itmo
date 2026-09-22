package myitmo_test

import (
	"encoding/json/v2"
	"net/http"
	"testing"
	"time"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestFinancesRoutes(t *testing.T) {
	from, to := myitmo.NewDate(2026, 3, 1), myitmo.NewDate(2026, 9, 1)
	tests := []struct {
		name   string
		call   func(c *myitmo.Client) error
		method string
		path   string
		query  string
	}{
		{"ScholarshipAccounts", func(c *myitmo.Client) error { _, err := c.Finances.ScholarshipAccounts(ctx); return err }, "GET", "/api/finances/scholarship/account", ""},
		{"ScholarshipIncome", func(c *myitmo.Client) error { _, err := c.Finances.ScholarshipIncome(ctx); return err }, "GET", "/api/finances/scholarship/income", ""},
		{"ScholarshipTotals", func(c *myitmo.Client) error { _, err := c.Finances.ScholarshipTotals(ctx, from, to); return err }, "GET", "/api/finances/scholarship/total", "dateFrom=2026-03-01&dateTo=2026-09-01"},
		{"ScholarshipTotalsAllTime", func(c *myitmo.Client) error {
			_, err := c.Finances.ScholarshipTotals(ctx, myitmo.Date{}, myitmo.Date{})
			return err
		}, "GET", "/api/finances/scholarship/total", ""},
		{"SalaryCurrent", func(c *myitmo.Client) error { _, err := c.Finances.SalaryCurrent(ctx); return err }, "GET", "/api/finances/salary/current", ""},
		{"SalaryPayoutPeriods", func(c *myitmo.Client) error { _, err := c.Finances.SalaryPayoutPeriods(ctx); return err }, "GET", "/api/finances/salary/payouts", ""},
		{"LastSalaryPayment", func(c *myitmo.Client) error { _, err := c.Finances.LastSalaryPayment(ctx); return err }, "GET", "/api/finances/salary/payouts/last", ""},
		{"SalaryTotals", func(c *myitmo.Client) error { _, err := c.Finances.SalaryTotals(ctx, from, to); return err }, "GET", "/api/finances/salary/payouts/total", "dateFrom=2026-03-01&dateTo=2026-09-01"},
		{"SalaryMonth", func(c *myitmo.Client) error { _, err := c.Finances.SalaryMonth(ctx, 2026, 4); return err }, "GET", "/api/finances/salary/payouts/month", "month=4&year=2026"},
		{"SalaryYear", func(c *myitmo.Client) error { _, err := c.Finances.SalaryYear(ctx, 2026); return err }, "GET", "/api/finances/salary/year", "year=2026"},
		{"EduPaymentsAvailable", func(c *myitmo.Client) error { _, err := c.Finances.EduPaymentsAvailable(ctx); return err }, "GET", "/api/finances/edupayments/availability", ""},
		{"EduContracts", func(c *myitmo.Client) error { _, err := c.Finances.EduContracts(ctx); return err }, "GET", "/api/finances/edupayments/contracts", ""},
		{"EduPayments", func(c *myitmo.Client) error { _, err := c.Finances.EduPayments(ctx); return err }, "GET", "/api/finances/edupayments/payments", ""},
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
		})
	}
}

func TestFinancesPayEdu(t *testing.T) {
	f, c := newFake(t)
	f.result(`"https://pay.example.test/form/abc"`)
	url := must[string](t)(c.Finances.PayEdu(ctx, myitmo.PayRequest{
		ContractID: "1001",
		Sum:        1500.5,
		SuccessURL: "https://my.itmo.ru/finances/education?payment=success",
		FailureURL: "https://my.itmo.ru/finances/education?payment=failure",
	}))
	r := f.expect(http.MethodPost, "/api/finances/edupayments/payments/pay")
	r.sameJSON(t, `{"contractId":1001,"sum":1500.5,"successUrl":"https://my.itmo.ru/finances/education?payment=success","failureUrl":"https://my.itmo.ru/finances/education?payment=failure"}`)
	if url != "https://pay.example.test/form/abc" {
		t.Errorf("url = %q", url)
	}
}

func TestFinancesScholarshipDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"sum":5000,"payment_items":[{"item_name":"Государственная академическая стипендия","sum":3000.5,"date_start":"2026-09-01","date_end":"2027-01-31T00:00:00"}]}]`)
	acc := must[[]myitmo.ScholarshipAccount](t)(c.Finances.ScholarshipAccounts(ctx))
	if len(acc) != 1 || acc[0].Sum != 5000 || acc[0].PaymentItems[0].Sum != 3000.5 ||
		acc[0].PaymentItems[0].DateEnd != myitmo.NewDate(2027, 1, 31) {
		t.Errorf("accounts = %+v", acc)
	}

	f.result(`[{"payment_date":"2026-08-01","payment":{"income":[{"item_name":"Стипендия","account":null,"sum":3000}],"income_details":[{"item_name":"Надбавка","sum":500}]}}]`)
	inc := must[[]myitmo.ScholarshipIncomeMonth](t)(c.Finances.ScholarshipIncome(ctx))
	if len(inc) != 1 || inc[0].PaymentDate != myitmo.NewDate(2026, 8, 1) || inc[0].Payment.Income[0].Account != "" ||
		inc[0].Payment.IncomeDetails[0].Sum != 500 {
		t.Errorf("income = %+v", inc)
	}

	f.result(`[{"category_id":3,"category_name":" Стипендии\n","sum":120000}]`)
	tot := must[[]myitmo.ScholarshipTotal](t)(c.Finances.ScholarshipTotals(ctx, myitmo.Date{}, myitmo.Date{}))
	if len(tot) != 1 || tot[0].CategoryID != 3 || tot[0].Sum != 120000 {
		t.Errorf("totals = %+v", tot)
	}
}

func TestFinancesSalaryDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"ist_name":"Бюджет","sum":40000}]`)
	cur := must[[]myitmo.SalaryComponent](t)(c.Finances.SalaryCurrent(ctx))
	if len(cur) != 1 || cur[0].IstName != "Бюджет" {
		t.Errorf("current = %+v", cur)
	}

	f.result(`[{"year":2026,"month":4},{"year":2026,"month":5}]`)
	periods := must[[]myitmo.SalaryPayoutPeriod](t)(c.Finances.SalaryPayoutPeriods(ctx))
	if len(periods) != 2 || periods[1].Year != 2026 {
		t.Errorf("periods = %+v", periods)
	}

	f.result(`{"sum":35000.25,"date":"2026-05-15","is_previous":true}`)
	last := must[*myitmo.SalaryLastPayment](t)(c.Finances.LastSalaryPayment(ctx))
	if last == nil || !last.IsPrevious || last.Date != myitmo.NewDate(2026, 5, 15) {
		t.Errorf("last = %+v", last)
	}
	f.result(`null`)
	if last := must[*myitmo.SalaryLastPayment](t)(c.Finances.LastSalaryPayment(ctx)); last != nil {
		t.Errorf("null last = %+v", last)
	}

	f.result(`[{"item_id":910,"sum":300000},{"item_id":701,"sum":45000}]`)
	totals := must[[]myitmo.SalaryTotalItem](t)(c.Finances.SalaryTotals(ctx, myitmo.Date{}, myitmo.Date{}))
	if len(totals) != 2 || totals[0].ItemID != myitmo.SalaryItemIncome || totals[1].ItemID != myitmo.SalaryItemTaxes {
		t.Errorf("totals = %+v", totals)
	}

	f.result(`[{"date":"2026-04-10","total":30000,"payouts":[{"category":"Salary","name":"Оплата труда","sum":34000,"salaries":[{"item_name":"Оклад","sum":34000}]},{"category":"Taxes","name":"НДФЛ","sum":4000,"salaries":[]}]}]`)
	month := must[[]myitmo.SalaryPayoutDay](t)(c.Finances.SalaryMonth(ctx, 2026, 4))
	if len(month) != 1 || len(month[0].Payouts) != 2 || month[0].Payouts[1].Category != myitmo.SalaryCategoryTaxes ||
		month[0].Payouts[0].Salaries[0].ItemName != "Оклад" {
		t.Errorf("month = %+v", month)
	}

	f.result(`[{"month":"2026-01-01","total":30000},{"month":"2026-03-01T00:00:00+03:00","total":31000}]`)
	year := must[[]myitmo.SalaryMonthTotal](t)(c.Finances.SalaryYear(ctx, 2026))
	if len(year) != 2 || year[1].Month.Month != time.March {
		t.Errorf("year = %+v", year)
	}
}

func TestFinancesEduDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`true`)
	if !must[bool](t)(c.Finances.EduPaymentsAvailable(ctx)) {
		t.Error("availability = false")
	}

	f.result(`[{"contractId":"A-17","number":"123/26","dateStart":"2026-09-01","dateEnd":"2030-08-31","active":1,"op":"Программная инженерия","period":"семестр","sum":"175000.00"},
		{"contractId":42,"number":"7/22","dateStart":"2022-09-01","dateEnd":"2026-08-31","active":false,"op":"Прикладная математика","period":"год","sum":300000}]`)
	contracts := must[[]myitmo.EduContract](t)(c.Finances.EduContracts(ctx))
	if len(contracts) != 2 || contracts[0].ContractID != "A-17" || !contracts[0].Active || contracts[0].Sum != 175000 ||
		contracts[1].ContractID != "42" || contracts[1].Active || contracts[1].Sum != 300000 {
		t.Errorf("contracts = %+v", contracts)
	}

	f.result(`[{"contractId":42,"name":"Обучение","number":"7/22","dateStart":"2022-09-01","dateEnd":"2026-08-31","active":true,"balance":"0",
		"payments":[{"period":"1 семестр","sum":"87500","paid":87500,"payUntil":"2026-09-15"},{"period":"2 семестр","sum":87500.5,"paid":"","payUntil":"2027-02-15"}]}]`)
	payments := must[[]myitmo.PaymentContract](t)(c.Finances.EduPayments(ctx))
	if len(payments) != 1 || len(payments[0].Payments) != 2 {
		t.Fatalf("payments = %+v", payments)
	}
	p := payments[0].Payments
	if p[0].Sum != 87500 || p[0].Paid != 87500 || p[1].Sum != 87500.5 || p[1].Paid != 0 || p[1].PayUntil != myitmo.NewDate(2027, 2, 15) {
		t.Errorf("installments = %+v", p)
	}
}

func TestFlexID(t *testing.T) {
	for in, want := range map[string]string{`"x-1"`: `"x-1"`, `17`: `17`, `"17"`: `17`, `"007"`: `"007"`} {
		var id myitmo.FlexID
		if err := json.Unmarshal([]byte(in), &id); err != nil {
			t.Fatalf("%s: %v", in, err)
		}
		out, err := json.Marshal(id)
		if err != nil || string(out) != want {
			t.Errorf("%s -> %s, %v; want %s", in, out, err, want)
		}
	}
	var id myitmo.FlexID
	if err := json.Unmarshal([]byte(`true`), &id); err == nil {
		t.Error("bool accepted as id")
	}
}
