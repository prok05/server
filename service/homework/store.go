package homework

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prok05/ecom/types"
	"github.com/prok05/ecom/utils"
	"log"
	"time"
)

type Store struct {
	dbpool *pgxpool.Pool
}

func NewStore(dbpool *pgxpool.Pool) *Store {
	return &Store{
		dbpool: dbpool,
	}
}

// Ok
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
			`INSERT INTO homework_solutions (homework_id, student_id) 
			 VALUES ($1, $2)`,
			homeworkID, studentID)
		if err != nil {
			log.Println("Failed to insert homework solution:", err)
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

// ok
func (s *Store) GetHomeworkTeacherFiles(homeworkID int) ([]types.File, error) {
	query := `SELECT id, filepath, filename, uploaded_at FROM homework_teacher_files WHERE homework_id=$1`

	rows, err := s.dbpool.Query(context.Background(), query, homeworkID)
	if err != nil {
		log.Printf("failed to get homework teacher files: %v", err)
		return nil, err
	}
	defer rows.Close()
	files, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.File])
	if err != nil {
		log.Printf("failed to collect homework teacher files: %v", err)
		return nil, err
	}

	return files, nil
}

// OK
func (s *Store) AssignSolution(data types.SolutionAssignment) error {
	tx, err := s.dbpool.Begin(context.Background())
	if err != nil {
		log.Println("Failed to start transaction: ", err)
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		}
	}()

	_, err = tx.Exec(context.Background(), `UPDATE homework_solutions SET status=$1, solution=$2 WHERE id=$3`,
		2, data.Solution, data.SolutionID)
	if err != nil {
		log.Println("Failed to insert into homeworks:", err)
		return err
	}

	for _, fhs := range data.SolutionFiles {
		path, err := utils.WriteFile(fhs)
		if err != nil {
			log.Println("Failed to save teacher file:", err)
			return err
		}
		_, err = tx.Exec(context.Background(),
			`INSERT INTO homework_files (homework_id, student_id, filepath, filename) 
			 VALUES ($1, $2, $3, $4)`,
			data.HomeworkID, data.StudentID, path, fhs.Filename)
		if err != nil {
			log.Println("Failed to insert teacher file:", err)
			return err
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		log.Println("Failed to commit transaction:", err)
		return err
	}

	return nil
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

func (s *Store) UpdateSolutionStatus(solutionID, status int) error {
	_, err := s.dbpool.Exec(context.Background(), `UPDATE homework_solutions SET status = $1 WHERE id = $2`, status, solutionID)
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

func (s *Store) GetHomeworkFilePathByID(fileID int) (string, error) {
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

func (s *Store) GetHomeworkTeacherFilePathByID(fileID int) (string, error) {
	var filepath string
	err := s.dbpool.QueryRow(context.Background(),
		`SELECT filepath FROM homework_teacher_files WHERE id = $1`, fileID).Scan(&filepath)

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

// ОК
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
		    homework_solutions hs ON h.id = hs.homework_id
		WHERE 
		    h.teacher_id = $1
		GROUP BY 
		    h.id`
	rows, err := s.dbpool.Query(context.Background(), query, teacherID)
	if err != nil {
		log.Printf("failed to get homeworks for teacher: %v", err)
		return nil, err
	}
	defer rows.Close()
	homeworks, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Homework])
	if err != nil {
		log.Printf("failed to collect homeworks for teacher: %v", err)
		return nil, err
	}

	return homeworks, nil
}

// OK
func (s *Store) GetHomeworksByStudentID(studentID int) ([]types.HomeworkStudent, error) {
	// Основной запрос для получения домашнего задания студента
	query := `
        SELECT 
            h.id, 
            h.lesson_id, 
            h.lesson_date, 
            h.lesson_topic,
            h.subject_title, 
            h.description, 
            h.created_at,
            hs.status
        FROM 
            homeworks h
        LEFT JOIN 
            homework_solutions hs ON h.id = hs.homework_id
        WHERE 
            hs.student_id = $1;
    `

	rows, err := s.dbpool.Query(context.Background(), query, studentID)
	if err != nil {
		log.Printf("failed to get homeworks for student: %v", err)
		return nil, err
	}
	defer rows.Close()

	homeworks := make([]types.HomeworkStudent, 0)

	// Сохраняем ID домашних заданий, чтобы получить файлы учителя
	homeworkIDs := make([]int, 0)
	homeworksMap := make(map[int]*types.HomeworkStudent)

	// Сканируем результаты основного запроса
	for rows.Next() {
		var homework types.HomeworkStudent
		err = rows.Scan(
			&homework.ID,
			&homework.LessonID,
			&homework.LessonDate,
			&homework.LessonTopic,
			&homework.SubjectTitle,
			&homework.Description,
			&homework.CreatedAt,
			&homework.Status,
		)
		if err != nil {
			log.Printf("failed to scan homework row: %v", err)
			return nil, err
		}

		// Инициализируем поле Files
		homework.Files = []types.File{}

		// Добавляем задание в массив и в карту
		homeworksMap[homework.ID] = &homework
		homeworkIDs = append(homeworkIDs, homework.ID)
	}

	// Если нет домашних заданий, возвращаем пустой массив
	if len(homeworkIDs) == 0 {
		return homeworks, nil
	}

	// Получаем файлы учителя
	filesQuery := `
        SELECT 
            tf.homework_id, 
            tf.id AS file_id, 
            tf.filename, 
            tf.filepath, 
            tf.uploaded_at
        FROM 
            homework_teacher_files tf
        WHERE 
            tf.homework_id = ANY($1);
    `

	fileRows, err := s.dbpool.Query(context.Background(), filesQuery, homeworkIDs)
	if err != nil {
		log.Printf("failed to get teacher files: %v", err)
		return nil, err
	}
	defer fileRows.Close()

	// Сканируем результаты запроса файлов учителя
	for fileRows.Next() {
		var file types.File
		var homeworkID int

		err = fileRows.Scan(
			&homeworkID,
			&file.ID,
			&file.FileName,
			&file.FilePath,
			&file.UploadedAt,
		)
		if err != nil {
			log.Printf("failed to scan teacher file row: %v", err)
			return nil, err
		}

		// Добавляем файл в соответствующее домашнее задание
		if homework, exists := homeworksMap[homeworkID]; exists {
			homework.Files = append(homework.Files, file)
		}
	}

	// Переносим все домашние задания из карты в массив
	for _, homework := range homeworksMap {
		homeworks = append(homeworks, *homework) // добавляем указатель на структуру в массив
	}

	return homeworks, nil
}

// OK
func (s *Store) GetHomeworkSolutions(homeworkID int) (*[]types.HomeworkSolution, error) {
	tx, err := s.dbpool.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	// 1. Получаем решения
	solutionsRows, err := tx.Query(context.Background(), `
        SELECT 
            hs.id AS solution_id,
            u.id AS student_id,
            COALESCE(u.last_name || ' ' || u.first_name, 'Неопознанный ученик') AS student_name,
            hs.solution,
            hs.status,
            hs.updated_at AS uploaded_at
        FROM 
            homework_solutions hs
        LEFT JOIN 
            users u 
        ON 
            hs.student_id = u.id
        WHERE 
            hs.homework_id = $1`, homeworkID)
	if err != nil {
		return nil, fmt.Errorf("failed to query solutions: %w", err)
	}
	defer solutionsRows.Close()

	type SolutionRow struct {
		ID          int            `db:"solution_id"`
		StudentID   sql.NullInt32  `db:"student_id"` // Изменено
		StudentName string         `db:"student_name"`
		Solution    sql.NullString `db:"solution"`
		Status      int            `db:"status"`
		UploadedAt  time.Time      `db:"uploaded_at"`
	}

	solutions, err := pgx.CollectRows(solutionsRows, pgx.RowToStructByName[SolutionRow])
	if err != nil {
		return nil, fmt.Errorf("failed to collect solutions: %w", err)
	}

	// Преобразуем решения в карту для удобства
	solutionMap := make(map[int]*types.HomeworkSolution)
	for _, sol := range solutions {
		studentID := 0
		studentName := sol.StudentName

		if sol.StudentID.Valid { // Проверяем, есть ли значение
			studentID = int(sol.StudentID.Int32)
		} else { // Пользователь отсутствует
			studentName = "Неопознанный ученик"
		}

		solutionMap[sol.ID] = &types.HomeworkSolution{
			ID:          sol.ID,
			StudentName: studentName,
			StudentID:   studentID,
			Solution:    sol.Solution.String,
			Status:      sol.Status,
			Files:       []types.SolutionFile{},
			UploadedAt:  sol.UploadedAt,
		}
	}

	// 2. Получаем файлы
	filesRows, err := tx.Query(context.Background(), `
        SELECT 
            hf.id AS file_id,
            hf.filename,
            hf.filepath,
            hf.student_id
        FROM 
            homework_files hf
        WHERE 
            hf.homework_id = $1`, homeworkID)
	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	defer filesRows.Close()

	type FileRow struct {
		FileID    int    `db:"file_id"`
		FileName  string `db:"filename"`
		FilePath  string `db:"filepath"`
		StudentID int    `db:"student_id"`
	}

	files, err := pgx.CollectRows(filesRows, pgx.RowToStructByName[FileRow])
	if err != nil {
		return nil, fmt.Errorf("failed to collect files: %w", err)
	}

	// 3. Привязываем файлы к решениям соответствующих студентов
	for _, file := range files {
		if solution, ok := solutionMap[file.StudentID]; ok { // Если решение существует для студента
			solution.Files = append(solution.Files, types.SolutionFile{
				ID:       file.FileID,
				FileName: file.FileName,
			})
		}
	}

	// Преобразуем карту решений в слайс
	var result []types.HomeworkSolution
	for _, solution := range solutionMap {
		result = append(result, *solution)
	}

	if len(result) == 0 {
		result = []types.HomeworkSolution{}
	}

	// Завершаем транзакцию
	if err := tx.Commit(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &result, nil
}

func (s *Store) GetHomeworkSolutionByStudent(homeworkID int, studentID int) (*types.HomeworkSolution, error) {
	tx, err := s.dbpool.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	// 1. Получаем решение
	row := tx.QueryRow(context.Background(), `
        SELECT 
            hs.id AS solution_id,
            u.last_name || ' ' || u.first_name AS student_name,
            hs.solution,
            hs.status,
            hs.updated_at AS uploaded_at
        FROM 
            homework_solutions hs
        JOIN 
            users u 
        ON 
            hs.student_id = u.id
        WHERE 
            hs.homework_id = $1 AND hs.student_id = $2`, homeworkID, studentID)

	type SolutionRow struct {
		ID          int            `db:"solution_id"`
		StudentName string         `db:"student_name"`
		Solution    sql.NullString `db:"solution"`
		Status      int            `db:"status"`
		UploadedAt  time.Time      `db:"uploaded_at"`
	}

	var sol SolutionRow
	if err := row.Scan(
		&sol.ID,
		&sol.StudentName,
		&sol.Solution,
		&sol.Status,
		&sol.UploadedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Если решения нет, возвращаем nil
		}
		return nil, fmt.Errorf("failed to query solution: %w", err)
	}

	solution := &types.HomeworkSolution{
		ID:          sol.ID,
		StudentName: sol.StudentName,
		Solution:    sol.Solution.String,
		Status:      sol.Status,
		Files:       []types.SolutionFile{},
		UploadedAt:  sol.UploadedAt,
	}

	// 2. Получаем файлы
	rows, err := tx.Query(context.Background(), `
        SELECT 
            hf.id AS file_id,
            hf.filename
        FROM 
            homework_files hf
        WHERE 
            hf.homework_id = $1 AND hf.student_id = $2`, homeworkID, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	type FileRow struct {
		FileID   int    `db:"file_id"`
		FileName string `db:"filename"`
	}

	for rows.Next() {
		var file FileRow
		if err := rows.Scan(&file.FileID, &file.FileName); err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}
		solution.Files = append(solution.Files, types.SolutionFile{
			ID:       file.FileID,
			FileName: file.FileName,
		})
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed during file rows iteration: %w", rows.Err())
	}

	// Завершаем транзакцию
	if err := tx.Commit(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return solution, nil
}
