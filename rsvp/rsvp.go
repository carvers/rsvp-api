package rsvp

import (
	"fmt"
	"net/http"

	"darlinggo.co/api"
	"darlinggo.co/trout"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

const scope = "https://www.googleapis.com/auth/userinfo.email"

func init() {
	var router trout.Router
	router.Endpoint("/parties").Methods("PUT", "OPTIONS").Handler(api.CORSMiddleware(http.HandlerFunc(putPartiesHandler)))
	router.Endpoint("/parties").Methods("GET", "OPTIONS").Handler(api.CORSMiddleware(http.HandlerFunc(getPartiesHandler)))
	router.Endpoint("/people").Methods("PUT", "OPTIONS").Handler(api.CORSMiddleware(http.HandlerFunc(putPeopleHandler)))
	router.Endpoint("/people").Methods("GET", "OPTIONS").Handler(api.CORSMiddleware(http.HandlerFunc(getPeopleHandler)))
	http.Handle("/", router)
}

type Response struct {
	Parties []Party            `json:"parties,omitempty"`
	People  []Person           `json:"people,omitempty"`
	Error   []api.RequestError `json:"errors,omitempty"`
}

func putPartiesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	var partiesReq struct {
		Parties []Party `json:"parties"`
	}
	err := api.Decode(r, &partiesReq)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
		return
	}
	// TODO(paddy): authenticate if it would create a new party
	// TODO(paddy): require codeword to update existing party
	parties, err := CreateParties(ctx, partiesReq.Parties)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
		return
	}
	api.Encode(w, r, http.StatusOK, Response{Parties: parties})
}

func putPeopleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	var peopleReq struct {
		People []Person `json:"people"`
	}
	err := api.Decode(r, &peopleReq)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
		return
	}
	// TODO(paddy): authenticate if it would create a new non-plus-one person
	// TODO(paddy): require codeword to update existing people
	people, err := CreatePeople(ctx, peopleReq.People)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
		return
	}
	api.Encode(w, r, http.StatusOK, Response{People: people})
}

func getPeopleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	var people []Person
	var err error
	personIDs := r.URL.Query()["person_id"]
	partyID := r.URL.Query().Get("party_id")
	var party *datastore.Key
	if partyID != "" {
		party = datastore.NewKey(ctx, partyKind, partyID, 0, nil)
	}
	switch {
	case len(personIDs) > 0:
		// TODO(paddy): authenticate; use the codeWord and only allow retrieving people from that party
		people, err = GetPeople(ctx, personIDs)
		if err != nil {
			fmt.Fprintf(w, "%+v\n", err)
			return
		}
	default:
		// TODO(paddy): authenticate
		people, err = ListPeople(ctx, party)
		if err != nil {
			fmt.Fprintf(w, "%+v\n", err)
			return
		}
	}
	api.Encode(w, r, http.StatusOK, Response{People: people})
}

func getPartiesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	var parties []Party
	var err error
	partyIDs := r.URL.Query()["party_id"]
	partyWord := r.URL.Query().Get("magic_word")
	switch {
	case len(partyIDs) > 0:
		// TODO(paddy): authenticate
		parties, err = GetParties(ctx, partyIDs)
		if err != nil {
			fmt.Fprintf(w, "%+v\n", err)
			return
		}
	case partyWord != "":
		party, err := GetPartyByMagicWord(ctx, partyWord)
		if err != nil {
			if err == ErrMagicWordNotFound {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "%+v\n", err)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%+v\n", err)
			return
		}
		parties = append(parties, party)
	default:
		// TODO(paddy): authenticate
		parties, err = ListParties(ctx)
		if err != nil {
			fmt.Fprintf(w, "%+v\n", err)
			return
		}
	}
	api.Encode(w, r, http.StatusOK, Response{Parties: parties})
}
