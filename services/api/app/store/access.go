package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/goravel/framework/support/carbon"

	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// Reader is who is asking to read docs: an optional signed-in user and an
// optional share-link token.
type Reader struct {
	User  *models.User
	Share string
}

// Access levels of a project for a reader.
const (
	AccessPublic = "public"
	AccessAdmin  = "admin"
	AccessMember = "member"
	AccessShare  = "share"
	AccessDomain = "domain"
)

// ReaderFrom resolves a bearer token (may be empty or invalid) and a share
// token into a Reader. Bad bearer tokens are ignored: public docs stay
// readable with a stale login.
func ReaderFrom(bearer, share string) Reader {
	r := Reader{Share: strings.TrimSpace(share)}
	if bearer != "" {
		if u, _, err := UserForToken(bearer); err == nil {
			r.User = &u
		}
	}
	return r
}

// ReadableProject loads a project if r may read it. Private projects the
// reader cannot see answer ErrNotFound, exactly like missing ones.
func ReadableProject(slug string, r Reader) (models.Project, error) {
	p, err := FindProject(slug, true)
	if err != nil {
		return p, err
	}
	access, err := accessOf(p, r)
	if err != nil {
		return p, err
	}
	if access == "" {
		return models.Project{}, ErrNotFound
	}
	p.Access = access
	return p, nil
}

func accessOf(p models.Project, r Reader) (string, error) {
	if p.Public {
		return AccessPublic, nil
	}
	if r.User != nil {
		if HasRole(*r.User, RoleAdmin) {
			return AccessAdmin, nil
		}
		n, err := facades.Orm().Query().Model(&models.ProjectMember{}).
			Where("project_id", p.ID).Where("user_id", r.User.ID).Count()
		if err != nil {
			return "", err
		}
		if n > 0 {
			return AccessMember, nil
		}
		if emailInDomains(r.User.Email, Domains(p)) {
			return AccessDomain, nil
		}
	}
	if r.Share != "" {
		if link, ok := validShare(p, r.Share); ok {
			_, _ = facades.Orm().Query().Exec(`UPDATE share_links SET last_used_at = now() WHERE id = ?`, link.ID)
			return AccessShare, nil
		}
	}
	return "", nil
}

func validShare(p models.Project, token string) (models.ShareLink, bool) {
	var link models.ShareLink
	if err := facades.Orm().Query().Where("token_hash", hashToken(token)).Where("project_id", p.ID).First(&link); err != nil || link.ID == 0 {
		return link, false
	}
	if link.ExpiresAt != nil && link.ExpiresAt.StdTime().Before(time.Now()) {
		return link, false
	}
	return link, true
}

// ReadableProjects lists the public projects plus the private ones r may
// read.
func ReadableProjects(r Reader) ([]ProjectView, error) {
	var ps []models.Project
	if err := facades.Orm().Query().Order("name asc").Get(&ps); err != nil {
		return nil, err
	}
	out := []ProjectView{}
	for _, p := range ps {
		access, err := accessOf(p, r)
		if err != nil {
			return nil, err
		}
		if access == "" {
			continue
		}
		p.Access = access
		out = append(out, ViewProject(p))
	}
	return out, nil
}

// Members.

type MemberView struct {
	UserID uint   `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

// Members lists the users who may read private project p.
func Members(p models.Project) ([]MemberView, error) {
	var rows []MemberView
	err := facades.Orm().Query().Raw(`
		SELECT u.id AS user_id, u.name, u.email, u.role FROM project_members m
		JOIN users u ON u.id = m.user_id WHERE m.project_id = ? ORDER BY u.name`, p.ID).Scan(&rows)
	if rows == nil {
		rows = []MemberView{}
	}
	for i := range rows {
		if !ValidRole(rows[i].Role) {
			rows[i].Role = RoleAdmin
		}
	}
	return rows, err
}

// AddMember gives user id read access to p. Adding twice is fine.
func AddMember(p models.Project, userID uint) error {
	var u models.User
	if err := facades.Orm().Query().Where("id", userID).First(&u); err != nil {
		return err
	}
	if u.ID == 0 {
		return ValidationError{"no such user"}
	}
	_, err := facades.Orm().Query().Exec(
		`INSERT INTO project_members (project_id, user_id, created_at, updated_at) VALUES (?, ?, now(), now())
		 ON CONFLICT (project_id, user_id) DO NOTHING`, p.ID, userID)
	return err
}

// RemoveMember takes read access away again.
func RemoveMember(p models.Project, userID uint) error {
	_, err := facades.Orm().Query().Exec(`DELETE FROM project_members WHERE project_id = ? AND user_id = ?`, p.ID, userID)
	return err
}

// Share links.

type ShareView struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	ExpiresAt  string `json:"expires_at"`
	LastUsedAt string `json:"last_used_at"`
	CreatedAt  string `json:"created_at"`
}

func iso(t *carbon.DateTime) string {
	if t == nil {
		return ""
	}
	return t.ToIso8601String()
}

// Shares lists p's share links.
func Shares(p models.Project) ([]ShareView, error) {
	var links []models.ShareLink
	if err := facades.Orm().Query().Where("project_id", p.ID).Order("id desc").Get(&links); err != nil {
		return nil, err
	}
	out := []ShareView{}
	for _, l := range links {
		out = append(out, ShareView{ID: l.ID, Name: l.Name, ExpiresAt: iso(l.ExpiresAt), LastUsedAt: iso(l.LastUsedAt), CreatedAt: iso(l.CreatedAt)})
	}
	return out, nil
}

// CreateShare makes a share link valid for days (0 = until revoked). The
// plain token is returned once.
func CreateShare(p models.Project, name string, days int, by models.User) (string, models.ShareLink, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", models.ShareLink{}, ValidationError{"name is required"}
	}
	if days < 0 || days > 3650 {
		return "", models.ShareLink{}, ValidationError{"expires_days must be between 0 and 3650"}
	}
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", models.ShareLink{}, err
	}
	plain := "jds_" + hex.EncodeToString(buf)
	link := models.ShareLink{ProjectID: p.ID, Name: name, TokenHash: hashToken(plain), CreatedBy: by.ID}
	if days > 0 {
		link.ExpiresAt = carbon.NewDateTime(carbon.Now().AddDays(days))
	}
	err := facades.Orm().Query().Create(&link)
	return plain, link, err
}

// ShareURL is the reader address that carries a share token.
func ShareURL(p models.Project, token string) string {
	return PageURL(p, "") + "?share=" + token
}

// RevokeShare deletes one of p's share links.
func RevokeShare(p models.Project, id uint) error {
	res, err := facades.Orm().Query().Where("project_id", p.ID).Where("id", id).Delete(&models.ShareLink{})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Email domains.

var domainRe = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$`)

// Domains lists p's allowed email domains.
func Domains(p models.Project) []string {
	out := []string{}
	for _, d := range strings.Split(p.AllowedDomains, ",") {
		if d = strings.TrimSpace(d); d != "" {
			out = append(out, d)
		}
	}
	return out
}

// NormalizeDomains validates and cleans a list like ["@Example.com", "example.org"].
func NormalizeDomains(in []string) ([]string, error) {
	seen := map[string]bool{}
	out := []string{}
	for _, d := range in {
		d = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(d), "@")))
		if d == "" || seen[d] {
			continue
		}
		if !domainRe.MatchString(d) {
			return nil, ValidationError{fmt.Sprintf("%q is not a domain like example.com", d)}
		}
		seen[d] = true
		out = append(out, d)
	}
	if len(out) > 20 {
		return nil, ValidationError{"at most 20 domains"}
	}
	return out, nil
}

// SetDomains replaces p's allowed email domains.
func SetDomains(p models.Project, domains []string) ([]string, error) {
	clean, err := NormalizeDomains(domains)
	if err != nil {
		return nil, err
	}
	_, err = facades.Orm().Query().Exec(`UPDATE projects SET allowed_domains = ? WHERE id = ?`, strings.Join(clean, ","), p.ID)
	return clean, err
}

// emailInDomains matches the part after the last "@" exactly (no
// subdomains), case-insensitively.
func emailInDomains(email string, domains []string) bool {
	i := strings.LastIndex(email, "@")
	if i < 0 || len(domains) == 0 {
		return false
	}
	host := strings.ToLower(strings.TrimSpace(email[i+1:]))
	for _, d := range domains {
		if host == d {
			return true
		}
	}
	return false
}
