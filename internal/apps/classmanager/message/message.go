package message

import "github.com/TheAmirhosssein/class-manager/internal/utils/messages"

var (
	SchoolRead    = messages.NewMessage(200_201, "")
	SchoolCreated = messages.NewMessage(201_201, "مدرسه ایجاد شد")
	SchoolUpdated = messages.NewMessage(200_202, "نام مدرسه به‌روز شد")
)
