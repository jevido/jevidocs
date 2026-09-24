package store

import (
	"net/mail"
	"strings"

	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

type UserView struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

func ViewUser(u models.User) UserView {
	v := UserView{ID: u.ID, Name: u.Name, Email: u.Email}
	if u.CreatedAt != nil {
		v.CreatedAt = u.CreatedAt.ToIso8601String()
	}
	return v
}

// ListUsers returns every admin.
func ListUsers() ([]UserView, error) {
	var us []models.User
	if err := facades.Orm().Query().Order("id asc").Get(&us); err != nil {
		return nil, err
	}
	out := make([]UserView, 0, len(us))
	for _, u := range us {
		out = append(out, ViewUser(u))
	}
	return out, nil
}

// ValidatePassword enforces the password policy.
func ValidatePassword(pw string) error {
	if len(pw) < 8 {
		return ValidationError{"password must be at least 8 characters"}
	}
	if len(pw) > 72 {
		return ValidationError{"password must be at most 72 characters"}
	}
	return nil
}

// CreateUser adds an admin.
func CreateUser(name, email, password string) (models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)
	var u models.User
	if _, err := mail.ParseAddress(email); err != nil || !strings.Contains(email, "@") {
		return u, ValidationError{"a valid email is required"}
	}
	if err := ValidatePassword(password); err != nil {
		return u, err
	}
	var other models.User
	_ = facades.Orm().Query().Where("email", email).First(&other)
	if other.ID != 0 {
		return u, ValidationError{"a user with that email already exists"}
	}
	if name == "" {
		name, _, _ = strings.Cut(email, "@")
	}
	hashed, err := facades.Hash().Make(password)
	if err != nil {
		return u, err
	}
	u = models.User{Name: name, Email: email, Password: hashed}
	err = facades.Orm().Query().Create(&u)
	return u, err
}

// DeleteUser removes an admin other than self, with their tokens.
func DeleteUser(self models.User, id uint) error {
	if id == self.ID {
		return ValidationError{"you cannot delete yourself"}
	}
	var u models.User
	if err := facades.Orm().Query().Where("id", id).First(&u); err != nil {
		return err
	}
	if u.ID == 0 {
		return ErrNotFound
	}
	if _, err := facades.Orm().Query().Where("user_id", u.ID).Delete(&models.ApiToken{}); err != nil {
		return err
	}
	_, err := facades.Orm().Query().Delete(&u)
	return err
}

// ChangePassword sets a new password after checking the current one.
func ChangePassword(u models.User, current, next string) error {
	var full models.User
	if err := facades.Orm().Query().Where("id", u.ID).First(&full); err != nil || full.ID == 0 {
		return ErrNotFound
	}
	if !facades.Hash().Check(current, full.Password) {
		return ValidationError{"current password is wrong"}
	}
	if err := ValidatePassword(next); err != nil {
		return err
	}
	hashed, err := facades.Hash().Make(next)
	if err != nil {
		return err
	}
	_, err = facades.Orm().Query().Exec(`UPDATE users SET password = ?, updated_at = now() WHERE id = ?`, hashed, full.ID)
	return err
}
