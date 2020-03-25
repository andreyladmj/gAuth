package mysql

import (
	"andreyladmj/gAuth/pkg/models"
	"database/sql"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Get(email string) (*models.User, error) {
	s := &models.User{}
	stmt := `SELECT  name, email, picture, gender, locale, created FROM users WHERE email = ?`
	err := m.DB.QueryRow(stmt, email).Scan(&s.Name, &s.Email, &s.Picture, &s.Gender, &s.Locale, &s.Created)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	return s, nil
}
func (m *UserModel) GetUserByToken(token string) (*models.User, error) {
	s := &models.User{}
	stmt := `SELECT  name, email, picture, gender, locale, u.created 
		FROM users u 
		JOIN tokens t ON t.user_id = u.id 
		WHERE token = ? AND t.updated > ADDDATE(UTC_TIMESTAMP(), INTERVAL -30 MINUTE)
		ORDER BY t.updated DESC
		LIMIT 1
		`
	err := m.DB.QueryRow(stmt, token).Scan(&s.Name, &s.Email, &s.Picture, &s.Gender, &s.Locale, &s.Created)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	stmt = `UPDATE tokens SET updated=UTC_TIMESTAMP() WHERE token=?;`
	_, err = m.DB.Exec(stmt, token)

	if err != nil {
		return nil, err
	}

	return s, nil
}

func (m *UserModel) Create(email, name, picture, gender, locale string) error {
	stmt := `INSERT INTO users (name, email, picture, gender, locale, created) VALUES (?,?,?,?,?,UTC_TIMESTAMP())`
	res, err := m.DB.Exec(stmt, name, email, picture, NewNullString(gender), locale)

	if err != nil {
		return err
	}

	_, err = res.RowsAffected()

	return err
}

func (m *UserModel) Update(email, name, picture, gender, locale string) error {
	stmt := `UPDATE users SET name=?, picture=?, gender=?, locale=? WHERE email=?;`
	res, err := m.DB.Exec(stmt, name, picture, NewNullString(gender), locale, email)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()

	return err
}


func (m *UserModel) UpdateOrCreate(email, name, picture, gender, locale string) (*models.User, error) {
	user, err := m.Get(email)

	if err != nil {
		if err != models.ErrNoRecord {
			return nil, err
		}

		err := m.Create(email, name, picture, gender, locale)

		if err != nil {
			return nil, err
		}

		return m.Get(email)
	}

	err = m.Update(email, name, picture, gender, locale)

	if err != nil {
		return nil, err
	}

	return user, nil
}


func (m *UserModel) CreateToken(email string) (string, error) {
	token, err := bcrypt.GenerateFromPassword([]byte(StringWithCharset(32, "")), 12)
	stmt := `INSERT INTO tokens (user_id, token, created, updated) VALUES (
		(SELECT id FROM users WHERE email=?), ?, UTC_TIMESTAMP(), UTC_TIMESTAMP()
	)`
	res, err := m.DB.Exec(stmt, email, string(token))

	if err != nil {
		return "", err
	}

	_, err = res.RowsAffected()

	return string(token), err
}