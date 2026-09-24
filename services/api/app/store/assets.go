package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"path"
	"strings"

	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// MaxAssetSize is the largest upload accepted.
const MaxAssetSize = 8 << 20

// AllowedAssetTypes are the content types that may be uploaded.
var AllowedAssetTypes = map[string]bool{
	"image/png": true, "image/jpeg": true, "image/gif": true, "image/webp": true,
	"image/svg+xml": true, "image/avif": true, "application/pdf": true, "text/plain": true,
}

var extTypes = map[string]string{
	".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".gif": "image/gif",
	".webp": "image/webp", ".svg": "image/svg+xml", ".avif": "image/avif", ".pdf": "application/pdf",
	".txt": "text/plain",
}

type AssetView struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	CreatedAt   string `json:"created_at"`
}

// AssetURL is the public address of an asset.
func AssetURL(a models.Asset) string {
	base := strings.TrimRight(facades.Config().GetString("http.url"), "/")
	return fmt.Sprintf("%s/api/assets/%d/%s", base, a.ID, url.PathEscape(a.Name))
}

func ViewAsset(a models.Asset) AssetView {
	v := AssetView{ID: a.ID, Name: a.Name, URL: AssetURL(a), ContentType: a.ContentType, Size: a.Size}
	if a.CreatedAt != nil {
		v.CreatedAt = a.CreatedAt.ToIso8601String()
	}
	return v
}

// cleanAssetName keeps a safe base name for URLs.
func cleanAssetName(name string) string {
	name = path.Base(strings.ReplaceAll(name, "\\", "/"))
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteByte('-')
		}
	}
	s := strings.Trim(b.String(), ".")
	if s == "" {
		s = "file"
	}
	if len(s) > 120 {
		s = s[len(s)-120:]
	}
	return s
}

// AssetContentType decides the stored content type from the declared one
// and the file extension; an extension match wins over a generic declared
// type like application/octet-stream.
func AssetContentType(name, declared string) string {
	declared = strings.ToLower(strings.TrimSpace(strings.Split(declared, ";")[0]))
	if AllowedAssetTypes[declared] {
		return declared
	}
	return extTypes[strings.ToLower(path.Ext(name))]
}

// SaveAsset stores data for p, returning the existing asset when the same
// bytes were uploaded before.
func SaveAsset(p models.Project, name, contentType string, data []byte) (models.Asset, error) {
	var a models.Asset
	if len(data) == 0 {
		return a, ValidationError{"file is empty"}
	}
	if len(data) > MaxAssetSize {
		return a, ValidationError{"file is larger than 8 MB"}
	}
	name = cleanAssetName(name)
	ct := AssetContentType(name, contentType)
	if ct == "" {
		return a, ValidationError{"file type not allowed: use png, jpeg, gif, webp, svg, avif, pdf or txt"}
	}
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])
	if err := facades.Orm().Query().Select("id", "project_id", "name", "content_type", "size", "sha256", "created_at", "updated_at").
		Where("project_id", p.ID).Where("sha256", hash).First(&a); err != nil {
		return a, err
	}
	if a.ID != 0 {
		return a, nil
	}
	a = models.Asset{ProjectID: p.ID, Name: name, ContentType: ct, Size: int64(len(data)), Sha256: hash, Data: data}
	err := facades.Orm().Query().Create(&a)
	return a, err
}

// ListAssets returns p's assets without their bytes, newest first.
func ListAssets(p models.Project) ([]AssetView, error) {
	var as []models.Asset
	if err := facades.Orm().Query().Select("id", "project_id", "name", "content_type", "size", "sha256", "created_at", "updated_at").
		Where("project_id", p.ID).Order("id desc").Get(&as); err != nil {
		return nil, err
	}
	out := make([]AssetView, 0, len(as))
	for _, a := range as {
		out = append(out, ViewAsset(a))
	}
	return out, nil
}

// DeleteAsset removes one asset of p.
func DeleteAsset(p models.Project, id uint) error {
	res, err := facades.Orm().Query().Where("project_id", p.ID).Where("id", id).Delete(&models.Asset{})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// FindAsset loads an asset with its bytes, if its project is public or all.
func FindAsset(id uint) (models.Asset, error) {
	var a models.Asset
	if err := facades.Orm().Query().Where("id", id).First(&a); err != nil {
		return a, err
	}
	if a.ID == 0 {
		return a, ErrNotFound
	}
	return a, nil
}
