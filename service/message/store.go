package message

import (
	"context"
	"fmt"
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

func (s *Store) SaveMessage(message *types.Message) (int, error) {
	var messageID int
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO messages (chat_id, sender_id, content) VALUES ($1, $2, $3) RETURNING id",
		message.ChatID, message.SenderID, message.Content).Scan(&messageID)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return messageID, nil
}

func (s *Store) GetMessages(chatID, limit, offset int) ([]*types.Message, error) {
	query := `SELECT id, chat_id, sender_id, content, created_at FROM messages WHERE chat_id = $1 ORDER BY created_at DESC OFFSET $2`
	rows, err := s.pool.Query(context.Background(), query, chatID, offset)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var messages []*types.Message
	for rows.Next() {
		var message types.Message
		if err := rows.Scan(&message.ID, &message.ChatID, &message.SenderID,
			&message.Content, &message.CreatedAt); err != nil {
			fmt.Println(err)
			return nil, err
		}
		messages = append(messages, &message)
	}
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

func (s *Store) IsUserInChat(chatID int, userID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2)`
	err := s.pool.QueryRow(context.Background(), query, chatID, userID).Scan(&exists)
	return exists, err
}
