package domain

import "context"

// Repository is the persistence boundary. The initial version is stateless;
// implement this interface when adding PostgreSQL or another datastore.
type Repository interface {
	SaveMany(context.Context, []Domain) error
	ListByServer(context.Context, string) ([]Domain, error)
}
