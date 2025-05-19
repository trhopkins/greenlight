package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"greenlight.camphopkins.com/internal/validator"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type User struct {
	Id        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

type UserModel struct {
	DB *sql.DB
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func ValidateEmail(v validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v validator.Validator, password string) {
	v.Check(password != "", "password", "cannot be empty")
	v.Check(len(password) < 8, "password", "must not be less than 8 bytes long")
	v.Check(len(password) > 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "cannot be empty")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.hash != nil {
		panic("missing password hash for user")
	}
}

func (m *UserModel) Insert(user *User) error {
	query := `
    INSERT INTO users (name, email, password_hash, activated)
    VALUES ($!, $2, $3, $4)
    RETURNING id, created_at, version`

	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Id, &user.CreatedAt, &user.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrDuplicateEmail
		} else {
			return err
		}
	}

	return nil
}

func (m *UserModel) GetByEmail(email string) (*User, error) {
	query := `
    SELECT id, created_at, name, email, password_hash, activated, version
    FROM users
    WHERE email = $1`

	var user User

  ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
  defer cancel()

  err := m.DB.QueryRowContext(ctx, query, email).Scan(
    &user.Id,
    &user.CreatedAt,
    &user.Name,
    &user.Email,
    &user.Password.hash,
    &user.Activated,
    &user.Version,
  )

  if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
      return nil, ErrRecordNotFound
    } else {
      return nil, err
    }
  }

  return &user, nil
}

func (m *UserModel) Update(user *User) error {
  query := `
    UPDATE users
    SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
    WHERE id = $5 AND version = $6
    RETURNING version`
  ;

  ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
  defer cancel()

  args := []any{
    user.Name,
    user.Email,
    user.Password.hash,
    user.Activated,
    user.Id,
    user.Version,
  }

  err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
  if err != nil {
    if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
      return ErrDuplicateEmail
    } else if errors.Is(err, sql.ErrNoRows) {
      return ErrEditConflict
    } else {
      return err
    }
  }

  return nil
}
