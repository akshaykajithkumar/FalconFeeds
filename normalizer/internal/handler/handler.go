package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"normalizer/internal/stix"
)

type APIHandler struct {
	collection *mongo.Collection
}

func NewAPIHandler(collection *mongo.Collection) *APIHandler {
	return &APIHandler{collection: collection}
}

func (h *APIHandler) RegisterRoutes() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/healthz", h.healthCheck).Methods("GET")
	r.HandleFunc("/indicators", h.getIndicators).Methods("GET")
	r.HandleFunc("/indicators/{id}", h.getIndicatorByID).Methods("GET")
	return r
}

func (h *APIHandler) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *APIHandler) getIndicators(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	value := query.Get("value")
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 {
		limit = 10
	}

	filter := bson.M{}
	if value != "" {
		filter["$or"] = []bson.M{
			{"objects.pattern": bson.M{"$regex": value}},
			{"objects.value": value},
			{"objects.hashes.SHA-256": value},
		}
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created", Value: -1}})

	cur, err := h.collection.Find(r.Context(), filter, opts)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer cur.Close(r.Context())

	var bundles []stix.Bundle
	if err := cur.All(r.Context(), &bundles); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, bundles)
}

func (h *APIHandler) getIndicatorByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var bundle stix.Bundle
	err := h.collection.FindOne(r.Context(), bson.M{"id": id}).Decode(&bundle)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "indicator not found")
		return
	}

	respondWithJSON(w, http.StatusOK, bundle)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
