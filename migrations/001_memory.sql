DO $$ BEGIN
  IF to_regclass('buddy_entries') IS NOT NULL AND to_regclass('cully_entries') IS NULL THEN
    ALTER TABLE buddy_entries RENAME TO cully_entries;
  ELSIF to_regclass('buddy_entries') IS NOT NULL AND to_regclass('cully_entries') IS NOT NULL THEN
    RAISE EXCEPTION 'Both source and destination memory tables exist; reconcile before migrating';
  END IF;
END $$;

-- Reuse the original indexes after the table rename, especially the HNSW
-- index, instead of rebuilding duplicate indexes over the existing data.
DO $$ DECLARE suffix text; BEGIN
  FOREACH suffix IN ARRAY ARRAY['search_idx','project_time_idx','type_idx','section_time_idx','owner_project_time_idx','embedding_idx'] LOOP
    IF to_regclass('buddy_entries_' || suffix) IS NOT NULL AND to_regclass('cully_entries_' || suffix) IS NULL THEN
      EXECUTE format('ALTER INDEX %I RENAME TO %I', 'buddy_entries_' || suffix, 'cully_entries_' || suffix);
    END IF;
  END LOOP;
  IF to_regclass('cully_entries') IS NOT NULL AND EXISTS(SELECT 1 FROM pg_constraint WHERE conrelid=to_regclass('cully_entries') AND conname='buddy_entries_section_check') THEN
    ALTER TABLE cully_entries RENAME CONSTRAINT buddy_entries_section_check TO cully_entries_section_check;
  END IF;
END $$;

CREATE EXTENSION IF NOT EXISTS vector;
CREATE TABLE IF NOT EXISTS cully_entries (
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
  ) STORED,
  embedding vector(1536)
);
ALTER TABLE cully_entries ADD COLUMN IF NOT EXISTS embedding vector(1536);
ALTER TABLE cully_entries ADD COLUMN IF NOT EXISTS owner_subject text;
ALTER TABLE cully_entries ADD COLUMN IF NOT EXISTS section text NOT NULL DEFAULT 'company';
ALTER TABLE cully_entries ADD COLUMN IF NOT EXISTS category text;
ALTER TABLE cully_entries ALTER COLUMN project_url DROP NOT NULL;
DO $$ BEGIN
  ALTER TABLE cully_entries ADD CONSTRAINT cully_entries_section_check CHECK (section IN ('personal','company'));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
CREATE INDEX IF NOT EXISTS cully_entries_search_idx ON cully_entries USING gin(search_vector);
CREATE INDEX IF NOT EXISTS cully_entries_project_time_idx ON cully_entries(project_url, occurred_at DESC);
CREATE INDEX IF NOT EXISTS cully_entries_type_idx ON cully_entries(entry_type, occurred_at DESC);
CREATE INDEX IF NOT EXISTS cully_entries_section_time_idx ON cully_entries(section, occurred_at DESC);
CREATE INDEX IF NOT EXISTS cully_entries_owner_project_time_idx ON cully_entries(owner_subject, project_url, occurred_at DESC);
CREATE INDEX IF NOT EXISTS cully_entries_embedding_idx ON cully_entries USING hnsw (embedding vector_cosine_ops) WHERE embedding IS NOT NULL;

ALTER TABLE cully_entries ALTER COLUMN owner_subject SET NOT NULL;
ALTER TABLE cully_entries ALTER COLUMN theme SET DEFAULT 'cully';
CREATE INDEX IF NOT EXISTS cully_entries_owner_time_idx ON cully_entries(owner_subject, occurred_at DESC, id);
CREATE INDEX IF NOT EXISTS cully_entries_owner_section_time_idx ON cully_entries(owner_subject, section, occurred_at DESC, id);
CREATE TABLE IF NOT EXISTS cully_mem0_jobs (
 owner_subject text NOT NULL,
 entry_id uuid NOT NULL,
 generation bigint NOT NULL DEFAULT 1,
 attempts integer NOT NULL DEFAULT 0,
 next_attempt timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(owner_subject,entry_id)
);
