package cloudstatus

import (
	"compress/gzip"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed fixtures/*.json.gz
var fixturesFS embed.FS

// Service represents a cloud status service entry.
type Service struct {
	Name  string `json:"name"`
	Group string `json:"group"`
}

// Index holds the parsed cloud status service data for searching.
type Index struct {
	providers []string
	services  []Service
}

// NewIndex loads and parses all embedded fixture files into a searchable index.
func NewIndex() (*Index, error) {
	entries, err := fs.ReadDir(fixturesFS, "fixtures")
	if err != nil {
		return nil, fmt.Errorf("reading fixtures dir: %w", err)
	}

	seen := make(map[string]struct{})
	var services []Service

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		svcs, err := loadFixture(fixturesFS, "fixtures/"+entry.Name())
		if err != nil {
			return nil, fmt.Errorf("loading %s: %w", entry.Name(), err)
		}
		for _, s := range svcs {
			if s.Name == "" || s.Group == "" {
				continue
			}
			key := s.Group + "\x00" + s.Name
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			services = append(services, s)
		}
	}

	sort.Slice(services, func(i, j int) bool {
		if services[i].Group != services[j].Group {
			return services[i].Group < services[j].Group
		}
		return services[i].Name < services[j].Name
	})

	providerSet := make(map[string]struct{})
	for _, s := range services {
		providerSet[s.Group] = struct{}{}
	}
	providers := make([]string, 0, len(providerSet))
	for p := range providerSet {
		providers = append(providers, p)
	}
	sort.Strings(providers)

	return &Index{
		providers: providers,
		services:  services,
	}, nil
}

func loadFixture(fsys embed.FS, path string) ([]Service, error) {
	f, err := fsys.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer gr.Close()

	var entries []Service
	if err := json.NewDecoder(gr).Decode(&entries); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}
	return entries, nil
}

// Providers returns all unique provider (group) names, sorted alphabetically.
func (idx *Index) Providers() []string {
	return idx.providers
}

// SearchResult holds a paginated search result.
type SearchResult struct {
	Services   []Service
	TotalCount int
}

// Search finds services matching the given criteria.
// Provider filters by exact group name. Query does case-insensitive substring match on service name.
// If both provider and query are empty, all services are returned (subject to pagination).
func (idx *Index) Search(provider, query string, page, pageSize int) SearchResult {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}

	queryLower := strings.ToLower(query)

	var matched []Service
	for _, s := range idx.services {
		if provider != "" && s.Group != provider {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(s.Name), queryLower) {
			continue
		}
		matched = append(matched, s)
	}

	total := len(matched)
	start := (page - 1) * pageSize
	if start >= total {
		return SearchResult{TotalCount: total}
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return SearchResult{
		Services:   matched[start:end],
		TotalCount: total,
	}
}
