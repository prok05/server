CREATE TABLE IF NOT EXISTS lesson_rates
(
    id          SERIAL PRIMARY KEY,
    student_id  BIGINT                   NOT NULL REFERENCES users (id),
    teacher_id  BIGINT                   NOT NULL REFERENCES users (id),
    lesson_id   BIGINT                   NOT NULL,
    lesson_date TIMESTAMP WITH TIME ZONE NOT NULL,
    rate        SMALLINT                 NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);