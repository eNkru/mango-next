package storage

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents a row in the users table.
type User struct {
	Username string
	Password string // bcrypt hash
	Token    string // nullable
	IsAdmin  bool
}

// ---------------------------------------------------------------------------
// User CRUD — matching Crystal Storage user methods
// ---------------------------------------------------------------------------

// randomStr generates a UUID v4 string without dashes, matching the Crystal
// random_str helper.
func randomStr() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// hashPassword returns the bcrypt hash of the given password, matching
// Crypto::Bcrypt::Password.create(pw).to_s in Crystal.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// verifyPassword compares a bcrypt hash with a plaintext password, matching
// Crypto::Bcrypt::Password.new(hash).verify(pw) in Crystal.
func verifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// validateUsername enforces the same rules as validate_username in
// src/util/validation.cr.
func validateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("username should contain at least 3 characters")
	}
	// Crystal: /^[a-zA-Z_][a-zA-Z0-9_\-]*$/
	re := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_\-]*$`)
	if !re.MatchString(username) {
		return fmt.Errorf("username can only contain alphanumeric characters, underscores, and hyphens")
	}
	return nil
}

// validatePassword enforces the same rules as validate_password in
// src/util/validation.cr.
func validatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password should contain at least 6 characters")
	}
	// Crystal: /^[[:ascii:]]+$/
	for _, r := range password {
		if r > 127 {
			return fmt.Errorf("password should contain ASCII characters only")
		}
	}
	return nil
}

// InitAdmin creates the initial admin user with a random password, matching
// the init_admin macro in storage.cr.
func (s *Storage) InitAdmin() error {
	pw := randomStr()
	hash, err := hashPassword(pw)
	if err != nil {
		return err
	}
	if _, err := s.db.Exec(
		"INSERT INTO users VALUES (?, ?, ?, ?)",
		"admin", hash, nil, 1,
	); err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}
	log.Printf("Initial user created. You can log in with {\"username\": \"admin\", \"password\": %q}", pw)
	return nil
}

// UsernameExists returns true if a user with the given username exists.
func (s *Storage) UsernameExists(username string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM users WHERE username = ?", username,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UsernameIsAdmin returns true if the given username has admin privileges.
func (s *Storage) UsernameIsAdmin(username string) (bool, error) {
	var admin int
	err := s.db.QueryRow(
		"SELECT admin FROM users WHERE username = ?", username,
	).Scan(&admin)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return admin > 0, nil
}

// VerifyUser checks the given username/password pair. On success, it ensures a
// token exists for the user (generating a new one if needed) and returns it.
// Returns an empty string and no error if the password doesn't match.
func (s *Storage) VerifyUser(username, password string) (string, error) {
	var hash string
	var token sql.NullString
	err := s.db.QueryRow(
		"SELECT password, token FROM users WHERE username = ?", username,
	).Scan(&hash, &token)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	if !verifyPassword(hash, password) {
		return "", nil
	}

	// Return existing token or generate a new one.
	if token.Valid && token.String != "" {
		return token.String, nil
	}

	newToken := randomStr()
	if _, err := s.db.Exec(
		"UPDATE users SET token = ? WHERE username = ?",
		newToken, username,
	); err != nil {
		return "", err
	}
	return newToken, nil
}

// VerifyToken returns the username associated with the given token, or an
// empty string if the token is invalid.
func (s *Storage) VerifyToken(token string) (string, error) {
	var username string
	err := s.db.QueryRow(
		"SELECT username FROM users WHERE token = ?", token,
	).Scan(&username)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return username, nil
}

// VerifyAdmin returns true if the given token belongs to an admin user.
func (s *Storage) VerifyAdmin(token string) (bool, error) {
	var admin int
	err := s.db.QueryRow(
		"SELECT admin FROM users WHERE token = ?", token,
	).Scan(&admin)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return admin > 0, nil
}

// ListUsers returns all users with their admin status.
func (s *Storage) ListUsers() ([]User, error) {
	rows, err := s.db.Query("SELECT username, admin FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		var admin int
		if err := rows.Scan(&u.Username, &admin); err != nil {
			return nil, err
		}
		u.IsAdmin = admin > 0
		users = append(users, u)
	}
	return users, rows.Err()
}

// NewUser creates a new user with the given username, password, and admin flag.
func (s *Storage) NewUser(username, password string, admin bool) error {
	if err := validateUsername(username); err != nil {
		return err
	}
	if err := validatePassword(password); err != nil {
		return err
	}

	hash, err := hashPassword(password)
	if err != nil {
		return err
	}

	adminInt := 0
	if admin {
		adminInt = 1
	}
	if _, err := s.db.Exec(
		"INSERT INTO users VALUES (?, ?, ?, ?)",
		username, hash, nil, adminInt,
	); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

// UpdateUser updates a user's details. If password is empty, it is not changed.
func (s *Storage) UpdateUser(originalUsername, username, password string, admin bool) error {
	if err := validateUsername(username); err != nil {
		return err
	}
	if password != "" {
		if err := validatePassword(password); err != nil {
			return err
		}
	}

	adminInt := 0
	if admin {
		adminInt = 1
	}

	// Check if removing last admin.
	if !admin {
		origAdmin, err := s.UsernameIsAdmin(originalUsername)
		if err != nil {
			return err
		}
		if origAdmin {
			count, err := s.adminCount()
			if err != nil {
				return err
			}
			if count <= 1 {
				return fmt.Errorf("cannot remove the last admin user")
			}
		}
	}

	if password == "" {
		_, err := s.db.Exec(
			"UPDATE users SET username = ?, admin = ? WHERE username = ?",
			username, adminInt, originalUsername,
		)
		return err
	}

	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		"UPDATE users SET username = ?, admin = ?, password = ? WHERE username = ?",
		username, adminInt, hash, originalUsername,
	)
	return err
}

// DeleteUser removes a user. It refuses to delete the last admin.
func (s *Storage) DeleteUser(username string) error {
	isAdmin, err := s.UsernameIsAdmin(username)
	if err != nil {
		return err
	}
	if isAdmin {
		count, err := s.adminCount()
		if err != nil {
			return err
		}
		if count <= 1 {
			return fmt.Errorf("cannot remove the last admin user")
		}
	}

	result, err := s.db.Exec("DELETE FROM users WHERE username = ?", username)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user %q not found", username)
	}
	return nil
}

// Logout clears the token for the given token value.
func (s *Storage) Logout(token string) error {
	_, err := s.db.Exec("UPDATE users SET token = NULL WHERE token = ?", token)
	return err
}

// adminCount returns the number of users with admin privileges.
func (s *Storage) adminCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE admin = 1").Scan(&count)
	return count, err
}

// CountUsers returns the total number of users.
func (s *Storage) CountUsers() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}
