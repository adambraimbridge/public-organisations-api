package organisation

import (
	"fmt"
	"net/http"

	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/gorilla/mux"
)

// PeopleDriver for cypher queries
var PeopleDriver Driver

// HealthCheck does something
func HealthCheck() v1a.Check {
	return v1a.Check{
		BusinessImpact:   "Unable to respond to Public People api requests",
		Name:             "Check connectivity to Neo4j - neoUrl is a parameter in hieradata for this service",
		PanicGuide:       "TODO - write panic guide",
		Severity:         1,
		TechnicalSummary: "Cannot connect to Neo4j a instance with at least one person loaded in it",
		Checker:          Checker,
	}
}

// Checker does more stuff
func Checker() (string, error) {
	err := PeopleDriver.CheckConnectivity()
	if err == nil {
		return "Connectivity to neo4j is ok", err
	}
	return "Error connecting to neo4j", err
}

// Ping says pong
func Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

// GetOrganisation is the public API
func GetOrganisation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if uuid == "" {
		http.Error(w, "uuid required", http.StatusBadRequest)
		return
	}
	/*	organisation, found, err := OrganisationDriver.Read(uuid)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "` + err.Error() + `"}`))
			return
		}
		if !found {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"Organisation not found."}`))
			return
		}
		Jason, _ := json.Marshal(organisation)
		log.Debugf("Organisation(uuid:%s): %s\n", Jason)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(organisation)*/
}
