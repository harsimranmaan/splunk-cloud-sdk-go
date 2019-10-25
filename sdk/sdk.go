package sdk

import (
	"github.com/splunk/splunk-cloud-sdk-go/v2/services"
	"github.com/splunk/splunk-cloud-sdk-go/v2/services/action"
	"github.com/splunk/splunk-cloud-sdk-go/v2/services/appregistry"
	"github.com/splunk/splunk-cloud-sdk-go/v2/services/catalog"
	"github.com/splunk/splunk-cloud-sdk-go/v2/services/collect"
	"github.com/splunk/splunk-cloud-sdk-go/v2/services/forwarders"
	"github.com/splunk/splunk-cloud-sdk-go/v2/services/identity"
	"github.com/splunk/splunk-cloud-sdk-go/v2/services/ingest"
	"github.com/splunk/splunk-cloud-sdk-go/v2/services/kvstore"
	"github.com/splunk/splunk-cloud-sdk-go/v2/services/ml"
	"github.com/splunk/splunk-cloud-sdk-go/v2/services/provisioner"
	"github.com/splunk/splunk-cloud-sdk-go/v2/services/search"
	"github.com/splunk/splunk-cloud-sdk-go/v2/services/streams"
)

// Client to communicate with Splunk Cloud service endpoints
type Client struct {
	*services.BaseClient
	// ActionService talks to Splunk Cloud action service
	ActionService *action.Service
	// CatalogService talks to the Splunk Cloud catalog service
	CatalogService *catalog.Service
	// CollectService talks to the Splunk Cloud collect service
	CollectService *collect.Service
	// IdentityService talks to Splunk Cloud IAC service
	IdentityService *identity.Service
	// IngestService talks to the Splunk Cloud ingest service
	IngestService *ingest.Service
	// KVStoreService talks to Splunk Cloud kvstore service
	KVStoreService *kvstore.Service
	// SearchService talks to the Splunk Cloud search service
	SearchService *search.Service
	// StreamsService talks to the Splunk Cloud streams service
	StreamsService *streams.Service
	// ForwardersService talks to the Splunk Cloud forwarders service
	ForwardersService *forwarders.Service
	// appRegistryService talks to the Splunk Cloud app registry service
	AppRegistryService *appregistry.Service
	// MachineLearningService talks to the Splunk Cloud machine learning service
	MachineLearningService *ml.Service
	// ProvisionerService talks to the Splunk Cloud provisioner service
	ProvisionerService *provisioner.Service
}

// NewClient returns a Splunk Cloud client for communicating with any service
func NewClient(config *services.Config) (*Client, error) {
	client, err := services.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Client{
		BaseClient:             client,
		ActionService:          &action.Service{Client: client},
		CatalogService:         &catalog.Service{Client: client},
		CollectService:         &collect.Service{Client: client},
		IdentityService:        &identity.Service{Client: client},
		IngestService:          &ingest.Service{Client: client},
		KVStoreService:         &kvstore.Service{Client: client},
		SearchService:          &search.Service{Client: client},
		StreamsService:         &streams.Service{Client: client},
		ForwardersService:      &forwarders.Service{Client: client},
		AppRegistryService:     &appregistry.Service{Client: client},
		MachineLearningService: &ml.Service{Client: client},
		ProvisionerService:     &provisioner.Service{Client: client},
	}, nil
}

// NewBatchEventsSenderWithMaxAllowedError is Deprecated: please use client.IngestService.NewBatchEventsSenderWithMaxAllowedError
func (c *Client) NewBatchEventsSenderWithMaxAllowedError(batchSize int, interval int64, maxErrorsAllowed int) (*ingest.BatchEventsSender, error) {
	return c.IngestService.NewBatchEventsSenderWithMaxAllowedError(batchSize, interval, 0, maxErrorsAllowed)
}

// NewBatchEventsSender is Deprecated: please use client.IngestService.NewBatchEventsSender
func (c *Client) NewBatchEventsSender(batchSize int, interval int64) (*ingest.BatchEventsSender, error) {
	return c.IngestService.NewBatchEventsSender(batchSize, interval, 0)
}
