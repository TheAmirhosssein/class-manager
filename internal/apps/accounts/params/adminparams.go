package params

type CreateAdminParams struct {
	Mobile   string
	Password string
	Name     string
}

type LoginAdminParams struct {
	Mobile   string
	Password string
}
