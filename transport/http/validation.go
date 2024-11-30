package http

type (
	validationErrors []string

	validatable interface {
		validate() validationErrors
	}
)

func (v *validationErrors) Add(e string) {
	*v = append(*v, e)
}
