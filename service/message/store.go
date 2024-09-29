package message

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prok05/ecom/types"
	"log"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool: pool,
	}
}

func (s *Store) SaveMessage(message *types.Message) error {
	_, err := s.pool.Exec(context.Background(),
		"INSERT INTO messages (chat_id, sender_id, content) VALUES ($1, $2, $3)",
		message.ChatID, message.SenderID, message.Content)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *Store) GetMessages(chatID, limit, offset int) ([]*types.Message, error) {
	query := `SELECT id, chat_id, sender_id, content, created_at FROM messages WHERE chat_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := s.pool.Query(context.Background(), query, chatID, limit, offset)
	if err != nil {
		return nil, err
	}
	var messages []*types.Message
	for rows.Next() {
		var message types.Message
		if err := rows.Scan(&message.ID, &message.ChatID, &message.SenderID,
			&message.Content, &message.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}
	return messages, nil
}
