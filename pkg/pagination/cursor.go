package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// Cursor represents a pagination cursor
type Cursor struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Offset    int       `json:"offset"`
}

// Encode encodes cursor to base64 string
func (c *Cursor) Encode() string {
	data, _ := json.Marshal(c)
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeCursor decodes cursor from base64 string
func DecodeCursor(cursorStr string) (*Cursor, error) {
	data, err := base64.URLEncoding.DecodeString(cursorStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor format: %w", err)
	}

	var cursor Cursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, fmt.Errorf("invalid cursor data: %w", err)
	}

	return &cursor, nil
}

// CursorPaginator handles cursor-based pagination
type CursorPaginator struct {
	Limit  int
	Cursor *Cursor
}

// NewCursorPaginator creates a new cursor paginator
func NewCursorPaginator(limit int, cursorStr string) (*CursorPaginator, error) {
	cp := &CursorPaginator{
		Limit: limit,
	}

	if cursorStr != "" {
		cursor, err := DecodeCursor(cursorStr)
		if err != nil {
			return nil, err
		}
		cp.Cursor = cursor
	}

	return cp, nil
}

// GetLimit returns the limit
func (cp *CursorPaginator) GetLimit() int {
	if cp.Limit <= 0 {
		return 20 // Default limit
	}
	if cp.Limit > 100 {
		return 100 // Max limit
	}
	return cp.Limit
}

// GetOffset returns the offset from cursor
func (cp *CursorPaginator) GetOffset() int {
	if cp.Cursor == nil {
		return 0
	}
	return cp.Cursor.Offset
}

// CreateNextCursor creates next cursor from last item
func CreateNextCursor(id string, timestamp time.Time, offset int) string {
	cursor := &Cursor{
		ID:        id,
		Timestamp: timestamp,
		Offset:    offset,
	}
	return cursor.Encode()
}

// ParseCursorFromRequest parses cursor from query parameter
func ParseCursorFromRequest(cursorStr string) (*Cursor, error) {
	if cursorStr == "" {
		return nil, nil
	}
	return DecodeCursor(cursorStr)
}
