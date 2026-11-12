package message

import "github.com/TheAmirhosssein/class-manager/internal/utils/messages"

var (
	UserNotFound         = messages.NewMessageByError(404_101, "کاربر یافت نشد")
	UserEmailExists      = messages.NewMessageByError(409_101, "این ایمیل قبلاً ثبت شده است")
	UserInvalidInput     = messages.NewMessageByError(400_101, "اطلاعات کاربر نامعتبر است")
	UserInvalidPassword  = messages.NewMessageByError(401_101, "ایمیل یا رمز عبور نادرست است")

	AdminNotFound        = messages.NewMessageByError(404_110, "مدیر یافت نشد")
	AdminEmailExists     = messages.NewMessageByError(409_110, "این ایمیل مدیر قبلاً ثبت شده است")
	AdminInvalidInput    = messages.NewMessageByError(400_110, "اطلاعات مدیر نامعتبر است")
	AdminInvalidPassword = messages.NewMessageByError(401_110, "ایمیل یا رمز عبور مدیر نادرست است")

	SessionInvalid = messages.NewMessageByError(401_200, "نشست نامعتبر است؛ دوباره وارد شوید.")
)
