package myitmo

import (
	"context"
	"iter"
	"time"

	"github.com/kewldan/go-itmo/internal/jsonx"
)

// AssistantService is the AI assistant chat (/api/assistant).
//
// The assistant is a separate backend: its JSON is not wrapped in the
// standard envelope. Chat calls need a workspace id, which is published
// in the feature flag my-itmo.web.ai-assistant.workspace-id.
type AssistantService struct{ c *Client }

// Assistant message roles.
const (
	AssistantRoleUser      = "user"
	AssistantRoleAssistant = "assistant"
)

// Assistant feedback ratings.
const (
	AssistantRatingUp   = "up"
	AssistantRatingDown = "down"
)

// Chat stream event types (AssistantChatEvent.Type).
const (
	// AssistantEventSession carries {"session_id": "..."}: the new or current session.
	AssistantEventSession = "session"
	// AssistantEventMessage is a text delta of the answer; see [AssistantChatEvent.Text].
	AssistantEventMessage = "message"
	// AssistantEventMessageID carries {"id": "...", "user_message_id": "..."}: server ids for feedback.
	AssistantEventMessageID = "message_id"
	// AssistantEventError is a failure; text matching "daily limit" means the daily quota is used up.
	AssistantEventError = "error"
	// AssistantEventClarification is a clarifying question appended to the answer.
	AssistantEventClarification = "clarification"
	// AssistantEventThought is a reasoning step with Title and Content.
	AssistantEventThought = "thought"
	// AssistantEventToolCall is a tool invocation; the name is in Name, Title or Content.
	AssistantEventToolCall = "tool_call"
	// AssistantEventToolResult is a tool result.
	AssistantEventToolResult = "tool_result"
	// AssistantEventSources lists cited sources in Sources or Content.
	AssistantEventSources = "sources"
	// AssistantEventTokenUsage carries an [AssistantTokenUsage] in Content.
	AssistantEventTokenUsage = "token_usage"
	// AssistantEventSessionUpdated means the session list changed (e.g. a new title).
	AssistantEventSessionUpdated = "session_updated"
)

// AssistantUser is the assistant profile of the current user.
type AssistantUser struct {
	// ConsentGivenAt is nil until the user accepts the terms ([AssistantService.AcceptConsent]).
	ConsentGivenAt *time.Time `json:"consent_given_at"`
}

// AssistantChatSession is a chat in the history sidebar.
type AssistantChatSession struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Title       string    `json:"title"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AssistantChatHistory is the message history of a session.
type AssistantChatHistory struct {
	Messages []AssistantChatMessage `json:"messages"`
}

// AssistantChatMessage is one message of a chat.
type AssistantChatMessage struct {
	ID string `json:"id"`
	// Role is AssistantRoleUser or AssistantRoleAssistant.
	Role             string `json:"role"`
	Content          string `json:"content"`
	FeedbackCategory string `json:"feedback_category"`
	FeedbackComment  string `json:"feedback_comment"`
	// FeedbackRating is AssistantRatingUp, AssistantRatingDown or empty.
	FeedbackRating string `json:"feedback_rating"`
	ErrorCode      string `json:"error_code"`
	// Sources is a list of [AssistantChatSource] or an object of them keyed by id.
	Sources    RawJSON              `json:"sources"`
	TokenUsage *AssistantTokenUsage `json:"token_usage"`
	CreatedAt  time.Time            `json:"created_at"`
}

// AssistantChatSource is a source cited by an answer; only type "web" carries a link.
type AssistantChatSource struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	URL      string `json:"url"`
	DocID    string `json:"doc_id"`
	Title    string `json:"title"`
	Filename string `json:"filename"`
}

// AssistantTokenUsage is the token accounting of an answer.
type AssistantTokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
	// Tools is a list of {name, calls, tokens} or an object of {calls, tokens} keyed by tool name.
	Tools RawJSON `json:"tools"`
}

// AssistantFeedback rates an assistant answer.
type AssistantFeedback struct {
	// Rating is AssistantRatingUp or AssistantRatingDown.
	Rating string `json:"rating"`
	// Category is one of the tags from the feature flag my-itmo.web.ai-assistant.feedback-tags.
	Category string `json:"category,omitzero"`
	// Comment is at most 1000 characters.
	Comment string `json:"comment,omitzero"`
}

// AssistantChatParams is a message to the assistant.
type AssistantChatParams struct {
	// WorkspaceID is required (feature flag my-itmo.web.ai-assistant.workspace-id).
	WorkspaceID string
	Content     string
	// SessionID continues a chat; empty starts a new one (see [AssistantEventSession]).
	SessionID string
	// EnableCitations asks for sources; usually false.
	EnableCitations bool
}

// AssistantChatEvent is one payload of the chat stream. Which fields are set
// depends on Type (see the AssistantEvent* constants).
type AssistantChatEvent struct {
	Type string `json:"type"`
	// Content is a string or an object depending on Type.
	Content RawJSON `json:"content"`
	Title   string  `json:"title"`
	Name    string  `json:"name"`
	// MessageID or ID may carry the assistant message id on message events.
	MessageID string  `json:"message_id"`
	ID        string  `json:"id"`
	Sources   RawJSON `json:"sources"`
	Data      RawJSON `json:"data"`
}

// Text returns the text of a message, clarification or error event: Content
// when it is a string, otherwise its content, text, message, question, detail
// or error field.
func (e AssistantChatEvent) Text() string {
	var s string
	if jsonx.Unmarshal(e.Content, &s) == nil {
		return s
	}
	var obj struct {
		Content  string `json:"content"`
		Text     string `json:"text"`
		Message  string `json:"message"`
		Question string `json:"question"`
		Detail   string `json:"detail"`
		Error    string `json:"error"`
	}
	if jsonx.Unmarshal(e.Content, &obj) != nil {
		return ""
	}
	for _, v := range []string{obj.Content, obj.Text, obj.Message, obj.Question, obj.Detail, obj.Error} {
		if v != "" {
			return v
		}
	}
	return ""
}

// SessionID returns content.session_id of a session event.
func (e AssistantChatEvent) SessionID() string {
	var v struct {
		SessionID string `json:"session_id"`
	}
	_ = jsonx.Unmarshal(e.Content, &v)
	return v.SessionID
}

// User returns the assistant profile; ConsentGivenAt tells whether the terms were accepted.
// GET /api/assistant/auth/users/me
func (s *AssistantService) User(ctx context.Context) (*AssistantUser, error) {
	return callRaw[*AssistantUser](ctx, s.c, get("api/assistant/auth/users/me", nil))
}

// AcceptConsent accepts the assistant terms of use, required before chatting.
// PUT /api/assistant/auth/users/me/consent
func (s *AssistantService) AcceptConsent(ctx context.Context) error {
	body := struct {
		Accepted bool `json:"accepted"`
	}{true}
	_, err := callRaw[RawJSON](ctx, s.c, put("api/assistant/auth/users/me/consent", body))
	return err
}

// Sessions returns a page of the user's chats. A typical limit is 30;
// the last page is shorter than limit.
// GET /api/assistant/chat/sessions
func (s *AssistantService) Sessions(ctx context.Context, workspaceID string, skip, limit int) ([]AssistantChatSession, error) {
	return callRaw[[]AssistantChatSession](ctx, s.c, get("api/assistant/chat/sessions",
		q().set("workspace_id", workspaceID).set("skip", skip).set("limit", limit)))
}

// Session returns the message history of a chat.
// GET /api/assistant/chat/sessions/{sessionId}
func (s *AssistantService) Session(ctx context.Context, sessionID string) (*AssistantChatHistory, error) {
	return callRaw[*AssistantChatHistory](ctx, s.c, get("api/assistant/chat/sessions/"+id(sessionID), nil))
}

// DeleteSession deletes a chat.
// DELETE /api/assistant/chat/sessions/{sessionId}
func (s *AssistantService) DeleteSession(ctx context.Context, sessionID string) error {
	_, err := callRaw[RawJSON](ctx, s.c, del("api/assistant/chat/sessions/"+id(sessionID), nil))
	return err
}

// Feedback rates an assistant message (id from AssistantChatMessage.ID or a message_id event).
// POST /api/assistant/chat/messages/{messageId}/feedback
func (s *AssistantService) Feedback(ctx context.Context, messageID string, fb AssistantFeedback) error {
	_, err := callRaw[RawJSON](ctx, s.c, post("api/assistant/chat/messages/"+id(messageID)+"/feedback", fb))
	return err
}

// Chat sends a message and streams the answer. Frames whose data is not JSON
// are skipped; a payload wrapped as {"data": {...}} is unwrapped. The stream
// ends when the server closes it.
// POST /api/assistant/chat
func (s *AssistantService) Chat(ctx context.Context, p AssistantChatParams) (iter.Seq2[AssistantChatEvent, error], error) {
	body := struct {
		WorkspaceID     string  `json:"workspace_id"`
		Content         string  `json:"content"`
		SessionID       *string `json:"session_id"`
		EnableCitations bool    `json:"enable_citations"`
	}{WorkspaceID: p.WorkspaceID, Content: p.Content, EnableCitations: p.EnableCitations}
	if p.SessionID != "" {
		body.SessionID = &p.SessionID
	}
	events, err := stream(ctx, s.c, post("api/assistant/chat", body))
	if err != nil {
		return nil, err
	}
	return func(yield func(AssistantChatEvent, error) bool) {
		for ev, err := range events {
			if err != nil {
				yield(AssistantChatEvent{}, err)
				return
			}
			var out AssistantChatEvent
			if jsonx.Unmarshal([]byte(ev.Data), &out) != nil {
				continue
			}
			if out.Type == "" && len(out.Data) > 0 {
				var inner AssistantChatEvent
				if jsonx.Unmarshal(out.Data, &inner) == nil && inner.Type != "" {
					out = inner
				}
			}
			if out.Type == "" {
				continue
			}
			if !yield(out, nil) {
				return
			}
		}
	}, nil
}
