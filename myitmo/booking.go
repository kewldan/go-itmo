package myitmo

import (
	"context"
	"encoding/json/v2"
	"time"
)

// BookingService is room booking (/api/booking).
type BookingService struct{ c *Client }

// BookingStatusID is the status_id of a booking. The meaning of some ids is
// not confirmed; prefer BookingStatus.StatusName for display.
type BookingStatusID int

// Booking statuses with a known meaning.
const (
	// BookingStatusApproved is an approved booking; virtual-room credentials
	// are shown only in this status.
	BookingStatusApproved BookingStatusID = 1
	// BookingStatusOnApproval is waiting for approval ("На согласовании").
	BookingStatusOnApproval BookingStatusID = 3
	// BookingStatusRejected is a rejected booking ("Отклонено").
	BookingStatusRejected BookingStatusID = 4
	// BookingStatusReturned is returned to the owner for edits; only in this
	// status may the owner call [BookingService.Update].
	BookingStatusReturned BookingStatusID = 5
)

// DefaultBookingOccupancyStatuses are the statuses that make up the
// occupancy grid (1, 2, 5, 6 and 8).
var DefaultBookingOccupancyStatuses = []BookingStatusID{1, 2, 5, 6, 8}

// BookingGroup is a top-level room group (coworking, study rooms, ...).
type BookingGroup struct {
	GroupID   int64  `json:"group_id"`
	GroupName string `json:"group_name"`
}

// BookingCategory is a category ("space") of rooms inside a group.
type BookingCategory struct {
	CategoryID   int64  `json:"category_id"`
	CategoryName string `json:"category_name"`
	// MinDays is the minimum number of days ahead a booking can be made.
	MinDays *int `json:"min_days"`
	// MaxDays is the maximum number of days ahead a booking can be made.
	MaxDays *int `json:"max_days"`
}

// BookingEquipment is equipment of a room or of a booking.
type BookingEquipment struct {
	EquipmentID int64 `json:"equipment_id"`
	// ID replaces EquipmentID in some booking responses.
	ID            int64  `json:"id,omitzero"`
	EquipmentName string `json:"equipment_name"`
	// Count is the available quantity of a room, or the requested quantity of a booking.
	Count int `json:"count"`
}

// BookingRoom is a bookable room. Search results and bookings carry only a
// part of the fields.
type BookingRoom struct {
	RoomID   int64  `json:"room_id"`
	RoomName string `json:"room_name"`
	// BuildingID links the room to [NavigatorService]; nil when unknown.
	BuildingID *int64 `json:"building_id"`
	PhotoLink  string `json:"photo_link"`
	// Floor is shown as is; its type is not known.
	Floor RawJSON `json:"floor"`
	// Area is shown as is; its type is not known.
	Area      RawJSON            `json:"area"`
	MinCap    *int               `json:"min_cap"`
	MaxCap    *int               `json:"max_cap"`
	Equipment []BookingEquipment `json:"equipment"`
	Address   string             `json:"address"`
	// Group carries at least group_id.
	Group *BookingGroup `json:"group"`
	// Category carries at least category_id.
	Category *BookingCategory `json:"category"`
}

// BookingStatus is a booking status.
type BookingStatus struct {
	StatusID   BookingStatusID `json:"status_id"`
	StatusName string          `json:"status_name"`
}

// BookingCoBooker is a co-organiser of a booking.
type BookingCoBooker struct {
	ISU int64 `json:"isu"`
	// ID replaces ISU in some responses.
	ID  int64  `json:"id,omitzero"`
	FIO string `json:"fio"`
}

// Booking is one of the user's bookings.
type Booking struct {
	BookingID int64 `json:"booking_id"`
	// Name is the event name.
	Name string `json:"name"`
	// Room carries room_id, room_name, address and group.
	Room *BookingRoom `json:"room"`
	// Category carries at least category_id; special ids are 267 (a Zoom
	// account: the login is the room name, the password is PasswordForRoom)
	// and 1296 (myQuiz).
	Category      *BookingCategory   `json:"category"`
	Status        BookingStatus      `json:"status"`
	StartDatetime time.Time          `json:"start_datetime"`
	EndDatetime   time.Time          `json:"end_datetime"`
	OwnerISU      int64              `json:"owner_isu"`
	OwnerFIO      string             `json:"owner_fio"`
	ContactPhone  string             `json:"contact_phone"`
	Participants  int                `json:"participants"`
	CoBookers     []BookingCoBooker  `json:"co_bookers"`
	Equipment     []BookingEquipment `json:"equipment"`
	// AdditionalInfo is the user's free-text note.
	AdditionalInfo             string `json:"additional_info"`
	TechSupport                bool   `json:"tech_support"`
	LinkToVirtualRoom          string `json:"link_to_virtual_room"`
	PasswordForVirtualRoom     string `json:"password_for_virtual_room"`
	LinkForInviteToVirtualRoom string `json:"link_for_invite_to_virtual_room"`
	PasswordForRoom            string `json:"password_for_room"`
	AdminComment               string `json:"admin_comment"`
	// Comment is the fallback when AdminComment is empty.
	Comment string `json:"comment"`
}

// BookingPage is a page of the user's bookings.
type BookingPage struct {
	// Count is the total number of matches.
	Count int       `json:"count"`
	List  []Booking `json:"list"`
}

// BookingRoomOccupancy is the occupancy of one room.
type BookingRoomOccupancy struct {
	RoomID   int64         `json:"room_id"`
	Bookings []BookingSlot `json:"bookings"`
}

// BookingSlot is an existing booking in the occupancy grid.
type BookingSlot struct {
	BookingID int64 `json:"booking_id"`
	// BookingName is shown for the user's own bookings.
	BookingName   string    `json:"booking_name"`
	StartDatetime time.Time `json:"start_datetime"`
	EndDatetime   time.Time `json:"end_datetime"`
	OwnerISU      *int64    `json:"owner_isu"`
	OwnerFIO      string    `json:"owner_fio"`
	// Status carries only status_id.
	Status BookingStatus `json:"status"`
}

// BookingUser is what the booking service knows about the current user.
// Only phone_number is known; other fields are ignored.
type BookingUser struct {
	PhoneNumber string `json:"phone_number"`
}

// BookingListParams filters [BookingService.My].
type BookingListParams struct {
	// DateStart is required.
	DateStart Date
	DateEnd   Date
	// Limit is the page size, typically 25.
	Limit *int
	// Offset is sent as is. It is a zero-based page index (page-1), not an
	// item offset.
	Offset *int
}

// bookingTimeLayout is how the booking service expects datetimes, in Moscow time.
const bookingTimeLayout = "2006-01-02 15:04"

func bookingTime(t time.Time) string { return t.In(MSK).Format(bookingTimeLayout) }

// BookingCreate is a new booking. Times are sent as "YYYY-MM-DD HH:mm" in
// Moscow time; the minimum length is 30 minutes and the grid covers 08:00–23:00.
type BookingCreate struct {
	// Name is the event name.
	Name           string
	AdditionalInfo string
	// Participants must be within the room's MinCap..MaxCap.
	Participants int
	// ContactPhone is formatted "+7 (XXX) XXX-XX-XX".
	ContactPhone string
	// EventID is usually null.
	EventID *int64
	// CoBookers are ISU numbers of co-organisers.
	CoBookers []int64
	Start     time.Time
	End       time.Time
	RoomID    int64
	// Equipment is the requested equipment; empty when not needed.
	Equipment   []BookingEquipment
	TechSupport bool
}

type bookingCreateWire struct {
	Name           string             `json:"name"`
	AdditionalInfo string             `json:"additional_info"`
	Participants   int                `json:"participants"`
	ContactPhone   string             `json:"contact_phone"`
	EventID        *int64             `json:"event_id"`
	CoBookers      []int64            `json:"co_bookers"`
	StartDatetime  string             `json:"start_datetime"`
	EndDatetime    string             `json:"end_datetime"`
	RoomID         int64              `json:"room_id"`
	Equipment      []BookingEquipment `json:"equipment"`
	TechSupport    bool               `json:"tech_support"`
}

// MarshalJSON encodes the booking in the wire format with formatted times.
func (b BookingCreate) MarshalJSON() ([]byte, error) {
	return json.Marshal(bookingCreateWire{
		Name:           b.Name,
		AdditionalInfo: b.AdditionalInfo,
		Participants:   b.Participants,
		ContactPhone:   b.ContactPhone,
		EventID:        b.EventID,
		CoBookers:      bookingSlice(b.CoBookers),
		StartDatetime:  bookingTime(b.Start),
		EndDatetime:    bookingTime(b.End),
		RoomID:         b.RoomID,
		Equipment:      bookingSlice(b.Equipment),
		TechSupport:    b.TechSupport,
	})
}

// BookingUpdate is an edit of an existing booking. Times are sent as
// "YYYY-MM-DD HH:mm" in Moscow time.
type BookingUpdate struct {
	BookingID int64
	RoomID    int64
	Start     time.Time
	End       time.Time
	// Name is the event name; it is sent both as name and as event_name.
	Name           string
	Participants   int
	ContactPhone   string
	AdditionalInfo string
	// CoBookers are ISU numbers of co-organisers.
	CoBookers   []int64
	Equipment   []BookingEquipment
	TechSupport bool
}

type bookingUpdateWire struct {
	BookingID      int64              `json:"booking_id"`
	RoomID         int64              `json:"room_id"`
	StartDatetime  string             `json:"start_datetime"`
	EndDatetime    string             `json:"end_datetime"`
	EventName      string             `json:"event_name"`
	Name           string             `json:"name"`
	Participants   int                `json:"participants"`
	ContactPhone   string             `json:"contact_phone"`
	AdditionalInfo string             `json:"additional_info"`
	CoBookers      []int64            `json:"co_bookers"`
	Equipment      []BookingEquipment `json:"equipment"`
	TechSupport    bool               `json:"tech_support"`
}

// MarshalJSON encodes the edit in the wire format with formatted times.
func (b BookingUpdate) MarshalJSON() ([]byte, error) {
	return json.Marshal(bookingUpdateWire{
		BookingID:      b.BookingID,
		RoomID:         b.RoomID,
		StartDatetime:  bookingTime(b.Start),
		EndDatetime:    bookingTime(b.End),
		EventName:      b.Name,
		Name:           b.Name,
		Participants:   b.Participants,
		ContactPhone:   b.ContactPhone,
		AdditionalInfo: b.AdditionalInfo,
		CoBookers:      bookingSlice(b.CoBookers),
		Equipment:      bookingSlice(b.Equipment),
		TechSupport:    b.TechSupport,
	})
}

// bookingSlice turns a nil slice into an empty one so it is sent as [].
func bookingSlice[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

// Groups returns the room groups.
// GET /api/booking/dictionary/rooms/groups
func (s *BookingService) Groups(ctx context.Context) ([]BookingGroup, error) {
	return call[[]BookingGroup](ctx, s.c, get("api/booking/dictionary/rooms/groups", nil))
}

// Categories returns the room categories of a group.
// GET /api/booking/dictionary/rooms/categories
func (s *BookingService) Categories(ctx context.Context, groupID int64) ([]BookingCategory, error) {
	return call[[]BookingCategory](ctx, s.c, get("api/booking/dictionary/rooms/categories", q().set("groupId", groupID)))
}

// Rooms returns the rooms of a category.
// GET /api/booking/rooms/roomsInCategory
func (s *BookingService) Rooms(ctx context.Context, categoryID int64) ([]BookingRoom, error) {
	return call[[]BookingRoom](ctx, s.c, get("api/booking/rooms/roomsInCategory", q().set("categoryId", categoryID)))
}

// Occupancy returns the occupancy of the rooms of a category on a date.
// statuses filters the bookings by status (the key repeats); nil means
// [DefaultBookingOccupancyStatuses], an empty non-nil slice sends no filter.
// GET /api/booking/rooms/roomBookings
func (s *BookingService) Occupancy(ctx context.Context, categoryID int64, date Date, statuses []BookingStatusID) ([]BookingRoomOccupancy, error) {
	if statuses == nil {
		statuses = DefaultBookingOccupancyStatuses
	}
	ids := make([]int64, len(statuses))
	for i, st := range statuses {
		ids[i] = int64(st)
	}
	return call[[]BookingRoomOccupancy](ctx, s.c, get("api/booking/rooms/roomBookings",
		q().set("status", ids).set("date", date).set("categoryId", categoryID)))
}

// SearchRooms finds rooms by name.
// GET /api/booking/rooms/byName
func (s *BookingService) SearchRooms(ctx context.Context, search string) ([]BookingRoom, error) {
	return call[[]BookingRoom](ctx, s.c, get("api/booking/rooms/byName", q().set("search", search)))
}

// My returns the user's bookings in a date range.
// GET /api/booking/bookings/my
func (s *BookingService) My(ctx context.Context, p BookingListParams) (*BookingPage, error) {
	return call[*BookingPage](ctx, s.c, get("api/booking/bookings/my",
		q().set("dateStart", p.DateStart).set("dateEnd", p.DateEnd).set("limit", p.Limit).set("offset", p.Offset)))
}

// Statuses returns the booking status dictionary.
// GET /api/booking/bookings/statuses
func (s *BookingService) Statuses(ctx context.Context) ([]BookingStatus, error) {
	return call[[]BookingStatus](ctx, s.c, get("api/booking/bookings/statuses", nil))
}

// Create books a room. The body is a one-element array.
// POST /api/booking/bookings/
func (s *BookingService) Create(ctx context.Context, b BookingCreate) error {
	return exec(ctx, s.c, post("api/booking/bookings/", []BookingCreate{b}))
}

// Update edits a booking; allowed for the owner in [BookingStatusReturned].
// The body is a one-element array.
// PATCH /api/booking/bookings/
func (s *BookingService) Update(ctx context.Context, b BookingUpdate) error {
	return exec(ctx, s.c, patch("api/booking/bookings/", []BookingUpdate{b}))
}

// Delete cancels a booking.
// DELETE /api/booking/bookings/{booking_id}
func (s *BookingService) Delete(ctx context.Context, bookingID int64) error {
	return exec(ctx, s.c, del("api/booking/bookings/"+id(bookingID), nil))
}

// User returns the booking service's data about the current user; nil when it has none.
// GET /api/booking/users/status
func (s *BookingService) User(ctx context.Context) (*BookingUser, error) {
	return call[*BookingUser](ctx, s.c, get("api/booking/users/status", nil))
}

// PopularEventNames returns the user's most frequent event names, typically 2.
// GET /api/booking/users/event/mostPopular
func (s *BookingService) PopularEventNames(ctx context.Context, counts int) ([]string, error) {
	return call[[]string](ctx, s.c, get("api/booking/users/event/mostPopular", q().set("counts", counts)))
}
