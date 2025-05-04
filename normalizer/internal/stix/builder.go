package stix

import "time"

// Bundle represents a STIX 2.1 Bundle
type Bundle struct {
	Type        string    `json:"type" bson:"type"`
	ID          string    `json:"id" bson:"id"`
	Objects     []any     `json:"objects" bson:"objects"`
	SpecVersion string    `json:"spec_version" bson:"spec_version"`
	Created     time.Time `json:"created" bson:"created"`
}

// Indicator represents a STIX 2.1 Indicator
type Indicator struct {
	Type        string    `json:"type" bson:"type"`
	ID          string    `json:"id" bson:"id"`
	Created     time.Time `json:"created" bson:"created"`
	Modified    time.Time `json:"modified" bson:"modified"`
	Pattern     string    `json:"pattern" bson:"pattern"`
	PatternType string    `json:"pattern_type" bson:"pattern_type"`
	ValidFrom   time.Time `json:"valid_from" bson:"valid_from"`
	Labels      []string  `json:"labels" bson:"labels"`
}

// ObservedData represents a STIX 2.1 Observed Data
type ObservedData struct {
	Type           string    `json:"type" bson:"type"`
	ID             string    `json:"id" bson:"id"`
	Created        time.Time `json:"created" bson:"created"`
	Modified       time.Time `json:"modified" bson:"modified"`
	FirstObserved  time.Time `json:"first_observed" bson:"first_observed"`
	LastObserved   time.Time `json:"last_observed" bson:"last_observed"`
	NumberObserved int       `json:"number_observed" bson:"number_observed"`
	ObjectRefs     []string  `json:"object_refs" bson:"object_refs"`
}

// Relationship represents a STIX 2.1 Relationship
type Relationship struct {
	Type             string    `json:"type" bson:"type"`
	ID               string    `json:"id" bson:"id"`
	Created          time.Time `json:"created" bson:"created"`
	Modified         time.Time `json:"modified" bson:"modified"`
	SourceRef        string    `json:"source_ref" bson:"source_ref"`
	TargetRef        string    `json:"target_ref" bson:"target_ref"`
	RelationshipType string    `json:"relationship_type" bson:"relationship_type"`
}

// DomainName represents a STIX 2.1 Domain Name observable
type DomainName struct {
	Type  string `json:"type" bson:"type"`
	ID    string `json:"id" bson:"id"`
	Value string `json:"value" bson:"value"`
}

// IPv4Addr represents a STIX 2.1 IPv4 Address observable
type IPv4Addr struct {
	Type  string `json:"type" bson:"type"`
	ID    string `json:"id" bson:"id"`
	Value string `json:"value" bson:"value"`
}

// File represents a STIX 2.1 File observable
type File struct {
	Type   string            `json:"type" bson:"type"`
	ID     string            `json:"id" bson:"id"`
	Hashes map[string]string `json:"hashes" bson:"hashes"`
}
