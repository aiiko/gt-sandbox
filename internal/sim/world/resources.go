package world

import "reflect"

// Los recursos son datos asociados al mundo que NO son estado autoritativo:
// índices derivados, cachés, espacios de trabajo de algoritmos. No entran en
// Hash() ni en snapshots, así que deben poder reconstruirse a partir del
// estado (y reconstruirse dentro de Step antes de usarse).

// SetResource asocia v al mundo, uno por tipo.
func SetResource[T any](w *World, v *T) {
	if w.resources == nil {
		w.resources = map[reflect.Type]any{}
	}
	w.resources[reflect.TypeFor[T]()] = v
}

// Resource devuelve el recurso de tipo T, o nil.
func Resource[T any](w *World) *T {
	v, _ := w.resources[reflect.TypeFor[T]()].(*T)
	return v
}
