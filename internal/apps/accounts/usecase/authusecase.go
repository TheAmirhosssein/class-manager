package usecase

import (
	"context"
	"strconv"

	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/delivery/http/handler/model"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/message"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/params"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/repository"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/validation"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/hash"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/jwt"
	"github.com/TheAmirhosssein/class-manager/internal/types"
	"github.com/TheAmirhosssein/class-manager/internal/utils/messages"
	"github.com/TheAmirhosssein/class-manager/pkg/log"
)

type AuthUsecase struct {
	userUC  *UserUsecase
	adminUC *AdminUsecase
	users   repository.UserRepository
	hash    hash.Hash
	tokens  *jwt.Service
}

func NewAuthUsecase(userUC *UserUsecase, adminUC *AdminUsecase, users repository.UserRepository, h hash.Hash, tokens *jwt.Service) *AuthUsecase {
	return &AuthUsecase{userUC: userUC, adminUC: adminUC, users: users, hash: h, tokens: tokens}
}

type LoginResult struct {
	User  model.UserJSON
	Token string
}

func (a *AuthUsecase) LoginTeacher(ctx context.Context, mobile, password string) (LoginResult, messages.Message) {
	if !validation.ValidMobile(mobile) || !validation.ValidPassword(password) {
		return LoginResult{}, message.UserInvalidPassword
	}
	user, err := a.users.ReadByMobile(ctx, mobile)
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
		Mobile:    user.Mobile,
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

func (a *AuthUsecase) LoginAdmin(ctx context.Context, mobile, password string) (LoginResult, messages.Message) {
	admin, msg := a.adminUC.Login(ctx, params.LoginAdminParams{Mobile: mobile, Password: password})
	if msg.Error != nil {
		return LoginResult{}, msg
	}
	token, err := a.tokens.Issue(jwt.IssueParams{
		SubjectID: admin.ID,
		Kind:      jwt.SubjectKindAdmin,
		Role:      "admin",
		Mobile:    admin.Mobile,
		Name:      admin.Name,
	})
	if err != nil {
		return LoginResult{}, messages.SharedServerError()
	}
	return LoginResult{
		User:  model.NewAdminJSON(admin),
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

func (a *AuthUsecase) SessionAdmin(ctx context.Context, token string) (model.UserJSON, messages.Message) {
	claims, err := a.tokens.Parse(token)
	if err != nil {
		return model.UserJSON{}, message.SessionInvalid
	}
	id, err := claims.SubjectID()
	if err != nil || claims.Kind != jwt.SubjectKindAdmin {
		return model.UserJSON{}, message.SessionInvalid
	}
	admin, msg := a.adminUC.ReadByID(ctx, id)
	if msg.Error != nil {
		return model.UserJSON{}, msg
	}
	return model.NewAdminJSON(admin), message.LoginSuccess
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
	if claims.Kind == jwt.SubjectKindAdmin {
		admin, msg := a.adminUC.ReadByID(ctx, id)
		if msg.Error != nil {
			return model.UserJSON{}, msg
		}
		return model.NewAdminJSON(admin), message.LoginSuccess
	}
	return model.UserJSON{
		ID:     strconv.FormatInt(int64(id), 10),
		Name:   claims.Name,
		Role:   claims.Role,
		Mobile: claims.Mobile,
	}, message.LoginSuccess
}

func ActorID(claims jwt.Claims) (types.ID, bool) {
	id, err := claims.SubjectID()
	if err != nil {
		return 0, false
	}
	return id, true
}
