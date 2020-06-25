package azuremonitor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/require"
	"github.com/xorcare/pointer"
)

func TestInsightsMetricsToFrame(t *testing.T) {
	tests := []struct {
		name          string
		testFile      string
		metric        string
		agg           string
		dimensions    []string
		expectedFrame func() *data.Frame
	}{
		{
			name:     "single series",
			testFile: "applicationinsights/4-application-insights-response-metrics-no-segment.json",
			metric:   "value",
			agg:      "avg",
			expectedFrame: func() *data.Frame {
				frame := data.NewFrame("",
					data.NewField("StartTime", nil, []time.Time{
						time.Date(2019, 9, 13, 1, 2, 3, 456789000, time.UTC),
						time.Date(2019, 9, 13, 2, 2, 3, 456789000, time.UTC),
					}),
					data.NewField("value", nil, []*float64{
						pointer.Float64(1),
						pointer.Float64(2),
					}),
				)
				return frame
			},
		},
		{
			name:       "segmented series",
			testFile:   "applicationinsights/4-application-insights-response-metrics-segmented.json",
			metric:     "value",
			agg:        "avg",
			dimensions: []string{"blob"},
			expectedFrame: func() *data.Frame {
				frame := data.NewFrame("",
					data.NewField("StartTime", nil, []time.Time{
						time.Date(2019, 9, 13, 1, 2, 3, 456789000, time.UTC),
						time.Date(2019, 9, 13, 2, 2, 3, 456789000, time.UTC),
					}),
					data.NewField("value", data.Labels{"blob": "a"}, []*float64{
						pointer.Float64(1),
						pointer.Float64(2),
					}),
					data.NewField("value", data.Labels{"blob": "b"}, []*float64{
						pointer.Float64(3),
						pointer.Float64(4),
					}),
				)
				return frame
			},
		},
		// {
		// 	name:       "segmented series",
		// 	testFile:   "applicationinsights/4-application-insights-response-metrics-multi-segmented.json",
		// 	metric:     "traces/count",
		// 	agg:        "sum",
		// 	dimensions: []string{"client/countryOrRegion", "client/city"},
		// 	expectedFrame: func() *data.Frame {
		// 		frame := data.NewFrame("") // data.NewField("StartTime", nil, []time.Time{
		// 		// 	time.Date(2019, 9, 13, 1, 2, 3, 456789000, time.UTC),
		// 		// 	time.Date(2019, 9, 13, 2, 2, 3, 456789000, time.UTC),
		// 		// }),
		// 		// data.NewField("value", data.Labels{"blob": "a"}, []*float64{
		// 		// 	pointer.Float64(1),
		// 		// 	pointer.Float64(2),
		// 		// }),
		// 		// data.NewField("value", data.Labels{"blob": "b"}, []*float64{
		// 		// 	pointer.Float64(3),
		// 		// 	pointer.Float64(4),
		// 		// }),

		// 		return frame
		// 	},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := loadInsightsMetricsResponse(tt.testFile)
			require.NoError(t, err)
			t.Log(err)

			frame, err := res.ToFrame(tt.metric, tt.agg, tt.dimensions)
			require.NoError(t, err)
			//t.Log(spew.Sdump(res))
			t.Log(frame.StringTable(-1, -1))
			if diff := cmp.Diff(tt.expectedFrame(), frame, data.FrameTestCompareOptions()...); diff != "" {
				t.Errorf("Result mismatch (-want +got):\n%s", diff)
			}

		})
	}

}

func loadInsightsMetricsResponse(name string) (MetricsResult, error) {
	var mr MetricsResult

	path := filepath.Join("testdata", name)
	f, err := os.Open(path)
	if err != nil {
		return mr, err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	err = d.Decode(&mr)
	return mr, err
}
