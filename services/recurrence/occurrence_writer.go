// Package recurrence — occurrence_writer.go
//
// OccurrenceWriter is a skeleton service that coordinates event expansion.
// In Phase 3 it will be wired to a repository and will write the returned
// Occurrence slices to the event_occurrence table via espyna use cases.
// For now it only returns []Occurrence to the caller and performs no I/O.
package recurrence

import (
	"fmt"
	"time"
)

const (
	// DefaultHorizon is the default look-ahead window for occurrence expansion.
	DefaultHorizon = 2 * 365 * 24 * time.Hour // ~2 years

	// DefaultBatchSize is the default number of occurrences to process per batch.
	DefaultBatchSize = 500
)

// OccurrenceWriterConfig controls expansion behaviour.
type OccurrenceWriterConfig struct {
	// Horizon is how far ahead to expand occurrences from dtStart.
	// Defaults to DefaultHorizon (~2 years) when zero.
	Horizon time.Duration

	// BatchSize is the maximum number of occurrences to write per persistence
	// batch. Defaults to DefaultBatchSize when zero.
	// (Unused in the current skeleton — reserved for Phase 3.)
	BatchSize int
}

// DefaultConfig returns an OccurrenceWriterConfig pre-filled with sensible
// production defaults.
func DefaultConfig() OccurrenceWriterConfig {
	return OccurrenceWriterConfig{
		Horizon:   DefaultHorizon,
		BatchSize: DefaultBatchSize,
	}
}

// OccurrenceWriter manages the lifecycle of occurrence expansion for a single
// event. It is stateless with respect to individual events and safe for
// concurrent use.
type OccurrenceWriter struct {
	expander *Expander
	config   OccurrenceWriterConfig
}

// NewOccurrenceWriter creates an OccurrenceWriter with the given config.
// Zero-value fields in config are replaced with defaults.
func NewOccurrenceWriter(config OccurrenceWriterConfig) *OccurrenceWriter {
	if config.Horizon <= 0 {
		config.Horizon = DefaultHorizon
	}
	if config.BatchSize <= 0 {
		config.BatchSize = DefaultBatchSize
	}
	return &OccurrenceWriter{
		expander: NewExpander(),
		config:   config,
	}
}

// ExpandEvent expands a single event's recurrence rule into a slice of
// Occurrence values.
//
// Parameters:
//   - eventID: the event's identifier (written into each Occurrence).
//   - workspaceID: the workspace the event belongs to.
//   - rruleString: RFC 5545 RRULE body (no "RRULE:" prefix).
//   - exdateString: comma-separated ISO 8601 dates/timestamps to exclude.
//   - dtStart: the event's start datetime (UTC).
//   - duration: the length of each occurrence (end = start + duration).
//
// The caller is responsible for persisting the returned occurrences.
// Phase 3 will introduce a Repo interface and move persistence here.
func (w *OccurrenceWriter) ExpandEvent(
	eventID string,
	workspaceID string,
	rruleString string,
	exdateString string,
	dtStart time.Time,
	duration time.Duration,
) ([]Occurrence, error) {
	if eventID == "" {
		return nil, fmt.Errorf("occurrence_writer: eventID is required")
	}
	if workspaceID == "" {
		return nil, fmt.Errorf("occurrence_writer: workspaceID is required")
	}

	exdates, err := ParseExdates(exdateString)
	if err != nil {
		return nil, fmt.Errorf("occurrence_writer: %w", err)
	}

	occs, err := w.expander.ExpandRRule(
		rruleString,
		dtStart,
		duration,
		w.config.Horizon,
		exdates,
	)
	if err != nil {
		return nil, fmt.Errorf("occurrence_writer: %w", err)
	}

	// Stamp each occurrence with the event and workspace identifiers.
	for i := range occs {
		occs[i].EventID = eventID
		occs[i].WorkspaceID = workspaceID
	}

	return occs, nil
}
