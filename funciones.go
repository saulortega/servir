package servir

import (
	"regexp"
	"strings"
)

func varExpDir(dir string) ([]string, string) {
	var porciones = []string{}
	var variables = []string{}

	var partes1 = strings.Split(dir, "{")

	for i, p1 := range partes1 {
		if i == 0 {
			porciones = append(porciones, regexp.QuoteMeta(p1))
			continue
		}

		partes2 := strings.Split(p1, "}")
		if len(partes2) != 2 || len(partes2[0]) == 0 {
			panic("Parámetro de recurso mal formateado: " + p1)
		}

		variables = append(variables, partes2[0])
		porciones = append(porciones, regexp.QuoteMeta(partes2[1]))
	}

	var exp = "^" + strings.Join(porciones, "([^/]*?)") + "$"

	return variables, exp
}
