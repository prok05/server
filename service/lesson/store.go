package lesson

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prok05/ecom/types"
	"log"
	"time"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool: pool,
	}
}

func (s *Store) SaveLessonRate(studentID, teacherID, lessonID int, lessonDate time.Time, rate int8) error {
	query := `INSERT INTO lesson_rates (student_id, teacher_id, lesson_id, lesson_date, rate) VALUES ($1, $2, $3, $4, $5)`
	_, err := s.pool.Exec(context.Background(),
		query,
		studentID, teacherID, lessonID, lessonDate, rate)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *Store) CheckRateExists(studentID, teacherID, lessonID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM lesson_rates 
			WHERE student_id = $1 AND teacher_id = $2 AND lesson_id = $3
		)`
	var exists bool
	err := s.pool.QueryRow(context.Background(), query, studentID, teacherID, lessonID).Scan(&exists)
	if err != nil {
		log.Println("Error checking rate existence:", err)
		return false, err
	}
	return exists, nil
}

func (s *Store) GetLessonRates() ([]types.LessonRate, error) {
	query := `
	SELECT 
        lr.id,
        lr.student_id,
        s.first_name AS student_first_name,
        s.last_name AS student_last_name,
        s.middle_name AS student_middle_name,
        lr.teacher_id,
        t.first_name AS teacher_first_name,
        t.last_name AS teacher_last_name,
        t.middle_name AS teacher_middle_name,
        lr.lesson_id,
        lr.lesson_date,
        lr.rate
    FROM 
        lesson_rates lr
    JOIN 
        users s ON lr.student_id = s.id
    JOIN 
        users t ON lr.teacher_id = t.id;
	`
	rows, err := s.pool.Query(context.Background(), query)
	if err != nil {
		log.Fatalf("Error fetching data: %v", err)
	}
	defer rows.Close()

	rates := make([]types.LessonRate, 0)

	for rows.Next() {
		var rate types.LessonRate
		err = rows.Scan(
			&rate.ID,
			&rate.StudentID,
			&rate.StudentFirstName,
			&rate.StudentLastName,
			&rate.StudentMiddleName,
			&rate.TeacherID,
			&rate.TeacherFirstName,
			&rate.TeacherLastName,
			&rate.TeacherMiddleName,
			&rate.LessonID,
			&rate.LessonDate,
			&rate.Rate,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		rates = append(rates, rate)
	}
	return rates, nil
}
