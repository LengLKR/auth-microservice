package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/LengLKR/auth-microservice/internal/domain"
	"github.com/LengLKR/auth-microservice/internal/repository"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// AuthService stub ของ service layer
type AuthService struct {
    repo      repository.UserRepository
    tokenRepo repository.TokenRepository
    jwtSecret string
	attempts  map[string][]time.Time
	mu 		  sync.Mutex	
}

// NewAuthService สร้าง AuthService พร้อม userRepo, tokenRepo, และ secret
func NewAuthService(r repository.UserRepository, t repository.TokenRepository, secret string) *AuthService {
    return &AuthService{
        repo:      r,
        tokenRepo: t,
        jwtSecret: secret,
		attempts: make(map[string][]time.Time),
    }
}

// Register สร้างบัญชีใหม่: hash, save, คืน JWT
func (s *AuthService) Register(ctx context.Context, email, password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    user := &domain.User{ Email: email, PasswordHash: string(hash) }
    if err := s.repo.Create(user); err != nil {
        return "", err
    }
    return s.generateToken(user.ID)
}

// Login ตรวจ credentials แล้วคืน JWT
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {

    // กดล็อก mutex
    s.mu.Lock()
    // ให้ UnLock อัตโนมัติเมื่อ return ออกจากบล็อกนี้
    defer s.mu.Unlock()

    // Rate limiting: สูงสุด 5 attempts ใน 1 นาที
    now := time.Now()
    arr := append(s.attempts[email], now)
    valid := make([]time.Time, 0, len(arr))
	
    for _, t0 := range arr {
        if now.Sub(t0) < time.Minute {
            valid = append(valid, t0)
        }
    }
    s.attempts[email] = valid
    if len(valid) >= 5 {
        return "", errors.New("too many login attempts; please try again later")
    }

    // ตรวจสอบ credential
    user, err := s.repo.FindByEmail(email)
    if err != nil {
        return "", errors.New("invalid credentials")
    }
    if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
        return "", errors.New("invalid credentials")
    }

    // ถ้า login สำเร็จ ให้เคลียร์ attempts (อยู่ภายนอก mutex หรือ lock ใหม่ก็ได้)
    delete(s.attempts, email)

    return s.generateToken(user.ID)
}


// Logout แปลง rawToken → ดึง expiresAt → บันทึกลง blacklist
func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
    tok, err := jwt.ParseWithClaims(rawToken, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
        return []byte(s.jwtSecret), nil
    })
    if err != nil {
        return err
    }
    claims, ok := tok.Claims.(*jwt.RegisteredClaims)
    if !ok || !tok.Valid {
        return errors.New("invalid token")
    }
    return s.tokenRepo.Blacklist(rawToken, claims.ExpiresAt.Time)
}

// generateToken สร้าง JWT ด้วย HS256
func (s *AuthService) generateToken(userID string) (string, error) {
    claims := jwt.RegisteredClaims{
        Subject:   userID,
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.jwtSecret))
}
