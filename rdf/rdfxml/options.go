package rdfxml

type config struct {
	base              string
	inferNumericTypes bool
}

type Option func(*config)

func WithBase(base string) Option {
	return func(c *config) { c.base = base }
}

func WithNumericInference(infer bool) Option {
	return func(c *config) { c.inferNumericTypes = infer }
}
