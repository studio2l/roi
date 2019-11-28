package roi

type NotFound struct {
	kind string
	id   string
}

func (e NotFound) Error() string {
	return e.kind + " not found: " + e.id
}
