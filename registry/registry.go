package registry

// To maintain the record of nodes
type Registry interface {

	Register() error
	FetchHealthyNode() (string, error)
}