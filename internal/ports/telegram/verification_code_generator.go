package tgports

type VerificationCodeGenerater interface {
	Generate() (string, error)
}
