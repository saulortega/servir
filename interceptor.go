package servir

import "net/http"

// Interceptor puede ser usada para escribir encabezados personalizados o realizar comprobaciones globales generales.
// Si el error retornado no es nil, se finaliza la función sin escribir nada en ResponseWriter.
// Deberá escribir en ResponseWriter si va a retornar un error no nil.
type Interceptor func(w http.ResponseWriter, r *http.Request) error
