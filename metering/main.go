package main

//import "C"
import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/option"
)

func main() {

}

// ------------------------------------------------------------------------------------

// New creates the new plugin.
var New = func() interface{} {
	return &Plugin{
		counters: new(sync.Map),
	}
}

// Plugin represents a usage storage which does nothing.
type Plugin struct {
	client   *datastore.Client
	counters *sync.Map
}

// Name returns the name of the provider.
func (s *Plugin) Name() string {
	return "google datastore"
}

// Configure configures the provider
func (s *Plugin) Configure(config map[string]interface{}) error {
	ctx := context.Background()
	project := config["project_id"].(string)

	// Save the config in a separate file for Google credentials
	if creds, err := json.Marshal(config); err == nil {
		if err := ioutil.WriteFile("creds.json", creds, os.ModePerm); err != nil {
			return err
		}
	}

	// Create a new client with the credentials file
	client, err := datastore.NewClient(ctx, project, option.WithCredentialsFile("creds.json"))
	if err != nil {
		println(err.Error())
		return err
	}

	s.client = client
	return nil
}

// Store stores the meters in some underlying usage storage.
func (s *Plugin) Store() (err error) {
	s.counters.Range(func(k, v interface{}) bool {
		err = s.increment(v.(*Counter))
		return err == nil
	})
	return
}

// Get retrieves a meter for a contract..
func (s *Plugin) Get(id uint32) interface{} {
	meter, _ := s.counters.LoadOrStore(id, NewMeter(id))
	return meter
}

// increment resets and increments a counter in the datastore.
func (s *Plugin) increment(c *Counter) error {
	contractKey := datastore.NameKey("contract", fmt.Sprintf("%v", c.contract), nil)
	meterKey := datastore.NameKey("meter", time.Now().UTC().Format("01/2006"), contractKey)

	inc := c.Reset()
	ctx := context.Background()
	_, err := s.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		var counter Counter
		if err := s.client.Get(ctx, meterKey, &counter); err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}

		// Increment the exising counter
		counter.MessageIn += inc.MessageIn
		counter.TrafficIn += inc.TrafficIn
		counter.MessageEg += inc.MessageEg
		counter.TrafficEg += inc.TrafficEg
		_, err := s.client.Put(ctx, meterKey, &counter)
		return err
	})
	return err
}

// ------------------------------------------------------------------------------------

// Meter represents a tracker for incoming and outgoing traffic.
type Meter interface {
	GetContract() uint32        // Returns the associated contract.
	AddIngress(size int64)      // Records the ingress message size.
	AddEgress(size int64)       // Records the egress message size.
	GetIngress() (int64, int64) // Returns the number of ingress messages and bytes recorded.
	GetEgress() (int64, int64)  // Returns the number of egress messages and bytes recorded.
}

// NewMeter constructs a new usage statistics instance.
func NewMeter(contract uint32) Meter {
	return &Counter{contract: contract}
}

// Counter represents a usage counter we store.
type Counter struct {
	contract  uint32 `datastore:"-"`
	MessageIn int64  `datastore:"MessageIn,noindex"`
	TrafficIn int64  `datastore:"TrafficIn,noindex"`
	MessageEg int64  `datastore:"MessageEg,noindex"`
	TrafficEg int64  `datastore:"TrafficEg,noindex"`
}

// GetContract returns the associated contract.
func (t *Counter) GetContract() uint32 {
	return t.contract
}

// AddIngress records the ingress message size.
func (t *Counter) AddIngress(size int64) {
	atomic.AddInt64(&t.MessageIn, 1)
	atomic.AddInt64(&t.TrafficIn, size)
}

// AddEgress records the egress message size.
func (t *Counter) AddEgress(size int64) {
	atomic.AddInt64(&t.MessageEg, 1)
	atomic.AddInt64(&t.TrafficEg, size)
}

// GetIngress returns the number of ingress messages and bytes recorded.
func (t *Counter) GetIngress() (int64, int64) {
	return atomic.LoadInt64(&t.MessageIn), atomic.LoadInt64(&t.TrafficIn)
}

// GetEgress returns the number of egress messages and bytes recorded.
func (t *Counter) GetEgress() (int64, int64) {
	return atomic.LoadInt64(&t.MessageEg), atomic.LoadInt64(&t.TrafficEg)
}

// Reset resets the tracker and returns old usage.
func (t *Counter) Reset() *Counter {
	var old Counter
	old.contract = t.contract
	old.MessageIn = atomic.SwapInt64(&t.MessageIn, 0)
	old.TrafficIn = atomic.SwapInt64(&t.TrafficIn, 0)
	old.MessageEg = atomic.SwapInt64(&t.MessageEg, 0)
	old.TrafficEg = atomic.SwapInt64(&t.TrafficEg, 0)
	return &old
}
