package secure_datastore

type SecureKVDataStore interface {
	Get(string) (map[string]interface{}, error)
	Put(string, map[string]interface{}) (map[string]interface{}, error)
}

type passwordProtectedSecureKVDataStore struct {
}
