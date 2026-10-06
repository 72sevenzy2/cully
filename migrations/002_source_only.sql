CREATE TABLE cully_entries (
  id uuid PRIMARY KEY,
  owner_subject text NOT NULL,
  section text NOT NULL DEFAULT 'company' CHECK (section IN ('personal','company')),
  project_url text,
  category text,
  theme text NOT NULL DEFAULT 'cully',
  entry_type text NOT NULL CHECK (entry_type IN ('work','issue','learning','decision')),
  summary text NOT NULL,
  approach text,
  outcome text,
  issue text,
  learning text,
  next_steps text,
  assistant text NOT NULL,
  tags text[] NOT NULL DEFAULT '{}',
  occurred_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  search_vector tsvector GENERATED ALWAYS AS (
    setweight(to_tsvector('english', coalesce(summary,'')), 'A') ||
    setweight(to_tsvector('english', coalesce(approach,'')), 'B') ||
    setweight(to_tsvector('english', coalesce(outcome,'')), 'B') ||
    setweight(to_tsvector('english', coalesce(issue,'')), 'A') ||
    setweight(to_tsvector('english', coalesce(learning,'')), 'A') ||
    setweight(to_tsvector('english', coalesce(next_steps,'')), 'C')
  ) STORED
);
CREATE INDEX cully_entries_search_idx ON cully_entries USING gin(search_vector);
CREATE INDEX cully_entries_project_time_idx ON cully_entries(project_url, occurred_at DESC);
CREATE INDEX cully_entries_type_idx ON cully_entries(entry_type, occurred_at DESC);
CREATE INDEX cully_entries_section_time_idx ON cully_entries(section, occurred_at DESC);
CREATE INDEX cully_entries_owner_project_time_idx ON cully_entries(owner_subject, project_url, occurred_at DESC);
CREATE INDEX cully_entries_owner_time_idx ON cully_entries(owner_subject, occurred_at DESC, id);
CREATE INDEX cully_entries_owner_section_time_idx ON cully_entries(owner_subject, section, occurred_at DESC, id);
CREATE TABLE cully_mem0_jobs (
 owner_subject text NOT NULL,
 entry_id uuid NOT NULL,
 generation bigint NOT NULL DEFAULT 1,
 attempts integer NOT NULL DEFAULT 0,
 next_attempt timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(owner_subject,entry_id)
);
