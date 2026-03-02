package auth

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// 密码验证错误
var (
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
)

// ValidRoles 有效的角色列表
var ValidRoles = map[string]bool{
	"admin":     true,
	"developer": true,
	"readonly":  true,
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

// ValidateRole 验证角色是否有效
func ValidateRole(role string) error {
	if role == "" {
		return nil // 空角色使用默认值
	}
	if !ValidRoles[role] {
		return fmt.Errorf("invalid role: %s, must be one of: admin, developer, readonly", role)
	}
	return nil
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
	// 返回副本，避免调用方修改内部状态
	cp := *user
	return &cp, nil
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
		// 返回副本，避免调用方修改内部状态
		cp := *user
		users = append(users, &cp)
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

	// 返回副本，避免调用方修改内部状态
	cp := *user
	return &cp, nil
}