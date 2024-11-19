package homework

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prok05/ecom/types"
	"log"
)

type Store struct {
	dbpool *pgxpool.Pool
}

func NewStore(dbpool *pgxpool.Pool) *Store {
	return &Store{
		dbpool: dbpool,
	}
}

func (s *Store) SaveHomework(lessonID, studentID, teacherID int, filepath string) error {
	_, err := s.dbpool.Exec(context.Background(),
		"INSERT INTO homeworks (lesson_id, student_id, teacher_id, filepath) VALUES ($1, $2, $3, $4)",
		lessonID, studentID, teacherID, filepath)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *Store) GetHomework(homeworkID int) (*types.Homework, error) {
	return nil, nil
}
