package password

import "golang.org/x/crypto/bcrypt"

func Hash(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
}

func HashString(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func HashAssign(password []byte, param *string) error {
	pwd, err := Hash(password)
	if err != nil {
		return err
	}
	*param = string(pwd)
	return nil
}

func HashAssignString(password string, param *string) error {
	return HashAssign([]byte(password), param)
}

func IsVerified(hashedPassword, password []byte) bool {
	return bcrypt.CompareHashAndPassword(hashedPassword, password) == nil
}

func IsVerifiedString(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
