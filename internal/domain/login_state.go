package domain

type LoginState struct {
	Username     string
	VerifyToken  []byte
	SharedSecret []byte
}
