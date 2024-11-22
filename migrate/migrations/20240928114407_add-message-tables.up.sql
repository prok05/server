CREATE TABLE IF NOT EXISTS chats (
                                     id SERIAL PRIMARY KEY,
                                     chat_type VARCHAR(50) NOT NULL,
                                     name VARCHAR(255) NOT NULL,
                                     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS chat_members (
                                        id SERIAL PRIMARY KEY,
                                        chat_id INTEGER NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
                                        user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                        joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                        unique (chat_id, user_id)
);

CREATE TABLE IF NOT EXISTS messages (
                                            id SERIAL PRIMARY KEY,
                                            chat_id INTEGER NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
                                            sender_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                            content TEXT NOT NULL,
                                            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_chat_type ON chats(chat_type);
CREATE INDEX idx_chat_members_chat_id ON chat_members(chat_id);
CREATE INDEX idx_chat_members_user_id ON chat_members(user_id);
CREATE INDEX idx_messages_chat_id ON messages(chat_id);
CREATE INDEX idx_messages_sender_id ON messages(sender_id);