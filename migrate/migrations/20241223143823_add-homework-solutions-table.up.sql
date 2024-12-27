CREATE TABLE IF NOT EXISTS homework_solutions
(
    id           SERIAL PRIMARY KEY,
    homework_id  BIGINT   NOT NULL REFERENCES homeworks (id),
    student_id   BIGINT   NOT NULL,
    status       SMALLINT NOT NULL        DEFAULT 3, -- 3 - не сдано, 1 - сдано, 2 - на проверке, 4 - отклонено
    review_notes TEXT,
    solution     TEXT,
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);