package servir

import (
	"net/http"
)

type Enrutador struct {
	Interceptor func(http.ResponseWriter, *http.Request, func(*http.Request))
	rutas       []*ruta
}

func NuevoEnrutador() *Enrutador {
	return &Enrutador{}
}

func (E *Enrutador) GET(dirección string, manejador func(http.ResponseWriter, *http.Request, map[string]string)) {
	ruta := nuevaRuta("GET", dirección, manejador)
	E.rutas = append(E.rutas, ruta)
}

func (E *Enrutador) PUT(dirección string, manejador func(http.ResponseWriter, *http.Request, map[string]string)) {
	ruta := nuevaRuta("PUT", dirección, manejador)
	E.rutas = append(E.rutas, ruta)
}

func (E *Enrutador) POST(dirección string, manejador func(http.ResponseWriter, *http.Request, map[string]string)) {
	ruta := nuevaRuta("POST", dirección, manejador)
	E.rutas = append(E.rutas, ruta)
}

func (E *Enrutador) DELETE(dirección string, manejador func(http.ResponseWriter, *http.Request, map[string]string)) {
	ruta := nuevaRuta("DELETE", dirección, manejador)
	E.rutas = append(E.rutas, ruta)
}

func (E *Enrutador) Receptor(w http.ResponseWriter, r *http.Request) {
	var coincidenciaDeRuta bool

	for _, ruta := range E.rutas {
		coincide, cncdeRuta, parámetros := ruta.comprobar(r.Method, r.URL.Path)
		if coincide {
			if E.Interceptor != nil {
				E.Interceptor(w, r, func(nr *http.Request) {
					ruta.manejador(w, nr, parámetros)
				})
			} else {
				ruta.manejador(w, r, parámetros)
			}
			return
		} else if cncdeRuta {
			coincidenciaDeRuta = true
		}
	}

	if coincidenciaDeRuta {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
