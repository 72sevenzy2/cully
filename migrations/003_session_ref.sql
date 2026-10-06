ALTER TABLE cully_entries ADD COLUMN session_ref text;
CREATE INDEX cully_entries_owner_session_time_idx ON cully_entries(owner_subject, session_ref, occurred_at DESC) WHERE session_ref IS NOT NULL;
