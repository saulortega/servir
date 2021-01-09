package servir

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Recurso interface {
	Obtener(r *http.Request, parámetros map[string]string) (Recurso, int, error)
	Listar(r *http.Request, parámetros map[string]string) ([]Recurso, int64, int, error)
	Crear(r *http.Request, parámetros map[string]string) (string, int, error)
	Editar(r *http.Request, parámetros map[string]string) (int, error)
	Eliminar(r *http.Request, parámetros map[string]string) (int, error)
}

type recurso struct {
	Recurso
	dirección *dirección
}

func (R *recurso) manejador(w http.ResponseWriter, r *http.Request) {
	var coincide, tipo = R.dirección.coincide(r.URL.Path)
	if !coincide {
		// Esto no debería ocurrir
		log.Println("Manejando un recurso cuya dirección no coincide. [1]")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var parámetros = R.dirección.parámetros(r.URL.Path)

	// Listar u obtener recursos
	if r.Method == http.MethodGet {

		if tipo == "singular" {

			Obj, cod, err := R.Obtener(r, parámetros)
			if err != nil {
				w.Header().Set("X-Error", err.Error())
				if cod >= 200 && cod < 300 {
					cod = 500
				}

				w.WriteHeader(cod)
				return
			}

			ObjJSON, err := json.Marshal(Obj)
			if err != nil {
				log.Println("[5925] Error codificando a JSON:", err)
				w.Header().Set("X-Error", "[5925] Ocurrió un error.")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if cod <= 0 {
				cod = http.StatusOK
			}

			w.WriteHeader(cod)
			w.Write(ObjJSON)
			return

		} else if tipo == "plural" {

			Objs, Ttl, cod, err := R.Listar(r, parámetros)
			if err != nil {
				w.Header().Set("X-Error", err.Error())
				if cod >= 200 && cod < 300 {
					cod = 500
				}

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
				log.Println("[5926] Error codificando a JSON:", err)
				w.Header().Set("X-Error", "[5926] Ocurrió un error.")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if cod <= 0 {
				cod = http.StatusOK
			}

			w.Header().Set("X-Total", strconv.FormatInt(Ttl, 10))
			w.WriteHeader(cod)
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
		if tipo != "plural" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		iden, cod, err := R.Crear(r, parámetros)
		if err != nil {
			w.Header().Set("X-Error", err.Error())
			if cod >= 200 && cod < 300 {
				cod = 500
			}
		}

		if cod <= 0 {
			cod = http.StatusCreated
		}

		w.Header().Set("X-Identificador", iden)
		w.WriteHeader(cod)
		return
	}

	// Editar recurso
	if r.Method == http.MethodPut {
		if tipo != "singular" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		cod, err := R.Editar(r, parámetros)
		if err != nil {
			w.Header().Set("X-Error", err.Error())
			if cod >= 200 && cod < 300 {
				cod = 500
			}
		}

		if cod <= 0 {
			cod = http.StatusOK
		}

		w.WriteHeader(cod)
		return
	}

	// Eliminar recurso
	if r.Method == http.MethodDelete {
		if tipo != "singular" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		cod, err := R.Eliminar(r, parámetros)
		if err != nil {
			w.Header().Set("X-Error", err.Error())
			if cod >= 200 && cod < 300 {
				cod = 500
			}
		}

		if cod <= 0 {
			cod = http.StatusOK
		}

		w.WriteHeader(cod)
		return
	}

	w.WriteHeader(http.StatusNotImplemented)
}
