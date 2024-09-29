package chat

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

func (s *Store) CreateChat(chat *types.Chat) error {
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (chat_type, name) VALUES ($1, $2) RETURNING id",
		chat.ChatType, chat.Name).Scan(&chat.ID)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *Store) GetAllChats() ([]*types.Chat, error) {
	var chats []*types.Chat
	query := "SELECT id, chat_type, name, created_at FROM chats"

	rows, err := s.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var chat types.Chat
		if err := rows.Scan(&chat.ID, &chat.ChatType, &chat.Name, &chat.CreatedAt); err != nil {
			return nil, err
		}
		chats = append(chats, &chat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

func (s *Store) GetChatByID(chatID int) (*types.Chat, error) {
	var chat types.Chat
	query := "SELECT id, chat_type, name, created_at FROM chats WHERE id = $1"
	err := s.pool.QueryRow(context.Background(), query, chatID).Scan(&chatID, &chat.ChatType, &chat.Name, &chat.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func (s *Store) DeleteChat(chatID int) error {
	query := "DELETE FROM chats WHERE id = $1"
	_, err := s.pool.Exec(context.Background(), query, chatID)
	if err != nil {
		return err
	}
	return nil
}
