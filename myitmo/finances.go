package myitmo

import "context"

// FinancesService is scholarships, tuition payments and salary (/api/finances).
type FinancesService struct{ c *Client }

// ScholarshipAccount is an assigned scholarship with its payment items.
type ScholarshipAccount struct {
	// Sum is the total of the assignment in roubles.
	Sum          float64           `json:"sum"`
	PaymentItems []ScholarshipItem `json:"payment_items"`
}

// ScholarshipItem is one scholarship of an assignment.
type ScholarshipItem struct {
	ItemName string `json:"item_name"`
	// Sum is in roubles.
	Sum       float64 `json:"sum"`
	DateStart Date    `json:"date_start"`
	DateEnd   Date    `json:"date_end"`
}

// ScholarshipIncomeMonth is the income of one month.
type ScholarshipIncomeMonth struct {
	// PaymentDate identifies the month (only month and year matter).
	PaymentDate Date                     `json:"payment_date"`
	Payment     ScholarshipIncomePayment `json:"payment"`
}

// ScholarshipIncomePayment is the breakdown of a month's income.
type ScholarshipIncomePayment struct {
	Income        []ScholarshipIncomeLine   `json:"income"`
	IncomeDetails []ScholarshipIncomeDetail `json:"income_details"`
}

// ScholarshipIncomeLine is one transfer of a month. The month total is the sum of Sum.
type ScholarshipIncomeLine struct {
	ItemName string `json:"item_name"`
	// Account is the bank account or card; empty when unknown.
	Account string  `json:"account"`
	Sum     float64 `json:"sum"`
}

// ScholarshipIncomeDetail is one accrual item of a month.
type ScholarshipIncomeDetail struct {
	ItemName string  `json:"item_name"`
	Sum      float64 `json:"sum"`
}

// ScholarshipTotal is the total of one finance category.
type ScholarshipTotal struct {
	CategoryID int64 `json:"category_id"`
	// CategoryName may carry stray spaces and line breaks.
	CategoryName string `json:"category_name"`
	// Sum is in roubles, not kopecks.
	Sum float64 `json:"sum"`
}

// SalaryComponent is a component of the assigned salary.
type SalaryComponent struct {
	// IstName is the component or funding source name.
	IstName string  `json:"ist_name"`
	Sum     float64 `json:"sum"`
}

// SalaryPayoutPeriod is an element of the payout period list; other members are unknown.
type SalaryPayoutPeriod struct {
	// Year is the only confirmed member; the list repeats it,
	// likely once per month with payouts.
	Year int `json:"year"`
}

// SalaryLastPayment is the last or the next salary payment.
type SalaryLastPayment struct {
	Sum  float64 `json:"sum"`
	Date Date    `json:"date"`
	// IsPrevious is true for a payment already made, false for an upcoming one.
	IsPrevious bool `json:"is_previous"`
}

// Salary total item ids of SalaryTotalItem.ItemID.
const (
	SalaryItemTaxes  = 701
	SalaryItemIncome = 910 // income after taxes
)

// SalaryTotalItem is a salary total by accrual item; other members are unknown.
type SalaryTotalItem struct {
	// ItemID is SalaryItemIncome, SalaryItemTaxes or another item.
	ItemID int64   `json:"item_id"`
	Sum    float64 `json:"sum"`
}

// SalaryPayoutDay is the payouts of one payout date.
type SalaryPayoutDay struct {
	Date    Date                   `json:"date"`
	Total   float64                `json:"total"`
	Payouts []SalaryPayoutCategory `json:"payouts"`
}

// SalaryCategoryTaxes is the SalaryPayoutCategory.Category of withheld taxes.
const SalaryCategoryTaxes = "Taxes"

// SalaryPayoutCategory is a group of payout lines.
type SalaryPayoutCategory struct {
	// Category is SalaryCategoryTaxes or another group key.
	Category string             `json:"category"`
	Name     string             `json:"name"`
	Sum      float64            `json:"sum"`
	Salaries []SalaryPayoutLine `json:"salaries"`
}

// SalaryPayoutLine is one line of a payout.
type SalaryPayoutLine struct {
	ItemName string  `json:"item_name"`
	Sum      float64 `json:"sum"`
}

// SalaryMonthTotal is the total of a month; Month is a date inside it.
type SalaryMonthTotal struct {
	Month Date    `json:"month"`
	Total float64 `json:"total"`
}

// EduContract is a tuition contract. The first one returned is the current one.
type EduContract struct {
	ContractID FlexID   `json:"contractId"`
	Number     string   `json:"number"`
	DateStart  Date     `json:"dateStart"`
	DateEnd    Date     `json:"dateEnd"`
	Active     FlexBool `json:"active"`
	// Op is the educational programme.
	Op string `json:"op"`
	// Period is the payment period label, e.g. "семестр".
	Period string `json:"period"`
	// Sum is the amount per period.
	Sum FlexNumber `json:"sum"`
}

// PaymentContract is a contract with its payment schedule (tuition and dormitory).
type PaymentContract struct {
	ContractID FlexID `json:"contractId"`
	Name       string `json:"name"`
	Number     string `json:"number"`
	DateStart  Date   `json:"dateStart"`
	DateEnd    Date   `json:"dateEnd"`
	Active     bool   `json:"active"`
	// Balance is the overpayment.
	Balance  FlexNumber           `json:"balance"`
	Payments []PaymentInstallment `json:"payments"`
}

// PaymentInstallment is one period of a payment schedule. Sum-Paid is still
// due; after PayUntil it is a debt.
type PaymentInstallment struct {
	Period   string     `json:"period"`
	Sum      FlexNumber `json:"sum"`
	Paid     FlexNumber `json:"paid"`
	PayUntil Date       `json:"payUntil"`
}

// PayRequest starts an online payment of a contract.
type PayRequest struct {
	ContractID FlexID `json:"contractId"`
	// Sum is in roubles with at most two decimals.
	Sum float64 `json:"sum"`
	// SuccessURL and FailureURL are where the gateway redirects afterwards.
	SuccessURL string `json:"successUrl"`
	FailureURL string `json:"failureUrl"`
}

// ScholarshipAccounts returns the assigned scholarships.
// GET /api/finances/scholarship/account
func (s *FinancesService) ScholarshipAccounts(ctx context.Context) ([]ScholarshipAccount, error) {
	return call[[]ScholarshipAccount](ctx, s.c, get("api/finances/scholarship/account", nil))
}

// ScholarshipIncome returns the monthly income history.
// GET /api/finances/scholarship/income
func (s *FinancesService) ScholarshipIncome(ctx context.Context) ([]ScholarshipIncomeMonth, error) {
	return call[[]ScholarshipIncomeMonth](ctx, s.c, get("api/finances/scholarship/income", nil))
}

// ScholarshipTotals returns totals per finance category between from and to;
// zero dates mean all time.
// GET /api/finances/scholarship/total
func (s *FinancesService) ScholarshipTotals(ctx context.Context, from, to Date) ([]ScholarshipTotal, error) {
	return call[[]ScholarshipTotal](ctx, s.c, get("api/finances/scholarship/total", q().set("dateFrom", from).set("dateTo", to)))
}

// SalaryCurrent returns the assigned salary components; empty when the user is not employed.
// Staff only.
// GET /api/finances/salary/current
func (s *FinancesService) SalaryCurrent(ctx context.Context) ([]SalaryComponent, error) {
	return call[[]SalaryComponent](ctx, s.c, get("api/finances/salary/current", nil))
}

// SalaryPayoutPeriods returns the periods that have salary payouts (the year selector).
// Staff only.
// GET /api/finances/salary/payouts
func (s *FinancesService) SalaryPayoutPeriods(ctx context.Context) ([]SalaryPayoutPeriod, error) {
	return call[[]SalaryPayoutPeriod](ctx, s.c, get("api/finances/salary/payouts", nil))
}

// LastSalaryPayment returns the last or next salary payment, or nil.
// Staff only.
// GET /api/finances/salary/payouts/last
func (s *FinancesService) LastSalaryPayment(ctx context.Context) (*SalaryLastPayment, error) {
	return call[*SalaryLastPayment](ctx, s.c, get("api/finances/salary/payouts/last", nil))
}

// SalaryTotals returns salary totals by item between from and to; zero dates mean all time.
// Staff only.
// GET /api/finances/salary/payouts/total
func (s *FinancesService) SalaryTotals(ctx context.Context, from, to Date) ([]SalaryTotalItem, error) {
	return call[[]SalaryTotalItem](ctx, s.c, get("api/finances/salary/payouts/total", q().set("dateFrom", from).set("dateTo", to)))
}

// SalaryMonth returns the payouts of a month (1..12) grouped by payout date.
// Staff only.
// GET /api/finances/salary/payouts/month
func (s *FinancesService) SalaryMonth(ctx context.Context, year, month int) ([]SalaryPayoutDay, error) {
	return call[[]SalaryPayoutDay](ctx, s.c, get("api/finances/salary/payouts/month", q().set("month", month).set("year", year)))
}

// SalaryYear returns monthly totals of a year; months without payouts may be missing.
// Staff only.
// GET /api/finances/salary/year
func (s *FinancesService) SalaryYear(ctx context.Context, year int) ([]SalaryMonthTotal, error) {
	return call[[]SalaryMonthTotal](ctx, s.c, get("api/finances/salary/year", q().set("year", year)))
}

// EduPaymentsAvailable reports whether the tuition payment section is available to the user.
// GET /api/finances/edupayments/availability
func (s *FinancesService) EduPaymentsAvailable(ctx context.Context) (bool, error) {
	return call[bool](ctx, s.c, get("api/finances/edupayments/availability", nil))
}

// EduContracts returns the tuition contracts; the first is current, the rest are archived.
// GET /api/finances/edupayments/contracts
func (s *FinancesService) EduContracts(ctx context.Context) ([]EduContract, error) {
	return call[[]EduContract](ctx, s.c, get("api/finances/edupayments/contracts", nil))
}

// EduPayments returns the payment schedule and debts of each tuition contract.
// GET /api/finances/edupayments/payments
func (s *FinancesService) EduPayments(ctx context.Context) ([]PaymentContract, error) {
	return call[[]PaymentContract](ctx, s.c, get("api/finances/edupayments/payments", nil))
}

// PayEdu starts an online tuition payment and returns the payment gateway URL.
// The gateway redirects back with the query payment=success or payment=failure.
// POST /api/finances/edupayments/payments/pay
func (s *FinancesService) PayEdu(ctx context.Context, req PayRequest) (string, error) {
	return call[string](ctx, s.c, post("api/finances/edupayments/payments/pay", req))
}
