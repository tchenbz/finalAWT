package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/tchenbz/test3AWT/internal/validator"

	"golang.org/x/crypto/bcrypt"
)

var ErrDuplicateEmail = errors.New("duplicate email")

type UserModel struct {
    DB *sql.DB
}
func (u UserModel) Insert(user *User) error {
	query := `
	INSERT INTO users (username, email, password_hash, activated) 
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version
   `
args := []any{user.Username, user.Email, user.Password.hash, user.Activated}

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
err := u.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
if err != nil {
	switch {
		case err.Error() == `pq: duplicate key value violates unique 
		constraint "users_email_key"`:
	return ErrDuplicateEmail
	default:
	return err
	}
}

return nil
}

var AnonymousUser = &User{}

type User struct {
    ID         int64	    `json:"id"`
    CreatedAt  time.Time   `json:"created_at"`
    Username   string      `json:"username"`
    Email      string      `json:"email"`
	Password   password   `json:"-"`
    Activated  bool       `json:"activated"`
    Version     int        `json:"-"`  
}

func (u *User) IsAnonymous() bool {
    return u == AnonymousUser
}

type password struct {
    plaintext *string
    hash      []byte
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
        switch {
            case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
                return false, nil
            default:
                return false, err
       }
  }

  return true, nil                
}

func (u UserModel) GetByEmail(email string) (*User, error) {
	query := `
	SELECT id, created_at, username, email, password_hash, activated, version
	FROM users
	WHERE email = $1
   `
var user User

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
err := u.DB.QueryRowContext(ctx, query, email).Scan(
	&user.ID,
	&user.CreatedAt,
	&user.Username,
	&user.Email,
	&user.Password.hash,
	&user.Activated,
	&user.Version,
)
if err != nil {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ErrRecordNotFound
	default:
		return nil, err
	}
}

return &user, nil
}

func ValidateEmail(v *validator.Validator, email string) {
v.Check(email != "", "email", "must be provided")
v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
    v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
    v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
    v.Check(user.Username != "", "username", "must be provided")
    v.Check(len(user.Username) <= 200, "username", "must not be more than 200 bytes long")
    ValidateEmail(v, user.Email)
    if user.Password.plaintext != nil {
        ValidatePasswordPlaintext(v, *user.Password.plaintext)
    }
    if user.Password.hash == nil {
        panic("missing password hash for user")
    }

}

func (u UserModel) Update (user *User) error { 
    query := `
        UPDATE users 
        SET username = $1, email = $2, password_hash = $3,
            activated = $4, version = version + 1
        WHERE id = $5 AND version = $6
        RETURNING version
		`
		args := []any{
			user.Username,
			user.Email,
			user.Password.hash,
			user.Activated,
			user.ID,
			user.Version,
		}
	
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
	
		err := u.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	 if err != nil {
        switch {
            case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
                return ErrDuplicateEmail
            case errors.Is(err, sql.ErrNoRows):
                return ErrEditConflict
            default:
                return err
            }
  }

    return nil
}

func (u UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
    tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	query := `
			SELECT users.id, users.created_at, users.username,
				users.email, users.password_hash, users.activated, users.version
			FROM users
			INNER JOIN tokens
			ON users.id = tokens.user_id
			WHERE tokens.hash = $1
			AND tokens.scope = $2 
			AND tokens.expiry > $3
		`
		args := []any{tokenHash[:], tokenScope, time.Now()}
		var user User
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err := u.DB.QueryRowContext(ctx, query, args...).Scan(
			 &user.ID,
			 &user.CreatedAt,
			 &user.Username,
			 &user.Email,
			 &user.Password.hash,
			 &user.Activated,
			 &user.Version,
		   )
		   if err != nil {
		   switch {
			 case errors.Is(err, sql.ErrNoRows):
				 return nil, ErrRecordNotFound
			 default:
				 return nil, err
			 }
		   }
 return &user, nil
}

func (u UserModel) GetByID(id int64) (*User, error) {
	query := `
		SELECT id, created_at, username, email, activated, version
		FROM users
		WHERE id = $1
	`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Username,
		&user.Email,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
