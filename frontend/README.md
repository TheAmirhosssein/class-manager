# کلاسینو — پنل مدیریت

رابط فارسی و راست‌چین. فرانت فقط از API بک‌اند داده می‌گیرد (مدرسه فعال است؛ سایر ماژول‌ها به‌زودی).

## اجرا

از ریشه مخزن:

```sh
go run ./cmd/main.go serve
```

→ `http://127.0.0.1:8080` (همان origin برای UI و `/api/v1`)

برای توسعه با پروکسی جدا: `make dev` یا `scripts/dev.sh`. برای اجرای نزدیک production (یک باینری، UI از Go): `make prod`.

## PWA

اپلیکیشن وب پیش‌رونده (نصب روی موبایل/دسکتاپ): `manifest.webmanifest` و `sw.js` در ریشهٔ `frontend/`؛ ثبت در `assets/pwa-register.js`. API همیشه از شبکه می‌رود؛ فقط پوسته و فایل‌های استاتیک کش می‌شوند. پس از تغییر `sw.js`، ثابت `CACHE` داخل آن را افزایش دهید.

## فایل‌ها

```text
index.html           پوسته پنل
assets/api.js        مرز API — تنها جایی که fetch انجام می‌شود
assets/app.js        تعاملات UI
assets/config.js     apiBase و CSRF (بدون راز)
docs/API.md          قرارداد API
```

## اتصال

`assets/config.js`: `apiBase` معمولاً `/api/v1` و `credentials: same-origin`. ورود با `POST /auth/login`؛ مدرسه با مسیرهای classmanager در `docs/API.md`.

بعد از `seed`: مدرس `09001110011` / `strong-password`.
