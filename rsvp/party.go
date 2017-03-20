package rsvp

import (
	"errors"

	"code.secondbit.org/uuid.hg"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

const (
	partyKind  = "Party"
	personKind = "Person"
)

var (
	ErrMagicWordNotFound = errors.New("magic word not found")
)

type Party struct {
	ID        string         `datastore:"-" json:"ID"`
	Lead      *datastore.Key `json:"-"`
	LeadID    string         `datastore:"-" json:"lead,omitempty"`
	Name      string         `json:"name,omitempty"`
	SortValue string         `json:"sortValue"`
	Address   string         `json:"address,omitempty"`
	MagicWord string         `json:"codeWord,omitempty"`
}

func (p Party) Key(ctx context.Context) *datastore.Key {
	return datastore.NewKey(ctx, partyKind, p.ID, 0, nil)
}

func (p Party) FillKeyIDs(ctx context.Context) Party {
	if p.Lead != nil && p.LeadID == "" {
		p.LeadID = p.Lead.StringID()
	}
	if p.Lead == nil && p.LeadID != "" {
		p.Lead = datastore.NewKey(ctx, partyKind, p.LeadID, 0, nil)
	}
	return p
}

func CreateParties(ctx context.Context, parties []Party) ([]Party, error) {
	var keys []*datastore.Key
	var err error
	for pos, party := range parties {
		if party.ID == "" {
			party.ID = uuid.NewID().String()
		}
		party = party.FillKeyIDs(ctx)
		parties[pos] = party
		keys = append(keys, party.Key(ctx))
	}
	_, err = datastore.PutMulti(ctx, keys, parties)
	return parties, err
}

type Person struct {
	ID                   string         `datastore:"-" json:"ID"`
	Party                *datastore.Key `json:"-"`
	PartyID              string         `datastore:"-" json:"party,omitempty"`
	Name                 string         `json:"name,omitempty"`
	Email                string         `json:"email,omitempty"`
	GetsPlusOne          bool           `json:"getsPlusOne"`
	PlusOne              *datastore.Key `json:"-"`
	PlusOneID            string         `datastore:"-" json:"plusOne,omitempty"`
	IsPlusOne            bool           `json:"isPlusOne"`
	IsPlusOneOf          *datastore.Key `json:"-"`
	IsPlusOneOfID        string         `datastore:"-" json:"isPlusOneOf,omitempty"`
	Replied              bool           `json:"replied"`
	Reply                bool           `json:"reply"`
	DietaryRestrictions  string         `json:"dietaryRestrictions,omitempty"`
	SongRequest          string         `json:"songRequest,omitempty"`
	IsChild              bool           `json:"isChild"`
	WillAccompany        *datastore.Key `json:"-"`
	WillAccompanyID      string         `datastore:"-" json:"willAccompany,omitempty"`
	BabysitterForWedding bool           `json:"babysitterForWedding"`
	BabysitterForEvents  bool           `json:"babysitterForEvents"`
}

func (p Person) Key(ctx context.Context) *datastore.Key {
	return datastore.NewKey(ctx, personKind, p.ID, 0, nil)
}

func (p Person) FillKeyIDs(ctx context.Context) Person {
	if p.Party != nil && p.PartyID == "" {
		p.PartyID = p.Party.StringID()
	}
	if p.Party == nil && p.PartyID != "" {
		p.Party = datastore.NewKey(ctx, partyKind, p.PartyID, 0, nil)
	}
	if p.PlusOne != nil && p.PlusOneID == "" {
		p.PlusOneID = p.PlusOne.StringID()
	}
	if p.PlusOne == nil && p.PlusOneID != "" {
		p.PlusOne = datastore.NewKey(ctx, personKind, p.PlusOneID, 0, nil)
	}
	if p.IsPlusOneOf != nil && p.IsPlusOneOfID == "" {
		p.IsPlusOneOfID = p.IsPlusOneOf.StringID()
	}
	if p.IsPlusOneOf == nil && p.IsPlusOneOfID != "" {
		p.IsPlusOneOf = datastore.NewKey(ctx, personKind, p.IsPlusOneOfID, 0, nil)
	}
	if p.WillAccompany != nil && p.WillAccompanyID == "" {
		p.WillAccompanyID = p.WillAccompany.StringID()
	}
	if p.WillAccompany == nil && p.WillAccompanyID != "" {
		p.WillAccompany = datastore.NewKey(ctx, personKind, p.WillAccompanyID, 0, nil)
	}
	return p
}

func CreatePeople(ctx context.Context, people []Person) ([]Person, error) {
	var keys []*datastore.Key
	var err error
	for pos, person := range people {
		if person.ID == "" {
			person.ID = uuid.NewID().String()
		}
		person = person.FillKeyIDs(ctx)
		people[pos] = person
		keys = append(keys, person.Key(ctx))
	}
	_, err = datastore.PutMulti(ctx, keys, people)
	return people, err
}

func ListParties(ctx context.Context) ([]Party, error) {
	q := datastore.NewQuery(partyKind).Order("Name")
	result := q.Run(ctx)
	var err error
	var parties []Party
	for err == nil {
		var party Party
		var key *datastore.Key
		key, err = result.Next(&party)
		if err == nil {
			party.ID = key.StringID()
			party = party.FillKeyIDs(ctx)
			parties = append(parties, party)
		}
	}
	if err == datastore.Done {
		err = nil
	}
	return parties, err
}

func GetParties(ctx context.Context, ids []string) ([]Party, error) {
	var keys []*datastore.Key
	for _, id := range ids {
		keys = append(keys, datastore.NewKey(ctx, partyKind, id, 0, nil))
	}
	parties := make([]Party, len(keys))
	err := datastore.GetMulti(ctx, keys, parties)
	if err != nil {
		return []Party{}, err
	}
	for pos, party := range parties {
		party.ID = keys[pos].StringID()
		party = party.FillKeyIDs(ctx)
		parties[pos] = party
	}
	return parties, nil
}

func GetPartyByMagicWord(ctx context.Context, word string) (Party, error) {
	q := datastore.NewQuery(partyKind).Filter("MagicWord =", word).Limit(1)
	var parties []Party
	keys, err := q.GetAll(ctx, &parties)
	if err != nil {
		return Party{}, err
	}
	if len(parties) < 1 {
		return Party{}, ErrMagicWordNotFound
	}
	party := parties[0]
	party.ID = keys[0].StringID()
	party = party.FillKeyIDs(ctx)
	return party, nil
}

func ListPeople(ctx context.Context, party *datastore.Key) ([]Person, error) {
	q := datastore.NewQuery("Person")
	if party != nil {
		q = q.Filter("Party =", party)
	}
	result := q.Run(ctx)
	var err error
	var people []Person
	for err == nil {
		var person Person
		var key *datastore.Key
		key, err = result.Next(&person)
		if err == nil {
			person.ID = key.StringID()
			person = person.FillKeyIDs(ctx)
			people = append(people, person)
		}
	}
	if err == datastore.Done {
		err = nil
	}
	return people, err
}

func GetPeople(ctx context.Context, ids []string) ([]Person, error) {
	var keys []*datastore.Key
	for _, id := range ids {
		keys = append(keys, datastore.NewKey(ctx, personKind, id, 0, nil))
	}
	people := make([]Person, len(keys))
	err := datastore.GetMulti(ctx, keys, people)
	if err != nil {
		return people, err
	}
	for pos, person := range people {
		person.ID = keys[pos].StringID()
		person = person.FillKeyIDs(ctx)
		people[pos] = person
	}
	return people, nil
}
