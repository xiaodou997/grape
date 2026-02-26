package auth

import (
	"errors"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type User struct {
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Password   string    `json:"-"` // 不暴露密码
	Role       string    `json:"role"`
	CreatedAt  time.Time `json:"createdAt"`
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

// MemoryUserStore 内存用户存储
type MemoryUserStore struct {
	users map[string]*User
	mu    sync.RWMutex
}

func NewMemoryUserStore() *MemoryUserStore {
	store := &MemoryUserStore{
		users: make(map[string]*User),
	}
	// 创建默认管理员账户
	store.createDefaultAdmin()
	return store
}

func (s *MemoryUserStore) createDefaultAdmin() {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	s.users["admin"] = &User{
		Username:  "admin",
		Email:     "admin@grape.local",
		Password:  string(hashedPassword),
		Role:      "admin",
		CreatedAt: time.Now(),
	}
}

func (s *MemoryUserStore) Get(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *MemoryUserStore) Create(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[user.Username]; ok {
		return ErrUserAlreadyExists
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	user.CreatedAt = time.Now()
	s.users[user.Username] = user
	return nil
}

func (s *MemoryUserStore) Update(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[user.Username]; !ok {
		return ErrUserNotFound
	}

	s.users[user.Username] = user
	return nil
}

func (s *MemoryUserStore) Delete(username string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[username]; !ok {
		return ErrUserNotFound
	}

	delete(s.users, username)
	return nil
}

func (s *MemoryUserStore) List() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users
}

func (s *MemoryUserStore) Validate(username, password string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[username]
	if !ok {
		return nil, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return user, nil
}
