package normalizer

import (
	"context"
	"encoding/json"
	"log"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"

	"normalizer/internal/stix"
)

var (
	ipv4Regex   = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	sha256Regex = regexp.MustCompile(`\b[A-Fa-f0-9]{64}\b`)
	domainRegex = regexp.MustCompile(`\b(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}\b`)
)

type Processor struct {
	redisClient *redis.Client
	mongoColl   *mongo.Collection
}

func NewProcessor(redisClient *redis.Client, mongoColl *mongo.Collection) *Processor {
	return &Processor{
		redisClient: redisClient,
		mongoColl:   mongoColl,
	}
}

func (p *Processor) Start(ctx context.Context) {
	log.Println("Starting IOC processor...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Processor stopping...")
			return
		default:
			p.ProcessMessages(ctx)
			time.Sleep(1 * time.Second)
		}
	}
}

func (p *Processor) ProcessMessages(ctx context.Context) {
	streams, err := p.redisClient.XRead(ctx, &redis.XReadArgs{
		Streams: []string{"raw-feeds", "0"},
		Block:   5 * time.Second,
	}).Result()

	if err != nil && err != redis.Nil {
		log.Printf("Redis XRead error: %v", err)
		return
	}

	for _, stream := range streams {
		for _, msg := range stream.Messages {
			p.ProcessMessage(ctx, msg)
		}
	}
}

func (p *Processor) ProcessMessage(ctx context.Context, msg redis.XMessage) {
	payload, ok := msg.Values["payload"].(string)
	if !ok {
		log.Printf("Invalid payload in message: %v", msg.ID)
		return
	}

	iocs := ExtractIOCs(payload)
	if len(iocs) == 0 {
		return
	}

	observedTime := GetObservedTime(msg)
	bundle := p.CreateSTIXBundle(iocs, observedTime)

	if err := p.PersistBundle(ctx, bundle); err != nil {
		log.Printf("Persistence error: %v", err)
	}

	if err := p.PublishToStream(ctx, bundle); err != nil {
		log.Printf("Stream publish error: %v", err)
	}
}

func (p *Processor) CreateSTIXBundle(iocs []string, observedTime time.Time) *stix.Bundle {
	bundle := &stix.Bundle{
		Type:        "bundle",
		ID:          "bundle--" + uuid.NewString(),
		SpecVersion: "2.1",
		Created:     time.Now().UTC(),
	}

	for _, ioc := range iocs {
		indicator, observedData, relationship, observable := p.CreateSTIXObjects(ioc, observedTime)
		if observable == nil {
			continue
		}

		bundle.Objects = append(bundle.Objects,
			indicator,
			observedData,
			observable,
			relationship,
		)
	}

	return bundle
}

func (p *Processor) CreateSTIXObjects(ioc string, observedTime time.Time) (*stix.Indicator, *stix.ObservedData, *stix.Relationship, interface{}) {
	observable := p.CreateObservable(ioc)
	if observable == nil {
		return nil, nil, nil, nil
	}

	indicator := &stix.Indicator{
		Type:        "indicator",
		ID:          "indicator--" + uuid.NewString(),
		Created:     observedTime,
		Modified:    observedTime,
		Pattern:     p.CreatePattern(observable),
		PatternType: "stix",
		ValidFrom:   observedTime,
		Labels:      []string{"malicious-activity"},
	}

	observedData := &stix.ObservedData{
		Type:           "observed-data",
		ID:             "observed-data--" + uuid.NewString(),
		Created:        observedTime,
		Modified:       observedTime,
		FirstObserved:  observedTime,
		LastObserved:   observedTime,
		NumberObserved: 1,
		ObjectRefs:     []string{p.GetObservableID(observable)},
	}

	relationship := &stix.Relationship{
		Type:             "relationship",
		ID:               "relationship--" + uuid.NewString(),
		Created:          observedTime,
		Modified:         observedTime,
		SourceRef:        indicator.ID,
		TargetRef:        observedData.ID,
		RelationshipType: "based-on",
	}

	return indicator, observedData, relationship, observable
}

func (p *Processor) CreateObservable(ioc string) interface{} {
	switch {
	case ipv4Regex.MatchString(ioc):
		return &stix.IPv4Addr{
			Type:  "ipv4-addr",
			ID:    "ipv4-addr--" + uuid.NewString(),
			Value: ioc,
		}
	case sha256Regex.MatchString(ioc):
		return &stix.File{
			Type: "file",
			ID:   "file--" + uuid.NewString(),
			Hashes: map[string]string{
				"SHA-256": ioc,
			},
		}
	case domainRegex.MatchString(ioc):
		return &stix.DomainName{
			Type:  "domain-name",
			ID:    "domain-name--" + uuid.NewString(),
			Value: ioc,
		}
	default:
		return nil
	}
}

func (p *Processor) GetObservableID(observable interface{}) string {
	switch obs := observable.(type) {
	case *stix.IPv4Addr:
		return obs.ID
	case *stix.File:
		return obs.ID
	case *stix.DomainName:
		return obs.ID
	default:
		return ""
	}
}

func (p *Processor) CreatePattern(observable interface{}) string {
	switch obs := observable.(type) {
	case *stix.IPv4Addr:
		return "[ipv4-addr:value = '" + obs.Value + "']"
	case *stix.File:
		return "[file:hashes.'SHA-256' = '" + obs.Hashes["SHA-256"] + "']"
	case *stix.DomainName:
		return "[domain-name:value = '" + obs.Value + "']"
	default:
		return ""
	}
}

func (p *Processor) PersistBundle(ctx context.Context, bundle *stix.Bundle) error {
	_, err := p.mongoColl.InsertOne(ctx, bundle)
	return err
}

func (p *Processor) PublishToStream(ctx context.Context, bundle *stix.Bundle) error {
	stixJSON, err := json.Marshal(bundle)
	if err != nil {
		return err
	}

	_, err = p.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: "stix-indicators",
		Values: map[string]interface{}{
			"bundle": string(stixJSON),
			"time":   time.Now().UTC().Format(time.RFC3339),
		},
	}).Result()

	return err
}

func ExtractIOCs(text string) []string {
	var iocs []string
	iocs = append(iocs, ipv4Regex.FindAllString(text, -1)...)
	iocs = append(iocs, sha256Regex.FindAllString(text, -1)...)
	iocs = append(iocs, domainRegex.FindAllString(text, -1)...)
	return iocs
}

func GetObservedTime(msg redis.XMessage) time.Time {
	if ts, ok := msg.Values["timestamp"].(int64); ok {
		return time.Unix(ts, 0).UTC()
	}
	return time.Now().UTC()
}
