package homework

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
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

func (s *Store) SaveHomework(lessonID, studentID, teacherID int) (int, error) {
	var homeworkID int
	err := s.dbpool.QueryRow(context.Background(),
		"INSERT INTO homeworks (lesson_id, student_id, teacher_id, status) VALUES ($1, $2, $3, $4) RETURNING id",
		lessonID, studentID, teacherID, 2).Scan(&homeworkID)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return homeworkID, nil
}

func (s *Store) SaveHomeworkFile(homeworkID int, filepath string) error {
	_, err := s.dbpool.Exec(context.Background(),
		"INSERT INTO homework_files (homework_id, filepath) VALUES ($1, $2)",
		homeworkID, filepath)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *Store) GetHomeworksByLessonAndStudentID(studentID int, lessonIDs []int) (map[int]*types.HomeworkInfo, error) {
	rows, err := s.dbpool.Query(context.Background(),
		`SELECT id, lesson_id, status FROM homeworks WHERE student_id = $1 AND lesson_id = ANY($2)`,
		studentID, lessonIDs)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	homeworks := make(map[int]*types.HomeworkInfo)

	for rows.Next() {
		var homeworkID *int
		var lessonID, status int
		if err := rows.Scan(&homeworkID, &lessonID, &status); err != nil {
			return nil, err
		}
		homeworks[lessonID] = &types.HomeworkInfo{
			ID:     homeworkID,
			Status: status,
		}
	}

	return homeworks, nil
}

//func (s *Store) GetHomework(homeworkID int) ([]types.HomeworkFile, error) {

func (s *Store) GetHomeworkFilesByHomeworkID(homeworkID int) ([]types.HomeworkFile, error) {
	rows, err := s.dbpool.Query(context.Background(),
		`SELECT id, filepath FROM homework_files WHERE homework_id = $1`, homeworkID)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	var homeworkFiles []types.HomeworkFile

	for rows.Next() {
		var homeworkFile types.HomeworkFile
		if err := rows.Scan(&homeworkFile.ID, &homeworkFile.FilePath); err != nil {
			return nil, err
		}

		homeworkFiles = append(homeworkFiles, homeworkFile)
	}
	return homeworkFiles, nil
}

func (s *Store) GetHomeworkPathByID(fileID int) (string, error) {
	var filepath string
	err := s.dbpool.QueryRow(context.Background(),
		`SELECT filepath FROM homework_files WHERE id = $1`, fileID).Scan(&filepath)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("file with ID %d not found", fileID)
		}
		log.Println("Error querying file path:", err)
		return "", err
	}
	return filepath, nil
}

func (s *Store) DeleteHomeworkFileByID(fileID int) error {
	_, err := s.dbpool.Exec(context.Background(),
		`DELETE from homework_files WHERE id = $1`, fileID)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
