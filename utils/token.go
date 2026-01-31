package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

var (
	AccessTokenTTL  = 24 * time.Minute
	RefreshTokenTTL = 30 * 24 * time.Hour

	ErrInvalidToken       = errors.New("invalid or expired refresh token")
	ErrTokenReused        = errors.New("refresh token revoked or reused")
	ErrInvalidAccessToken = errors.New("invalid or expired access token")
	ErrNoActiveSessions   = errors.New("no active sessions found")
)

type TokenRotator struct {
	Redis              *redis.Client
	AccessTokenSecret  []byte
	RefreshTokenSecret []byte
}

type RotateResult struct {
	AccessToken  string
	RefreshToken string
}

func NewTokenRotator() *TokenRotator {
	return &TokenRotator{
		Redis:              GetRedisClient(),
		AccessTokenSecret:  []byte(GetConfigValue("JWT_ACCESS_SECRET")),
		RefreshTokenSecret: []byte(GetConfigValue("JWT_REFRESH_SECRET")),
	}
}

func generateJTI() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (t *TokenRotator) Rotate(
	ctx context.Context,
	refreshToken string,
) (*RotateResult, error) {

	parsed, err := jwt.Parse(refreshToken, func(token *jwt.Token) (any, error) {
		return t.RefreshTokenSecret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	userId, ok1 := claims["userId"].(string)
	email, ok2 := claims["email"].(string)
	oldJTI, ok3 := claims["jti"].(string)

	if !ok1 || !ok2 || !ok3 {
		return nil, ErrInvalidToken
	}

	key := "refresh:" + oldJTI
	exists, err := t.Redis.Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, ErrTokenReused
	}

	pipe := t.Redis.TxPipeline()
	pipe.Del(ctx, key)
	pipe.SRem(ctx, "user_sessions:"+userId, oldJTI)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	newJTI, err := generateJTI()
	if err != nil {
		return nil, err
	}

	refreshClaims := jwt.MapClaims{
		"userId": userId,
		"email":  email,
		"jti":    newJTI,
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(RefreshTokenTTL).Unix(),
	}

	newRefreshToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		refreshClaims,
	).SignedString(t.RefreshTokenSecret)
	if err != nil {
		return nil, err
	}

	pipe = t.Redis.TxPipeline()
	pipe.Set(ctx, "refresh:"+newJTI, userId, RefreshTokenTTL)
	pipe.SAdd(ctx, "user_sessions:"+userId, newJTI)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	accessClaims := jwt.MapClaims{
		"userId": userId,
		"email":  email,
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(AccessTokenTTL).Unix(),
	}

	newAccessToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		accessClaims,
	).SignedString(t.AccessTokenSecret)
	if err != nil {
		return nil, err
	}

	return &RotateResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}


func (t *TokenRotator) Revoke(ctx context.Context, accessToken string) error {
	parsed, err := jwt.Parse(accessToken, func(token *jwt.Token) (any, error) {
		return t.AccessTokenSecret, nil
	})
	if err != nil || !parsed.Valid {
		return ErrInvalidAccessToken
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return ErrInvalidAccessToken
	}

	userId, ok := claims["userId"].(string)
	if !ok {
		return ErrInvalidAccessToken
	}

	sessionKey := "user_sessions:" + userId
	jtis, err := t.Redis.SMembers(ctx, sessionKey).Result()
	if err != nil {
		return err
	}

	if len(jtis) == 0 {
		return ErrNoActiveSessions
	}

	pipe := t.Redis.TxPipeline()
	for _, jti := range jtis {
		pipe.Del(ctx, "refresh:"+jti)
	}
	pipe.Del(ctx, sessionKey)

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return nil
}
