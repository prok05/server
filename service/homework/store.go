package homework

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prok05/ecom/types"
	"github.com/prok05/ecom/utils"
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

func (s *Store) AssignHomework(data types.HomeworkAssignment) (int, error) {
	tx, err := s.dbpool.Begin(context.Background())
	if err != nil {
		log.Println("Failed to start transaction: ", err)
		return 0, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		}
	}()

	var homeworkID int
	err = tx.QueryRow(context.Background(),
		`INSERT INTO homeworks (lesson_id, lesson_date, lesson_topic, teacher_id, subject_title, description) 
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		data.LessonID, data.LessonDate, data.LessonTopic, data.TeacherID, data.SubjectTitle, data.Description).Scan(&homeworkID)
	if err != nil {
		log.Println("Failed to insert into homeworks:", err)
		return 0, err
	}

	for _, fhs := range data.TeacherFiles {
		path, err := utils.WriteFile(fhs)
		if err != nil {
			log.Println("Failed to save teacher file:", err)
			return 0, err
		}
		_, err = tx.Exec(context.Background(),
			`INSERT INTO homework_teacher_files (homework_id, filepath, filename) 
			 VALUES ($1, $2, $3)`,
			homeworkID, path, fhs.Filename)
		if err != nil {
			log.Println("Failed to insert teacher file:", err)
			return 0, err
		}
	}

	for _, studentID := range data.StudentIDs {
		_, err = tx.Exec(context.Background(),
			`INSERT INTO homework_status (homework_id, student_id, status) 
			 VALUES ($1, $2, $3)`,
			homeworkID, studentID, 3)
		if err != nil {
			log.Println("Failed to insert homework status:", err)
			return 0, err
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		log.Println("Failed to commit transaction:", err)
		return 0, err
	}

	return homeworkID, nil
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

func (s *Store) DeleteHomework(homeworkID int) error {
	_, err := s.dbpool.Exec(context.Background(), `DELETE FROM homeworks WHERE id = $1`, homeworkID)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *Store) UpdateHomeworkStatus(homeworkID, status int) error {
	_, err := s.dbpool.Exec(context.Background(), `UPDATE homeworks SET status = $1 WHERE id = $2`, status, homeworkID)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *Store) CountHomeworksWithStatus(lessonID, teacherID, status int) (int, error) {
	var count int
	err := s.dbpool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM homeworks WHERE lesson_id = $1 AND teacher_id = $2 AND status = $3",
		lessonID, teacherID, status).Scan(&count)
	if err != nil {
		log.Println("Error counting homeworks:", err)
		return 0, err
	}
	return count, nil
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

func (s *Store) GetHomeworkByLessonID(lessonID int) (*types.Homework, error) {
	var homework types.Homework
	err := s.dbpool.QueryRow(context.Background(), `SELECT * from homeworks WHERE lesson_id = $1`, lessonID).Scan(
		&homework.ID,
		&homework.LessonID,
		&homework.LessonDate,
		&homework.LessonTopic,
		&homework.TeacherID,
		&homework.SubjectTitle,
		&homework.Description,
		&homework.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			log.Println(pgErr.Message)
			return nil, err
		}
	}
	return &homework, nil
}

func (s *Store) GetHomeworksByTeacherID(teacherID int) ([]types.Homework, error) {
	query := `SELECT 
		    h.id, 
		    h.lesson_id, 
		    h.lesson_date, 
		    h.lesson_topic, 
		    h.teacher_id, 
		    h.subject_title, 
		    h.description, 
		    h.created_at, 
		    COALESCE(COUNT(hs.id) FILTER (WHERE hs.status = 2), 0) AS under_review_count
		FROM 
		    homeworks h
		LEFT JOIN 
		    homework_status hs ON h.id = hs.homework_id
		WHERE 
		    h.teacher_id = $1
		GROUP BY 
		    h.id`
	rows, err := s.dbpool.Query(context.Background(), query, teacherID)
	if err != nil {
		log.Printf("failed to get homeworks for teacher: %v", err)
		return nil, err
	}
	homeworks, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Homework])
	if err != nil {
		log.Printf("failed to collect homeworks for teacher: %v", err)
		return nil, err
	}

	return homeworks, nil
}
