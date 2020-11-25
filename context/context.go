package context

import (
	"context"
	"go-web-dev/models/user"
)

const (
	userKey privateKey = "user"
)

type privateKey string

func WithUser(ctx context.Context, user *user.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func User(ctx context.Context) *user.User {
	if temp := ctx.Value(userKey); temp != nil {
		if usr, ok := temp.(*user.User); ok {
			return usr
		}
	}
	return nil
}
