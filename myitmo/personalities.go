package myitmo

import (
	"context"
	"iter"
)

// PersonalitiesService is the people directory (/api/personalities).
type PersonalitiesService struct{ c *Client }

// Person is the full public profile of a person.
type Person struct {
	ISU       int64       `json:"isu"`
	FIO       string      `json:"fio"`
	Gender    string      `json:"gender"`
	PhotoURL  string      `json:"photo"`
	Contacts  []Contact   `json:"contacts"`
	Rooms     []Room      `json:"rooms"`
	Positions []Position  `json:"positions"`
	Education []Education `json:"education"`
	// Powers are the person's responsibilities.
	Powers []PersonPower `json:"powers"`
	// Levels is the academic rank and degree; nil when the person has none.
	Levels *PersonLevels `json:"levels"`
	// Activities is present on the wire but may be incomplete; use
	// [PersonalitiesService.Activities].
	Activities       RawJSON `json:"activities,omitzero"`
	ExchangeTraining bool    `json:"exchange_training"`
}

// PersonPower is a responsibility of a person in a department.
type PersonPower struct {
	PowerID   int64  `json:"power_id"`
	PowerName string `json:"power_name"`
	DepName   string `json:"dep_name"`
	DepLink   string `json:"dep_link"`
}

// PersonLevels is an academic rank and degree.
type PersonLevels struct {
	// Rank is empty when the person has a degree but no rank.
	Rank   string `json:"rank"`
	Degree string `json:"degree"`
}

// Contact is a group of contacts of one kind.
type Contact struct {
	Contact []string `json:"contact"`
	// ContactAlias is a machine key such as phone, email, email_corp, web,
	// telegram or vkontakte.
	ContactAlias string `json:"contact_alias"`
}

// Room is an office linked to a person.
type Room struct {
	RoomNumber   string `json:"room_number"`
	BuildingName string `json:"bld_name"`
}

// Position is an employee position. The shape of the vacation fields and
// Sorting is not confirmed.
type Position struct {
	DepartmentName string `json:"department_name"`
	DepartmentLink string `json:"department_link"`
	PositionName   string `json:"position_name"`
	// Sorting orders the positions; a number or a string.
	Sorting FlexID `json:"sorting"`
	// Vacation is non-null while the person is on vacation; its shape is unknown.
	Vacation      RawJSON `json:"vacation,omitzero"`
	StartVacation Date    `json:"start_vacation"`
	EndVacation   Date    `json:"end_vacation"`
}

// OnVacation reports whether the position is marked as being on vacation.
func (p Position) OnVacation() bool {
	return len(p.Vacation) > 0 && string(p.Vacation) != "null"
}

// Education is a study record of a person.
type Education struct {
	Course      string `json:"course"`
	FacultyName string `json:"faculty_name"`
	Group       string `json:"group"`
}

// PersonSummary is a search result.
type PersonSummary struct {
	// ID usually equals the ISU number.
	ID     int64  `json:"id"`
	FIO    string `json:"fio"`
	Gender string `json:"gender"`
	Phone  string `json:"phone"`
	Email  string `json:"email"`
	Work   string `json:"work"`
	// Education is a study summary shown when Work is empty.
	Education string `json:"education"`
	PhotoURL  string `json:"photo"`
}

// Activity types in PersonActivity.Type and PersonActivitiesParams.Type.
const (
	ActivityProject = "project"
	ActivityEvent   = "event"
	ActivityArticle = "article"
	// ActivityRID is a registered result of intellectual activity (patent, software).
	ActivityRID = "rid"
)

// PersonActivity is a project, event, publication or intellectual property item.
// Dates are preformatted strings.
type PersonActivity struct {
	// Type is one of the Activity* constants or "other".
	Type      string `json:"type"`
	TypeName  string `json:"type_name"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	DateBegin string `json:"date_begin"`
	DateEnd   string `json:"date_end"`
	DateDoc   string `json:"date_doc"`
	// Year is a number or a string; empty when unknown.
	Year    FlexID `json:"year"`
	Info    string `json:"info"`
	Authors string `json:"authors"`
}

// PersonActivitiesParams filters [PersonalitiesService.Activities].
type PersonActivitiesParams struct {
	// Query is the search text.
	Query string
	// Type is one of the Activity* constants; empty means all.
	Type string
	// Limit is the page size (typically 20); 0 leaves it to the server.
	Limit int
	// Offset is an item offset.
	Offset int
}

// Get returns the profile of a person by ISU number.
// GET /api/personalities/persons/{person_id}
func (s *PersonalitiesService) Get(ctx context.Context, isu int64) (*Person, error) {
	return call[*Person](ctx, s.c, get("api/personalities/persons/"+id(isu), nil))
}

// Search finds people by name, ISU number or other attributes.
// GET /api/personalities/persons
func (s *PersonalitiesService) Search(ctx context.Context, query string, limit, offset int) (*Page[PersonSummary], error) {
	return call[*Page[PersonSummary]](ctx, s.c, get("api/personalities/persons", q().set("q", query).set("limit", limit).set("offset", offset)))
}

// SearchAll iterates over every match, fetching pages of pageSize lazily.
func (s *PersonalitiesService) SearchAll(ctx context.Context, query string, pageSize int) iter.Seq2[PersonSummary, error] {
	return func(yield func(PersonSummary, error) bool) {
		for offset := 0; ; {
			page, err := s.Search(ctx, query, pageSize, offset)
			if err != nil {
				yield(PersonSummary{}, err)
				return
			}
			for _, p := range page.Data {
				if !yield(p, nil) {
					return
				}
			}
			offset += len(page.Data)
			if len(page.Data) == 0 || offset >= page.Count {
				return
			}
		}
	}
}

// Activities returns a page of a person's activities.
// GET /api/personalities/persons/{person_id}/activities
func (s *PersonalitiesService) Activities(ctx context.Context, isu int64, p PersonActivitiesParams) (*Page[PersonActivity], error) {
	qv := q().set("q", p.Query).set("type", p.Type).set("offset", p.Offset)
	if p.Limit > 0 {
		qv.set("limit", p.Limit)
	}
	return call[*Page[PersonActivity]](ctx, s.c, get("api/personalities/persons/"+id(isu)+"/activities", qv))
}
