CREATE TABLE IF NOT EXISTS homeworks
(
    id            SERIAL PRIMARY KEY,
    lesson_id     BIGINT                   NOT NULL,
    lesson_date   TIMESTAMP WITH TIME ZONE NOT NULL,
    lesson_topic  TEXT,
    teacher_id    BIGINT                   NOT NULL REFERENCES users (id),
    subject_title TEXT                     NOT NULL,
    description   TEXT                     NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);