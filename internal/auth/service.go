package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	repo "github.com/rahul810050/go-store/internal/adapters/postgresql/sqlc"
)

type Service interface {
	Signup(ctx context.Context, input SignupInput) (repo.CreateUserRow, error)
	Signin(ctx context.Context, input SigninInput) (SigninResponse, error)
	Refresh(ctx context.Context, input RefreshInput) (SigninResponse, error)
	Logout(ctx context.Context, input RefreshInput) (error)
}

type service struct {
	repo repo.Querier
	db *pgxpool.Pool
}

type SigninResponse struct {
	User repo.GetUserByIDRow
	AccessToken string
	RefreshToken string
}

func NewService(db *pgxpool.Pool) Service {
	return &service{
		repo: repo.New(db),
		db: db,
	}
}

func (s *service) Signup(ctx context.Context, input SignupInput) (repo.CreateUserRow, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	_, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err == nil {
		return repo.CreateUserRow{}, ErrEmailAlreadyExists
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return repo.CreateUserRow{}, err
	}

	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		return repo.CreateUserRow{}, err
	}

	user, err := s.repo.CreateUser(ctx, repo.CreateUserParams{
		Name: input.Name,
		Email: input.Email,
		PasswordHash: hashedPassword,
	})
	
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repo.CreateUserRow{}, ErrEmailAlreadyExists
		}
		return repo.CreateUserRow{}, err
	}
	return user, nil
}


func (s *service) Signin(ctx context.Context, input SigninInput) (SigninResponse, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SigninResponse{}, ErrInvalidCredentials
		}
		return SigninResponse{}, err
	}
	valid, err := verifyPassword(input.Password, user.PasswordHash)
	if err != nil {
		return SigninResponse{}, err
	}

	if !valid {
		return SigninResponse{}, ErrInvalidCredentials
	}
	
	safeUser, err := s.repo.GetUserByID(ctx, user.ID)
	if err != nil {
		return SigninResponse{}, err
	}
	accessToken, err := generateAccessToken(user.ID, user.Role)
	if err != nil {
		return SigninResponse{}, err
	}
	familyID, err := newTokenFamilyID()
	if err != nil {
		return SigninResponse{}, err
	}
	refreshToken, err := s.saveRefreshToken(ctx, user.ID, familyID)
	if err != nil {
		return SigninResponse{}, err
	}

	return SigninResponse{
		User: safeUser,
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *service) Refresh(ctx context.Context, input RefreshInput) (SigninResponse, error) {
	tokenHash := hashRefreshToken(input.RefreshToken)

	oldToken, err := s.repo.GetValidRefreshToken(ctx, tokenHash)
	if err != nil {
		return SigninResponse{}, ErrInvalidRefreshToken
	}

	user, err := s.repo.GetUserByID(ctx, oldToken.UserID)
	if err != nil {
		return SigninResponse{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return SigninResponse{}, err
	}

	defer tx.Rollback(ctx)

	qtx := repo.New(tx)

	err = qtx.RevokeRefreshToken(ctx, tokenHash)
	if err != nil {
		return SigninResponse{}, err
	}

	newRefreshToken, err := s.saveRefreshToken(ctx, user.ID, oldToken.FamilyID)
	if err != nil {
		return SigninResponse{}, err
	}

	accessToken, err := generateAccessToken(user.ID, user.Role)
	if err != nil {
		return SigninResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return SigninResponse{}, err
	}

	return SigninResponse{
		User: user,
		AccessToken: accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *service) Logout(ctx context.Context, input RefreshInput) error {
	tokenHash := hashRefreshToken(input.RefreshToken)
	
	refreshToken, err := s.repo.GetValidRefreshToken(ctx, tokenHash)
	if err != nil {
		return ErrInvalidRefreshToken
	}
	return s.repo.RevokeRefreshTokenFamily(ctx, refreshToken.FamilyID)
}