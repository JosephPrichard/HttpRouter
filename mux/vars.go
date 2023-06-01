package mux

import (
	"context"
	"net/http"
)

type key int

var id = key(1)

func setVar(r *http.Request, key string, value string) {
	vars := Vars(r)
	vars[key] = value
	ctx := context.WithValue(r.Context(), id, vars)
	*r = *r.WithContext(ctx)
}

func Vars(r *http.Request) map[string]string {
	val := r.Context().Value(id)
	if val == nil {
		return make(map[string]string)
	}
	return val.(map[string]string)
}