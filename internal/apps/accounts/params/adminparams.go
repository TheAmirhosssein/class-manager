package params

type CreateAdminParams struct {
	Email    string
	Password string
	Name     string
}

type LoginAdminParams struct {
	Email    string
	Password string
}
