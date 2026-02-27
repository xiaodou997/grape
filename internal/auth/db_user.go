package auth

import (
	"errors"
	"time"

	"github.com/graperegistry/grape/internal/db"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type User struct {
	ID         uint       `json:"id"`
	Username   string     `json:"username"`
	Email      string     `json:"email"`
	Password   string     `json:"-"`
	Role       string     `json:"role"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastLogin  *time.Time `json:"lastLogin,omitempty"`
}

type UserStore interface {
	Get(username string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(username string) error
	List() []*User
	Validate(username, password string) (*User, error)
}

// DBUserStore 数据库用户存储
type DBUserStore struct{}

func NewDBUserStore() *DBUserStore {
	return &DBUserStore{}
}

func (s *DBUserStore) Get(username string) (*User, error) {
	var dbUser db.User
	result := db.DB.Where("username = ?", username).First(&dbUser)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}

	return s.toUser(&dbUser), nil
}

func (s *DBUserStore) Create(user *User) error {
	// 检查用户是否已存在
	var count int64
	db.DB.Model(&db.User{}).Where("username = ?", user.Username).Count(&count)
	if count > 0 {
		return ErrUserAlreadyExists
	}

	dbUser := &db.User{
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}

	if err := dbUser.SetPassword(user.Password); err != nil {
		return err
	}

	result := db.DB.Create(dbUser)
	if result.Error != nil {
		return result.Error
	}

	user.CreatedAt = dbUser.CreatedAt
	return nil
}

func (s *DBUserStore) Update(user *User) error {
	var dbUser db.User
	result := db.DB.Where("username = ?", user.Username).First(&dbUser)
	if result.Error != nil {
		return ErrUserNotFound
	}

	dbUser.Email = user.Email
	dbUser.Role = user.Role
	dbUser.LastLogin = user.LastLogin

	// 如果 Password 非空，则更新密码
	if user.Password != "" {
		if err := dbUser.SetPassword(user.Password); err != nil {
			return err
		}
	}

	return db.DB.Save(&dbUser).Error
}

func (s *DBUserStore) Delete(username string) error {
	result := db.DB.Where("username = ?", username).Delete(&db.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (s *DBUserStore) List() []*User {
	var dbUsers []db.User
	db.DB.Find(&dbUsers)

	users := make([]*User, 0, len(dbUsers))
	for _, dbUser := range dbUsers {
		users = append(users, s.toUser(&dbUser))
	}
	return users
}

func (s *DBUserStore) Validate(username, password string) (*User, error) {
	var dbUser db.User
	result := db.DB.Where("username = ?", username).First(&dbUser)
	if result.Error != nil {
		return nil, ErrUserNotFound
	}

	if !dbUser.CheckPassword(password) {
		return nil, ErrInvalidPassword
	}

	return s.toUser(&dbUser), nil
}

func (s *DBUserStore) toUser(dbUser *db.User) *User {
	return &User{
		ID:        dbUser.ID,
		Username:  dbUser.Username,
		Email:     dbUser.Email,
		Role:      dbUser.Role,
		CreatedAt: dbUser.CreatedAt,
		LastLogin: dbUser.LastLogin,
	}
}
