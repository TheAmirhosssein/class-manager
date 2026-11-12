package message

import "github.com/TheAmirhosssein/class-manager/internal/utils/messages"

var (
	SchoolRead    = messages.NewMessage(200_201, "")
	SchoolCreated = messages.NewMessage(201_201, "مدرسه ایجاد شد")
	SchoolUpdated = messages.NewMessage(200_202, "نام مدرسه به‌روز شد")

	ClassRead    = messages.NewMessage(200_301, "")
	ClassCreated = messages.NewMessage(201_301, "کلاس ایجاد شد")
	ClassUpdated = messages.NewMessage(200_302, "کلاس به‌روز شد")
	ClassDeleted = messages.NewMessage(200_303, "کلاس حذف شد")

	StudentRead    = messages.NewMessage(200_401, "")
	StudentCreated = messages.NewMessage(201_401, "دانش‌آموز ثبت شد")
	StudentUpdated = messages.NewMessage(200_402, "دانش‌آموز به‌روز شد")
	StudentDeleted = messages.NewMessage(200_403, "دانش‌آموز حذف شد")

	SessionRead    = messages.NewMessage(200_501, "")
	SessionCreated = messages.NewMessage(201_501, "جلسه به برنامه اضافه شد")
	SessionUpdated = messages.NewMessage(200_502, "جلسه به‌روز شد")
	SessionDeleted = messages.NewMessage(200_503, "جلسه حذف شد")

	AttendanceRead  = messages.NewMessage(200_601, "")
	AttendanceSaved = messages.NewMessage(200_602, "")

	DisciplineRead    = messages.NewMessage(200_701, "")
	DisciplineCreated = messages.NewMessage(201_701, "یادداشت انضباطی ثبت شد")
	DisciplineUpdated = messages.NewMessage(200_702, "یادداشت به‌روز شد")
	DisciplineDeleted = messages.NewMessage(200_703, "یادداشت حذف شد")
)
