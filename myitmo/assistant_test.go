package myitmo_test

import (
	"iter"
	"net/http"
	"testing"

	"github.com/kewldan/go-itmo/myitmo"
)

func TestAssistantUserAndConsent(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusOK, `{"id":"u-1","consent_given_at":"2026-03-01T10:00:00Z"}`)
	u := must[*myitmo.AssistantUser](t)(c.Assistant.User(t.Context()))
	f.expect("GET", "/api/assistant/auth/users/me")
	if u.ConsentGivenAt == nil || u.ConsentGivenAt.Year() != 2026 {
		t.Fatalf("user = %+v", u)
	}

	f.reply(http.StatusOK, `{"consent_given_at":null}`)
	if u = must[*myitmo.AssistantUser](t)(c.Assistant.User(t.Context())); u.ConsentGivenAt != nil {
		t.Fatalf("user = %+v", u)
	}

	f.reply(http.StatusOK, `{"ok":true}`)
	if err := c.Assistant.AcceptConsent(t.Context()); err != nil {
		t.Fatal(err)
	}
	f.expect("PUT", "/api/assistant/auth/users/me/consent").sameJSON(t, `{"accepted":true}`)
}

func TestAssistantSessions(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusOK, `[{"id":"s-1","workspace_id":"ws-1","title":"Exam dates","created_at":"2026-03-01T10:00:00Z","updated_at":"2026-03-02T11:00:00Z"}]`)
	list := must[[]myitmo.AssistantChatSession](t)(c.Assistant.Sessions(t.Context(), "ws-1", 30, 30))
	r := f.expect("GET", "/api/assistant/chat/sessions")
	if r.Query.Get("workspace_id") != "ws-1" || r.Query.Get("skip") != "30" || r.Query.Get("limit") != "30" {
		t.Fatalf("query = %v", r.Query)
	}
	if len(list) != 1 || list[0].ID != "s-1" || list[0].Title != "Exam dates" || list[0].UpdatedAt.Day() != 2 {
		t.Fatalf("sessions = %+v", list)
	}
}

func TestAssistantSessionHistory(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusOK, `{"messages":[
		{"id":"m-1","role":"user","content":"When is the exam?","created_at":"2026-03-01T10:00:00Z"},
		{"id":"m-2","role":"assistant","content":"On Monday.","feedback_rating":"up","feedback_category":null,"feedback_comment":null,"error_code":null,
		 "sources":{"src-1":{"type":"web","url":"https://example.org","title":"Schedule"}},
		 "token_usage":{"input_tokens":10,"output_tokens":5,"total_tokens":15,"tools":[{"name":"search","calls":1,"tokens":3}]},
		 "created_at":"2026-03-01T10:00:05Z"}]}`)
	h := must[*myitmo.AssistantChatHistory](t)(c.Assistant.Session(t.Context(), "s 1"))
	f.expect("GET", "/api/assistant/chat/sessions/s%201")
	if len(h.Messages) != 2 {
		t.Fatalf("history = %+v", h)
	}
	m := h.Messages[1]
	if m.Role != myitmo.AssistantRoleAssistant || m.FeedbackRating != myitmo.AssistantRatingUp || m.TokenUsage == nil || m.TokenUsage.TotalTokens != 15 || len(m.Sources) == 0 {
		t.Fatalf("message = %+v", m)
	}
	if h.Messages[0].TokenUsage != nil {
		t.Fatal("token usage of a user message must be nil")
	}
}

func TestAssistantDeleteAndFeedback(t *testing.T) {
	f, c := newFake(t)
	f.reply(http.StatusNoContent, ``)
	if err := c.Assistant.DeleteSession(t.Context(), "s-1"); err != nil {
		t.Fatal(err)
	}
	f.expect("DELETE", "/api/assistant/chat/sessions/s-1")

	f.reply(http.StatusOK, `{"status":"ok"}`)
	if err := c.Assistant.Feedback(t.Context(), "m-2", myitmo.AssistantFeedback{Rating: myitmo.AssistantRatingDown, Category: "wrong"}); err != nil {
		t.Fatal(err)
	}
	f.expect("POST", "/api/assistant/chat/messages/m-2/feedback").sameJSON(t, `{"rating":"down","category":"wrong"}`)
}

func TestAssistantChatStream(t *testing.T) {
	f, c := newFake(t)
	body := "event: message\ndata: {\"type\":\"session\",\"content\":{\"session_id\":\"s-9\"}}\n\n" +
		"data: {\"type\":\"message\",\"content\":\"Hel\"}\n\n" +
		"data: not json\n\n" +
		"data: {\"data\":{\"type\":\"message\",\"content\":{\"text\":\"lo\",\"message_id\":\"m-9\"}}}\n\n" +
		"data: {\"type\":\"message_id\",\n" +
		"data: \"content\":{\"id\":\"m-9\",\"user_message_id\":\"m-8\"}}\n\n" +
		"data: {\"type\":\"error\",\"content\":{\"detail\":\"Daily limit reached\"}}\n\n"
	f.replyWith(http.StatusOK, http.Header{"Content-Type": {"text/event-stream"}}, body)
	events := must[iter.Seq2[myitmo.AssistantChatEvent, error]](t)(
		c.Assistant.Chat(t.Context(), myitmo.AssistantChatParams{WorkspaceID: "ws-1", Content: "Hello"}))
	var got []myitmo.AssistantChatEvent
	for ev, err := range events {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, ev)
	}
	r := f.expect("POST", "/api/assistant/chat")
	r.sameJSON(t, `{"workspace_id":"ws-1","content":"Hello","session_id":null,"enable_citations":false}`)
	if r.Header.Get("Accept") != "text/event-stream" {
		t.Fatalf("Accept = %q", r.Header.Get("Accept"))
	}
	if len(got) != 5 {
		t.Fatalf("events = %+v", got)
	}
	if got[0].Type != myitmo.AssistantEventSession || got[0].SessionID() != "s-9" {
		t.Fatalf("session event = %+v", got[0])
	}
	if got[1].Text()+got[2].Text() != "Hello" {
		t.Fatalf("text = %q + %q", got[1].Text(), got[2].Text())
	}
	if got[3].Type != myitmo.AssistantEventMessageID || got[4].Type != myitmo.AssistantEventError || got[4].Text() != "Daily limit reached" {
		t.Fatalf("events = %+v", got[3:])
	}

	f.replyWith(http.StatusOK, http.Header{"Content-Type": {"text/event-stream"}}, "")
	for range must[iter.Seq2[myitmo.AssistantChatEvent, error]](t)(
		c.Assistant.Chat(t.Context(), myitmo.AssistantChatParams{WorkspaceID: "ws-1", Content: "More", SessionID: "s-9", EnableCitations: true})) {
	}
	f.expect("POST", "/api/assistant/chat").sameJSON(t, `{"workspace_id":"ws-1","content":"More","session_id":"s-9","enable_citations":true}`)
}
