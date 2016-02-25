package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"bosun.org/cmd/bosun/database"
)

func RateQueryString(m MetricMetaTagKeys) string {
	if m.Rate == "counter" {
		return "rate{counter,,1}:"
	}
	return ""
}

var (
	flagBaseURL      = flag.String("b", "http://bosun", "bosun root url")
	flagMetricRoot   = flag.String("m", "haproxy.server.", "get all metrics that start with this string")
	flagPerRow       = flag.Int("p", 6, "number of graph panels per row")
	flagTemplateVars = flag.String("t", "", "csv of template vars with an initial value, i.e. host=foo,group=baz. Would be referenced as $host and $group in query")
	flagQuery        = flag.String("q", `q("sum:$ds-avg:%s%s{}{host=ny-lb05|ny-lb06}", "$start", "")`, "First %s is metric, second is counter string")
)

func main() {
	flag.Parse()
	metricRoot := *flagMetricRoot
	baseURL := *flagBaseURL
	perRow := *flagPerRow
	query := *flagQuery
	templates := []Template{}
	splitTemplateVars := strings.Split(*flagTemplateVars, ",")
	if splitTemplateVars[0] != "" {
		for _, entry := range splitTemplateVars {
			kv := strings.Split(entry, "=")
			if len(kv) != 2 {
				log.Fatal("Template vars must have an initial value")
			}
			templates = append(templates, NewTemplate(kv[0], kv[1]))
		}
	}
	metrics, err := GetMetadataMetrics(baseURL)
	if err != nil {
		log.Fatal(err)
	}
	filteredMetrics := metrics.MetricsStartsWith(metricRoot)
	gd := GrafanaDashBoard{}
	if len(templates) != 0 {
		gd.Templating.List = templates
	}
	gd.Title = "Gen Dashboard"
	gd.Style = "dark"
	gd.Timezone = "utc"
	gd.Editable = true

	span := 12 / perRow

	var row Row
	var panels []Panel
	for i, m := range filteredMetrics {
		if i != 0 && i%perRow == 0 {
			row.Panels = panels
			gd.Rows = append(gd.Rows, row)
			log.Printf("Appending row with %v panels", len(panels))
			row = Row{}
			panels = []Panel{}
		}
		t := Target{
			Expr:  fmt.Sprintf(query, RateQueryString(m), m.Metric),
			RefId: "A",
		}
		panel := NewPanel()
		panel.Title = fmt.Sprintf("%s", m.Metric)
		panel.Datasource = "Bosun"
		panel.Targets = []Target{t}
		panel.Span = span
		ii := i
		panel.ID = ii
		panel.LeftYAxisLabel = m.Unit
		if m.Desc != "" {
			panel.Links = append(panel.Links, Link{Type: "Absolute", Title: m.Desc})
		}
		panels = append(panels, panel)
	}
	b, err := json.MarshalIndent(&gd, "", "\t")
	if err != nil {
		log.Fatal("failed to unmarshal dashbaord")
	}
	fmt.Println(string(b))
}

type Metrics []MetricMetaTagKeys

func (metrics Metrics) MetricsStartsWith(prefix string) Metrics {
	filteredMetrics := Metrics{}
	for _, m := range metrics {
		if strings.HasPrefix(m.Metric, prefix) {
			if m.MetricMetadata == nil {
				log.Printf("No metadata for %s, skipping", m.Metric)
				continue
			}
			filteredMetrics = append(filteredMetrics, m)
		}
	}
	return filteredMetrics
}

type MetricMetaTagKeys struct {
	Metric string
	*database.MetricMetadata
	TagKeys []string
}

func GetMetadataMetrics(baseURL string) (Metrics, error) {
	var metrics Metrics
	u := fmt.Sprintf("%s/api/metadata/metrics", baseURL)
	res, err := http.Get(u)
	if err != nil {
		return metrics, fmt.Errorf("failed to get metrics")
	}
	defer res.Body.Close()
	var mm map[string]MetricMetaTagKeys
	d := json.NewDecoder(res.Body)
	err = d.Decode(&mm)
	if err != nil {
		return metrics, fmt.Errorf("unable to decode metric response")
	}
	for k, v := range mm {
		v.Metric = k
		metrics = append(metrics, v)
	}
	return metrics, nil
}

type GrafanaDashBoard struct {
	// Annotations struct {
	// 	List []interface{} `json:"list"`
	// } `json:"annotations"`
	Editable     bool        `json:"editable"`
	HideControls bool        `json:"hideControls"`
	ID           interface{} `json:"id"`
	//OriginalTitle   string        `json:"originalTitle"`
	Rows            []Row         `json:"rows"`
	SchemaVersion   int           `json:"schemaVersion"`
	SharedCrosshair bool          `json:"sharedCrosshair"`
	Style           string        `json:"style"`
	Tags            []interface{} `json:"tags"`
	Templating      struct {
		List []Template `json:"list"`
	} `json:"templating"`
	// Templating      struct {
	// 	List []interface{} `json:"list"`
	// } `json:"templating"`
	// Time struct {
	// 	From string `json:"from"`
	// 	To   string `json:"to"`
	// } `json:"time"`
	// Timepicker struct {
	// 	RefreshIntervals []string `json:"refresh_intervals"`
	// 	TimeOptions      []string `json:"time_options"`
	// } `json:"timepicker"`
	Timezone string `json:"timezone"`
	Title    string `json:"title"`
	Version  int    `json:"version"`
}

type Row struct {
	Collapse bool `json:"collapse"`
	Editable bool `json:"editable"`
	//Height   string  `json:"height"`
	Panels []Panel `json:"panels"`
	Title  string  `json:"title"`
}

type Panel struct {
	//AliasColors struct{} `json:"aliasColors"`
	Bars       bool    `json:"bars"`
	Datasource string  `json:"datasource"`
	Editable   bool    `json:"editable"`
	Error      bool    `json:"error"`
	Fill       float64 `json:"fill"`
	// Grid        struct {
	// 	LeftLogBase     int     `json:"leftLogBase"`
	// 	LeftMax         interface{} `json:"leftMax"`
	// 	LeftMin         interface{} `json:"leftMin"`
	// 	RightLogBase    int     `json:"rightLogBase"`
	// 	RightMax        interface{} `json:"rightMax"`
	// 	RightMin        interface{} `json:"rightMin"`
	// 	Threshold1      interface{} `json:"threshold1"`
	// 	Threshold1Color string      `json:"threshold1Color"`
	// 	Threshold2      interface{} `json:"threshold2"`
	// 	Threshold2Color string      `json:"threshold2Color"`
	// } `json:"grid"`
	ID            int     `json:"id"`
	IsNew         bool    `json:"isNew"`
	Legend        Legend  `json:"legend"`
	Lines         bool    `json:"lines"`
	Linewidth     float64 `json:"linewidth"`
	Links         []Link  `json:"links"`
	NullPointMode string  `json:"nullPointMode"`
	Percentage    bool    `json:"percentage"`
	Pointradius   float64 `json:"pointradius"`
	Points        bool    `json:"points"`
	Renderer      string  `json:"renderer"`
	//SeriesOverrides []interface{} `json:"seriesOverrides"`
	Span        int      `json:"span"`
	Stack       bool     `json:"stack"`
	SteppedLine bool     `json:"steppedLine"`
	Targets     []Target `json:"targets"`
	//TimeFrom        interface{}   `json:"timeFrom"`
	//TimeShift       interface{}   `json:"timeShift"`
	Title string `json:"title"`
	//Tooltip         struct {
	//	Shared    bool   `json:"shared"`
	//	ValueType string `json:"value_type"`
	//} `json:"tooltip"`
	Type            string   `json:"type"`
	X_Axis          bool     `json:"x-axis"`
	Y_Axis          bool     `json:"y-axis"`
	YFormats        []string `json:"y_formats"`
	LeftYAxisLabel  string   `json:"leftYAxisLabel"`
	RightYAxisLabel string   `json:"rightYAxisLabel"`
}

func NewPanel() Panel {
	return Panel{
		Renderer:  "flot",
		Type:      "graph",
		X_Axis:    true,
		Y_Axis:    true,
		YFormats:  []string{"short", "short"},
		Lines:     true,
		Linewidth: 2,
		Legend:    Legend{Show: true},
	}
}

type Target struct {
	Aggregator           string   `json:"aggregator"`
	DownsampleAggregator string   `json:"downsampleAggregator"`
	Errors               struct{} `json:"errors"`
	Expr                 string   `json:"expr"`
	RefId                string   `json:"refId"`
}

type Legend struct {
	Avg     bool `json:"avg"`
	Current bool `json:"current"`
	Max     bool `json:"max"`
	Min     bool `json:"min"`
	Show    bool `json:"show"`
	Total   bool `json:"total"`
	Values  bool `json:"values"`
}

type Link struct {
	Type  string `json:"type"`
	Title string `json:"title"`
}

type Template struct {
	AllFormat string  `json:"allFormat"`
	Current   Current `json:"current"`
	//Datasource  interface{} `json:"datasource"`
	IncludeAll  bool             `json:"includeAll"`
	Multi       bool             `json:"multi"`
	MultiFormat string           `json:"multiFormat"`
	Name        string           `json:"name"`
	Options     []TemplateOption `json:"options"`
	Query       string           `json:"query"`
	Refresh     bool             `json:"refresh"`
	//Regex       string           `json:"regex"`
	Type string `json:"type"`
}

type Current struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

type TemplateOption struct {
	Selected bool   `json:"selected"`
	Text     string `json:"text"`
	Value    string `json:"value"`
}

func NewTemplate(name, value string) Template {
	return Template{
		AllFormat:   "glob",
		MultiFormat: "glob",
		Current:     Current{value, value},
		Name:        name,
		Options: []TemplateOption{
			TemplateOption{
				Selected: true,
				Text:     value,
				Value:    value,
			},
		},
		Query: value,
		Type:  "custom",
	}
}
