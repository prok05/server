CREATE TYPE user_role AS ENUM('teacher', 'student', 'supervisor');

CREATE TABLE IF NOT EXISTS users (
                                     id BIGINT PRIMARY KEY,
                                     phone VARCHAR(15) NOT NULL UNIQUE,
                                     password VARCHAR(255) NOT NULL,
                                     first_name VARCHAR(255) NOT NULL,
                                     last_name VARCHAR(255) NOT NULL,
                                     middle_name VARCHAR(255) NOT NULL,
                                     user_role user_role NOT NULL,
                                     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);