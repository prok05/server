CREATE TABLE IF NOT EXISTS homework_teacher_files
(
    id          SERIAL PRIMARY KEY,
    homework_id BIGINT       NOT NULL REFERENCES homeworks(id),
    filepath    VARCHAR(255) NOT NULL,
    filename   VARCHAR(255) NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);