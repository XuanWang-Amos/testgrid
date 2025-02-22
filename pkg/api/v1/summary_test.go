/*
Copyright 2023 The TestGrid Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	apipb "github.com/GoogleCloudPlatform/testgrid/pb/api/v1"
	configpb "github.com/GoogleCloudPlatform/testgrid/pb/config"
	summarypb "github.com/GoogleCloudPlatform/testgrid/pb/summary"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestListTabSummaries(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]*configpb.Configuration
		summaries   map[string]*summarypb.DashboardSummary
		req         *apipb.ListTabSummariesRequest
		want        *apipb.ListTabSummariesResponse
		expectError bool
	}{
		{
			name: "Returns an error when there's no dashboard in config",
			config: map[string]*configpb.Configuration{
				"gs://default/config": {},
			},
			req: &apipb.ListTabSummariesRequest{
				Dashboard: "missing",
			},
			expectError: true,
		},
		{
			name: "Returns an error when there's no summary for dashboard yet",
			config: map[string]*configpb.Configuration{
				"gs://default/config": {
					Dashboards: []*configpb.Dashboard{
						{
							Name: "ACME",
							DashboardTab: []*configpb.DashboardTab{
								{
									Name:          "me-me",
									TestGroupName: "testgroupname",
								},
							},
						},
					},
				},
			},
			req: &apipb.ListTabSummariesRequest{
				Dashboard: "acme",
			},
			expectError: true,
		},

		{
			name: "Returns correct tab summaries for a dashboard",
			config: map[string]*configpb.Configuration{
				"gs://default/config": {
					Dashboards: []*configpb.Dashboard{
						{
							Name: "Marco",
							DashboardTab: []*configpb.DashboardTab{
								{
									Name:          "polo-1",
									TestGroupName: "cheesecake",
								},
								{
									Name:          "polo-2",
									TestGroupName: "tiramisu",
								},
								{
									Name:          "polo-3",
									TestGroupName: "donut",
								},
								{
									Name:          "polo-4",
									TestGroupName: "brownie",
								},
							},
						},
					},
				},
			},
			summaries: map[string]*summarypb.DashboardSummary{
				"gs://default/summary/summary-marco": {
					TabSummaries: []*summarypb.DashboardTabSummary{
						{
							DashboardName:       "Marco",
							DashboardTabName:    "polo-1",
							Status:              "1/7 tests are passing!",
							OverallStatus:       summarypb.DashboardTabSummary_FLAKY,
							LatestGreen:         "Hulk",
							LastUpdateTimestamp: float64(915166800.916166782),
							LastRunTimestamp:    float64(915166800.916166782),
						},
						{
							DashboardName:       "Marco",
							DashboardTabName:    "polo-2",
							Status:              "1/7 tests are passing!",
							OverallStatus:       summarypb.DashboardTabSummary_ACCEPTABLE,
							LatestGreen:         "Lantern",
							LastUpdateTimestamp: float64(0.1),
							LastRunTimestamp:    float64(0.1),
						},
						{
							DashboardName:       "Marco",
							DashboardTabName:    "polo-3",
							Status:              "1/7 tests are passing!",
							OverallStatus:       summarypb.DashboardTabSummary_ACCEPTABLE,
							LatestGreen:         "Hulk",
							LastUpdateTimestamp: float64(916166800),
							LastRunTimestamp:    float64(916166800),
						},
						{
							DashboardName:       "Marco",
							DashboardTabName:    "polo-4",
							Status:              "1/7 tests are failing!",
							OverallStatus:       summarypb.DashboardTabSummary_ACCEPTABLE,
							LatestGreen:         "Lantern",
							LastUpdateTimestamp: float64(0.916166782),
							LastRunTimestamp:    float64(0.916166782),
						},
					},
				},
			},
			req: &apipb.ListTabSummariesRequest{
				Dashboard: "marco",
			},
			want: &apipb.ListTabSummariesResponse{
				TabSummaries: []*apipb.TabSummary{
					{
						DashboardName:         "Marco",
						TabName:               "polo-1",
						DetailedStatusMessage: "1/7 tests are passing!",
						OverallStatus:         "FLAKY",
						LatestPassingBuild:    "Hulk",
						LastRunTimestamp: &timestamp.Timestamp{
							Seconds: 915166800,
							Nanos:   916166782,
						},
						LastUpdateTimestamp: &timestamp.Timestamp{
							Seconds: 915166800,
							Nanos:   916166782,
						},
					},
					{
						DashboardName:         "Marco",
						TabName:               "polo-2",
						DetailedStatusMessage: "1/7 tests are passing!",
						OverallStatus:         "ACCEPTABLE",
						LatestPassingBuild:    "Lantern",
						LastRunTimestamp: &timestamp.Timestamp{
							Nanos: 100000000,
						},
						LastUpdateTimestamp: &timestamp.Timestamp{
							Nanos: 100000000,
						},
					},
					{
						DashboardName:         "Marco",
						TabName:               "polo-3",
						DetailedStatusMessage: "1/7 tests are passing!",
						OverallStatus:         "ACCEPTABLE",
						LatestPassingBuild:    "Hulk",
						LastRunTimestamp: &timestamp.Timestamp{
							Seconds: 916166800,
						},
						LastUpdateTimestamp: &timestamp.Timestamp{
							Seconds: 916166800,
						},
					},
					{
						DashboardName:         "Marco",
						TabName:               "polo-4",
						DetailedStatusMessage: "1/7 tests are failing!",
						OverallStatus:         "ACCEPTABLE",
						LatestPassingBuild:    "Lantern",
						LastRunTimestamp: &timestamp.Timestamp{
							Nanos: 916166782,
						},
						LastUpdateTimestamp: &timestamp.Timestamp{
							Nanos: 916166782,
						},
					},
				},
			},
		},
		{
			name: "Server error with unreadable config",
			config: map[string]*configpb.Configuration{
				"gs://welp/config": {},
			},
			req: &apipb.ListTabSummariesRequest{
				Dashboard: "doesntmatter",
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := setupTestServer(t, tc.config, nil, tc.summaries)
			got, err := server.ListTabSummaries(context.Background(), tc.req)
			switch {
			case err != nil:
				if !tc.expectError {
					t.Errorf("got unexpected error: %v", err)
				}
			case tc.expectError:
				t.Error("failed to receive an error")
			default:
				if diff := cmp.Diff(tc.want, got, protocmp.Transform()); diff != "" {
					t.Errorf("got unexpected diff (-want +got):\n%s", diff)
				}
			}
		})

	}

}

func GetTabSummary(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]*configpb.Configuration
		summaries   map[string]*summarypb.DashboardSummary
		req         *apipb.GetTabSummaryRequest
		want        *apipb.GetTabSummaryResponse
		expectError bool
	}{
		{
			name: "Returns an error when there's no dashboard in config",
			config: map[string]*configpb.Configuration{
				"gs://default/config": {},
			},
			req: &apipb.GetTabSummaryRequest{
				Dashboard: "missing",
				Tab:       "Carpe Noctem",
			},
			expectError: true,
		},
		{
			name: "Returns an error when there's no tab in config",
			config: map[string]*configpb.Configuration{
				"gs://default/config": {
					Dashboards: []*configpb.Dashboard{
						{
							Name: "Aurora",
							DashboardTab: []*configpb.DashboardTab{
								{
									Name: "Borealis",
								},
							},
						},
					},
				},
			},
			req: &apipb.GetTabSummaryRequest{
				Dashboard: "Aurora",
				Tab:       "Noctem",
			},
			expectError: true,
		},
		{
			name: "Returns an error when there's no summary for dashboard yet",
			config: map[string]*configpb.Configuration{
				"gs://default/config": {
					Dashboards: []*configpb.Dashboard{
						{
							Name: "ACME",
							DashboardTab: []*configpb.DashboardTab{
								{
									Name:          "me-me",
									TestGroupName: "testgroupname",
								},
							},
						},
					},
				},
			},
			req: &apipb.GetTabSummaryRequest{
				Dashboard: "acme",
				Tab:       "me-me",
			},
			expectError: true,
		},
		{
			name: "Returns an error when there's no summary for a tab",
			config: map[string]*configpb.Configuration{
				"gs://default/config": {
					Dashboards: []*configpb.Dashboard{
						{
							Name: "Marco",
							DashboardTab: []*configpb.DashboardTab{
								{
									Name:          "polo-1",
									TestGroupName: "cheesecake",
								},
								{
									Name:          "polo-2",
									TestGroupName: "tiramisu",
								},
							},
						},
					},
				},
			},
			summaries: map[string]*summarypb.DashboardSummary{
				"gs://default/summary/summary-marco": {
					TabSummaries: []*summarypb.DashboardTabSummary{
						{
							DashboardName:       "Marco",
							DashboardTabName:    "polo-1",
							Status:              "1/7 tests are passing!",
							OverallStatus:       summarypb.DashboardTabSummary_FLAKY,
							LatestGreen:         "Hulk",
							LastUpdateTimestamp: float64(915166800),
							LastRunTimestamp:    float64(915166800),
						},
					},
				},
			},
			req: &apipb.GetTabSummaryRequest{
				Dashboard: "marco",
				Tab:       "polo-2",
			},
			expectError: true,
		},
		{
			name: "Returns correct tab summary for a dashboard-tab",
			config: map[string]*configpb.Configuration{
				"gs://default/config": {
					Dashboards: []*configpb.Dashboard{
						{
							Name: "Marco",
							DashboardTab: []*configpb.DashboardTab{
								{
									Name:          "polo-1",
									TestGroupName: "cheesecake",
								},
								{
									Name:          "polo-2",
									TestGroupName: "tiramisu",
								},
							},
						},
					},
				},
			},
			summaries: map[string]*summarypb.DashboardSummary{
				"gs://default/summary/summary-marco": {
					TabSummaries: []*summarypb.DashboardTabSummary{
						{
							DashboardName:       "Marco",
							DashboardTabName:    "polo-1",
							Status:              "1/7 tests are passing!",
							OverallStatus:       summarypb.DashboardTabSummary_FLAKY,
							LatestGreen:         "Hulk",
							LastUpdateTimestamp: float64(915166800),
							LastRunTimestamp:    float64(915166800),
						},
						{
							DashboardName:       "Marco",
							DashboardTabName:    "polo-2",
							Status:              "1/7 tests are failing!",
							OverallStatus:       summarypb.DashboardTabSummary_ACCEPTABLE,
							LatestGreen:         "Lantern",
							LastUpdateTimestamp: float64(916166800),
							LastRunTimestamp:    float64(916166800),
						},
					},
				},
			},
			req: &apipb.GetTabSummaryRequest{
				Dashboard: "marco",
				Tab:       "polo-1",
			},
			want: &apipb.GetTabSummaryResponse{
				TabSummary: &apipb.TabSummary{
					DashboardName:         "Marco",
					TabName:               "polo-1",
					DetailedStatusMessage: "1/7 tests are passing!",
					OverallStatus:         "FLAKY",
					LatestPassingBuild:    "Hulk",
					LastRunTimestamp: &timestamp.Timestamp{
						Seconds: 915166800,
					},
					LastUpdateTimestamp: &timestamp.Timestamp{
						Seconds: 915166800,
					},
				},
			},
		},
		{
			name: "Server error with unreadable config",
			config: map[string]*configpb.Configuration{
				"gs://welp/config": {},
			},
			req: &apipb.GetTabSummaryRequest{
				Dashboard: "non refert",
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := setupTestServer(t, tc.config, nil, tc.summaries)
			got, err := server.GetTabSummary(context.Background(), tc.req)
			switch {
			case err != nil:
				if !tc.expectError {
					t.Errorf("got unexpected error: %v", err)
				}
			case tc.expectError:
				t.Error("failed to receive an error")
			default:
				if diff := cmp.Diff(tc.want, got, protocmp.Transform()); diff != "" {
					t.Errorf("got unexpected diff (-want +got):\n%s", diff)
				}
			}
		})

	}

}

func TestListTabSummariesHTTP(t *testing.T) {
	tests := []struct {
		name             string
		config           map[string]*configpb.Configuration
		summaries        map[string]*summarypb.DashboardSummary
		endpoint         string
		params           string
		expectedCode     int
		expectedResponse *apipb.ListTabSummariesResponse
	}{
		{
			name: "Returns an error when there's no dashboard in config",
			config: map[string]*configpb.Configuration{
				"gs://default/config": {},
			},
			endpoint:     "/dashboards/whatever/tab-summaries",
			expectedCode: http.StatusNotFound,
		},
		{
			name: "Returns an error when there's no summary for dashboard yet",
			config: map[string]*configpb.Configuration{
				"gs://default/config": {
					Dashboards: []*configpb.Dashboard{
						{
							Name: "ACME",
							DashboardTab: []*configpb.DashboardTab{
								{
									Name:          "me-me",
									TestGroupName: "testgroupname",
								},
							},
						},
					},
				},
			},
			endpoint:     "/dashboards/acme/tab-summaries",
			expectedCode: http.StatusNotFound,
		},
		{
			name: "Returns correct tab summaries for a dashboard",
			config: map[string]*configpb.Configuration{
				"gs://default/config": {
					Dashboards: []*configpb.Dashboard{
						{
							Name: "Marco",
							DashboardTab: []*configpb.DashboardTab{
								{
									Name:          "polo-1",
									TestGroupName: "cheesecake",
								},
								{
									Name:          "polo-2",
									TestGroupName: "tiramisu",
								},
							},
						},
					},
				},
			},
			summaries: map[string]*summarypb.DashboardSummary{
				"gs://default/summary/summary-marco": {
					TabSummaries: []*summarypb.DashboardTabSummary{
						{
							DashboardName:       "Marco",
							DashboardTabName:    "polo-1",
							Status:              "1/7 tests are passing!",
							OverallStatus:       summarypb.DashboardTabSummary_FLAKY,
							LatestGreen:         "Hulk",
							LastUpdateTimestamp: float64(915166800),
							LastRunTimestamp:    float64(915166800),
						},
						{
							DashboardName:       "Marco",
							DashboardTabName:    "polo-2",
							Status:              "1/7 tests are failing!",
							OverallStatus:       summarypb.DashboardTabSummary_ACCEPTABLE,
							LatestGreen:         "Lantern",
							LastUpdateTimestamp: float64(916166800),
							LastRunTimestamp:    float64(916166800),
						},
					},
				},
			},
			endpoint:     "/dashboards/marco/tab-summaries",
			expectedCode: http.StatusOK,
			expectedResponse: &apipb.ListTabSummariesResponse{
				TabSummaries: []*apipb.TabSummary{
					{
						DashboardName:         "Marco",
						TabName:               "polo-1",
						OverallStatus:         "FLAKY",
						DetailedStatusMessage: "1/7 tests are passing!",
						LatestPassingBuild:    "Hulk",
						LastUpdateTimestamp: &timestamp.Timestamp{
							Seconds: 915166800,
						},
						LastRunTimestamp: &timestamp.Timestamp{
							Seconds: 915166800,
						},
					},
					{
						DashboardName:         "Marco",
						TabName:               "polo-2",
						OverallStatus:         "ACCEPTABLE",
						DetailedStatusMessage: "1/7 tests are failing!",
						LatestPassingBuild:    "Lantern",
						LastUpdateTimestamp: &timestamp.Timestamp{
							Seconds: 916166800,
						},
						LastRunTimestamp: &timestamp.Timestamp{
							Seconds: 916166800,
						},
					},
				},
			},
		},
		{
			name: "Returns correct tab summaries for a dashboard with an updated scope",
			config: map[string]*configpb.Configuration{
				"gs://k9s/config": {
					Dashboards: []*configpb.Dashboard{
						{
							Name: "Marco",
							DashboardTab: []*configpb.DashboardTab{
								{
									Name:          "polo-1",
									TestGroupName: "cheesecake",
								},
							},
						},
					},
				},
			},
			summaries: map[string]*summarypb.DashboardSummary{
				"gs://k9s/summary/summary-marco": {
					TabSummaries: []*summarypb.DashboardTabSummary{
						{
							DashboardName:       "Marco",
							DashboardTabName:    "polo-1",
							Status:              "1/7 tests are passing!",
							OverallStatus:       summarypb.DashboardTabSummary_FLAKY,
							LatestGreen:         "Hulk",
							LastUpdateTimestamp: float64(915166800),
							LastRunTimestamp:    float64(915166800),
						},
					},
				},
			},
			endpoint:     "/dashboards/marco/tab-summaries?scope=gs://k9s",
			expectedCode: http.StatusOK,
			expectedResponse: &apipb.ListTabSummariesResponse{
				TabSummaries: []*apipb.TabSummary{
					{
						DashboardName:         "Marco",
						TabName:               "polo-1",
						OverallStatus:         "FLAKY",
						DetailedStatusMessage: "1/7 tests are passing!",
						LatestPassingBuild:    "Hulk",
						LastUpdateTimestamp: &timestamp.Timestamp{
							Seconds: 915166800,
						},
						LastRunTimestamp: &timestamp.Timestamp{
							Seconds: 915166800,
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := Route(nil, setupTestServer(t, test.config, nil, test.summaries))
			request, err := http.NewRequest("GET", test.endpoint, nil)
			if err != nil {
				t.Fatalf("Can't form request: %v", err)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != test.expectedCode {
				t.Errorf("Expected %d, but got %d", test.expectedCode, response.Code)
			}

			if response.Code == http.StatusOK {
				var ts apipb.ListTabSummariesResponse
				if err := protojson.Unmarshal(response.Body.Bytes(), &ts); err != nil {
					t.Fatalf("Failed to unmarshal json message into a proto message: %v", err)
				}
				if diff := cmp.Diff(test.expectedResponse, &ts, protocmp.Transform()); diff != "" {
					t.Errorf("Obtained unexpected  diff (-want +got):\n%s", diff)
				}
			}
		})
	}
}
