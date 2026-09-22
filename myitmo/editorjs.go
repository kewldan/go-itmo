package myitmo

// EditorJS is a rich-text document in the Editor.js format
// (https://editorjs.io). Services store it as is and send it back.
type EditorJS struct {
	// Time is the save time in Unix milliseconds, when the editor set it.
	Time    int64           `json:"time,omitzero"`
	Blocks  []EditorJSBlock `json:"blocks"`
	Version string          `json:"version,omitzero"`
}

// EditorJSBlock is one block of an [EditorJS] document.
type EditorJSBlock struct {
	ID string `json:"id,omitzero"`
	// Type is the block tool: "header", "list", "paragraph" were seen.
	Type string `json:"type"`
	// Data depends on Type, e.g. {"text": "..."} for a paragraph,
	// {"text": "...", "level": 2} for a header, {"style": "ordered", "items": [...]} for a list.
	Data RawJSON `json:"data,omitzero"`
}

// IsEmpty reports whether d is nil or has no blocks.
func (d *EditorJS) IsEmpty() bool { return d == nil || len(d.Blocks) == 0 }
