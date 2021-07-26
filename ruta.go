package servir

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type ruta struct {
	método    string
	dirección string
	dirExpReg *regexp.Regexp
	variables []string
	manejador func(http.ResponseWriter, *http.Request, map[string]string)
}

func nuevaRuta(método string, dirección string, manejador func(http.ResponseWriter, *http.Request, map[string]string)) *ruta {
	var R = ruta{
		método:    método,
		dirección: dirección,
		manejador: manejador,
	}

	dirección = strings.TrimSpace(dirección)
	dirección = strings.TrimSuffix(dirección, "/")
	if !strings.HasPrefix(dirección, "/") {
		dirección = "/" + dirección
	}

	var partes = []string{}
	for _, p := range strings.Split(dirección, "/") {
		if strings.HasPrefix(p, ":") {
			R.variables = append(R.variables, strings.TrimPrefix(p, ":"))
			partes = append(partes, "(.+?)")
		} else {
			partes = append(partes, regexp.QuoteMeta(p))
		}
	}

	R.dirExpReg = regexp.MustCompile(fmt.Sprintf("^%s$", strings.Join(partes, "/")))

	return &R
}

func (r *ruta) comprobar(mtd string, dir string) (bool, bool, map[string]string) {
	dir = strings.TrimSpace(dir)
	dir = strings.TrimSuffix(dir, "/")
	variablesURL := r.dirExpReg.FindStringSubmatch(dir)
	if len(variablesURL) == 0 {
		return false, false, nil
	}

	if len(r.variables) != len(variablesURL)-1 {
		return false, false, nil
	}

	if r.método != mtd {
		return false, true, nil
	}

	var parámetros = map[string]string{}
	for i, v := range r.variables {
		parámetros[v] = variablesURL[i+1]
	}

	return true, true, parámetros
}
