package store

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// A handoff code carries an admin-app session over to the docs site, which
// lives on another origin and cannot read the admin's token. The admin asks
// for a code, opens the docs site with ?handoff=<code>, and the site trades
// it for its own session. Codes live one minute and work once.
const handoffTTL = time.Minute

var usedHandoffs = struct {
	sync.Mutex
	m map[string]time.Time
}{m: map[string]time.Time{}}

// handoffCode is "<userid>.<expiryunix>.<nonce>.<sig>".
func handoffCode(key []byte, userID uint, expires time.Time) (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	nonce := hex.EncodeToString(buf)
	msg := fmt.Sprintf("handoff:%d.%d.%s", userID, expires.Unix(), nonce)
	return fmt.Sprintf("%d.%d.%s.%s", userID, expires.Unix(), nonce, sign(key, msg)), nil
}

func checkHandoffCode(key []byte, code string, now time.Time) (uint, bool) {
	parts := strings.Split(code, ".")
	if len(parts) != 4 {
		return 0, false
	}
	id, err1 := strconv.ParseUint(parts[0], 10, 64)
	exp, err2 := strconv.ParseInt(parts[1], 10, 64)
	if err1 != nil || err2 != nil || now.Unix() > exp {
		return 0, false
	}
	want := sign(key, fmt.Sprintf("handoff:%s.%s.%s", parts[0], parts[1], parts[2]))
	if !hmac.Equal([]byte(want), []byte(parts[3])) {
		return 0, false
	}
	return uint(id), true
}

// IssueHandoff makes a one-minute, single-use code for u.
func IssueHandoff(u models.User) (string, error) {
	return handoffCode(appKey(), u.ID, time.Now().Add(handoffTTL))
}

// RedeemHandoff trades a handoff code for a new session token.
func RedeemHandoff(code string) (string, models.User, error) {
	now := time.Now()
	id, ok := checkHandoffCode(appKey(), code, now)
	if !ok {
		return "", models.User{}, ErrNotFound
	}
	usedHandoffs.Lock()
	for c, t := range usedHandoffs.m {
		if now.Sub(t) > 2*handoffTTL {
			delete(usedHandoffs.m, c)
		}
	}
	if _, used := usedHandoffs.m[code]; used {
		usedHandoffs.Unlock()
		return "", models.User{}, ErrNotFound
	}
	usedHandoffs.m[code] = now
	usedHandoffs.Unlock()

	var u models.User
	if err := facades.Orm().Query().Where("id", id).First(&u); err != nil || u.ID == 0 {
		return "", u, ErrNotFound
	}
	plain, _, err := IssueToken(u, "docs site session", "session")
	return plain, u, err
}
