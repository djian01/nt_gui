package testengine

import (
	"database/sql"
	"errors"
)

// HTTP records created before the recording option omitted type and were always recorded.
func recordingEnabled(c Config) bool { return c.Type == "" || c.Recording }

// Unrecorded probes live in an in-memory SQLite store using the same indexed
// chart summaries. No metadata or samples reach the durable database until Record is enabled.
// Recorded tests also keep a lightweight current-session summary here for shared pagination.
func (r *Runner) prepareLiveResults(s Session) error {
	if r.liveResults == nil {
		db, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			return err
		}
		db.SetMaxOpenConns(1)
		store := &Store{db: db}
		if err := store.initialize(); err != nil {
			db.Close()
			return err
		}
		r.liveResults = store
	}
	if err := r.liveResults.Save(s, nil); err != nil {
		return err
	}
	if !recordingEnabled(s.Config) {
		r.transient[s.ID] = true
	}
	return nil
}

func (r *Runner) removeLiveResults(id string) {
	if r.liveResults != nil {
		// Internal removal also handles an aborted batch whose copy is running.
		_, _ = r.liveResults.db.Exec("DELETE FROM tests WHERE id=?", id)
		_, _ = r.liveResults.db.Exec("DELETE FROM current_tests WHERE id=?", id)
		delete(r.transient, id)
	}
}

func (r *Runner) resultsStore(id string) *Store {
	if r.transient[id] {
		return r.liveResults
	}
	return r.store
}

func (r *Runner) saveProbe(s Session, sample *Sample) error {
	var liveTx *sql.Tx
	if r.liveResults != nil {
		var err error
		liveTx, err = r.liveResults.db.Begin()
		if err != nil {
			return err
		}
		defer liveTx.Rollback()
		liveSample := sample
		if !r.transient[s.ID] {
			liveSample = nil
		}
		if err := saveTest(liveTx, s, liveSample); err != nil {
			return err
		}
	}
	if recordingEnabled(s.Config) {
		if err := r.store.Save(s, sample); err != nil {
			return err
		}
	}
	if liveTx != nil {
		return liveTx.Commit()
	}
	return nil
}

func (r *Runner) Timeline(id string, from, to int64) (Timeline, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.resultsStore(id).Timeline(id, from, to)
}

// Recording is one-way for a running test. Earlier live
// samples remain visible but only future probes are written to disk.
func (r *Runner) Record(id string) (Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.sessions[id]
	if s == nil || !s.Running {
		return Session{}, errors.New("Recording requires a running test")
	}
	if recordingEnabled(s.Config) {
		return s.Session, nil
	}
	next := s.Session
	next.Config.Recording = true
	next.Revision++
	if err := r.saveProbe(next, nil); err != nil {
		return Session{}, err
	}
	s.Session = next
	r.emitLocked(s)
	return next, nil
}
