package chat

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
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

func (s *Store) CreateChat(chat *types.Chat, members []int) error {
	err := s.pool.QueryRow(context.Background(),
		"INSERT INTO chats (chat_type, name) VALUES ($1, $2) RETURNING id",
		chat.ChatType, chat.Name).Scan(&chat.ID)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, memberID := range members {
		_, err := s.pool.Exec(context.Background(),
			`INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2)`, chat.ID, memberID)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func (s *Store) GetAllChatsByUserID(userID int) ([]types.AllChatsItem, error) {
	chats := make([]types.AllChatsItem, 0)
	query := `
		SELECT c.id, c.chat_type, c.name, COALESCE(m.content, ''), COALESCE(m.created_at, c.created_at)
		FROM chats c
		JOIN chat_members cm on c.id = cm.chat_id
		LEFT JOIN messages m on m.id = (
		    SELECT id FROM messages WHERE chat_id = c.id ORDER BY created_at DESC LIMIT 1
		)
		WHERE cm.user_id = $1`

	rows, err := s.pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var chat types.AllChatsItem
		var lastMessage types.Message
		if err := rows.Scan(&chat.ID, &chat.ChatType, &chat.Name, &lastMessage.Content, &lastMessage.CreatedAt); err != nil {
			return nil, err
		}
		chat.LastMessage = lastMessage

		participants, err := s.GetChatParticipants(chat.ID)
		if err != nil {
			return nil, err
		}
		chat.Participants = participants
		chats = append(chats, chat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

func (s *Store) GetChatParticipants(chatID int) ([]types.Participant, error) {
	var participants []types.Participant
	query := `SELECT u.id, u.first_name, u.last_name FROM users u 
    JOIN chat_members cm on u.id = cm.user_id WHERE cm.chat_id = $1`
	rows, err := s.pool.Query(context.Background(), query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var participant types.Participant
		if err := rows.Scan(&participant.UserID, &participant.FirstName, &participant.LastName); err != nil {
			return nil, err
		}
		participants = append(participants, participant)
	}

	return participants, nil
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

func (s *Store) GetChatByUserIDs(user1ID, user2ID int) (*types.Chat, error) {
	query := `
		SELECT c.id, c.chat_type, c.name, c.created_at
		FROM chats c
		JOIN chat_members cm1 ON c.id = cm1.chat_id
		JOIN chat_members cm2 ON c.id = cm2.chat_id
		WHERE cm1.user_id = $1 AND cm2.user_id = $2
		LIMIT 1;
	`

	var chat types.Chat
	err := s.pool.QueryRow(context.Background(), query, user1ID, user2ID).Scan(
		&chat.ID, &chat.ChatType, &chat.Name, &chat.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
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
