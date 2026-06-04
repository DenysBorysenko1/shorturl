// Package context provides context utilities for managing
// user authentication context throughout the application.
package context

import "context"

type userIDKey struct{}

var ctxKeyUserID = userIDKey{}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, userID)
}

func UserID(ctx context.Context) (string, bool) {
	s, ok := ctx.Value(ctxKeyUserID).(string)
	return s, ok
}
