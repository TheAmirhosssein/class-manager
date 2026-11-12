package message

import "github.com/TheAmirhosssein/class-manager/internal/utils/messages"

var (
	UserCreated = messages.NewMessage(201_101, "کاربر ایجاد شد")
	UserRead    = messages.NewMessage(200_101, "")
	AdminCreated = messages.NewMessage(201_110, "مدیر ایجاد شد")
	AdminRead    = messages.NewMessage(200_110, "")
	LoginSuccess = messages.NewMessage(200_120, "ورود موفق")
)
