package servir

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/saulortega/auten"
	//"github.com/saulortega/utilidades/notificación"
)

type Manejador struct {
	BDU                *sql.DB
	Recursos           map[string]Recurso
	ExpiraciónDeSesión time.Duration
	Autenticación      bool
	Registro           bool
}

type Recurso interface {
	dirección() string
	Obtener(r *http.Request, i string) (Recurso, int, error)
	Listar(r *http.Request) ([]Recurso, int64, int, error)
	Crear(r *http.Request) (string, int, error)
	Editar(r *http.Request, i string) (int, error)
	Eliminar(r *http.Request, i string) (int, error)
}

func manejador() *Manejador {
	var m = new(Manejador)
	m.Recursos = make(map[string]Recurso)
	m.ExpiraciónDeSesión = 4 * time.Hour
	m.Autenticación = true
	m.Registro = true
	return m
}

func NuevoManejador() *Manejador {
	return manejador()
}

func (M *Manejador) Receptor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, X-Requested-With, Content-Type, Accept")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE")
	w.Header().Add("Access-Control-Expose-Headers", "X-Llave")
	w.Header().Add("Access-Control-Expose-Headers", "X-Identificador")
	w.Header().Add("Access-Control-Expose-Headers", "X-Total")
	w.Header().Add("Access-Control-Expose-Headers", "X-Notificaciones")
	w.Header().Add("Access-Control-Expose-Headers", "X-Token")

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

	// Comprobación de ingreso
	if r.URL.Path == "/ingreso" && M.Autenticación {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var D, err = auten.Ingreso(r, M.BDU, M.ExpiraciónDeSesión)
		if err != nil {
			//Pendiente responder con notificaciones...
			//w.Header().Set("X-Notificaciones", notificación.Error(err.Error()).Base64())
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		//Pendiente responder datos de usuario...
		w.Header().Set("X-Token", D.Token)
		w.WriteHeader(http.StatusOK)

		return
	}

	// Comprobación de registro
	/*
		if r.URL.Path == "/registro" && M.Autenticación && M.Registro {
			if r.Method == "POST" {
				w.WriteHeader(http.StatusBadRequest)
				// ... pendiente ---------------------------------
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}
	*/

	//Comprobar token
	if M.Autenticación {
		_, err := auten.Sesión(r, M.BDU)
		if err != nil {
			//Pendiente responder con notificaciones...
			//w.Header().Set("X-Notificaciones", notificación.Error(err.Error()).Base64())
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	// Listar u obtener recursos
	if r.Method == http.MethodGet {
		R, en := M.Recursos[r.URL.Path]

		// Listar
		if en {
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
		}

		R, I := M.buscarRecursoConIdentificador(r)
		if R == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Obtener
		Obj, cod, err := R.Obtener(r, I)
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
	}

	// Crear recurso
	if r.Method == http.MethodPost {
		R, en := M.Recursos[r.URL.Path]
		if !en {
			w.WriteHeader(http.StatusNotFound)
			return
		}

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
		R, I := M.buscarRecursoConIdentificador(r)
		if R == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		cod, err := R.Editar(r, I)
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
		R, I := M.buscarRecursoConIdentificador(r)
		if R == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		cod, err := R.Eliminar(r, I)
		if err != nil {
			// Pendiente notificación......
			w.WriteHeader(cod)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	// Esto no debería ocurrir
	w.WriteHeader(http.StatusNotImplemented)
}

func (M *Manejador) Recurso(R Recurso) {
	if len(R.dirección()) <= 1 {
		return
	}

	M.Recursos[R.dirección()] = R
}

func (M *Manejador) buscarRecursoConIdentificador(r *http.Request) (Recurso, string) {
	for dir, R := range M.Recursos {
		if regexp.MustCompile(fmt.Sprintf("^%s/(.+)", dir)).MatchString(r.URL.Path) {
			return R, strings.Split(r.URL.Path, dir+"/")[1]
		}
	}

	return nil, ""
}
