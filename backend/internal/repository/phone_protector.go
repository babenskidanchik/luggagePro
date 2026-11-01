package repository

type PhoneProtector interface {
	Encrypt(phone string) ([]byte, error)
	Decrypt(encrypted []byte) (string, error)
	LookupHash(phone string) []byte
}
