// Service layer for business logic

package services

import "context"

type UserService interface {
    GetUser(ctx context.Context, id int64) (*User, error)
    CreateUser(ctx context.Context, email, firstName, lastName string) (*User, error)
    UpdateUser(ctx context.Context, id int64, user *UserUpdate) (*User, error)
    DeleteUser(ctx context.Context, id int64) error
    ListUsers(ctx context.Context, offset, limit int) ([]*User, error)
}

type ApiKeyService interface {
    CreateKey(ctx context.Context, userID int64) (*ApiKey, error)
    GetKey(ctx context.Context, id int64) (*ApiKey, error)
    RevokeKey(ctx context.Context, id int64) error
    ListKeys(ctx context.Context, userID int64) ([]*ApiKey, error)
    ValidateKey(ctx context.Context, key, secret string) (*ApiKey, error)
}

type RequestService interface {
    LogRequest(ctx context.Context, req *RequestLog) error
    GetRequestMetrics(ctx context.Context, userID int64) (*Metrics, error)
}

// Implementation
type userService struct {
    db Repository
}

func (s *userService) GetUser(ctx context.Context, id int64) (*User, error) {
    return s.db.GetUser(ctx, id)
}

func (s *userService) CreateUser(ctx context.Context, email, firstName, lastName string) (*User, error) {
    // Validation, business logic here
    return s.db.CreateUser(ctx, email, firstName, lastName)
}
