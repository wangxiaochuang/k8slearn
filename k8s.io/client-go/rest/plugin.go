package rest

type AuthProviderConfigPersister interface {
	Persist(map[string]string) error
}
