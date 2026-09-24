package store

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/goravel/framework/support/carbon"

	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// SessionLifetime is how long an admin sign-in lasts. API tokens (kind
// "api") do not expire; they are revoked in the admin.
const SessionLifetime = 30 * 24 * time.Hour

// ErrBadCredentials is returned by Login for an unknown email or a wrong
// password; which of the two is not revealed.
var ErrBadCredentials = errors.New("invalid email or password")

func hashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// IssueToken creates a bearer token for user. The plain token is returned
// once; only its hash is stored.
func IssueToken(u models.User, name, kind string) (string, models.ApiToken, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", models.ApiToken{}, err
	}
	plain := "jd_" + hex.EncodeToString(buf)
	t := models.ApiToken{UserID: u.ID, Name: name, Kind: kind, TokenHash: hashToken(plain)}
	err := facades.Orm().Query().Create(&t)
	return plain, t, err
}

// UserForToken resolves a bearer token to its user and token row.
func UserForToken(plain string) (models.User, models.ApiToken, error) {
	var t models.ApiToken
	var u models.User
	plain = strings.TrimSpace(plain)
	if plain == "" {
		return u, t, ErrNotFound
	}
	if err := facades.Orm().Query().Where("token_hash", hashToken(plain)).First(&t); err != nil || t.ID == 0 {
		return u, t, ErrNotFound
	}
	if t.Kind == "session" && t.CreatedAt != nil && t.CreatedAt.StdTime().Before(time.Now().Add(-SessionLifetime)) {
		_, _ = facades.Orm().Query().Delete(&t)
		return u, t, ErrNotFound
	}
	if err := facades.Orm().Query().Where("id", t.UserID).First(&u); err != nil || u.ID == 0 {
		return u, t, ErrNotFound
	}
	_, _ = facades.Orm().Query().Exec(`UPDATE api_tokens SET last_used_at = now() WHERE id = ?`, t.ID)
	return u, t, nil
}

// ErrTooManyAttempts is returned by Login when an address tried too often.
var ErrTooManyAttempts = errors.New("too many sign-in attempts, try again later")

// loginLimiter slows password guessing: 10 failed attempts per 15 minutes
// per client address, in memory (one API process).
var loginLimiter = &RateLimiter{Limit: 10, Window: 15 * time.Minute}

// Login checks credentials and issues a session token.
func Login(email, password, ip string) (string, models.User, error) {
	if loginLimiter.Exceeded(ip, time.Now()) {
		return "", models.User{}, ErrTooManyAttempts
	}
	var u models.User
	_ = facades.Orm().Query().Where("email", strings.ToLower(strings.TrimSpace(email))).First(&u)
	if u.ID == 0 || !facades.Hash().Check(password, u.Password) {
		loginLimiter.Allow(ip, time.Now())
		return "", u, ErrBadCredentials
	}
	plain, _, err := IssueToken(u, "admin session", "session")
	return plain, u, err
}

// RevokeToken deletes one token of user.
func RevokeToken(u models.User, id uint) error {
	res, err := facades.Orm().Query().Where("user_id", u.ID).Where("id", id).Delete(&models.ApiToken{})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

type TokenView struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	LastUsedAt string `json:"last_used_at"`
	CreatedAt  string `json:"created_at"`
}

// APITokens lists user's API (non-session) tokens.
func APITokens(u models.User) ([]TokenView, error) {
	var ts []models.ApiToken
	if err := facades.Orm().Query().Where("user_id", u.ID).Where("kind", "api").Order("id desc").Get(&ts); err != nil {
		return nil, err
	}
	out := []TokenView{}
	for _, t := range ts {
		v := TokenView{ID: t.ID, Name: t.Name}
		if t.LastUsedAt != nil {
			v.LastUsedAt = t.LastUsedAt.ToIso8601String()
		}
		if t.CreatedAt != nil {
			v.CreatedAt = t.CreatedAt.ToIso8601String()
		}
		out = append(out, v)
	}
	return out, nil
}

// EnsureAdmin creates the first admin from ADMIN_EMAIL/ADMIN_PASSWORD when
// there are no users yet.
func EnsureAdmin() error {
	email := strings.ToLower(strings.TrimSpace(facades.Config().GetString("app.admin_email")))
	password := facades.Config().GetString("app.admin_password")
	if email == "" || password == "" {
		return nil
	}
	count, err := facades.Orm().Query().Model(&models.User{}).Count()
	if err != nil || count > 0 {
		return err
	}
	hashed, err := facades.Hash().Make(password)
	if err != nil {
		return err
	}
	name, _, _ := strings.Cut(email, "@")
	return facades.Orm().Query().Create(&models.User{Name: name, Email: email, Password: hashed, Role: RoleAdmin})
}

// Now is the current time, for timestamps set by hand.
func Now() *carbon.DateTime { return carbon.NewDateTime(carbon.Now()) }
