package servir

import (
	"log"
	"net/http"
	//"github.com/saulortega/utilidades/notificación"
)

type Manejador struct {
	recursos     []recurso
	Autenticador Autenticador
	Interceptor  Interceptor
}

func manejador() *Manejador {
	var m = new(Manejador)
	m.recursos = []recurso{}

	m.Interceptor = func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}

	m.Autenticador = func(*http.Request) (bool, string, int) {
		return false, "No autorizado.", http.StatusUnauthorized
	}

	return m
}

func NuevoManejador() *Manejador {
	return manejador()
}

func (M *Manejador) Receptor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, X-Requested-With, Content-Type, Accept")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE")
	w.Header().Add("Access-Control-Expose-Headers", "X-Identificador")
	w.Header().Add("Access-Control-Expose-Headers", "X-Total")
	w.Header().Add("Access-Control-Expose-Headers", "X-Notificaciones")

	var err = M.Interceptor(w, r)
	if err != nil {
		return
	}

	// Responder Options
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GET, POST, PUT, DELETE"))
		return
	}

	// Comprobar método
	if r.Method != http.MethodGet && r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	var aut, msj, cod = M.Autenticador(r)
	if !aut {
		if cod <= 0 {
			cod = http.StatusUnauthorized
		}

		if msj != "" {
			w.Header().Set("X-Notificaciones", msj)
		}

		w.WriteHeader(cod)
		return
	}

	err = r.ParseMultipartForm(32 << 20)
	if err == http.ErrNotMultipart {
		err = r.ParseForm()
	}
	if err != nil {
		// Pendiente agregar respuesta en encabezado...
		log.Println("[57320]", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var recursoRegistrado bool
	for _, recurso := range M.recursos {
		var coincide, _ = recurso.dirección.coincide(r.URL.Path)
		if coincide {
			recursoRegistrado = true
			recurso.manejador(w, r)
			break
		}
	}

	if !recursoRegistrado {
		w.WriteHeader(http.StatusNotFound)
		return
	}

}

// Recurso agrega un recurso al manejador
func (M *Manejador) Recurso(url string, R Recurso) {
	var r = recurso{}
	r.dirección = new(dirección)
	r.dirección.inicializar(url)
	r.Recurso = R

	M.recursos = append(M.recursos, r)
}
