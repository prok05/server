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

func (s *Store) UpdateHomeworkStatus(homeworkID, status int) error {
	_, err := s.dbpool.Exec(context.Background(), `UPDATE homeworks SET status = $1 WHERE id = $2`, status, homeworkID)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
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

	homeworkFiles := make([]types.HomeworkFile, 0)

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

func (s *Store) DeleteHomeworkFileByID(fileID int) (*int, error) {
	var homeworkID int
	err := s.dbpool.QueryRow(context.Background(),
		`DELETE from homework_files WHERE id = $1 RETURNING homework_id`, fileID).Scan(&homeworkID)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &homeworkID, nil
}

func (s *Store) GetHomeworksByTeacherAndLessonID(lessonID, teacherID int, studentIDs []int) ([]types.HomeworkResponse, error) {
	query := `
	SELECT 
		u.first_name,
		u.last_name,
		h.id AS homework_id,
		h.status AS homework_status,
		COALESCE(array_agg(hf.id) FILTER (WHERE hf.id IS NOT NULL), '{}') AS file_ids
	FROM 
		homeworks h
	JOIN 
		users u ON h.student_id = u.id
	LEFT JOIN 
		homework_files hf ON h.id = hf.homework_id
	WHERE 
		h.lesson_id = $1
		AND h.teacher_id = $2
		AND (h.student_id = ANY($3))
	GROUP BY 
		u.first_name, u.last_name, h.id, h.status;
	`

	rows, err := s.dbpool.Query(context.Background(), query, lessonID, teacherID, studentIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	results := make([]types.HomeworkResponse, 0)

	for rows.Next() {
		var result types.HomeworkResponse
		var fileIDs []int

		err := rows.Scan(
			&result.FirstName,
			&result.LastName,
			&result.HomeworkID,
			&result.HomeworkStatus,
			&fileIDs)

		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result.Files = fileIDs
		results = append(results, result)
	}

	return results, nil
}
