package message

import "github.com/TheAmirhosssein/class-manager/internal/utils/messages"

var (
	UserNotFound        = messages.NewMessageByError(404_101, "کاربر یافت نشد")
	UserMobileExists    = messages.NewMessageByError(409_101, "این شماره موبایل قبلاً ثبت شده است")
	UserInvalidInput    = messages.NewMessageByError(400_101, "اطلاعات کاربر نامعتبر است")
	UserInvalidPassword = messages.NewMessageByError(401_101, "شماره موبایل یا رمز عبور نادرست است")

	AdminNotFound        = messages.NewMessageByError(404_110, "مدیر یافت نشد")
	AdminMobileExists    = messages.NewMessageByError(409_110, "این شماره موبایل مدیر قبلاً ثبت شده است")
	AdminInvalidInput    = messages.NewMessageByError(400_110, "اطلاعات مدیر نامعتبر است")
	AdminInvalidPassword = messages.NewMessageByError(401_110, "شماره موبایل یا رمز عبور مدیر نادرست است")

	SessionInvalid = messages.NewMessageByError(401_200, "نشست نامعتبر است؛ دوباره وارد شوید.")
)
