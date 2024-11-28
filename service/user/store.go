package user

import (
	"context"
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

func (s *Store) GetAllTeachers() ([]*types.UserDTO, error) {
	query := `SELECT id, first_name, last_name, middle_name, user_role FROM users WHERE user_role='teacher'`
	rows, err := s.dbpool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teachers []*types.UserDTO
	for rows.Next() {
		var teacher types.UserDTO
		if err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.MiddleName, &teacher.Role); err != nil {
			return nil, err
		}
		teachers = append(teachers, &teacher)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return teachers, nil
}

func (s *Store) FindUserByEmail(email string) (*types.User, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Store) FindUserByPhone(phone string) (*types.User, error) {
	rows, err := s.dbpool.Query(context.Background(),
		"SELECT * FROM users WHERE phone = $1", phone)

	if err != nil {
		return nil, err
	}

	u := new(types.User)
	for rows.Next() {
		u, err = scanRowsIntoUser(rows)
		if err != nil {
			return nil, err
		}
	}
	if u.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}

func (s *Store) FindUserByID(id int) (*types.UserDTO, error) {
	rows, err := s.dbpool.Query(context.Background(),
		"SELECT id, phone, first_name, middle_name, last_name, user_role FROM users WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	u := new(types.UserDTO)
	for rows.Next() {
		if err := rows.Scan(&u.ID, &u.Phone, &u.FirstName, &u.MiddleName, &u.LastName, &u.Role); err != nil {
			return nil, err
		}
	}
	if u.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}

func (s *Store) FindUsersByIDs(id int) ([]*types.UserDTO, error) {
	rows, err := s.dbpool.Query(context.Background(),
		"SELECT id, phone, first_name, middle_name, last_name, user_role FROM users WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	users := make([]types.UserDTO, 0)

	for rows.Next() {
		u :=
		if err := rows.Scan(&u.ID, &u.Phone, &u.FirstName, &u.MiddleName, &u.LastName, &u.Role); err != nil {
			return nil, err
		}
	}
	if u.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}

func (s *Store) CreateUser(user types.User) error {
	_, err := s.dbpool.Exec(context.Background(),
		"INSERT INTO users (id, phone, password, first_name, last_name, middle_name, user_role) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		user.ID, user.Phone, user.Password, user.FirstName, user.LastName, user.MiddleName, user.Role)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func scanRowsIntoUser(rows pgx.Rows) (*types.User, error) {
	user := new(types.User)

	err := rows.Scan(
		&user.ID,
		&user.Phone,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.MiddleName,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}
