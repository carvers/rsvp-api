package main

import (
	"net/http"
	"strings"

	"darlinggo.co/api"
	"darlinggo.co/trout"
)

func (deps Dependencies) Handler() http.Handler {
	var router trout.Router
	router.Endpoint("/parties").Methods("PUT", "OPTIONS").Handler(CORSMiddleware(http.HandlerFunc(deps.putPartiesHandler)))
	router.Endpoint("/parties").Methods("GET", "OPTIONS").Handler(CORSMiddleware(http.HandlerFunc(deps.getPartiesHandler)))
	router.Endpoint("/people").Methods("PUT", "OPTIONS").Handler(CORSMiddleware(http.HandlerFunc(deps.putPeopleHandler)))
	router.Endpoint("/people").Methods("GET", "OPTIONS").Handler(CORSMiddleware(http.HandlerFunc(deps.getPeopleHandler)))
	return router
}

func CORSMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Header.Get("Origin")
		host = strings.TrimPrefix(host, "http://")
		host = strings.TrimPrefix(host, "https://")
		if strings.HasSuffix(host, ".local") || host == "wedding.carvers.co" || host == "wedding.carvers.house" || host == "192.168.86.123" {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if strings.ToLower(r.Method) == "options" {
			methods := strings.Join(r.Header[http.CanonicalHeaderKey("Trout-Methods")], ", ")
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Allow", methods)
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func userIsAdmin(deps Dependencies, r *http.Request) bool {
	token := r.Header.Get("Authorization")
	if token == "" {
		return false
	}
	// TODO(paddy): decrypt token and check validity
	return false
}

func doWordsMatch(word string, parties []Party) bool {
	if len(parties) != 1 {
		return false
	}
	return parties[0].MagicWord == word
}

func reqCreatesNewNonPlusOnes(admin bool, existing, new []Person) bool {
	getPlusOnes := map[string]string{}
	for _, person := range existing {
		if person.GetsPlusOne {
			getPlusOnes[person.ID] = person.PartyID
		}
	}
	if admin {
		for _, person := range new {
			if person.GetsPlusOne {
				getPlusOnes[person.ID] = person.PartyID
			}
		}
	}
	for _, person := range new {
		var found bool
		for _, exist := range existing {
			if exist.ID == person.ID {
				if exist.IsPlusOne {
					delete(getPlusOnes, exist.IsPlusOneOfID)
				}
				found = true
				break
			}
		}
		if found {
			continue
		}
		if party, ok := getPlusOnes[person.IsPlusOneOfID]; !ok {
			return true
		} else {
			if party != person.PartyID {
				return true
			}
			delete(getPlusOnes, person.IsPlusOneOfID)
		}
	}
	return false
}

type Response struct {
	Parties []Party            `json:"parties,omitempty"`
	People  []Person           `json:"people,omitempty"`
	Errors  []api.RequestError `json:"errors,omitempty"`
}

func (deps Dependencies) putPartiesHandler(w http.ResponseWriter, r *http.Request) {
	var partiesReq struct {
		Parties []Party `json:"parties"`
	}
	err := api.Decode(r, &partiesReq)
	if err != nil {
		deps.log.Printf("%+v\n", err)
		api.Encode(w, r, http.StatusBadRequest, Response{Errors: api.InvalidFormatError})
		return
	}
	if !userIsAdmin(deps, r) {
		api.Encode(w, r, http.StatusUnauthorized, Response{Errors: api.AccessDeniedError})
		return
	}
	parties, err := deps.CreateParties(r.Context(), partiesReq.Parties)
	if err != nil {
		deps.log.Printf("%+v\n", err)
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	api.Encode(w, r, http.StatusOK, Response{Parties: parties})
}

func (deps Dependencies) putPeopleHandler(w http.ResponseWriter, r *http.Request) {
	var peopleReq struct {
		People []Person `json:"people"`
	}
	err := api.Decode(r, &peopleReq)
	if err != nil {
		deps.log.Printf("%+v\n", err)
		api.Encode(w, r, http.StatusBadRequest, Response{Errors: api.InvalidFormatError})
		return
	}

	// retrieve the people we're updating
	peopleIDs := make([]string, 0, len(peopleReq.People))
	for _, person := range peopleReq.People {
		peopleIDs = append(peopleIDs, person.ID)
	}
	people, err := deps.GetPeople(r.Context(), peopleIDs)
	if err != nil {
		deps.log.Printf("Error retrieving people: %+v\n", err)
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}

	// check if we're adding new people that aren't +1s of existing people
	newNonPlusOnes := reqCreatesNewNonPlusOnes(userIsAdmin(deps, r), people, peopleReq.People)
	if newNonPlusOnes && !userIsAdmin(deps, r) {
		api.Encode(w, r, http.StatusUnauthorized, Response{Errors: api.AccessDeniedError})
		return
	}

	// figure out what code word we need to be using for all these people
	partyIDmap := map[string]struct{}{}
	for _, person := range peopleReq.People {
		partyIDmap[person.PartyID] = struct{}{}
	}

	// if there's more than one code word required and the user isn't an admin
	// this is a bad request, bail out
	if len(partyIDmap) != 1 && !userIsAdmin(deps, r) {
		api.Encode(w, r, http.StatusBadRequest, Response{Errors: []api.RequestError{{Slug: api.RequestErrOverflow, Field: "/people/partyID"}}})
		return
	}

	// still need to get all the parties, though, in case user is admin
	partyIDs := make([]string, 0, len(partyIDmap))
	for id := range partyIDmap {
		partyIDs = append(partyIDs, id)
	}
	parties, err := deps.GetParties(r.Context(), partyIDs)
	if err != nil {
		deps.log.Printf("%+v\n", err)
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}

	// sanity check, in case we have a bad party ID or something
	if len(parties) != len(partyIDs) {
		deps.log.Printf("Expected %d results, got %d\n", len(partyIDs), len(parties))
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}

	// if we're not adding new people and not an admin
	// we need to identify ourselves with the code word
	if !newNonPlusOnes && !userIsAdmin(deps, r) {
		if !doWordsMatch(r.Header.Get("Code-Word"), parties) {
			api.Encode(w, r, http.StatusForbidden, Response{Errors: api.AccessDeniedError})
			return
		}
	}

	// finally, write everything back in
	people, err = deps.CreatePeople(r.Context(), peopleReq.People)
	if err != nil {
		deps.log.Printf("%+v\n", err)
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.AccessDeniedError})
		return
	}
	api.Encode(w, r, http.StatusOK, Response{People: people})
}

func (deps Dependencies) getPeopleHandler(w http.ResponseWriter, r *http.Request) {
	var people []Person
	var err error
	personIDs := r.URL.Query()["person_id"]
	partyID := r.URL.Query().Get("party_id")
	switch {
	case len(personIDs) > 0:
		// TODO(paddy): authenticate; use the codeWord and only allow retrieving people from that party
		people, err = deps.GetPeople(r.Context(), personIDs)
		if err != nil {
			deps.log.Printf("%+v\n", err)
			return
		}
	default:
		if !userIsAdmin(deps, r) && partyID == "" {
			api.Encode(w, r, http.StatusUnauthorized, Response{Errors: api.AccessDeniedError})
			return
		} else if partyID != "" {
			par, err := deps.GetParties(r.Context(), []string{partyID})
			if err != nil {
				deps.log.Printf("%+v\n", err)
				api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
				return
			}
			if len(par) < 1 || par[0].MagicWord != r.Header.Get("Code-Word") {
				api.Encode(w, r, http.StatusUnauthorized, Response{Errors: api.AccessDeniedError})
				return
			}
		}
		people, err = deps.ListPeople(r.Context(), partyID)
		if err != nil {
			deps.log.Printf("%+v\n", err)
			return
		}
	}
	api.Encode(w, r, http.StatusOK, Response{People: people})
}

func (deps Dependencies) getPartiesHandler(w http.ResponseWriter, r *http.Request) {
	var parties []Party
	var err error
	partyIDs := r.URL.Query()["party_id"]
	partyWord := r.URL.Query().Get("magic_word")
	switch {
	case len(partyIDs) > 0:
		if !userIsAdmin(deps, r) {
			api.Encode(w, r, http.StatusUnauthorized, Response{Errors: api.AccessDeniedError})
			return
		}
		parties, err = deps.GetParties(r.Context(), partyIDs)
		if err != nil {
			deps.log.Printf("%+v\n", err)
			return
		}
	case partyWord != "":
		party, err := deps.GetPartyByMagicWord(r.Context(), partyWord)
		if err != nil {
			if err == ErrMagicWordNotFound {
				w.WriteHeader(http.StatusNotFound)
				deps.log.Printf("%+v\n", err)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			deps.log.Printf("%+v\n", err)
			return
		}
		parties = append(parties, party)
	default:
		if !userIsAdmin(deps, r) {
			api.Encode(w, r, http.StatusUnauthorized, Response{Errors: api.AccessDeniedError})
			return
		}
		parties, err = deps.ListParties(r.Context())
		if err != nil {
			deps.log.Printf("%+v\n", err)
			api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
			return
		}
	}
	api.Encode(w, r, http.StatusOK, Response{Parties: parties})
}
