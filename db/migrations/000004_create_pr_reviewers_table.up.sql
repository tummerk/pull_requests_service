CREATE TABLE pr_reviewers (
                              id SERIAL PRIMARY KEY,
                              pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
                              reviewer_id VARCHAR(255) NOT NULL REFERENCES users(id),
                              assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                              replaced_at TIMESTAMP,
                              is_current BOOLEAN NOT NULL DEFAULT TRUE,
                              UNIQUE(pull_request_id, reviewer_id)
);