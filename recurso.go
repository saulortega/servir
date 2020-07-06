package servir

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Recurso interface {
	Obtener(r *http.Request, parámetros map[string]string) (Recurso, int, error)
	Listar(r *http.Request) ([]Recurso, int64, int, error)
	Crear(r *http.Request) (string, int, error)
	Editar(r *http.Request, parámetros map[string]string) (int, error)
	Eliminar(r *http.Request, parámetros map[string]string) (int, error)
}

type recurso struct {
	Recurso
	dirección *dirección
}

func (R *recurso) manejador(w http.ResponseWriter, r *http.Request) {
	// Listar u obtener recursos
	if r.Method == http.MethodGet {
		var coincide, tipo = R.dirección.coincide(r.URL.Path)
		if !coincide {
			// Esto no debería ocurrir
			log.Println("Manejando un recurso cuya dirección no coincide. [1]")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if tipo == "singular" {

			Obj, cod, err := R.Obtener(r, R.dirección.parámetros(r.URL.Path))
			if err != nil {
				// Pendiente notificación......
				w.WriteHeader(cod)
				return
			}

			ObjJSON, err := json.Marshal(Obj)
			if err != nil {
				log.Println("Error codificando a JSON ")
				//w.Header().Set("X-Notificaciones", notificación.Error("Ocurrió un error. Intente nuevamente.").Base64())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(ObjJSON)
			return

		} else if tipo == "plural" {

			Objs, Ttl, cod, err := R.Listar(r)
			if err != nil {
				// Pendiente notificación......
				w.WriteHeader(cod)
				return
			}

			if len(Objs) == 0 {
				w.WriteHeader(http.StatusNoContent)
				w.Write([]byte("[]"))
				return
			}

			ObjsJSON, err := json.Marshal(Objs)
			if err != nil {
				log.Println("Error codificando a JSON")
				//w.Header().Set("X-Notificaciones", notificación.Error("Ocurrió un error. Intente nuevamente.").Base64())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("X-Total", strconv.FormatInt(Ttl, 10))
			w.WriteHeader(http.StatusOK)
			w.Write(ObjsJSON)
			return

		} else {
			// Esto no debería ocurrir
			log.Println("Manejando un recurso cuya dirección no coincide. [2]")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}

	// Crear recurso
	if r.Method == http.MethodPost {
		iden, cod, err := R.Crear(r)
		if err != nil {
			// Pendiente notificación......
			w.WriteHeader(cod)
			return
		}

		w.Header().Set("X-Identificador", iden)
		w.WriteHeader(http.StatusCreated)
		return
	}

	// Editar recurso
	if r.Method == http.MethodPut {
		cod, err := R.Editar(r, R.dirección.parámetros(r.URL.Path))
		if err != nil {
			// Pendiente notificación......
			w.WriteHeader(cod)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	// Eliminar recurso
	if r.Method == http.MethodDelete {
		cod, err := R.Eliminar(r, R.dirección.parámetros(r.URL.Path))
		if err != nil {
			// Pendiente notificación......
			w.WriteHeader(cod)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusNotImplemented)
}
