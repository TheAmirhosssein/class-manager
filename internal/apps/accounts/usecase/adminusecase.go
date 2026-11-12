package usecase

import (
	"context"
	"strings"

	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/entity"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/message"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/params"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/repository"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/validation"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/hash"
	"github.com/TheAmirhosssein/class-manager/internal/types"
	"github.com/TheAmirhosssein/class-manager/internal/utils/messages"
	"github.com/TheAmirhosssein/class-manager/pkg/log"
)

type AdminUsecase struct {
	admins repository.AdminRepository
	hash   hash.Hash
}

func NewAdminUsecase(admins repository.AdminRepository, h hash.Hash) *AdminUsecase {
	return &AdminUsecase{admins: admins, hash: h}
}

func (u *AdminUsecase) ReadByID(ctx context.Context, id types.ID) (entity.Admin, messages.Message) {
	admin, err := u.admins.ReadByID(ctx, id)
	if err != nil {
		log.Printf("admin read by id: %v", err)
		return entity.Admin{}, messages.SharedServerError()
	}
	if !admin.ID.Valid() {
		return entity.Admin{}, message.AdminNotFound
	}
	return admin, message.AdminRead
}

func (u *AdminUsecase) Create(ctx context.Context, p params.CreateAdminParams) (entity.Admin, messages.Message) {
	if !validation.ValidEmail(p.Email) || !validation.ValidUserName(p.Name) || !validation.ValidPassword(p.Password) {
		return entity.Admin{}, message.AdminInvalidInput
	}
	exists, err := u.admins.ExistByEmail(ctx, p.Email)
	if err != nil {
		log.Printf("admin create exist check: %v", err)
		return entity.Admin{}, messages.SharedServerError()
	}
	if exists {
		return entity.Admin{}, message.AdminEmailExists
	}
	hashVal, err := u.hash.Encrypt(p.Password)
	if err != nil {
		log.Printf("admin create hash: %v", err)
		return entity.Admin{}, messages.SharedServerError()
	}
	in := entity.Admin{
		Email:        strings.TrimSpace(p.Email),
		PasswordHash: hashVal,
		Name:         strings.TrimSpace(p.Name),
	}
	created, err := u.admins.Create(ctx, in)
	if err != nil {
		log.Printf("admin create insert: %v", err)
		return entity.Admin{}, messages.SharedServerError()
	}
	return created, message.AdminCreated
}

func (u *AdminUsecase) Login(ctx context.Context, p params.LoginAdminParams) (entity.Admin, messages.Message) {
	if !validation.ValidEmail(p.Email) || !validation.ValidPassword(p.Password) {
		return entity.Admin{}, message.AdminInvalidPassword
	}
	admin, err := u.admins.ReadByEmail(ctx, strings.TrimSpace(p.Email))
	if err != nil {
		log.Printf("admin login read: %v", err)
		return entity.Admin{}, messages.SharedServerError()
	}
	if !admin.ID.Valid() {
		return entity.Admin{}, message.AdminInvalidPassword
	}
	if err := u.hash.Compare(admin.PasswordHash, p.Password); err != nil {
		return entity.Admin{}, message.AdminInvalidPassword
	}
	return admin, message.LoginSuccess
}
