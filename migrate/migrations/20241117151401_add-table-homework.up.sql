CREATE TABLE IF NOT EXISTS homeworks
(
    id          SERIAL PRIMARY KEY,
    student_id  BIGINT       NOT NULL,
    lesson_id   BIGINT       NOT NULL,
    teacher_id  BIGINT       NOT NULL,
    status      SMALLINT     NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);