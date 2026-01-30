package engine

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate() string {
	return "Quibit CLI initialized. Project generation coming soon."
}
