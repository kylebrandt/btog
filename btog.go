package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func main() {
	metricRoot := "linux.net.stat.tcp."
	baseURL := "http://bosun"
	perRow := 4

	metrics, err := GetMetrics(baseURL, metricRoot)
	if err != nil {
		log.Fatal(err)
	}
	filteredMetrics := metrics.MetricsStartsWith(metricRoot)
	gd := GrafanaDashBoard{}
    gd.Title = "Gen Dashboard"
    gd.Style = "dark"
    gd.Timezone = "browser"
    gd.Editable = true
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
			Expr: fmt.Sprintf(`q("sum:$ds-avg:rate{counter,,1}:%s", "$start", "")`, m),
            RefId: "A",
		}
        panel := NewPanel()
		panel.Title = m
        panel.Datasource = "Bosun"
        panel.Targets = []Target{t}
        ii := i
        panel.ID = ii
		panels = append(panels, panel)
	}
	b, err := json.MarshalIndent(&gd, "", "\t")
	if err != nil {
		log.Fatal("failed to unmarshal dashbaord")
	}
	fmt.Println(string(b))
}

type Metrics []string

func (metrics Metrics) MetricsStartsWith(prefix string) []string {
	filteredMetrics := []string{}
	for _, m := range metrics {
		if strings.HasPrefix(m, prefix) {
			filteredMetrics = append(filteredMetrics, m)
		}
	}
	return filteredMetrics
}

func GetMetrics(baseURL, metricRoot string) (Metrics, error) {
	var metrics Metrics
	u := fmt.Sprintf("%s/api/metric", baseURL)
	log.Println(u)
	res, err := http.Get(u)
	if err != nil {
		return metrics, fmt.Errorf("failed to get metrics")
	}
	defer res.Body.Close()
	d := json.NewDecoder(res.Body)
	err = d.Decode(&metrics)
	if err != nil {
		return metrics, fmt.Errorf("unable to decode metric response")
	}
	return metrics, nil
}

type GrafanaDashBoard struct {
	// Annotations struct {
	// 	List []interface{} `json:"list"`
	// } `json:"annotations"`
	Editable        bool          `json:"editable"`
	HideControls    bool          `json:"hideControls"`
	ID              interface{}   `json:"id"`
	//Links           []interface{} `json:"links"`
	//OriginalTitle   string        `json:"originalTitle"`
	Rows            []Row         `json:"rows"`
	SchemaVersion   int       `json:"schemaVersion"`
	SharedCrosshair bool          `json:"sharedCrosshair"`
	Style           string        `json:"style"`
	Tags            []interface{} `json:"tags"`
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
	Timezone string  `json:"timezone"`
	Title    string  `json:"title"`
	Version  int `json:"version"`
}

type Row struct {
	Collapse bool    `json:"collapse"`
	Editable bool    `json:"editable"`
	//Height   string  `json:"height"`
	Panels   []Panel `json:"panels"`
	Title    string  `json:"title"`
}

type Panel struct {
	//AliasColors struct{} `json:"aliasColors"`
	Bars        bool     `json:"bars"`
	Datasource  string   `json:"datasource"`
	Editable    bool     `json:"editable"`
	Error       bool     `json:"error"`
	Fill        float64  `json:"fill"`
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
	ID              int       `json:"id"`
	IsNew           bool          `json:"isNew"`
	Legend          Legend        `json:"legend"`
	Lines           bool          `json:"lines"`
	Linewidth       float64       `json:"linewidth"`
	NullPointMode   string        `json:"nullPointMode"`
	Percentage      bool          `json:"percentage"`
	Pointradius     float64       `json:"pointradius"`
	Points          bool          `json:"points"`
	Renderer        string        `json:"renderer"`
	//SeriesOverrides []interface{} `json:"seriesOverrides"`
	Span            int       `json:"span"`
	Stack           bool          `json:"stack"`
	SteppedLine     bool          `json:"steppedLine"`
	Targets         []Target      `json:"targets"`
	//TimeFrom        interface{}   `json:"timeFrom"`
	//TimeShift       interface{}   `json:"timeShift"`
	Title           string        `json:"title"`
	//Tooltip         struct {
	//	Shared    bool   `json:"shared"`
	//	ValueType string `json:"value_type"`
	//} `json:"tooltip"`
	Type     string   `json:"type"`
	X_Axis   bool     `json:"x-axis"`
	Y_Axis   bool     `json:"y-axis"`
	YFormats []string `json:"y_formats"`
}

func NewPanel() Panel {
	return Panel{
		Renderer:  "flot",
        Type: "graph",
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