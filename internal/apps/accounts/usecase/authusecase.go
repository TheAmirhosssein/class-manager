package usecase

import (
	"context"
	"strconv"
	"strings"

	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/delivery/http/handler/model"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/message"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/repository"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/validation"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/hash"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/jwt"
	"github.com/TheAmirhosssein/class-manager/internal/types"
	"github.com/TheAmirhosssein/class-manager/internal/utils/messages"
	"github.com/TheAmirhosssein/class-manager/pkg/log"
)

type AuthUsecase struct {
	userUC *UserUsecase
	users  repository.UserRepository
	hash   hash.Hash
	tokens *jwt.Service
}

func NewAuthUsecase(userUC *UserUsecase, users repository.UserRepository, h hash.Hash, tokens *jwt.Service) *AuthUsecase {
	return &AuthUsecase{userUC: userUC, users: users, hash: h, tokens: tokens}
}

type LoginResult struct {
	User  model.UserJSON
	Token string
}

func (a *AuthUsecase) LoginTeacher(ctx context.Context, email, password string) (LoginResult, messages.Message) {
	if !validation.ValidEmail(email) || !validation.ValidPassword(password) {
		return LoginResult{}, message.UserInvalidPassword
	}
	user, err := a.users.ReadByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		log.Printf("auth login read: %v", err)
		return LoginResult{}, messages.SharedServerError()
	}
	if !user.ID.Valid() {
		return LoginResult{}, message.UserInvalidPassword
	}
	if err := a.hash.Compare(user.PasswordHash, password); err != nil {
		return LoginResult{}, message.UserInvalidPassword
	}
	token, err := a.tokens.Issue(jwt.IssueParams{
		SubjectID: user.ID,
		Kind:      jwt.SubjectKindUser,
		Role:      string(user.Role),
		Email:     user.Email,
		Name:      user.Name,
	})
	if err != nil {
		return LoginResult{}, messages.SharedServerError()
	}
	return LoginResult{
		User:  model.NewUserJSON(user),
		Token: token,
	}, message.LoginSuccess
}

func (a *AuthUsecase) Session(ctx context.Context, token string) (model.UserJSON, messages.Message) {
	claims, err := a.tokens.Parse(token)
	if err != nil {
		return model.UserJSON{}, message.SessionInvalid
	}
	id, err := claims.SubjectID()
	if err != nil || claims.Kind != jwt.SubjectKindUser {
		return model.UserJSON{}, message.SessionInvalid
	}
	user, msg := a.userUC.ReadByID(ctx, id)
	if msg.Error != nil {
		return model.UserJSON{}, msg
	}
	return model.NewUserJSON(user), message.LoginSuccess
}

func (a *AuthUsecase) SessionFromClaims(ctx context.Context, claims jwt.Claims) (model.UserJSON, messages.Message) {
	id, err := claims.SubjectID()
	if err != nil {
		return model.UserJSON{}, message.SessionInvalid
	}
	if claims.Kind == jwt.SubjectKindUser {
		user, msg := a.userUC.ReadByID(ctx, id)
		if msg.Error != nil {
			return model.UserJSON{}, msg
		}
		return model.NewUserJSON(user), message.LoginSuccess
	}
	return model.UserJSON{
		ID:    strconv.FormatInt(int64(id), 10),
		Name:  claims.Name,
		Role:  claims.Role,
		Email: claims.Email,
	}, message.LoginSuccess
}

func ActorID(claims jwt.Claims) (types.ID, bool) {
	id, err := claims.SubjectID()
	if err != nil {
		return 0, false
	}
	return id, true
}
