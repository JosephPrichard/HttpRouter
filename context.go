package httprouter

import (
	"context"
	"net/http"
)

type varskey int

// the id will always be unique since the varskey type only exists in this package
var id = varskey(1)

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
