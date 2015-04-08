package arch

import (
	"fmt"
)

//Generator type represent a function that returns a Linkage
type Generator func(*LinkDescriptor) (Linkage, error)

//Factory provides a base of map factory of generating predefined
//structs based on data passed in
type Factory struct {
	generators map[string]Generator
}

//NewFactory creates a new factory object
func NewFactory() *Factory {
	return &Factory{make(map[string]Generator)}
}

//Resolve takes a map[string]interface{} and based on the proto id
//it generates a corresponding arch.Linkage
func (f *Factory) Resolve(m *LinkDescriptor) (Linkage, error) {
	gen, ok := f.generators[m.Proto]

	if !ok {
		return nil, fmt.Errorf("%s not found %v", m.Proto, m)
	}

	return gen(m)
}

//Provide adds a corresponding function generator into the factory
func (f *Factory) Provide(key string, gen Generator) bool {
	_, ok := f.generators[key]

	if !ok {
		f.generators[key] = gen
		return true
	}

	return false
}

//Has checks if a generator with the specific key exists
func (f *Factory) Has(key string) bool {
	_, ok := f.generators[key]
	return ok
}

//Remove deletes a generator with the specific key
func (f *Factory) Remove(key string) {
	delete(f.generators, key)
}

//ProvideOverwrite overwrites a gen key with another generation function
func (f *Factory) ProvideOverwrite(key string, gen Generator) {
	f.generators[key] = gen
}
