package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/rest"

	"github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/cmd"
	"golang.org/x/net/publicsuffix"
)

var GroupName = os.Getenv("GROUP_NAME")

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	cmd.RunWebhookServer(GroupName,
		&nextlayerDNSProviderSolver{},
	)
}

// customDNSProviderSolver implements the provider-specific logic needed to
// 'present' an ACME challenge TXT record for your own DNS provider.
// To do so, it must implement the `github.com/jetstack/cert-manager/pkg/acme/webhook.Solver`
// interface.
type nextlayerDNSProviderSolver struct {
}

type nextlayerDNSProviderConfig struct {
	APIKey string `json:"apiKey"`
}

type Record struct {
	Name     string `json:"name"`
	Content  string `json:"content"`
	Disabled bool   `json:"disabled"`
}

type RRSet struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Changetype string   `json:"changetype"`
	TTL        int      `json:"ttl"`
	Records    []Record `json:"records"`
}

type RRSetUpdateRequest struct {
	RRSets []RRSet `json:"rrsets"`
}

func (c *nextlayerDNSProviderSolver) Name() string {
	return "nextlayer"
}

func (c *nextlayerDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}
	client := &http.Client{}

	name, domain, zone := c.getDomainAndEntry(ch)
	fmt.Println("name: ", name)
	fmt.Println("domain: ", domain)
	fmt.Println("zone: ", zone)

	rrsetRequest := &RRSetUpdateRequest{
		RRSets: []RRSet{
			{
				Name:       name + "." + domain + ".",
				Type:       "TXT",
				Changetype: "REPLACE",
				TTL:        60,
				Records: []Record{
					{
						Name:     name + "." + domain + ".",
						Content:  `"` + ch.Key + `"`,
						Disabled: false,
					},
				},
			},
		},
	}

	entries, err := json.Marshal(rrsetRequest)
	fmt.Println("Request that is being sent to PowerDNS: ", string(entries))

	requestBody := bytes.NewBuffer(entries)

	// Create new record by patching zone
	req, err := http.NewRequest("PATCH", "https://dns.nextlayer.at/api/v1/servers/localhost/zones/"+zone, requestBody)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-API-Key", cfg.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failure : ", err)
	}

	// Display Results
	fmt.Println("response Status : ", resp.Status)
	fmt.Println("response Headers : ", resp.Header)
	respBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body : ", string(respBody))

	return nil
}

func (c *nextlayerDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}
	client := &http.Client{}

	name, domain, zone := c.getDomainAndEntry(ch)
	fmt.Println("name: ", name)
	fmt.Println("domain: ", domain)
	fmt.Println("zone: ", zone)

	rrsetRequest := &RRSetUpdateRequest{
		RRSets: []RRSet{
			{
				Name:       name + "." + domain + ".",
				Type:       "TXT",
				Changetype: "DELETE",
				TTL:        60,
				Records:    []Record{},
			},
		},
	}

	entries, err := json.Marshal(rrsetRequest)
	fmt.Println("Request that is being sent to PowerDNS: ", string(entries))

	requestBody := bytes.NewBuffer(entries)

	// Create new record by patching zone
	req, err := http.NewRequest("PATCH", "https://dns.nextlayer.at/api/v1/servers/localhost/zones/"+zone, requestBody)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-API-Key", cfg.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failure : ", err)
	}

	// Display Results
	fmt.Println("response Status : ", resp.Status)
	fmt.Println("response Headers : ", resp.Header)
	respBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body : ", string(respBody))

	return nil
}

// Initialize will be called when the webhook first starts.
// This method can be used to instantiate the webhook, i.e. initialising
// connections or warming up caches.
// Typically, the kubeClientConfig parameter is used to build a Kubernetes
// client that can be used to fetch resources from the Kubernetes API, e.g.
// Secret resources containing credentials used to authenticate with DNS
// provider accounts.
// The stopCh can be used to handle early termination of the webhook, in cases
// where a SIGTERM or similar signal is sent to the webhook process.
func (c *nextlayerDNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	///// UNCOMMENT THE BELOW CODE TO MAKE A KUBERNETES CLIENTSET AVAILABLE TO
	///// YOUR CUSTOM DNS PROVIDER

	//cl, err := kubernetes.NewForConfig(kubeClientConfig)
	//if err != nil {
	//	return err
	//}
	//
	//c.client = cl

	///// END OF CODE TO MAKE KUBERNETES CLIENTSET AVAILABLE
	return nil
}

// loadConfig is a small helper function that decodes JSON configuration into
// the typed config struct.
func loadConfig(cfgJSON *extapi.JSON) (nextlayerDNSProviderConfig, error) {

	cfg := nextlayerDNSProviderConfig{}
	// handle the 'base case' where no configuration has been provided
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}

type tld map[string]*tld

func (c *nextlayerDNSProviderSolver) getDomainAndEntry(ch *v1alpha1.ChallengeRequest) (string, string, string) {
	entry := strings.TrimSuffix(ch.ResolvedFQDN, ch.ResolvedZone)
	entry = strings.TrimSuffix(entry, ".")
	domain := strings.TrimSuffix(ch.ResolvedZone, ".")

	zone, err := publicsuffix.EffectiveTLDPlusOne(entry + "." + domain)
	if err != nil {
		fmt.Println("Failure : ", err)
	}

	return entry, domain, zone
}
