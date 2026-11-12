# قرارداد API — classmanager (مدرسه)

Base: `/api/v1`. JSON، UTF-8. شناسه‌ها در JSON به صورت رشته‌اند.

## احراز هویت

همهٔ مسیرهای زیر نیاز به **نشست مدرس** (`role: teacher`) دارند: کوکی `access_token` یا هدر `Authorization: Bearer …`.

درخواست‌های `POST` و `PATCH` علاوه بر نشست، کوکی `csrftoken` و هدر `X-CSRFToken` با همان مقدار را می‌خواهند (پس از یک `GET` روی API کوکی تنظیم می‌شود).

## مسیرهای مدرسه (ماژول classmanager)

| متد | مسیر | پاسخ |
| --- | --- | --- |
| GET | `/schools` | `School[]` — فقط مدرسه‌های ساخته‌شده توسط کاربر جاری |
| GET | `/school?schoolId=` اختیاری | `{ "id", "name" }` — بدون `schoolId` اولین مدرسهٔ کاربر؛ بدون مدرسه ۴۰۴ |
| POST | `/school` | `{ "id", "name" }` — ایجاد مدرسه جدید (مالک = کاربر جاری) |
| PATCH | `/school?schoolId=` یا `"id"` در body | `{ "id", "name" }` — فقط مدرسهٔ متعلق به کاربر |

### School

```json
{"id":"1","name":"مدرسه نمونه"}
```

POST / PATCH بدنه:

```json
{"name":"نام مدرسه"}
```

PATCH می‌تواند `"id"` داشته باشد وقتی چند مدرسه دارید.

نام پس از trim باید ۲ تا ۸۰ کاراکتر باشد. برای هر حساب مدرس، نام مدرسه یکتا است (دو مدرسه با نام یکسان برای یک کاربر مجاز نیست).

### خطاها

```json
{"message":"اجازه دسترسی به این مدرسه را ندارید."}
```

۴۰۱ نشست، ۴۰۳ مدرسهٔ دیگران، ۴۰۴ مدرسهٔ خود کاربر وجود ندارد، ۴۲۲ نام نامعتبر.

## حساب کاربری (`accounts`)

| متد | مسیر | دسترسی |
| --- | --- | --- |
| POST | `/auth/login` | عمومی |
| GET | `/auth/session` | نشست |
| POST | `/auth/logout` | نشست |

ورود: `{"mobile":"09…","password":"…"}` → `{"user":{…},"token":…}`.

## مسیرهای کلاس (ماژول classmanager)

فقط برای مدرسه‌ای که **مالک آن** کاربر جاری است. در UI صفحهٔ کلاس‌ها، `schoolId` با dropdown همان صفحه به query وصل می‌شود (ذخیرهٔ جداگانه لازم نیست). بدون `schoolId` در API، اولین مدرسهٔ کاربر استفاده می‌شود.

| متد | مسیر | پاسخ |
| --- | --- | --- |
| GET | `/classes?schoolId=` اختیاری | `Class[]` |
| GET | `/class?classId=` | یک `Class` |
| POST | `/class?schoolId=` اختیاری | `Class` — ایجاد در مدرسهٔ متعلق به کاربر |
| PATCH | `/class?classId=` یا `"id"` در body | `Class` |
| DELETE | `/class?classId=` | `{ "message": "کلاس حذف شد" }` |

### Class

```json
{"id":"1","schoolId":"2","title":"ریاضی ۳","grade":"12","room":"کلاس ۱۰۲","studentCount":0}
```

POST بدنه:

```json
{"title":"ریاضی ۳","grade":"12","room":"کلاس ۱۰۲"}
```

`grade` یکی از `7`، `8`، `9`، `10`، `11`، `12`. `title` بعد از trim بین ۲ تا ۱۰۰ کاراکتر؛ `room` حداکثر ۸۰ کاراکتر (می‌تواند خالی باشد).

### خطاهای کلاس

۴۰۳ مدرسهٔ دیگران، ۴۰۴ کلاس یا مدرسهٔ کاربر یافت نشد، ۴۲۲ ورودی نامعتبر.

## مسیرهای دانش‌آموز (ماژول classmanager)

هر دانش‌آموز دقیقاً به **یک کلاس** تعلق دارد. فقط کلاس‌های مدرسه‌ای که کاربر جاری مالک آن است.

| متد | مسیر | پاسخ |
| --- | --- | --- |
| GET | `/students?classId=` | `Student[]` |
| GET | `/student?studentId=` | یک `Student` |
| POST | `/student?classId=` | `Student` |
| PATCH | `/student?studentId=` یا `"id"` در body | `Student` |
| DELETE | `/student?studentId=` | `{ "message": "دانش‌آموز حذف شد" }` |

### Student

```json
{
  "id": "1",
  "classId": "3",
  "firstName": "علی",
  "lastName": "رضایی",
  "name": "علی رضایی",
  "nationalCode": "1234567891",
  "fatherName": "محمد",
  "studentNumber": "14031234"
}
```

POST بدنه:

```json
{
  "firstName": "علی",
  "lastName": "رضایی",
  "nationalCode": "1234567891",
  "fatherName": "محمد",
  "studentNumber": "14031234"
}
```

`firstName` و `lastName` بعد از trim بین ۲ تا ۶۰ کاراکتر؛ `nationalCode` **اختیاری** — اگر مقدار دارد باید ده رقم معتبر (الگوریتم کد ملی ایران) باشد؛ `fatherName` و `studentNumber` اختیاری (حداکثر ۶۰ و ۲۰ کاراکتر). کدهای ملی غیرخالی در کل سیستم یکتا هستند.

### خطاهای دانش‌آموز

۴۰۳ کلاس متعلق به دیگران، ۴۰۴ دانش‌آموز یا کلاس یافت نشد، ۴۰۹ کد ملی تکراری، ۴۲۲ ورودی نامعتبر.

## مسیرهای جلسه / برنامه هفتگی (sessions)

هر جلسه به **یک کلاس** و **یک تاریخ** (تقویم میلادی `YYYY-MM-DD` در منطقهٔ Asia/Tehran) وصل است. روز هفته (۰=شنبه … ۶=جمعه) باید با تاریخ در تهران هم‌خوان باشد.

| متد | مسیر | پاسخ |
| --- | --- | --- |
| GET | `/sessions?schoolId=` یا `classId=` | `Session[]` — یکی از پارامترها لازم است |
| GET | `/session?sessionId=` | یک `Session` |
| POST | `/session?classId=` | `Session` |
| PATCH | `/session?sessionId=` یا `"id"` در body | `Session` |
| DELETE | `/session?sessionId=` | `{ "message": "جلسه حذف شد" }` |

### Session

```json
{
  "id": "1",
  "classId": "3",
  "sessionDate": "2025-11-12",
  "day": 3,
  "start": "08:00",
  "end": "09:30",
  "subject": "ریاضی"
}
```

`subject` اختیاری (حداکثر ۱۰۰ کاراکتر). ساعات بدون offset و با منطقهٔ زمانی مدرسه (این نمونه: Asia/Tehran)؛ تاریخ در UI با تقویم شمسی نمایش داده می‌شود. ترکیب `(classId, sessionDate, start)` یکتا است؛ تداخل بازهٔ زمانی در همان تاریخ برای مدرسه ممنوع است.

### حضور و غیاب (بر اساس جلسهٔ برنامه)

| متد | مسیر | توضیح |
|-----|------|--------|
| GET | `/attendance?sessionId=` | `{ sessionId, classId, records: [{ studentId, status }] }` |
| PATCH | `/attendance?sessionId=&studentId=` | body: `{ "status": "present" \| "absent" \| "late" \| "excused" }` — ذخیرهٔ فوری یک دانش‌آموز |

`sessionId` همان شناسهٔ جلسه در جدول `sessions` (برنامه هفتگی) است.

سایر ماژول‌ها (انضباط، …) هنوز روی این سرور ثبت نشده‌اند.
