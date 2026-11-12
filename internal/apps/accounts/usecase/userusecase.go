package usecase

import (
	"context"
	"strings"

	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/entity"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/enum"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/message"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/params"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/repository"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/validation"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/hash"
	"github.com/TheAmirhosssein/class-manager/internal/types"
	"github.com/TheAmirhosssein/class-manager/internal/utils/messages"
	"github.com/TheAmirhosssein/class-manager/pkg/log"
)

type UserUsecase struct {
	users  repository.UserRepository
	admins repository.AdminRepository
	hash   hash.Hash
}

func NewUserUsecase(users repository.UserRepository, admins repository.AdminRepository, h hash.Hash) *UserUsecase {
	return &UserUsecase{users: users, admins: admins, hash: h}
}

func (u *UserUsecase) ReadByID(ctx context.Context, id types.ID) (entity.User, messages.Message) {
	user, err := u.users.ReadByID(ctx, id)
	if err != nil {
		log.Printf("user read by id: %v", err)
		return entity.User{}, messages.SharedServerError()
	}
	if !user.ID.Valid() {
		return entity.User{}, message.UserNotFound
	}
	return user, message.UserRead
}

func (u *UserUsecase) CreateByAdmin(ctx context.Context, p params.CreateUserUsecaseParams) (entity.User, messages.Message) {
	if !p.ActorAdminID.Valid() {
		return entity.User{}, message.AdminNotFound
	}
	actor, err := u.admins.ReadByID(ctx, p.ActorAdminID)
	if err != nil {
		log.Printf("user create actor read: %v", err)
		return entity.User{}, messages.SharedServerError()
	}
	if !actor.ID.Valid() {
		return entity.User{}, message.AdminNotFound
	}
	if !validation.ValidMobile(p.Mobile) || !validation.ValidUserName(p.Name) || !validation.ValidPassword(p.Password) {
		return entity.User{}, message.UserInvalidInput
	}
	exists, err := u.users.ExistByMobile(ctx, p.Mobile)
	if err != nil {
		log.Printf("user create exist check: %v", err)
		return entity.User{}, messages.SharedServerError()
	}
	if exists {
		return entity.User{}, message.UserMobileExists
	}
	hashVal, err := u.hash.Encrypt(p.Password)
	if err != nil {
		log.Printf("user create hash: %v", err)
		return entity.User{}, messages.SharedServerError()
	}
	in := entity.User{
		Mobile:       validation.NormalizeMobile(p.Mobile),
		PasswordHash: hashVal,
		Name:         strings.TrimSpace(p.Name),
		Role:         enum.RoleTeacher,
	}
	created, err := u.users.Create(ctx, in)
	if err != nil {
		log.Printf("user create insert: %v", err)
		return entity.User{}, messages.SharedServerError()
	}
	return created, message.UserCreated
}
