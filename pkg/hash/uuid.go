package hash

import "github.com/google/uuid"

// UUID from namespace + id
func UUID(source string, salt string) (string, error) {
	namespace, err := uuid.Parse(source)
	if err != nil {
		return "", err
	}
	id := uuid.NewSHA1(namespace, []byte(salt))

	return id.String(), nil
}
