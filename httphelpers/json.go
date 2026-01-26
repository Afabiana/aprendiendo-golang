package httphelpers

import (
	"encoding/json"
	"log"
	"net/http"
)

func RespondError(w http.ResponseWriter, msg string, status int) {
	RespondJSON(w, status, map[string]string{
		"error": msg,
	})
}

func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// No se puede volver a escribir headers acá
		// Solo loguear
		log.Println("error encoding JSON:", err)
	}
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // para evitar campos tipo { "hack": "ll..ñk" } Decoder compara: keys del JSON vs campos exportados del struct + tags json

	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}
