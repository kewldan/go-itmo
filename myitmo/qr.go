package myitmo

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kewldan/go-itmo/internal/rest"
)

// QRService is the digital building pass served by qr.itmo.su.
type QRService struct{ c *Client }

// Pass is the building pass QR payload.
type Pass struct {
	// Hex is the pass in hexadecimal. The QR code encodes this text itself
	// (byte mode, error correction L), not the decoded bytes.
	Hex string `json:"qr_hex"`
}

// Pass returns the current building pass. The service has its own response
// format ({"response": {...}}) and needs the MyITMO access token.
func (s *QRService) Pass(ctx context.Context) (*Pass, error) {
	target, err := s.c.qrBase.Parse("v1/user/pass")
	if err != nil {
		return nil, fmt.Errorf("myitmo: qr url: %w", err)
	}
	out, err := callRaw[struct {
		Response *Pass `json:"response"`
	}](ctx, s.c, &rest.Request{Method: http.MethodGet, Path: target.String()})
	if err != nil {
		return nil, err
	}
	if out.Response == nil {
		return nil, &Error{Method: http.MethodGet, Path: "v1/user/pass", StatusCode: http.StatusOK, Message: "empty pass response"}
	}
	return out.Response, nil
}
