package myitmo

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kewldan/go-itmo/internal/jsonx"
	"github.com/kewldan/go-itmo/internal/rest"
)

// requestsExecLoose sends r for a route whose response body is not known to be an
// envelope: any 2xx succeeds unless the body is an envelope with a non-zero
// error_code.
func requestsExecLoose(ctx context.Context, c *Client, r *rest.Request) error {
	resp, err := c.rest.Send(ctx, r) //nolint:bodyclose // closed by errorFrom or rest.ReadAll
	if err != nil {
		return err
	}
	if resp.StatusCode/100 != 2 {
		return errorFrom(r, resp)
	}
	body, err := rest.ReadAll(resp, maxBody)
	if err != nil {
		return fmt.Errorf("myitmo: read %s %s: %w", r.Method, r.Path, err)
	}
	var env envelope[jsonValue]
	if jsonx.Unmarshal(body, &env) == nil && env.ErrorCode != 0 {
		return &Error{Method: r.Method, Path: r.Path, StatusCode: http.StatusOK, Code: env.ErrorCode, Message: env.ErrorMessage, Details: details(env.Result)}
	}
	return nil
}
