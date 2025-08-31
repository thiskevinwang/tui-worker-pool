package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/chdb-io/chdb-go/chdb"
)

// ClickHouseHandler is a slog.Handler that inserts log records into ClickHouse.
type ClickHouseHandler struct {
	session *chdb.Session
}

// For how to query chdb, see: https://github.com/chdb-io/chdb-go?tab=readme-ov-file#chdb-go-cli
//
// INTERACTIVE
// $ chdb-go --path ./tmp/chdb
//
// SIMPLE
// $ chdb-go --path ./tmp/chdb "SELECT * FROM log_table ORDER BY timestamp DESC LIMIT 10;"

// SELECT message::JSON FROM log_table ORDER BY timestamp DESC LIMIT 10 FORMAT JSON

// SELECT
// 	toStartOfInterval(timestamp, INTERVAL 1 SECOND) AS bucket,
// 	count(*) AS count
// FROM log_table
// GROUP BY bucket
// ORDER BY bucket DESC
// WITH FILL
// FROM toDateTime64(now64(), 3)
// TO toDateTime64(now64() - INTERVAL 1 minute, 3)
// STEP -toIntervalMillisecond(1000);

func (h *ClickHouseHandler) InitTable() error {
	_, err := h.session.Query(`
	CREATE TABLE IF NOT EXISTS log_table
	(
		timestamp DateTime,
		message_type String,
		message JSON
	)
	ENGINE = MergeTree()
	ORDER BY timestamp`)
	return err
}

// NewClickHouseHandler creates a new ClickHouseHandler with the given session.
func NewClickHouseHandler(session *chdb.Session) *ClickHouseHandler {
	return &ClickHouseHandler{session: session}
}

// Handle inserts the log record into the ClickHouse table.
func (h *ClickHouseHandler) Handle(ctx context.Context, r slog.Record) error {
	// Prepare the data for insertion
	timestamp := r.Time.Format("2006-01-02 15:04:05")
	messageType := r.Level.String()

	// Serialize the message and attributes as JSON
	messageData := map[string]interface{}{
		"message": r.Message,
	}
	r.Attrs(func(a slog.Attr) bool {
		messageData[a.Key] = a.Value.Any()
		return true
	})
	messageJSON, err := json.Marshal(messageData)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Insert into ClickHouse
	query := fmt.Sprintf(`
        INSERT INTO log_table (timestamp, message_type, message)
        VALUES ('%s', '%s', '%s'::JSON)
    `, timestamp, messageType, messageJSON)

	_, err = h.session.Query(query)
	if err != nil {
		return fmt.Errorf("failed to insert log: %w", err)
	}

	return nil
}

// WithAttrs returns a new handler with additional attributes (not implemented for simplicity).
func (h *ClickHouseHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For simplicity, return the same handler. In a full implementation, you'd handle attributes.
	return h
}

// WithGroup returns a new handler with a group (not implemented for simplicity).
func (h *ClickHouseHandler) WithGroup(name string) slog.Handler {
	// For simplicity, return the same handler.
	return h
}

// Enabled checks if the level is enabled (always true for simplicity).
func (h *ClickHouseHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}
