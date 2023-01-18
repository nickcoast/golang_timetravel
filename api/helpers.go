package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

var (
	ErrInternal = errors.New("internal error")
)

// logs an error if it's not nil
func logError(err error) {
	if err != nil {
		log.Printf("error: %v", err)
	}
}

// writeJSON writes the data as json.
func writeJSON(w http.ResponseWriter, data interface{}, statusCode int) error {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	return err
}

// writeError writes the message as an error
func writeError(w http.ResponseWriter, message string, statusCode int) error {
	log.Printf("response errored: %s", message)
	return writeJSON(
		w,
		map[string]string{"error": message},
		statusCode,
	)
}

// handle erroneous api paths
func resourceNameFromSynonym(resourceSynonym string) (resourceName string, err error) {
	// TODO: send response with correct API path
	synonyms := map[string]string{
		"addresses":         "insured_addresses",
		"address":           "insured_addresses",
		"insured_addresses": "insured_addresses",
		"employee":          "employees",
		"employees":         "employees",
		"insureds":          "insured",
		"insured":           "insured",
	}
	resourceName, ok := synonyms[resourceSynonym]
	if !ok {
		return "", errors.New("No endpoint for" + resourceSynonym + ". Please use 'insured', 'addresses', or 'employees'")
	}
	return resourceName, nil
}