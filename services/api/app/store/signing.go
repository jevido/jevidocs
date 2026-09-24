package store

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"dev.jevido/jevidocs/services/api/app/docs"
	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// PreviewTTL is how long a draft preview link works.
const PreviewTTL = 7 * 24 * time.Hour

func appKey() []byte { return []byte(facades.Config().GetString("app.key")) }

func sign(key []byte, msg string) string {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(msg))
	return hex.EncodeToString(m.Sum(nil))[:32]
}

// previewToken is "<pageid>.<expiryunix>.<sig>".
func previewToken(key []byte, pageID uint, expires time.Time) string {
	msg := fmt.Sprintf("preview:%d.%d", pageID, expires.Unix())
	return fmt.Sprintf("%d.%d.%s", pageID, expires.Unix(), sign(key, msg))
}

// checkPreviewToken returns the page ID a valid, unexpired token is for.
func checkPreviewToken(key []byte, token string, now time.Time) (uint, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, false
	}
	id, err1 := strconv.ParseUint(parts[0], 10, 64)
	exp, err2 := strconv.ParseInt(parts[1], 10, 64)
	if err1 != nil || err2 != nil || now.Unix() > exp {
		return 0, false
	}
	want := sign(key, fmt.Sprintf("preview:%d.%d", id, exp))
	if !hmac.Equal([]byte(want), []byte(parts[2])) {
		return 0, false
	}
	return uint(id), true
}

func assetSig(key []byte, assetID uint) string { return sign(key, fmt.Sprintf("asset:%d", assetID)) }

// CheckAssetSig reports whether sig authorises reading a private project's asset.
func CheckAssetSig(assetID uint, sig string) bool {
	return hmac.Equal([]byte(assetSig(appKey(), assetID)), []byte(sig))
}

// PreviewLink is a shareable URL that shows pg even while unpublished.
func PreviewLink(p models.Project, pg models.Page) string {
	return PageURL(p, pg.Slug) + "?preview=" + previewToken(appKey(), pg.ID, time.Now().Add(PreviewTTL))
}

// ViewPreview shows the page a preview token points at, published or not.
func ViewPreview(p models.Project, slug, token string) (PageView, error) {
	id, ok := checkPreviewToken(appKey(), token, time.Now())
	if !ok {
		return PageView{}, ErrNotFound
	}
	pg, err := FindPageByID(p, id)
	if err != nil {
		return PageView{}, err
	}
	if pg.Slug != docs.NormalizeSlug(slug) {
		return PageView{}, ErrNotFound
	}
	v, err := viewOf(p, pg)
	v.Draft = !pg.Published
	return v, err
}

// AssetReadable: public projects' assets are open; private ones need sig.
func AssetReadable(a models.Asset, sig string) bool {
	var p models.Project
	if err := facades.Orm().Query().Where("id", a.ProjectID).First(&p); err != nil || p.ID == 0 {
		return false
	}
	return p.Public || CheckAssetSig(a.ID, sig)
}
