package homework

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prok05/ecom/types"
)

type Store struct {
	dbpool *pgxpool.Pool
}

func NewStore(dbpool *pgxpool.Pool) *Store {
	return &Store{
		dbpool: dbpool,
	}
}

func (s *Store) SaveHomework(homework *types.Homework) error {
	return nil
}

func (s *Store) GetHomework(homeworkID int) (*types.Homework, error) {
	return nil, nil
}
