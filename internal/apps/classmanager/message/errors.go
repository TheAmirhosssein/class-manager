package message

import "github.com/TheAmirhosssein/class-manager/internal/utils/messages"

var (
	Forbidden = messages.NewMessageByError(403_201, "اجازه دسترسی به این مدرسه را ندارید.")
	NoSchool  = messages.NewMessageByError(404_201, "مدرسه‌ای برای حساب شما ثبت نشده است. ابتدا یک مدرسه بسازید.")
	NotFound  = messages.NewMessageByError(404_202, "مدرسه یافت نشد.")
	InvalidInput = messages.NewMessageByError(422_201, "نام مدرسه نامعتبر است.")
	NameTaken    = messages.NewMessageByError(409_201, "مدرسه‌ای با این نام قبلاً ثبت کرده‌اید. نام دیگری انتخاب کنید.")

	ClassNotFound     = messages.NewMessageByError(404_301, "کلاس یافت نشد.")
	InvalidClassInput = messages.NewMessageByError(422_301, "اطلاعات کلاس نامعتبر است.")
)
