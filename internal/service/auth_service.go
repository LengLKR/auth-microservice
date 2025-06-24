package service

import (
    "context"
    "errors"
    "log"
    "strings"
    "sync"
    "time"

    "github.com/LengLKR/auth-microservice/internal/domain"
    repo "github.com/LengLKR/auth-microservice/internal/repository"
    pr "github.com/LengLKR/auth-microservice/internal/repository/password_reset"

    "github.com/golang-jwt/jwt/v4"
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
    "google.golang.org/grpc/metadata"
)


// AuthService stub ของ service layer
type AuthService struct {
    repo      repo.UserRepository
    tokenRepo repo.TokenRepository
    resetRepo pr.PasswordResetRepository
    jwtSecret string
    attempts  map[string][]time.Time
    mu        sync.Mutex
}

// NewAuthService สร้าง AuthService พร้อม userRepo, tokenRepo, resetRepo, และ secret
func NewAuthService(
    r repo.UserRepository,
    t repo.TokenRepository,
    rr pr.PasswordResetRepository,
    secret string,
) *AuthService {
    return &AuthService{
        repo:      r,
        tokenRepo: t,
        resetRepo: rr,
        jwtSecret: secret,
        attempts:  make(map[string][]time.Time),
    }
}

// Register สร้างบัญชีใหม่: hash, save, คืน JWT
func (s *AuthService) Register(ctx context.Context, email, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	user := &domain.User{Email: email, PasswordHash: string(hash)}
	if err := s.repo.Create(user); err != nil {
		return "", err
	}
	return s.generateToken(user.ID)
}

// Login ตรวจ credentials แล้วคืน JWT
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	s.mu.Lock()
	// ปลดล็อกเมื่อออกจากฟังก์ชันเสมอ
	defer s.mu.Unlock()

	// Rate limiting: สูงสุด 5 failed attempts ใน 1 นาที
	now := time.Now()
	attempts := s.attempts[email]
	// กรองเฉพาะ attempts ที่ยังไม่เกิน 1 นาที
	var recent []time.Time
	for _, ts := range attempts {
		if now.Sub(ts) < time.Minute {
			recent = append(recent, ts)
		}
	}
	// อัปเดต map ด้วย recent attempts
	s.attempts[email] = recent

	if len(recent) >= 5 {
		return "", errors.New("too many login attempts; please try again later")
	}

	// ตรวจสอบ credentials
	user, err := s.repo.FindByEmail(email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		// เพิ่ม failed attempt ทันทีภายใต้ล็อก
		s.attempts[email] = append(s.attempts[email], now)
		return "", errors.New("invalid credentials")
	}

	// ถ้าสำเร็จ ลบประวัติ attempts ทั้งหมด
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

// ListUsers ดึงรายชื่อผู้ใช้พร้อม filter + pagination
func (s *AuthService) ListUsers(ctx context.Context, filterName, filterEmail string, page, size int) ([]domain.User, int64, error) {
	//ถ้าอยากจำกัดเฉพาะ admin ให้ เช็ค metadata จาก ctx ตรงนี้
	users, total, err := s.repo.FindAll(filterName, filterEmail, page, size)
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.User, len(users))
	for i, u := range users {
		out[i] = *u
	}
	return out, total, nil
}

// GetProfile ดึง profile ของตัวเอง
func (s *AuthService) GetProfile(ctx context.Context, id string) (domain.User, error) {
	sub, err := s.subjectFromCtx(ctx)
	if err != nil {
		return domain.User{}, err
	}
	if sub != id {
		return domain.User{}, errors.New("permission denied")
	}
	u, err := s.repo.FindByID(id)
	if err != nil {
		return domain.User{}, err
	}
	return *u, nil
}

// UpdateProfile ให้แก้ไข email (หรือ field อื่นได้ตามต้องการ)
func (s *AuthService) UpdateProfile(ctx context.Context, id, email string) (domain.User, error) {
	sub, err := s.subjectFromCtx(ctx)
	if err != nil {
		return domain.User{}, err
	}
	if sub != id {
		return domain.User{}, errors.New("permission denied")
	}
	u, err := s.repo.FindByID(id)
	if err != nil {
		return domain.User{}, err
	}
	u.Email = email
	if err := s.repo.Update(u); err != nil {
		return domain.User{}, err
	}
	return *u, nil
}

// DeleteProfile ทำ soft delete
func (s *AuthService) DeleteProfile(ctx context.Context, id string) error {
	sub, err := s.subjectFromCtx(ctx)

	if err != nil {
		return err
	}
	if sub != id {
		return errors.New("permission denied")
	}
	return s.repo.SoftDelete(id)

}

// subjectFromCtx ดึง JWT subject จาก metadata: "authorization: Bearer <token>"
func (s *AuthService) subjectFromCtx(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("missing metadata")
	}
	auth := md["authorization"]
	if len(auth) == 0 {
		return "", errors.New("missing authorization header")
	}
	parts := strings.SplitN(auth[0], " ", 2)
	if len(parts) != 2 {
		return "", errors.New("invalid authorization format")
	}
	raw := parts[1]
	tok, err := jwt.ParseWithClaims(raw, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := tok.Claims.(*jwt.RegisteredClaims)
	if !ok || !tok.Valid {
		return "", errors.New("invalid token")
	}
	return claims.Subject, nil
}

// RequestPasswordReset สั่งสร้าง reset token
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
    user, err := s.repo.FindByEmail(email)
    if err != nil {
        // แกล้งทำเหมือนสำเร็จ เพื่อไม่บอกว่ามีหรือไม่มี user
        return nil
    }
    token := uuid.NewString()
    expires := time.Now().Add(15 * time.Minute)
    if err := s.resetRepo.Create(token, user.ID, expires); err != nil {
        return err
    }
    log.Printf("Password reset token for %s: %s\n", email, token)
    return nil
}

// ResetPassword ตรวจ token, เปลี่ยนรหัสผ่าน และลบ token ทิ้ง
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
    userID, err := s.resetRepo.Verify(token)
    if err != nil {
        return errors.New("invalid or expired reset token")
    }
    hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u, err := s.repo.FindByID(userID)
    if err != nil {
        return err
    }
    u.PasswordHash = string(hashed)
    if err := s.repo.Update(u); err != nil {
        return err
    }
    return s.resetRepo.Delete(token)
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

