CREATE TABLE IF NOT EXISTS homework_files
(
    id          SERIAL PRIMARY KEY,
    homework_id BIGINT       NOT NULL REFERENCES homeworks (id),
    student_id  BIGINT       NOT NULL REFERENCES users (id),
    filepath    VARCHAR(255) NOT NULL,
    filename   VARCHAR(255) NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);