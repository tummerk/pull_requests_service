CREATE TABLE pull_requests (
                               id SERIAL PRIMARY KEY,
                               title VARCHAR(500) NOT NULL,
                               description TEXT,
                               author_id INTEGER NOT NULL REFERENCES users(id),
                               status VARCHAR(20) NOT NULL DEFAULT 'OPEN', -- OPEN, MERGED
                               need_more_reviewers BOOLEAN NOT NULL DEFAULT FALSE,
                               created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                               updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                               merged_at TIMESTAMP,
                               CHECK (status IN ('OPEN', 'MERGED'))
);
