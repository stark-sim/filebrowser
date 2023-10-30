package http

import (
	"net/http"
)

var bdLogin = withHashFile(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	//loginInfo := d.raw.(*bd.BDLoginCode)

	return renderJSON(w, r, d)
})
