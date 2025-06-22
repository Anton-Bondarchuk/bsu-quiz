package ports

type VerificationCodeGenerater interface {
	Generate() (string, error)
}
