# go-itmo agent guide

Working agreement for anyone (human or agent) changing this repository. It
states the rules that are not obvious from the code. User-facing documentation
lives in `README.md`.

## What this is

An unofficial Go client for ITMO University services:

| Package | Service | Auth |
|---|---|---|
| `itmoid` | ITMO.ID (Keycloak realm `itmo`) | password form, OTP, SSO cookies, PKCE, refresh, token exchange |
| `myitmo` | MyITMO cabinet `https://my.itmo.ru/api/...` and the pass service `qr.itmo.su` | ITMO.ID access token (client `student-personal-cabinet`) |
| `bars` | BARS grade book `https://bars.itmo.ru/backend/rest/...` | own `Bearer` session from an ITMO.ID code (client `bars`) |
| `internal/rest` | HTTP plumbing: requests, multipart, downloads, SSE | — |
| `internal/jsonx` | JSON v2 options: lenient timestamps (RFC 3339, no seconds, no offset, unix ms) | — |

The goal is **complete route coverage** of MyITMO (student and staff
sections) and BARS. New routes are added here as the services grow.

## Hard rules

- Go 1.27, `encoding/json/v2`. Dependencies: `golang.org/x/oauth2`,
  `golang.org/x/net` only. Anything else needs a strong reason.
- JSON tags match the wire exactly (json/v2 is case-sensitive). One Go field
  per wire field; never rename a tag to "fix" server spelling.
- Nullability: pointers (`*int64`, `*float64`, `*bool`, `*time.Time`, `*T`)
  only where null or absence is meaningful and was seen or is plausible;
  strings stay `string` (null decodes to ""). Timestamps are `time.Time`,
  calendar dates are `myitmo.Date`.
- Values the server sends as either a number or a string use the shared
  lenient types in `myitmo/flex.go`: `FlexID` (string id), `FlexInt`,
  `FlexNumber`, `FlexBool` (bool or 0/1). Do not add per-service copies.
- Unknown shapes are `RawJSON` with a comment saying what is known. Never
  `any`, never `map[string]any` for a structured object. Free-form server
  maps may be `map[string]T` when the value type is known.
- Enum-like values get typed constants only when the meaning is known; the
  field type stays the wire type so unknown values still decode.
- Tokens, cookies, passwords, codes, ISU numbers of real people and grades
  never appear in logs, errors, tests, fixtures or docs. `rest.Client`
  logs method, path, status and latency only. BARS errors carry no body.
- Tokens are sent only to configured hosts (`myitmo.Client.prepare`).
- No global state; every client is safe for concurrent use.
- No AI attribution anywhere: commits, code, docs.
- Comments and docs describe the API itself, never where the knowledge
  about it came from.

## encoding/json/v2 in Go 1.27

- Tags are case-sensitive; a field matches only its exact wire key (there
  are keys like `HeadItmoIsu`, `Chairman`, `Type`, and one `total_сount`
  with a Cyrillic "с" in the GIA service — keep them byte-for-byte).
- `omitzero` for optional request fields. There is no `format:emitnull`: when
  the server wants an explicit `null`, use a pointer or a small type with
  `MarshalJSON`.
- Extra members are captured with `json:",embed"` on a `jsontext.Value` or
  map field; `,inline` and `,unknown` are silently treated as field names.
- `time.Time` fields decode through `internal/jsonx`, which accepts RFC 3339,
  timestamps without seconds or offset (read as Moscow time), date-only
  strings, unix milliseconds, `""` and `null`.

## myitmo conventions

One file per API group (`sport.go`, `requests_v2.go`, ...) with one service
type `XxxService struct{ c *Client }` registered in `client.go`. Methods:

```go
// Filters returns the values of the schedule filters.
func (s *SportService) Filters(ctx context.Context) (*SportFilters, error) {
	return call[*SportFilters](ctx, s.c, get("api/sport/sign/schedule/filters", nil))
}
```

- Paths are relative (`api/...`, no leading slash); path segments go through
  `id(v)` which escapes them.
- Envelope helpers in `call.go`: `call[T]` (`{error_code,error_message,result}`),
  `callData[T]` (`{code,message,data}`, schedule service), `callRaw[T]`
  (no envelope), `exec` (result ignored), `download` (file body, returns
  `*File`, caller closes), `stream` (Server-Sent Events).
- Request builders: `get(path, q().set("k", v)...)`, `post/put/patch/del(path, body)`,
  `multipart(method, path, fields, uploads...)`, `withQuery`, `withHeader`.
  `q().set` skips empty strings and nil pointers; ints and bools are always sent.
- Naming: `List`/`Get`/`Create`/`Update`/`Delete` for CRUD, otherwise the
  noun the route returns (`Specializations`, `Filters`) or the verb it
  performs (`SignIn`, `Approve`). Keep the wire term in the doc comment
  together with the HTTP method and path: `// POST /api/sport/sign/schedule/lessons`.
- More than three optional inputs → a `XxxParams` struct. Request bodies are
  structs with exact tags and `omitzero` for optional fields.
- Return `*T` for objects, `[]T` for lists, `error` alone when the result is
  meaningless.
- Staff-only routes stay in the same service; note the audience in the doc
  comment (`Staff only.`). The server answers 403 otherwise (`Error.IsForbidden`).

## Tests

- `myitmo`: external package `myitmo_test`, fake server from
  `helpers_test.go` (`newFake`, `f.result(json)`, `f.expect(method, path)`,
  `r.sameJSON`, `must`). Every method is exercised: method, escaped path,
  query and body are asserted, and a representative response decodes.
- `itmoid`: fake Keycloak over TLS in `itmoid_test.go`.
- `bars`: fake server, session renewal and period selection.
- Fixtures are synthetic. Never paste a real response without replacing
  names, ISU numbers, grades and tokens.
- Live checks live behind the `integration` build tag
  (`myitmo/integration_test.go`), call read-only endpoints only and take a
  refresh token from `ITMO_REFRESH_TOKEN_FILE` (rewritten in place, since
  ITMO.ID rotates tokens) or `ITMO_REFRESH_TOKEN`. They log counts, never
  payloads, and never run in CI. A `*DecodeError` there is a model bug; a
  `*Error` is a server answer (often 403 for staff sections).

## Build and verify

```bash
go build ./... && go vet ./...
go test -race ./...
golangci-lint run          # v2, config in .golangci.yml
go test -tags integration -count=1 ./...   # optional, needs a token
```

All three must pass before a change is done. `gofmt` is enforced by the linter.

## Git

- Conventional, imperative commit subjects (`myitmo: add dormitory payments`).
- No generated files, no IDE metadata, no real responses with personal data.
- No AI attribution or co-author trailers.
