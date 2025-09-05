package utils

import (
    "encoding/json"
    "net/http"
)

func ReadJSON(r *http.Request, v interface{}) error {
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    return dec.Decode(v)
}

func RespondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(payload)
}

func RespondWithError(w http.ResponseWriter, status int, msg string) {
    RespondWithJSON(w, status, map[string]interface{}{"error": msg})
}
