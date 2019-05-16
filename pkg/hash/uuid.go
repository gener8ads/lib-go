package hash

import "github.com/google/uuid"

// UUID from namespace + id
func UUID(source string, salt string) string {
	namespace, _ := uuid.FromBytes([]byte(source))
	id := uuid.NewSHA1(namespace, []byte(salt))

	return id.String()
}
