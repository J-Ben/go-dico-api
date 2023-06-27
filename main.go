package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"dictionnaire/dictionary"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

func main() {
	// Crée un nouveau dictionnaire en utilisant le chemin du fichier de base de données
	dict, err := dictionary.NewDictionary("C:/Users/bmamfoumbi/Documents/ESTIAM/GoLang/dictionnaire-main/database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer dict.Close()

	router := mux.NewRouter()

	// Définit la route pour obtenir et supprimer un mot spécifique
	router.HandleFunc("/mot/{mot}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		mot := params["mot"]

		switch r.Method {
		case http.MethodGet:
			// Récupère l'entrée correspondant au mot du dictionnaire
			entry, err := dict.GetWord(mot)
			if err != nil {
				if err == bolt.ErrBucketNotFound {
					http.NotFound(w, r)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(entry)

		case http.MethodDelete:
			// Supprime l'entrée correspondant au mot du dictionnaire
			err := dict.DeleteWord(mot)
			if err != nil {
				if err == bolt.ErrBucketNotFound {
					http.NotFound(w, r)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	}).Methods(http.MethodGet, http.MethodDelete)

	// Définit la route pour ajouter un nouveau mot au dictionnaire
	router.HandleFunc("/mot", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var entry dictionary.Entry
			err := json.NewDecoder(r.Body).Decode(&entry)
			if err != nil {
				http.Error(w, "Chargement de la requête invalide", http.StatusBadRequest)
				return
			}

			entry.CreatedAt = time.Now()

			err = dict.AddWord(entry.Word, entry.Definition, entry.CreatedAt)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(entry)

		default:
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	}).Methods(http.MethodPost)

	// Définit la route pour obtenir tous les mots du dictionnaire
	router.HandleFunc("/mots", func(w http.ResponseWriter, r *http.Request) {
		entries, err := dict.GetAllWords()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(entries)
	}).Methods(http.MethodGet)

	log.Println("Écoute sur le port 5060...")
	log.Fatal(http.ListenAndServe(":5060", router))
}
