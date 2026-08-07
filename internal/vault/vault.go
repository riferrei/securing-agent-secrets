// Package vault resolves secrets from 1Password via a service account.
package vault

import (
	"context"
	"fmt"
	"net"
	"net/url"

	onepassword "github.com/1password/onepassword-sdk-go"
)

// ResolveRedisURL resolves the host, port, user, and password references from
// 1Password and assembles the Redis connection string.
func ResolveRedisURL(ctx context.Context, token, hostRef, portRef, userRef, passRef string) (string, error) {
	client, err := onepassword.NewClient(ctx,
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo("securing-agent-secrets", "v1.0.0"),
	)
	if err != nil {
		return "", fmt.Errorf("creating 1password client: %w", err)
	}

	host, err := client.Secrets().Resolve(ctx, hostRef)
	if err != nil {
		return "", fmt.Errorf("resolving %s: %w", hostRef, err)
	}
	port, err := client.Secrets().Resolve(ctx, portRef)
	if err != nil {
		return "", fmt.Errorf("resolving %s: %w", portRef, err)
	}
	user, err := client.Secrets().Resolve(ctx, userRef)
	if err != nil {
		return "", fmt.Errorf("resolving %s: %w", userRef, err)
	}
	password, err := client.Secrets().Resolve(ctx, passRef)
	if err != nil {
		return "", fmt.Errorf("resolving %s: %w", passRef, err)
	}

	u := url.URL{Scheme: "redis", User: url.UserPassword(user, password), Host: net.JoinHostPort(host, port)}
	return u.String(), nil
}
