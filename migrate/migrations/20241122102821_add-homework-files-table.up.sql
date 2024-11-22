CREATE TABLE IF NOT EXISTS homework_files
(
    id          SERIAL PRIMARY KEY,
    homework_id BIGINT       NOT NULL REFERENCES homeworks(id),
    filepath    VARCHAR(255) NOT NULL
);