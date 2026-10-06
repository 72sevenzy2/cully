// Package postgres implements owner-scoped source records with PostgreSQL.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mcp-runtime/cully/internal/mem0"
	"github.com/mcp-runtime/cully/internal/memory"
)

type Store struct {
	Pool        *pgxpool.Pool
	Mem0Enabled bool
	Mem0        *mem0.Client
}

const record = "to_jsonb(e) - 'owner_subject' - 'search_vector'"

func decode(data []byte) (*memory.Entry, error) {
	var e memory.Entry
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}
	e.NormalizeTimes()
	return &e, nil
}
func one(ctx context.Context, tx pgx.Tx, sql string, args ...any) (*memory.Entry, error) {
	var data []byte
	err := tx.QueryRow(ctx, sql, args...).Scan(&data)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return decode(data)
}
func (s *Store) Execute(ctx context.Context, owner string, r memory.Request) (memory.Result, error) {
	if r.Operation == "recall" {
		return s.Recall(ctx, owner, *r.Search)
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return memory.Result{}, memory.ErrUnavailable
	}
	defer tx.Rollback(ctx)
	result, err := s.execute(ctx, tx, owner, r)
	if err != nil {
		return memory.Result{}, fmt.Errorf("%w: database operation failed", memory.ErrUnavailable)
	}
	if s.Mem0Enabled && (r.Operation == "log" || (r.Operation == "update" && result.Entry != nil) || (r.Operation == "delete" && result.Deleted)) {
		entryID := ""
		if result.Entry != nil {
			entryID = result.Entry.ID
		} else {
			entryID = r.ID.EntryID
		}
		_, err = tx.Exec(ctx, `INSERT INTO cully_mem0_jobs(owner_subject,entry_id) VALUES($1,$2::uuid) ON CONFLICT(owner_subject,entry_id) DO UPDATE SET generation=cully_mem0_jobs.generation+1, attempts=0, next_attempt=now()`, owner, entryID)
		if err != nil {
			return memory.Result{}, memory.ErrUnavailable
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return memory.Result{}, memory.ErrUnavailable
	}
	return result, nil
}
func (s *Store) execute(ctx context.Context, tx pgx.Tx, owner string, r memory.Request) (memory.Result, error) {
	out := memory.Result{}
	switch r.Operation {
	case "log":
		v := r.Log
		when := time.Now()
		if v.OccurredAt != "" {
			when, _ = time.Parse(time.RFC3339, v.OccurredAt)
		}
		e, err := one(ctx, tx, `INSERT INTO cully_entries AS e (id,owner_subject,section,project_url,session_ref,category,entry_type,summary,approach,outcome,issue,learning,next_steps,assistant,tags,occurred_at) VALUES ($1::uuid,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16) RETURNING `+record, uuid.NewString(), owner, v.Section, v.ProjectURL, v.SessionRef, v.Category, v.EntryType, v.Summary, v.Approach, v.Outcome, v.Issue, v.Learning, v.NextSteps, v.Assistant, v.Tags, when)
		out.Entry = e
		return out, err
	case "get":
		e, err := one(ctx, tx, "SELECT "+record+" FROM cully_entries e WHERE owner_subject=$1 AND id=$2::uuid", owner, r.ID.EntryID)
		out.Entry = e
		return out, err
	case "delete":
		tag, err := tx.Exec(ctx, "DELETE FROM cully_entries WHERE owner_subject=$1 AND id=$2::uuid", owner, r.ID.EntryID)
		out.Deleted = tag.RowsAffected() > 0
		return out, err
	case "update":
		v := r.Update
		args := []any{owner, v.EntryID}
		assignments := []string{}
		add := func(field string, value any, cast string) {
			args = append(args, value)
			assignments = append(assignments, fmt.Sprintf("%s=$%d%s", field, len(args), cast))
		}
		fields := []struct {
			name  string
			value *string
		}{{"summary", v.Summary}, {"approach", v.Approach}, {"outcome", v.Outcome}, {"issue", v.Issue}, {"learning", v.Learning}, {"next_steps", v.NextSteps}, {"section", v.Section}, {"category", v.Category}}
		for _, f := range fields {
			if f.value != nil {
				if *f.value == "" && f.name != "summary" && f.name != "section" {
					add(f.name, nil, "")
				} else {
					add(f.name, *f.value, "")
				}
			}
		}
		if v.Tags != nil {
			add("tags", *v.Tags, "")
		}
		assignments = append(assignments, "updated_at=now()")
		e, err := one(ctx, tx, "UPDATE cully_entries e SET "+strings.Join(assignments, ",")+" WHERE owner_subject=$1 AND id=$2::uuid RETURNING "+record, args...)
		out.Entry = e
		return out, err
	case "recent":
		v := r.Recent
		where, args := filters(owner, v.ProjectURL, v.SessionRef, v.Section, v.Category, v.EntryType, "")
		args = append(args, v.Limit)
		return rows(ctx, tx, "SELECT "+record+" FROM cully_entries e WHERE "+where+fmt.Sprintf(" ORDER BY occurred_at DESC,id LIMIT $%d", len(args)), args...)
	case "search":
		v := r.Search
		where, args := filters(owner, v.ProjectURL, v.SessionRef, v.Section, v.Category, v.EntryType, v.Since)
		args = append(args, v.Query, v.Limit)
		q, count := len(args)-1, len(args)
		sql := fmt.Sprintf(`SELECT %s FROM cully_entries e WHERE %s AND search_vector @@ websearch_to_tsquery('english',$%d) ORDER BY ts_rank_cd(search_vector,websearch_to_tsquery('english',$%d)) DESC,e.occurred_at DESC,e.id LIMIT $%d`, record, where, q, q, count)
		return rows(ctx, tx, sql, args...)
	case "projects":
		v := r.Projects
		rs, err := tx.Query(ctx, `SELECT project_url,section,count(*),max(occurred_at) FROM cully_entries WHERE owner_subject=$1 AND project_url IS NOT NULL AND ($2::text IS NULL OR section=$2) GROUP BY project_url,section ORDER BY max(occurred_at) DESC,project_url LIMIT $3`, owner, v.Section, v.Limit)
		if err != nil {
			return out, err
		}
		defer rs.Close()
		out.Projects = []memory.Project{}
		for rs.Next() {
			var p memory.Project
			if err = rs.Scan(&p.ProjectURL, &p.Section, &p.EntryCount, &p.LastActivity); err != nil {
				return out, err
			}
			p.LastActivity = p.LastActivity.In(memory.IST)
			out.Projects = append(out.Projects, p)
		}
		return out, rs.Err()
	}
	return out, memory.ErrInvalid
}
func filters(owner string, project, session, section, category, kind *string, since string) (string, []any) {
	args := []any{owner}
	parts := []string{"owner_subject=$1"}
	for _, f := range []struct {
		name  string
		value *string
	}{{"project_url", project}, {"session_ref", session}, {"section", section}, {"category", category}, {"entry_type", kind}} {
		if f.value != nil {
			args = append(args, *f.value)
			parts = append(parts, fmt.Sprintf("%s=$%d", f.name, len(args)))
		}
	}
	if since != "" {
		t, _ := time.Parse(time.RFC3339, since)
		args = append(args, t)
		parts = append(parts, fmt.Sprintf("occurred_at >= $%d", len(args)))
	}
	return strings.Join(parts, " AND "), args
}
func rows(ctx context.Context, tx pgx.Tx, sql string, args ...any) (memory.Result, error) {
	out := memory.Result{Entries: []memory.Entry{}}
	rs, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return out, err
	}
	defer rs.Close()
	for rs.Next() {
		var data []byte
		if err = rs.Scan(&data); err != nil {
			return out, err
		}
		e, err := decode(data)
		if err != nil {
			return out, err
		}
		out.Entries = append(out.Entries, *e)
	}
	return out, rs.Err()
}
