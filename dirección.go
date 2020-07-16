package servir

import (
	"regexp"
	"strings"
)

type dirección struct {
	dir              string
	variables        []string
	expPlural        *regexp.Regexp
	expSingular      *regexp.Regexp
	puedeSerSingular bool
}

// inicializar analiza la dirección registrada y crea las expresiones regulares que se usarán en la escucha de entradas
func (D *dirección) inicializar(url string) {
	url = strings.TrimSpace(url)
	url = strings.TrimRight(url, "/")

	if !strings.HasPrefix(url, "/") {
		panic("La dirección del recurso debe empezar con una barra diagonal (/)")
	}

	D.dir = url
	D.puedeSerSingular = strings.HasSuffix(url, "}")

	var expVariables *regexp.Regexp

	if D.puedeSerSingular {
		partesBarra := strings.Split(url, "/")
		urlPlural := strings.Join(partesBarra[:len(partesBarra)-1], "/")
		if strings.HasSuffix(urlPlural, "}") {
			panic("La dirección del recurso «" + url + "» no parece cumplir el estándar REST. Use el formato «/usuarios/{usuario}/mensajes/{mensaje}»")
		}

		var expSingular, expPlural string
		D.variables, expSingular = varExpDir(url)
		_, expPlural = varExpDir(urlPlural)
		D.expSingular = regexp.MustCompile(expSingular)
		D.expPlural = regexp.MustCompile(expPlural)
		expVariables = D.expSingular
	} else {
		var expPlural string
		D.variables, expPlural = varExpDir(url)
		D.expPlural = regexp.MustCompile(expPlural)
		expVariables = D.expPlural
	}

	if len(expVariables.FindStringSubmatch(url)) != len(D.variables)+1 {
		panic("La dirección del recurso «" + url + "» no parece tener las variables bien formadas")
	}

}

// coincide verifica si la URL corresponde a la registrada en el recurso
// El primer valor retornado indica si la URL coincide con el recurso
// El segundo parámetro retornado indica si es singular o plural
func (D *dirección) coincide(url string) (bool, string) {
	url = strings.TrimRight(url, "/")

	if D.puedeSerSingular {
		if D.expSingular.MatchString(url) {
			return true, "singular"
		}
	}

	if D.expPlural.MatchString(url) {
		return true, "plural"
	}

	return false, ""
}

// parámetros obtiene un mapa de parámetros variables en la dirección URL
func (D *dirección) parámetros(url string) map[string]string {
	var parámetros = map[string]string{}
	var pedazos = []string{}

	if D.puedeSerSingular {
		pedazos = D.expSingular.FindStringSubmatch(url)
	}

	if len(pedazos) == 0 {
		pedazos = D.expPlural.FindStringSubmatch(url)
	}

	for i, p := range pedazos {
		if i == 0 {
			continue
		}

		parámetros[D.variables[i-1]] = p
	}

	return parámetros
}
