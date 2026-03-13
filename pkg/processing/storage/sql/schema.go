package sql

import (
	"context"
	"database/sql"
)

const schema = `
CREATE TABLE IF NOT EXISTS process_definitions (
    key_id          INTEGER PRIMARY KEY,
    bpmn_process_id TEXT    NOT NULL,
    name            TEXT    NOT NULL DEFAULT '',
    version         INTEGER NOT NULL,
    content_hash    BLOB    NOT NULL,
    content         BLOB    NOT NULL,
    deployed_at     INTEGER NOT NULL,
    deleted_at      INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_pd_process_id ON process_definitions(bpmn_process_id);
CREATE INDEX IF NOT EXISTS idx_pd_hash ON process_definitions(bpmn_process_id, content_hash);
CREATE INDEX IF NOT EXISTS idx_pd_version ON process_definitions(bpmn_process_id, version);

CREATE TABLE IF NOT EXISTS process_instances (
    key_id               INTEGER PRIMARY KEY,
    process_definition_key INTEGER NOT NULL,
    bpmn_process_id      TEXT    NOT NULL,
    parent_key           INTEGER NOT NULL DEFAULT 0,
    parent_element_key   INTEGER NOT NULL DEFAULT 0,
    state                INTEGER NOT NULL,
    created_at           INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_pi_state ON process_instances(state);

CREATE TABLE IF NOT EXISTS element_instances (
    key_id               INTEGER PRIMARY KEY,
    process_instance_key INTEGER NOT NULL,
    process_definition_key INTEGER NOT NULL,
    element_id           TEXT    NOT NULL,
    element_type         TEXT    NOT NULL,
    flow_scope_key       INTEGER NOT NULL,
    state                INTEGER NOT NULL,
    created_at           INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_ei_pi ON element_instances(process_instance_key);
CREATE INDEX IF NOT EXISTS idx_ei_state ON element_instances(process_instance_key, state);

CREATE TABLE IF NOT EXISTS variables (
    key_id               INTEGER PRIMARY KEY,
    process_instance_key INTEGER NOT NULL,
    scope_key            INTEGER NOT NULL,
    name                 TEXT    NOT NULL,
    value                BLOB    NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_var_scope_name ON variables(scope_key, name);
CREATE INDEX IF NOT EXISTS idx_var_scope ON variables(scope_key);

CREATE TABLE IF NOT EXISTS jobs (
    key_id               INTEGER PRIMARY KEY,
    process_instance_key INTEGER NOT NULL,
    element_instance_key INTEGER NOT NULL,
    process_definition_key INTEGER NOT NULL,
    type                 TEXT    NOT NULL,
    state                INTEGER NOT NULL,
    retries              INTEGER NOT NULL,
    worker               TEXT    NOT NULL DEFAULT '',
    deadline             INTEGER NOT NULL DEFAULT 0,
    error_message        TEXT    NOT NULL DEFAULT '',
    error_code           TEXT    NOT NULL DEFAULT '',
    variables            BLOB,
    created_at           INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_job_type_state ON jobs(type, state);
CREATE INDEX IF NOT EXISTS idx_job_pi ON jobs(process_instance_key);

CREATE TABLE IF NOT EXISTS timers (
    key_id               INTEGER PRIMARY KEY,
    process_instance_key INTEGER NOT NULL,
    element_instance_key INTEGER NOT NULL,
    process_definition_key INTEGER NOT NULL,
    state                INTEGER NOT NULL,
    due_date             INTEGER NOT NULL,
    repetitions          INTEGER NOT NULL DEFAULT 0,
    cycle_duration_ns    INTEGER NOT NULL DEFAULT 0,
    created_at           INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_timer_due ON timers(state, due_date);
CREATE INDEX IF NOT EXISTS idx_timer_pi ON timers(process_instance_key);

CREATE TABLE IF NOT EXISTS message_subscriptions (
    key_id               INTEGER PRIMARY KEY,
    process_instance_key INTEGER NOT NULL,
    element_instance_key INTEGER NOT NULL,
    message_name         TEXT    NOT NULL,
    correlation_key      TEXT    NOT NULL,
    state                INTEGER NOT NULL,
    created_at           INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_ms_match ON message_subscriptions(message_name, correlation_key, state);
CREATE INDEX IF NOT EXISTS idx_ms_pi ON message_subscriptions(process_instance_key);

CREATE TABLE IF NOT EXISTS message_buffer (
    key_id          INTEGER PRIMARY KEY,
    message_name    TEXT    NOT NULL,
    correlation_key TEXT    NOT NULL,
    variables       BLOB,
    expires_at      INTEGER NOT NULL,
    created_at      INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_mb_match ON message_buffer(message_name, correlation_key);
CREATE INDEX IF NOT EXISTS idx_mb_expires ON message_buffer(expires_at);

CREATE TABLE IF NOT EXISTS incidents (
    key_id               INTEGER PRIMARY KEY,
    process_instance_key INTEGER NOT NULL,
    element_instance_key INTEGER NOT NULL DEFAULT 0,
    job_key              INTEGER NOT NULL DEFAULT 0,
    type                 TEXT    NOT NULL,
    state                INTEGER NOT NULL,
    error_message        TEXT    NOT NULL DEFAULT '',
    created_at           INTEGER NOT NULL,
    resolved_at          INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_inc_state ON incidents(state);
CREATE INDEX IF NOT EXISTS idx_inc_pi ON incidents(process_instance_key);

CREATE TABLE IF NOT EXISTS forms (
    key_id       INTEGER PRIMARY KEY,
    form_id      TEXT    NOT NULL,
    version      INTEGER NOT NULL,
    content      BLOB    NOT NULL,
    content_hash BLOB    NOT NULL,
    deployed_at  INTEGER NOT NULL,
    deleted_at   INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_form_id ON forms(form_id, version);
CREATE INDEX IF NOT EXISTS idx_form_hash ON forms(form_id, content_hash);
`

// Migrate creates all tables and indexes.
func Migrate(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, schema)
	return err
}
