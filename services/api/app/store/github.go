package store

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"testing/fstest"
	"time"

	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// Limits for a GitHub source download.
const (
	sourceTimeout  = 20 * time.Second
	sourceMaxBytes = 50 << 20
	sourceMaxFiles = 2000
	sourceMaxFile  = 1 << 20
)

var (
	repoRe    = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
	refRe     = regexp.MustCompile(`^[A-Za-z0-9_./-]{1,200}$`)
	srcPathRe = regexp.MustCompile(`^[A-Za-z0-9_./ -]{0,300}$`)
)

// SourceView is a project's GitHub source as admins see it.
type SourceView struct {
	Repo       string `json:"repo"`
	Ref        string `json:"ref"`
	Path       string `json:"path"`
	Secret     string `json:"secret"`
	SyncedAt   string `json:"synced_at"`
	Status     string `json:"status"`
	WebhookURL string `json:"webhook_url"`
}

// SourceInput sets a project's GitHub source. An empty repo disconnects it.
type SourceInput struct {
	Repo string `json:"repo"`
	Ref  string `json:"ref"`
	Path string `json:"path"`
}

func ViewSource(p models.Project) SourceView {
	v := SourceView{Repo: p.SourceRepo, Ref: p.SourceRef, Path: p.SourcePath, Secret: p.SourceSecret, Status: p.SourceStatus,
		WebhookURL: strings.TrimRight(facades.Config().GetString("http.url"), "/") + "/api/hooks/github/" + p.Slug}
	if p.SourceSyncedAt != nil {
		v.SyncedAt = p.SourceSyncedAt.ToIso8601String()
	}
	return v
}

func cleanSourcePath(s string) (string, error) {
	s = strings.Trim(strings.TrimSpace(s), "/")
	if !srcPathRe.MatchString(s) {
		return "", ValidationError{"path may only contain letters, digits, _ . - / and spaces"}
	}
	for _, seg := range strings.Split(s, "/") {
		if seg == ".." || seg == "." {
			return "", ValidationError{"path must not contain . or .. segments"}
		}
	}
	return s, nil
}

// SaveSource validates and stores a project's GitHub source, generating the
// webhook secret the first time.
func SaveSource(p models.Project, in SourceInput) (models.Project, error) {
	repo := strings.TrimSpace(in.Repo)
	repo = strings.TrimSuffix(strings.TrimPrefix(repo, "https://github.com/"), ".git")
	ref := strings.TrimSpace(in.Ref)
	if ref == "" {
		ref = "main"
	}
	dir, err := cleanSourcePath(in.Path)
	if err != nil {
		return p, err
	}
	if repo != "" && !repoRe.MatchString(repo) {
		return p, ValidationError{"repo must look like owner/name"}
	}
	if !refRe.MatchString(ref) || strings.Contains(ref, "..") {
		return p, ValidationError{"ref must be a branch or tag name"}
	}
	p.SourceRepo, p.SourceRef, p.SourcePath = repo, ref, dir
	if repo != "" && p.SourceSecret == "" {
		buf := make([]byte, 16)
		if _, err := rand.Read(buf); err != nil {
			return p, err
		}
		p.SourceSecret = hex.EncodeToString(buf)
	}
	return p, facades.Orm().Query().Save(&p)
}

// VerifyGitHubSignature checks an X-Hub-Signature-256 header against body.
func VerifyGitHubSignature(secret string, body []byte, header string) bool {
	if secret == "" || !strings.HasPrefix(header, "sha256=") {
		return false
	}
	got, err := hex.DecodeString(strings.TrimPrefix(header, "sha256="))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hmac.Equal(got, mac.Sum(nil))
}

// ExtractMarkdown reads a GitHub tarball (gzip'd tar with one top-level
// directory) and returns the .md/.mdx files under dir, keyed by their path
// relative to dir.
func ExtractMarkdown(r io.Reader, dir string) (fstest.MapFS, error) {
	gz, err := gzip.NewReader(io.LimitReader(r, sourceMaxBytes))
	if err != nil {
		return nil, fmt.Errorf("not a gzip archive: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	out := fstest.MapFS{}
	dir = strings.Trim(dir, "/")
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read archive: %w", err)
		}
		if h.Typeflag != tar.TypeReg {
			continue
		}
		// Drop the "<repo>-<sha>/" top directory GitHub adds.
		_, name, ok := strings.Cut(h.Name, "/")
		if !ok {
			continue
		}
		name = path.Clean(name)
		if strings.HasPrefix(name, "../") || name == ".." {
			continue
		}
		rel := name
		if dir != "" {
			if !strings.HasPrefix(name, dir+"/") {
				continue
			}
			rel = strings.TrimPrefix(name, dir+"/")
		}
		ext := path.Ext(rel)
		if ext != ".md" && ext != ".mdx" || h.Size > sourceMaxFile {
			continue
		}
		if len(out) >= sourceMaxFiles {
			return nil, fmt.Errorf("more than %d Markdown files", sourceMaxFiles)
		}
		rel = repoPathToPage(rel)
		if rel == "" {
			continue
		}
		data, err := io.ReadAll(io.LimitReader(tr, sourceMaxFile+1))
		if err != nil {
			return nil, err
		}
		out[rel] = &fstest.MapFile{Data: data}
	}
	return out, nil
}

// sourceLocks keeps two syncs of one project from running at once.
var sourceLocks sync.Map

// SyncFromGitHub downloads the project's source folder and syncs its pages
// (pruning pages without a file). The outcome is recorded on the project.
func SyncFromGitHub(p models.Project) (SyncResult, error) {
	var res SyncResult
	if p.SourceRepo == "" {
		return res, ValidationError{"project has no GitHub source"}
	}
	mu, _ := sourceLocks.LoadOrStore(p.ID, &sync.Mutex{})
	mu.(*sync.Mutex).Lock()
	defer mu.(*sync.Mutex).Unlock()

	res, err := syncFromGitHub(p)
	status := fmt.Sprintf("ok: %d created, %d updated, %d unchanged, %d deleted", res.Created, res.Updated, res.Unchanged, res.Deleted)
	if err != nil {
		status = "error: " + err.Error()
	}
	if _, uerr := facades.Orm().Query().Exec(`UPDATE projects SET source_status = ?, source_synced_at = now() WHERE id = ?`, status, p.ID); uerr != nil && err == nil {
		err = uerr
	}
	return res, err
}

func syncFromGitHub(p models.Project) (SyncResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sourceTimeout)
	defer cancel()
	url := fmt.Sprintf("https://codeload.github.com/%s/tar.gz/%s", p.SourceRepo, p.SourceRef)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return SyncResult{}, err
	}
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return SyncResult{}, fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return SyncResult{}, fmt.Errorf("download %s@%s: HTTP %d", p.SourceRepo, p.SourceRef, resp.StatusCode)
	}
	files, err := ExtractMarkdown(resp.Body, p.SourcePath)
	if err != nil {
		return SyncResult{}, err
	}
	if len(files) == 0 {
		return SyncResult{}, fmt.Errorf("no .md or .mdx files under %q", p.SourcePath)
	}
	return SyncPages(p, files, true)
}

// SyncFromGitHubAsync starts a sync in the background; its result shows up
// in the project's source status.
func SyncFromGitHubAsync(p models.Project) {
	go func() {
		if _, err := SyncFromGitHub(p); err != nil {
			facades.Log().Warningf("github sync %s: %v", p.Slug, err)
		}
	}()
}

var unsafeSegRe = regexp.MustCompile(`[^a-z0-9._-]+`)

// repoPathToPage turns a file path from a repository into one whose slug is
// valid: route-group folders like "(framework)" (Next.js, fumadocs) are
// dropped, and segments are lowercased with other characters as dashes.
func repoPathToPage(rel string) string {
	ext := path.Ext(rel)
	parts := strings.Split(strings.TrimSuffix(rel, ext), "/")
	var out []string
	for _, seg := range parts {
		if strings.HasPrefix(seg, "(") && strings.HasSuffix(seg, ")") {
			continue
		}
		seg = strings.Trim(unsafeSegRe.ReplaceAllString(strings.ToLower(seg), "-"), "-._")
		if seg == "" {
			return ""
		}
		out = append(out, seg)
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, "/") + ext
}
