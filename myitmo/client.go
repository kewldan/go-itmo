// Package myitmo is a client for the MyITMO personal cabinet API (https://my.itmo.ru).
//
// Create a client from an ITMO.ID token source (see package itmoid) and use the
// service fields, grouped the same way the web cabinet groups its sections:
//
//	auth := itmoid.New()
//	tok, err := auth.Login(ctx, itmoid.MyITMO, itmoid.Credentials{Username: u, Password: p})
//	client := myitmo.New(auth.TokenSource(ctx, itmoid.MyITMO, tok, save))
//	days, err := client.Schedule.Personal(ctx, myitmo.Today(), myitmo.Today().AddDays(7))
//
// Every method returns *[Error] when the server answers with an HTTP error or
// with a non-zero error_code inside a 200 response.
package myitmo

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/kewldan/go-itmo/internal/rest"
)

// Default endpoints.
const (
	DefaultBaseURL   = "https://my.itmo.ru/"
	DevBaseURL       = "https://dev.my.itmo.su/"
	DefaultQRBaseURL = "https://qr.itmo.su/"
)

// Client talks to MyITMO. It is safe for concurrent use.
type Client struct {
	rest       rest.Client
	qrBase     *url.URL
	tokens     oauth2.TokenSource
	language   string
	authorized map[string]bool

	Schedule       *ScheduleService
	RecordBook     *RecordBookService
	StudyPlan      *StudyPlanService
	IndividualPlan *IndividualPlanService
	Personalities  *PersonalitiesService
	Sport          *SportService
	System         *SystemService
	Requests       *RequestsService
	RequestsV2     *RequestsV2Service
	Finances       *FinancesService
	Election       *ElectionService
	QR             *QRService
	Booking        *BookingService
	Navigator      *NavigatorService
	Queues         *QueuesService
	Dormitory      *DormitoryService
	DMS            *DMSService
	Agreements     *AgreementsService
	Sign           *SignService
	Assistant      *AssistantService
	Intro          *ElectiveCourseService
	Facultative    *ElectiveCourseService
	Practices      *PracticesService
	Employments    *EmploymentsService
	GIA            *GIAService
	GIAStudents    *GIAStudentsService
	Adviser        *AdviserService
	Checklist      *ChecklistService
	Constructor    *ConstructorService
	ConstructorEP  *ConstructorEPService
	Vacation       *VacationService
}

type config struct {
	httpClient *http.Client
	baseURL    string
	qrBaseURL  string
	language   string
	userAgent  string
	logger     *slog.Logger
}

// Option configures a [Client].
type Option func(*config)

// WithHTTPClient sets the underlying HTTP client (timeouts, proxies, transport).
func WithHTTPClient(c *http.Client) Option { return func(o *config) { o.httpClient = c } }

// WithBaseURL points the client at another MyITMO deployment, for example [DevBaseURL].
func WithBaseURL(u string) Option { return func(o *config) { o.baseURL = u } }

// WithQRBaseURL overrides the pass service address.
func WithQRBaseURL(u string) Option { return func(o *config) { o.qrBaseURL = u } }

// WithLanguage sets Accept-Language (default "ru"). MyITMO localises some
// payloads by it: people search returns transliterated names without it.
func WithLanguage(lang string) Option { return func(o *config) { o.language = lang } }

// WithUserAgent sets the User-Agent header.
func WithUserAgent(ua string) Option { return func(o *config) { o.userAgent = ua } }

// WithLogger logs every request at debug level: method, path, status and
// latency only. Tokens, queries and bodies are never logged.
func WithLogger(l *slog.Logger) Option { return func(o *config) { o.logger = l } }

// New returns a client authorised by tokens, usually from
// itmoid.Authenticator.TokenSource. For a fixed access token use
// oauth2.StaticTokenSource.
func New(tokens oauth2.TokenSource, opts ...Option) *Client {
	cfg := config{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    DefaultBaseURL,
		qrBaseURL:  DefaultQRBaseURL,
		language:   "ru",
		userAgent:  "go-itmo",
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	base := mustParse(cfg.baseURL)
	qr := mustParse(cfg.qrBaseURL)

	c := &Client{
		qrBase:     qr,
		tokens:     oauth2.ReuseTokenSource(nil, tokens),
		language:   cfg.language,
		authorized: map[string]bool{base.Host: true, qr.Host: true},
	}
	c.rest = rest.Client{HTTP: cfg.httpClient, BaseURL: base, UserAgent: cfg.userAgent, Logger: cfg.logger, Prepare: c.prepare}

	c.Schedule = &ScheduleService{c}
	c.RecordBook = &RecordBookService{c}
	c.StudyPlan = &StudyPlanService{c}
	c.IndividualPlan = &IndividualPlanService{c}
	c.Personalities = &PersonalitiesService{c}
	c.Sport = &SportService{c}
	c.System = &SystemService{c}
	c.Requests = &RequestsService{c}
	c.RequestsV2 = &RequestsV2Service{c}
	c.Finances = &FinancesService{c}
	c.Election = &ElectionService{c}
	c.QR = &QRService{c}
	c.Booking = &BookingService{c}
	c.Navigator = &NavigatorService{c}
	c.Queues = &QueuesService{c}
	c.Dormitory = &DormitoryService{c}
	c.DMS = &DMSService{c}
	c.Agreements = &AgreementsService{c}
	c.Sign = &SignService{c}
	c.Assistant = &AssistantService{c}
	c.Intro = &ElectiveCourseService{c, "intro"}
	c.Facultative = &ElectiveCourseService{c, "facultative"}
	c.Practices = &PracticesService{c}
	c.Employments = &EmploymentsService{c}
	c.GIA = &GIAService{c}
	c.GIAStudents = &GIAStudentsService{c}
	c.Adviser = &AdviserService{c}
	c.Checklist = &ChecklistService{c}
	c.Constructor = &ConstructorService{c}
	c.ConstructorEP = &ConstructorEPService{c}
	c.Vacation = &VacationService{c}
	return c
}

func mustParse(raw string) *url.URL {
	if !strings.HasSuffix(raw, "/") {
		raw += "/"
	}
	u, err := url.Parse(raw)
	if err != nil {
		panic("myitmo: invalid base URL: " + err.Error())
	}
	return u
}

// prepare authorises only MyITMO hosts: the token must never leak to a
// foreign host even if a caller passes an absolute URL to [Client.Do].
func (c *Client) prepare(req *http.Request) error {
	if !c.authorized[req.URL.Host] {
		return nil
	}
	tok, err := c.tokens.Token()
	if err != nil {
		return &TokenError{Err: err}
	}
	tok.SetAuthHeader(req)
	if req.Header.Get("Accept-Language") == "" && c.language != "" {
		req.Header.Set("Accept-Language", c.language)
	}
	return nil
}

// TokenError means the access token could not be obtained or refreshed; the
// user has to sign in again.
type TokenError struct{ Err error }

func (e *TokenError) Error() string { return "myitmo: token unavailable: " + e.Err.Error() }
func (e *TokenError) Unwrap() error { return e.Err }

// Do sends an arbitrary request to MyITMO and decodes the "result" field of the
// standard envelope into out (which may be nil). Use it for endpoints this
// package does not model yet; path is relative to the base URL.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	res, err := call[jsonValue](ctx, c, &rest.Request{Method: method, Path: path, Query: query, JSON: body})
	if err != nil || out == nil {
		return err
	}
	return decodeInto(res, out)
}
