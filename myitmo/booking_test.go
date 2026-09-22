package myitmo_test

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestBookingRequests(t *testing.T) {
	day := myitmo.NewDate(2026, time.March, 2)
	tests := []struct {
		name   string
		call   func(c *myitmo.Client) error
		method string
		path   string
		query  url.Values
	}{
		{"Groups", func(c *myitmo.Client) error { _, err := c.Booking.Groups(ctx); return err },
			http.MethodGet, "/api/booking/dictionary/rooms/groups", url.Values{}},
		{"Categories", func(c *myitmo.Client) error { _, err := c.Booking.Categories(ctx, 3); return err },
			http.MethodGet, "/api/booking/dictionary/rooms/categories", url.Values{"groupId": {"3"}}},
		{"Rooms", func(c *myitmo.Client) error { _, err := c.Booking.Rooms(ctx, 12); return err },
			http.MethodGet, "/api/booking/rooms/roomsInCategory", url.Values{"categoryId": {"12"}}},
		{"OccupancyDefault", func(c *myitmo.Client) error { _, err := c.Booking.Occupancy(ctx, 12, day, nil); return err },
			http.MethodGet, "/api/booking/rooms/roomBookings",
			url.Values{"status": {"1", "2", "5", "6", "8"}, "date": {"2026-03-02"}, "categoryId": {"12"}}},
		{"OccupancyNoFilter", func(c *myitmo.Client) error {
			_, err := c.Booking.Occupancy(ctx, 12, day, []myitmo.BookingStatusID{})
			return err
		}, http.MethodGet, "/api/booking/rooms/roomBookings", url.Values{"date": {"2026-03-02"}, "categoryId": {"12"}}},
		{"SearchRooms", func(c *myitmo.Client) error { _, err := c.Booking.SearchRooms(ctx, "1405"); return err },
			http.MethodGet, "/api/booking/rooms/byName", url.Values{"search": {"1405"}}},
		{"My", func(c *myitmo.Client) error {
			_, err := c.Booking.My(ctx, myitmo.BookingListParams{DateStart: day, DateEnd: day.AddDays(30), Limit: myitmo.Ptr(25), Offset: myitmo.Ptr(1)})
			return err
		}, http.MethodGet, "/api/booking/bookings/my",
			url.Values{"dateStart": {"2026-03-02"}, "dateEnd": {"2026-04-01"}, "limit": {"25"}, "offset": {"1"}}},
		{"MyMinimal", func(c *myitmo.Client) error {
			_, err := c.Booking.My(ctx, myitmo.BookingListParams{DateStart: day})
			return err
		}, http.MethodGet, "/api/booking/bookings/my", url.Values{"dateStart": {"2026-03-02"}}},
		{"Statuses", func(c *myitmo.Client) error { _, err := c.Booking.Statuses(ctx); return err },
			http.MethodGet, "/api/booking/bookings/statuses", url.Values{}},
		{"Delete", func(c *myitmo.Client) error { return c.Booking.Delete(ctx, 777) },
			http.MethodDelete, "/api/booking/bookings/777", url.Values{}},
		{"User", func(c *myitmo.Client) error { _, err := c.Booking.User(ctx); return err },
			http.MethodGet, "/api/booking/users/status", url.Values{}},
		{"PopularEventNames", func(c *myitmo.Client) error { _, err := c.Booking.PopularEventNames(ctx, 2); return err },
			http.MethodGet, "/api/booking/users/event/mostPopular", url.Values{"counts": {"2"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, c := newFake(t)
			if err := tt.call(c); err != nil {
				t.Fatal(err)
			}
			r := f.expect(tt.method, tt.path)
			if !reflect.DeepEqual(r.Query, tt.query) {
				t.Errorf("query = %v, want %v", r.Query, tt.query)
			}
			if tt.method == http.MethodGet && len(r.Body) != 0 {
				t.Errorf("unexpected body %q", r.Body)
			}
		})
	}
}

func TestBookingCreate(t *testing.T) {
	f, c := newFake(t)
	start := time.Date(2026, time.March, 2, 7, 30, 0, 0, time.UTC) // 10:30 MSK
	err := c.Booking.Create(ctx, myitmo.BookingCreate{
		Name:         "Study group",
		Participants: 4,
		ContactPhone: "+7 (900) 000-00-00",
		CoBookers:    []int64{100001},
		Start:        start,
		End:          start.Add(90 * time.Minute),
		RoomID:       55,
		TechSupport:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	r := f.expect(http.MethodPost, "/api/booking/bookings/")
	if ct := r.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
	r.sameJSON(t, `[{"name":"Study group","additional_info":"","participants":4,
		"contact_phone":"+7 (900) 000-00-00","event_id":null,"co_bookers":[100001],
		"start_datetime":"2026-03-02 10:30","end_datetime":"2026-03-02 12:00",
		"room_id":55,"equipment":[],"tech_support":true}]`)
}

func TestBookingUpdate(t *testing.T) {
	f, c := newFake(t)
	start := time.Date(2026, time.March, 2, 18, 0, 0, 0, myitmo.MSK)
	err := c.Booking.Update(ctx, myitmo.BookingUpdate{
		BookingID:      777,
		RoomID:         55,
		Start:          start,
		End:            start.Add(time.Hour),
		Name:           "Seminar",
		Participants:   10,
		ContactPhone:   "+7 (900) 000-00-00",
		AdditionalInfo: "projector",
		Equipment:      []myitmo.BookingEquipment{{EquipmentID: 9, EquipmentName: "Projector", Count: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	f.expect(http.MethodPatch, "/api/booking/bookings/").sameJSON(t, `[{"booking_id":777,"room_id":55,
		"start_datetime":"2026-03-02 18:00","end_datetime":"2026-03-02 19:00",
		"event_name":"Seminar","name":"Seminar","participants":10,
		"contact_phone":"+7 (900) 000-00-00","additional_info":"projector","co_bookers":[],
		"equipment":[{"equipment_id":9,"equipment_name":"Projector","count":1}],"tech_support":false}]`)
}

func TestBookingCreateError(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusOK, `{"error_code":5,"error_message":"Room is busy","result":null}`)
	err := c.Booking.Create(ctx, myitmo.BookingCreate{Start: time.Now(), End: time.Now()})
	if myitmo.ErrorCode(err) != 5 {
		t.Fatalf("err = %v", err)
	}
}

func TestBookingDecode(t *testing.T) {
	f, c := newFake(t)
	f.result(`[{"category_id":12,"category_name":"Coworking","min_days":null,"max_days":14}]`)
	cats := must[[]myitmo.BookingCategory](t)(c.Booking.Categories(ctx, 3))
	if len(cats) != 1 || cats[0].MinDays != nil || *cats[0].MaxDays != 14 {
		t.Errorf("categories = %+v", cats)
	}

	f.result(`[{"room_id":55,"room_name":"1405","building_id":2,"photo_link":null,"floor":4,"area":"30 m2",
		"min_cap":2,"max_cap":12,"address":"Example st. 1",
		"equipment":[{"equipment_id":9,"equipment_name":"Projector","count":2}],
		"group":{"group_id":1},"category":{"category_id":12}}]`)
	rooms := must[[]myitmo.BookingRoom](t)(c.Booking.Rooms(ctx, 12))
	if len(rooms) != 1 || *rooms[0].BuildingID != 2 || rooms[0].Group.GroupID != 1 ||
		rooms[0].Equipment[0].Count != 2 || string(rooms[0].Floor) != "4" {
		t.Errorf("rooms = %+v", rooms)
	}

	f.result(`[{"room_id":55,"bookings":[{"booking_id":1,"booking_name":"","start_datetime":"2026-03-02T10:00:00+03:00",
		"end_datetime":"2026-03-02T11:30:00+03:00","owner_isu":null,"owner_fio":"Test User","status":{"status_id":1}}]}]`)
	occ := must[[]myitmo.BookingRoomOccupancy](t)(c.Booking.Occupancy(ctx, 12, myitmo.NewDate(2026, 3, 2), nil))
	b := occ[0].Bookings[0]
	if b.Status.StatusID != myitmo.BookingStatusApproved || b.OwnerISU != nil || b.EndDatetime.Sub(b.StartDatetime) != 90*time.Minute {
		t.Errorf("occupancy = %+v", occ)
	}

	f.result(`{"count":1,"list":[{"booking_id":777,"name":"Seminar",
		"room":{"room_id":55,"room_name":"1405","address":"Example st. 1","group":{"group_id":1}},
		"category":{"category_id":12},"status":{"status_id":5,"status_name":"Returned"},
		"start_datetime":"2026-03-02T18:00:00+03:00","end_datetime":"2026-03-02T19:00:00+03:00",
		"owner_isu":100000,"owner_fio":"Test User","contact_phone":"+7 (900) 000-00-00","participants":10,
		"co_bookers":[{"isu":100001,"fio":"Other User"}],
		"equipment":[{"id":9,"equipment_name":"Projector","count":1}],
		"additional_info":null,"tech_support":false,"link_to_virtual_room":null,"password_for_virtual_room":null,
		"link_for_invite_to_virtual_room":null,"password_for_room":null,"admin_comment":"Fix the time","comment":null}]}`)
	page := must[*myitmo.BookingPage](t)(c.Booking.My(ctx, myitmo.BookingListParams{DateStart: myitmo.NewDate(2026, 3, 1)}))
	if page.Count != 1 {
		t.Fatalf("page = %+v", page)
	}
	bk := page.List[0]
	if bk.Status.StatusID != myitmo.BookingStatusReturned || bk.Room.Address != "Example st. 1" ||
		bk.CoBookers[0].ISU != 100001 || bk.Equipment[0].ID != 9 || bk.AdminComment != "Fix the time" ||
		bk.StartDatetime.Hour() != 18 {
		t.Errorf("booking = %+v", bk)
	}

	f.result(`[{"status_id":1,"status_name":"Approved"},{"status_id":4,"status_name":"Rejected"}]`)
	st := must[[]myitmo.BookingStatus](t)(c.Booking.Statuses(ctx))
	if len(st) != 2 || st[1].StatusID != myitmo.BookingStatusRejected {
		t.Errorf("statuses = %+v", st)
	}

	f.result(`{"phone_number":"+7 (900) 000-00-00"}`)
	if u := must[*myitmo.BookingUser](t)(c.Booking.User(ctx)); u.PhoneNumber != "+7 (900) 000-00-00" {
		t.Errorf("user = %+v", u)
	}
	f.result(`null`)
	if u := must[*myitmo.BookingUser](t)(c.Booking.User(ctx)); u != nil {
		t.Errorf("user = %+v, want nil", u)
	}

	f.result(`["Seminar","Study group"]`)
	if names := must[[]string](t)(c.Booking.PopularEventNames(ctx, 2)); len(names) != 2 {
		t.Errorf("names = %v", names)
	}

	f.result(`[{"group_id":1,"group_name":"Coworking"}]`)
	if g := must[[]myitmo.BookingGroup](t)(c.Booking.Groups(ctx)); g[0].GroupName != "Coworking" {
		t.Errorf("groups = %+v", g)
	}
}
