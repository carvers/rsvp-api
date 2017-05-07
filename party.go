package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"darlinggo.co/pan"

	"github.com/coreos/go-oidc"
	"github.com/pborman/uuid"
)

var (
	ErrMagicWordNotFound = errors.New("magic word not found")
)

type Dependencies struct {
	db       *sql.DB
	log      *log.Logger
	verifier *oidc.IDTokenVerifier
}

type Party struct {
	ID        string `json:"ID"`
	LeadID    string `json:"lead,omitempty"`
	Name      string `json:"name,omitempty"`
	SortValue string `json:"sortValue"`
	Address   string `json:"address,omitempty"`
	MagicWord string `json:"codeWord,omitempty"`
}

func (p Party) GetSQLTableName() string {
	return "parties"
}

func (p Party) getUpsertAction() string {
	columns := pan.Columns(p)
	for pos, column := range columns {
		columns[pos] = fmt.Sprintf("%s = EXCLUDED.%s", column, column)
	}
	return columns.String()
}

func (deps Dependencies) CreateParties(ctx context.Context, parties []Party) ([]Party, error) {
	tabler := make([]pan.SQLTableNamer, 0, len(parties))
	for _, party := range parties {
		if party.ID == "" {
			party.ID = uuid.New()
		}
		tabler = append(tabler, party)
	}
	query := pan.Insert(tabler...).Flush(", ").Expression("ON CONFLICT (" + pan.Column(parties[0], "ID") + ") DO UPDATE SET").Flush(" ")
	query.Expression(parties[0].getUpsertAction()).Flush(", ")
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return nil, err
	}
	deps.log.Println(queryStr)
	_, err = deps.db.Exec(queryStr, query.Args()...)
	if err != nil {
		return nil, err
	}
	return parties, err
}

type Person struct {
	ID                  string `json:"ID"`
	PartyID             string `json:"party,omitempty"`
	Name                string `json:"name,omitempty"`
	Email               string `json:"email,omitempty"`
	GetsPlusOne         bool   `json:"getsPlusOne"`
	PlusOneID           string `json:"plusOne,omitempty"`
	IsPlusOne           bool   `json:"isPlusOne"`
	IsPlusOneOfID       string `json:"isPlusOneOf,omitempty"`
	Replied             bool   `json:"replied"`
	Reply               bool   `json:"reply"`
	DietaryRestrictions string `json:"dietaryRestrictions,omitempty"`
	SongRequest         string `json:"songRequest,omitempty"`
	IsChild             bool   `json:"isChild"`
	WillAccompanyID     string `json:"willAccompany,omitempty"`

	Hiking     bool `json:"hiking"`
	Kayaking   bool `json:"kayaking"`
	Jetski     bool `json:"jetski"`
	Fishing    bool `json:"fishing"`
	Hanford    bool `json:"hanford"`
	Ligo       bool `json:"ligo"`
	Reach      bool `json:"reach"`
	Bechtel    bool `json:"bechtel"`
	Wine       bool `json:"wine"`
	EscapeRoom bool `json:"escapeRoom"`
}

func (p Person) GetSQLTableName() string {
	return "people"
}

func (p Person) getUpsertAction() string {
	columns := pan.Columns(p)
	for pos, column := range columns {
		columns[pos] = fmt.Sprintf("%s = EXCLUDED.%s", column, column)
	}
	return columns.String()
}

func (deps Dependencies) CreatePeople(ctx context.Context, people []Person) ([]Person, error) {
	tabler := make([]pan.SQLTableNamer, 0, len(people))
	for _, person := range people {
		if person.ID == "" {
			person.ID = uuid.New()
		}
		tabler = append(tabler, person)
	}
	query := pan.Insert(tabler...).Flush(", ").Expression("ON CONFLICT (" + pan.Column(people[0], "ID") + ") DO UPDATE SET").Flush(" ")
	query.Expression(people[0].getUpsertAction()).Flush(", ")
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return nil, err
	}
	deps.log.Println(queryStr)
	_, err = deps.db.Exec(queryStr, query.Args()...)
	if err != nil {
		return nil, err
	}
	return people, nil
}

func (deps Dependencies) ListParties(ctx context.Context) ([]Party, error) {
	var p Party
	query := pan.New("SELECT " + pan.Columns(p).String() + " FROM " + pan.Table(p)).OrderBy("sort_value").Flush(" ")
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return nil, err
	}
	rows, err := deps.db.Query(queryStr, query.Args()...)
	if err != nil {
		return nil, err
	}
	var parties []Party
	for rows.Next() {
		var party Party
		err := pan.Unmarshal(rows, &party)
		if err != nil {
			return nil, err
		}
		parties = append(parties, party)
	}
	return parties, nil
}

func (deps Dependencies) GetParties(ctx context.Context, ids []string) ([]Party, error) {
	var p Party
	ifIDs := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		ifIDs = append(ifIDs, id)
	}
	query := pan.New("SELECT " + pan.Columns(p).String() + " FROM " + pan.Table(p)).Where()
	query.In(p, "ID", ifIDs...).OrderBy("sort_value").Flush(" ")
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return nil, err
	}
	rows, err := deps.db.Query(queryStr, query.Args()...)
	if err != nil {
		return nil, err
	}
	var parties []Party
	for rows.Next() {
		var party Party
		err := pan.Unmarshal(rows, &party)
		if err != nil {
			return nil, err
		}
		parties = append(parties, party)
	}
	return parties, nil
}

func (deps Dependencies) GetPartyByMagicWord(ctx context.Context, word string) (Party, error) {
	var p Party
	query := pan.New("SELECT " + pan.Columns(p).String() + " FROM " + pan.Table(p)).Where()
	query.Comparison(p, "MagicWord", "=", word).Limit(1).Flush(" ")
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return p, err
	}
	rows, err := deps.db.Query(queryStr, query.Args()...)
	if err != nil {
		return p, err
	}
	var found bool
	for rows.Next() {
		err := pan.Unmarshal(rows, &p)
		if err != nil {
			return p, err
		}
		found = true
	}
	if !found {
		return Party{}, ErrMagicWordNotFound
	}
	return p, nil
}

func (deps Dependencies) ListPeople(ctx context.Context, party string) ([]Person, error) {
	var p Person
	query := pan.New("SELECT " + pan.Columns(p).String() + " FROM " + pan.Table(p))
	if party != "" {
		query.Where()
		query.Comparison(p, "PartyID", "=", party)
	}
	query.Flush(" ")
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return nil, err
	}

	rows, err := deps.db.Query(queryStr, query.Args()...)
	if err != nil {
		return nil, err
	}
	var people []Person
	for rows.Next() {
		var person Person
		err := pan.Unmarshal(rows, &person)
		if err != nil {
			return nil, err
		}
		people = append(people, person)
	}
	return people, nil
}

func (deps Dependencies) GetPeople(ctx context.Context, ids []string) ([]Person, error) {
	var p Person
	ifIDs := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		ifIDs = append(ifIDs, id)
	}

	query := pan.New("SELECT "+pan.Columns(p).String()+" FROM "+pan.Table(p)).Where().In(p, "ID", ifIDs...).Flush(" ")
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return nil, err
	}
	deps.log.Println("GetPeople query string:", queryStr)

	rows, err := deps.db.Query(queryStr, query.Args()...)
	if err != nil {
		return nil, err
	}

	var people []Person
	for rows.Next() {
		var person Person
		err := pan.Unmarshal(rows, &person)
		if err != nil {
			return nil, err
		}
		people = append(people, person)
	}
	return people, nil
}
