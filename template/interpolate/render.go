package interpolate

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/reflectwalk"
)

// RenderFilter is an option for filtering what gets rendered and
// doesn't within an interface.
type RenderFilter struct {
	Include []string
	Exclude []string

	once       sync.Once
	excludeSet map[string]struct{}
	includeSet map[string]struct{}
}

// RenderMap renders all the strings in the given interface. The
// interface must decode into a map[string]interface{}, but is left
// as an interface{} type to ease backwards compatibility with the way
// arguments are passed around in Packer.
func RenderMap(v interface{}, ctx *Context, f *RenderFilter) (map[string]interface{}, error) {
	// First decode it into the map
	var m map[string]interface{}
	if err := mapstructure.Decode(v, &m); err != nil {
		return nil, err
	}

	// Now go through each value and render it
	for k, raw := range m {
		// Always validate every field
		if err := ValidateInterface(raw, ctx); err != nil {
			return nil, fmt.Errorf("invalid '%s': %s", k, err)
		}

		if !f.include(k) {
			continue
		}

		raw, err := RenderInterface(raw, ctx)
		if err != nil {
			return nil, fmt.Errorf("render '%s': %s", k, err)
		}

		m[k] = raw
	}

	return m, nil
}

// RenderInterface renders any value and returns the resulting value.
func RenderInterface(v interface{}, ctx *Context) (interface{}, error) {
	f := func(v string) (string, error) {
		return Render(v, ctx)
	}

	walker := &renderWalker{
		F:       f,
		Replace: true,
	}
	err := reflectwalk.Walk(v, walker)
	if err != nil {
		return nil, err
	}

	if walker.Top != nil {
		v = walker.Top
	}
	return v, nil
}

// ValidateInterface renders any value and returns the resulting value.
func ValidateInterface(v interface{}, ctx *Context) error {
	f := func(v string) (string, error) {
		return v, Validate(v, ctx)
	}

	walker := &renderWalker{
		F:       f,
		Replace: false,
	}
	err := reflectwalk.Walk(v, walker)
	if err != nil {
		return err
	}

	return nil
}

// Include checks whether a key should be included.
func (f *RenderFilter) include(k string) bool {
	if f == nil {
		return true
	}

	k = strings.ToLower(k)

	f.once.Do(f.init)
	if len(f.includeSet) > 0 {
		_, ok := f.includeSet[k]
		return ok
	}
	if len(f.excludeSet) > 0 {
		_, ok := f.excludeSet[k]
		return !ok
	}

	return true
}

func (f *RenderFilter) init() {
	f.includeSet = make(map[string]struct{})
	for _, v := range f.Include {
		f.includeSet[strings.ToLower(v)] = struct{}{}
	}

	f.excludeSet = make(map[string]struct{})
	for _, v := range f.Exclude {
		f.excludeSet[strings.ToLower(v)] = struct{}{}
	}
}

// renderWalker implements interfaces for the reflectwalk package
// (github.com/mitchellh/reflectwalk) that can be used to automatically
// execute a callback for an interpolation.
type renderWalker struct {
	// F is the function to call for every interpolation. It can be nil.
	//
	// If Replace is true, then the return value of F will be used to
	// replace the interpolation.
	F       renderWalkerFunc
	Replace bool

	// ContextF is an advanced version of F that also receives the
	// location of where it is in the structure. This lets you do
	// context-aware validation.
	ContextF renderWalkerContextFunc

	// Top is the top value of the walk. This might get replaced if the
	// top value needs to be modified. It is valid to read after any walk.
	// If it is nil, it means the top wasn't replaced.
	Top interface{}

	key         []string
	lastValue   reflect.Value
	loc         reflectwalk.Location
	cs          []reflect.Value
	csKey       []reflect.Value
	csData      interface{}
	sliceIndex  int
	unknownKeys []string
}

// renderWalkerFunc is the callback called by interpolationWalk.
// It is called with any interpolation found. It should return a value
// to replace the interpolation with, along with any errors.
//
// If Replace is set to false in renderWalker, then the replace
// value can be anything as it will have no effect.
type renderWalkerFunc func(string) (string, error)

// renderWalkerContextFunc is called by interpolationWalk if
// ContextF is set. This receives both the interpolation and the location
// where the interpolation is.
//
// This callback can be used to validate the location of the interpolation
// within the configuration.
type renderWalkerContextFunc func(reflectwalk.Location, string)

func (w *renderWalker) Enter(loc reflectwalk.Location) error {
	w.loc = loc
	return nil
}

func (w *renderWalker) Exit(loc reflectwalk.Location) error {
	w.loc = reflectwalk.None

	switch loc {
	case reflectwalk.Map:
		w.cs = w.cs[:len(w.cs)-1]
	case reflectwalk.MapValue:
		w.key = w.key[:len(w.key)-1]
		w.csKey = w.csKey[:len(w.csKey)-1]
	case reflectwalk.Slice:
		// Split any values that need to be split
		w.cs = w.cs[:len(w.cs)-1]
	case reflectwalk.SliceElem:
		w.csKey = w.csKey[:len(w.csKey)-1]
	}

	return nil
}

func (w *renderWalker) Map(m reflect.Value) error {
	w.cs = append(w.cs, m)
	return nil
}

func (w *renderWalker) MapElem(m, k, v reflect.Value) error {
	w.csData = k
	w.csKey = append(w.csKey, k)
	w.key = append(w.key, k.String())
	w.lastValue = v
	return nil
}

func (w *renderWalker) Slice(s reflect.Value) error {
	w.cs = append(w.cs, s)
	return nil
}

func (w *renderWalker) SliceElem(i int, elem reflect.Value) error {
	w.csKey = append(w.csKey, reflect.ValueOf(i))
	w.sliceIndex = i
	return nil
}

func (w *renderWalker) Primitive(v reflect.Value) error {
	setV := v

	// We only care about strings
	if v.Kind() == reflect.Interface {
		setV = v
		v = v.Elem()
	}
	if v.Kind() != reflect.String {
		return nil
	}

	strV := v.String()
	if w.ContextF != nil {
		w.ContextF(w.loc, strV)
	}

	if w.F == nil {
		return nil
	}

	replaceVal, err := w.F(strV)
	if err != nil {
		return fmt.Errorf(
			"%s in:\n\n%s",
			err, v.String())
	}

	if w.Replace {
		resultVal := reflect.ValueOf(replaceVal)
		switch w.loc {
		case reflectwalk.MapKey:
			m := w.cs[len(w.cs)-1]

			// Delete the old value
			var zero reflect.Value
			m.SetMapIndex(w.csData.(reflect.Value), zero)

			// Set the new key with the existing value
			m.SetMapIndex(resultVal, w.lastValue)

			// Set the key to be the new key
			w.csData = resultVal
		case reflectwalk.MapValue:
			// If we're in a map, then the only way to set a map value is
			// to set it directly.
			m := w.cs[len(w.cs)-1]
			mk := w.csData.(reflect.Value)
			m.SetMapIndex(mk, resultVal)
		case reflectwalk.WalkLoc:
			// At the root element, we can't write that, so we just save it
			w.Top = resultVal.Interface()
		default:
			// Otherwise, we should be addressable
			setV.Set(resultVal)
		}
	}

	return nil
}

func (w *renderWalker) removeCurrent() {
	// Append the key to the unknown keys
	w.unknownKeys = append(w.unknownKeys, strings.Join(w.key, "."))

	for i := 1; i <= len(w.cs); i++ {
		c := w.cs[len(w.cs)-i]
		switch c.Kind() {
		case reflect.Map:
			// Zero value so that we delete the map key
			var val reflect.Value

			// Get the key and delete it
			k := w.csData.(reflect.Value)
			c.SetMapIndex(k, val)
			return
		}
	}

	panic("No container found for removeCurrent")
}

func (w *renderWalker) replaceCurrent(v reflect.Value) {
	c := w.cs[len(w.cs)-2]
	switch c.Kind() {
	case reflect.Map:
		// Get the key and delete it
		k := w.csKey[len(w.csKey)-1]
		c.SetMapIndex(k, v)
	}
}
