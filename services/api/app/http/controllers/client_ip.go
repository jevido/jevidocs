package controllers

import (
	"net"
	"strings"

	"github.com/goravel/framework/contracts/http"
)

// clientIP is the address rate limits key on. Gin's ClientIP trusts every
// proxy by default, so it returns the left-most X-Forwarded-For entry, which
// any client can set. Traefik is our only proxy and appends the address it
// saw, so the right-most entry is the one we can trust; without the header
// the TCP peer is used.
func clientIP(ctx http.Context) string {
	return rightmostHop(ctx.Request().Header("X-Forwarded-For", ""), ctx.Request().Origin().RemoteAddr)
}

func rightmostHop(xff, remoteAddr string) string {
	if xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[len(parts)-1]); ip != "" {
			return ip
		}
	}
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	return remoteAddr
}
