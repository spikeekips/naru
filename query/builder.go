package query

type Builder interface {
	Query() Query
	Build() (interface{}, error)
}
