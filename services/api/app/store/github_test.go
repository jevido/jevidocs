package store

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func tarball(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	_ = tw.WriteHeader(&tar.Header{Name: "repo-abc123/", Typeflag: tar.TypeDir, Mode: 0o755})
	for name, body := range files {
		if err := tw.WriteHeader(&tar.Header{Name: "repo-abc123/" + name, Typeflag: tar.TypeReg, Mode: 0o644, Size: int64(len(body))}); err != nil {
			t.Fatal(err)
		}
		_, _ = tw.Write([]byte(body))
	}
	_ = tw.Close()
	_ = gz.Close()
	return buf.Bytes()
}

func TestExtractMarkdown(t *testing.T) {
	data := tarball(t, map[string]string{
		"README.md":               "root readme",
		"docs/index.md":           "# Home",
		"docs/guides/install.mdx": "# Install",
		"docs/logo.png":           "png",
		"docsx/other.md":          "not in docs",
	})
	fs, err := ExtractMarkdown(bytes.NewReader(data), "/docs/")
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 2 || fs["index.md"] == nil || fs["guides/install.mdx"] == nil {
		t.Errorf("files = %v", keys(fs))
	}
	all, _ := ExtractMarkdown(bytes.NewReader(data), "")
	if len(all) != 4 {
		t.Errorf("whole repo = %v", keys(all))
	}
}

func keys[M ~map[string]V, V any](m M) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestVerifyGitHubSignature(t *testing.T) {
	body := []byte(`{"ref":"refs/heads/main"}`)
	mac := hmac.New(sha256.New, []byte("s3cret"))
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if !VerifyGitHubSignature("s3cret", body, sig) {
		t.Error("valid signature rejected")
	}
	if VerifyGitHubSignature("other", body, sig) || VerifyGitHubSignature("s3cret", body, "sha256=00") ||
		VerifyGitHubSignature("", body, sig) || VerifyGitHubSignature("s3cret", body, "") {
		t.Error("invalid signature accepted")
	}
}

func TestCleanSourcePath(t *testing.T) {
	for _, bad := range []string{"../etc", "docs/../../x", "a;b"} {
		if _, err := cleanSourcePath(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	if got, _ := cleanSourcePath("/apps/docs/content/docs/"); got != "apps/docs/content/docs" {
		t.Errorf("got %q", got)
	}
}

func TestRepoPathToPage(t *testing.T) {
	for in, want := range map[string]string{
		"(framework)/comparisons.mdx": "comparisons.mdx",
		"UI/Theme Options.md":         "ui/theme-options.md",
		"(a)/(b)/index.md":            "index.md",
		"___.md":                      "",
	} {
		if got := repoPathToPage(in); got != want {
			t.Errorf("repoPathToPage(%q) = %q, want %q", in, got, want)
		}
	}
}
