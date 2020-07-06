package servir

import "net/http"

// Autenticador es la función que debe verificar las credenciales de acceso.
// El primer parámetro retornado debe ser un booleano indicando si la solicitud tiene permisos de acceso al recurso.
// El segundo parámetro retornado debe ser un texto opcional que se enviará en el encabezado X-Notificaciones en caso de que la autenticación falle.
// El tercer parámetro retornado debe ser el código de estado HTTP que se retornará al cliente. Si es cero se enviará http.StatusUnauthorized.
type Autenticador func(*http.Request) (bool, string, int)
