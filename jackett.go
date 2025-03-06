package jackett

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"golang.org/x/net/context"
)

var (
	apiURL string
	apiKey string
)

type Settings struct {
	ApiURL string
	ApiKey string
	Client *http.Client
}

type FetchRequest struct {
	Query      string
	Trackers   []string
	Categories []uint
}

type FetchResponse struct {
	Results  []Result
	Indexers []Indexer
}

type jackettTime struct {
	time.Time
}

func (jt *jackettTime) UnmarshalJSON(b []byte) (err error) {
	str := strings.Trim(string(b), `"`)
	if str == "0001-01-01T00:00:00" {
	} else if len(str) == 19 {
		jt.Time, err = time.Parse(time.RFC3339, str+"Z")
	} else {
		jt.Time, err = time.Parse(time.RFC3339, str)
	}
	return
}

type Result struct {
	BannerUrl            string
	BlackholeLink        string
	Category             []uint
	CategoryDesc         string
	Comments             string
	Description          string
	DownloadVolumeFactor float32
	Files                uint
	FirstSeen            jackettTime
	Gain                 float32
	Grabs                uint
	Guid                 string
	Imdb                 uint
	InfoHash             string
	Link                 string
	MagnetUri            string
	MinimumRatio         float32
	MinimumSeedTime      uint
	Peers                uint
	PublishDate          jackettTime
	RageID               uint
	Seeders              uint
	Size                 uint
	TMDb                 uint
	TVDBId               uint
	Title                string
	Tracker              string
	TrackerId            string
	UploadVolumeFactor   float32
}

type Indexer struct {
	Error   string
	ID      string
	Name    string
	Results uint
	Status  uint
}

type Jackett struct {
	settings *Settings
}

func NewJackett(s *Settings) *Jackett {
	if s.ApiURL == "" && apiURL != "" {
		s.ApiURL = apiURL
	}
	if s.ApiKey == "" && apiKey != "" {
		s.ApiKey = apiKey
	}
	if s.Client == nil {
		s.Client = http.DefaultClient
	}
	return &Jackett{settings: s}
}

func (j *Jackett) generateFetchURL(fr *FetchRequest) (string, error) {
	u, err := url.Parse(j.settings.ApiURL)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse apiURL %q", j.settings.ApiURL)
	}
	u.Path = "/api/v2.0/indexers/all/results"
	q := u.Query()
	q.Set("apikey", j.settings.ApiKey)
	for _, t := range fr.Trackers {
		q.Add("Tracker[]", t)
	}
	for _, c := range fr.Categories {
		q.Add("Category[]", fmt.Sprintf("%v", c))
	}
	if fr.Query != "" {
		q.Add("Query", fr.Query)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (j *Jackett) Fetch(ctx context.Context, fr *FetchRequest) (*FetchResponse, error) {
	u, err := j.generateFetchURL(fr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate fetch url")
	}
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make fetch request")
	}
	res, err := j.settings.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to invoke fetch request")
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read fetch data")
	}
	var fres FetchResponse
	err = json.Unmarshal(data, &fres)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal fetch data with url=%v and data=%v", u, string(data))
	}
	return &fres, nil
}

func (j *Jackett) AddIndexer(ctx context.Context, indexerID string) error {
	u, err := url.Parse(j.settings.ApiURL)
	if err != nil {
		return errors.Wrapf(err, "failed to parse apiURL %q", j.settings.ApiURL)
	}
	u.Path = fmt.Sprintf("/api/v2.0/indexers/%s", indexerID)
	q := u.Query()
	q.Set("apikey", j.settings.ApiKey)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), nil)
	if err != nil {
		return errors.Wrap(err, "failed to make add indexer request")
	}
	res, err := j.settings.Client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to invoke add indexer request")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add indexer, status code: %d", res.StatusCode)
	}
	return nil
}

func (j *Jackett) RemoveIndexer(ctx context.Context, indexerID string) error {
	u, err := url.Parse(j.settings.ApiURL)
	if err != nil {
		return errors.Wrapf(err, "failed to parse apiURL %q", j.settings.ApiURL)
	}
	u.Path = fmt.Sprintf("/api/v2.0/indexers/%s", indexerID)
	q := u.Query()
	q.Set("apikey", j.settings.ApiKey)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "DELETE", u.String(), nil)
	if err != nil {
		return errors.Wrap(err, "failed to make remove indexer request")
	}
	res, err := j.settings.Client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to invoke remove indexer request")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to remove indexer, status code: %d", res.StatusCode)
	}
	return nil
}

func (j *Jackett) UpdateIndexer(ctx context.Context, indexerID string, settings map[string]interface{}) error {
	u, err := url.Parse(j.settings.ApiURL)
	if err != nil {
		return errors.Wrapf(err, "failed to parse apiURL %q", j.settings.ApiURL)
	}
	u.Path = fmt.Sprintf("/api/v2.0/indexers/%s", indexerID)
	q := u.Query()
	q.Set("apikey", j.settings.ApiKey)
	u.RawQuery = q.Encode()

	body, err := json.Marshal(settings)
	if err != nil {
		return errors.Wrap(err, "failed to marshal indexer settings")
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", u.String(), strings.NewReader(string(body)))
	if err != nil {
		return errors.Wrap(err, "failed to make update indexer request")
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := j.settings.Client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to invoke update indexer request")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update indexer, status code: %d", res.StatusCode)
	}
	return nil
}

func init() {
	if v, ok := os.LookupEnv("JACKETT_API_URL"); ok {
		apiURL = v
	}
	if v, ok := os.LookupEnv("JACKETT_API_KEY"); ok {
		apiKey = v
	}
}
