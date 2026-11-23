// Package content empaqueta los datos del juego (definiciones de objetos,
// etc.) para que los binarios no dependan del directorio de trabajo. La
// carga y validación están en internal/content.
package content

import "embed"

// FS contiene objects/*.json y names.txt.
//
//go:embed objects/*.json names.txt
var FS embed.FS
