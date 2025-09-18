package mcp

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	registryv0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
)

func TestServersService_List(t *testing.T) {
	updatedSince := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		opts           *ServerListOptions
		expectedQuery  values
		responseBody   string
		expectedResult *registryv0.ServerListResponse
		expectedCursor string
	}{
		{
			name: "basic list with options",
			opts: &ServerListOptions{
				ListOptions: ListOptions{
					Limit:  10,
					Cursor: "abc123",
				},
				Search: "github",
			},
			expectedQuery: values{
				"limit":  "10",
				"cursor": "abc123",
				"search": "github",
			},
			responseBody: `{
				"servers": [
					{
						"name": "test-server",
						"version": "1.0.0",
						"description": "A test server",
						"repository": {
							"url": "https://github.com/example/test-server"
						}
					}
				],
				"metadata": {
					"next_cursor": "next123"
				}
			}`,
			expectedResult: &registryv0.ServerListResponse{
				Servers: []registryv0.ServerJSON{
					{
						Name:        "test-server",
						Version:     "1.0.0",
						Description: "A test server",
						Repository: model.Repository{
							URL: "https://github.com/example/test-server",
						},
					},
				},
				Metadata: &registryv0.Metadata{
					NextCursor: "next123",
				},
			},
			expectedCursor: "next123",
		},
		{
			name: "list with updated_since filter",
			opts: &ServerListOptions{
				UpdatedSince: &updatedSince,
			},
			expectedQuery: values{
				"updated_since": "2024-01-01T00:00:00Z",
			},
			responseBody: `{"servers": [], "metadata": {}}`,
			expectedResult: &registryv0.ServerListResponse{
				Servers:  []registryv0.ServerJSON{},
				Metadata: &registryv0.Metadata{},
			},
			expectedCursor: "",
		},
		{
			name:          "empty list with no options",
			opts:          nil,
			expectedQuery: values{},
			responseBody:  `{"servers": [], "metadata": {}}`,
			expectedResult: &registryv0.ServerListResponse{
				Servers:  []registryv0.ServerJSON{},
				Metadata: &registryv0.Metadata{},
			},
			expectedCursor: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux, _, teardown := setup()
			defer teardown()

			mux.HandleFunc("/v0/servers", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, "GET")
				testFormValues(t, r, tt.expectedQuery)

				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, tt.responseBody)
			})

			ctx := context.Background()
			servers, resp, err := client.Servers.List(ctx, tt.opts)
			if err != nil {
				t.Errorf("Servers.List returned error: %v", err)
			}

			if !reflect.DeepEqual(servers, tt.expectedResult) {
				t.Errorf("Servers.List returned %+v, want %+v", servers, tt.expectedResult)
			}

			if resp.NextCursor != tt.expectedCursor {
				t.Errorf("Response.NextCursor = %v, want %v", resp.NextCursor, tt.expectedCursor)
			}
		})
	}
}

func TestServersService_Get(t *testing.T) {
	publishedAt, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2024-01-02T00:00:00Z")

	tests := []struct {
		name           string
		serverID       string
		statusCode     int
		responseBody   string
		expectedResult *registryv0.ServerJSON
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:       "successful get by ID",
			serverID:   "test-id",
			statusCode: http.StatusOK,
			responseBody: `{
				"name": "test-server",
				"version": "1.0.0",
				"description": "A test server",
				"repository": {
					"url": "https://github.com/example/test-server"
				},
				"_meta": {
					"io.modelcontextprotocol.registry/official": {
						"id": "test-id",
						"published_at": "2024-01-01T00:00:00Z",
						"updated_at": "2024-01-02T00:00:00Z",
						"is_latest": true
					}
				}
			}`,
			expectedResult: &registryv0.ServerJSON{
				Name:        "test-server",
				Version:     "1.0.0",
				Description: "A test server",
				Repository: model.Repository{
					URL: "https://github.com/example/test-server",
				},
				Meta: &registryv0.ServerMeta{
					Official: &registryv0.RegistryExtensions{
						ID:          "test-id",
						PublishedAt: publishedAt,
						UpdatedAt:   updatedAt,
						IsLatest:    true,
					},
				},
			},
			expectError: false,
		},
		{
			name:           "server not found",
			serverID:       "nonexistent",
			statusCode:     http.StatusNotFound,
			responseBody:   `{"message": "Server not found"}`,
			expectedResult: nil,
			expectError:    true,
			expectedErrMsg: "Server not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux, _, teardown := setup()
			defer teardown()

			mux.HandleFunc(fmt.Sprintf("/v0/servers/%s", tt.serverID), func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, "GET")
				w.WriteHeader(tt.statusCode)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, tt.responseBody)
			})

			ctx := context.Background()
			server, resp, err := client.Servers.Get(ctx, tt.serverID)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if resp.StatusCode != tt.statusCode {
					t.Errorf("Expected status code %d, got %d", tt.statusCode, resp.StatusCode)
				}
				if errResp, ok := err.(*ErrorResponse); ok {
					if errResp.Message != tt.expectedErrMsg {
						t.Errorf("Expected error message %q, got %q", tt.expectedErrMsg, errResp.Message)
					}
				} else {
					t.Errorf("Expected ErrorResponse, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Servers.Get returned error: %v", err)
				}
				if !reflect.DeepEqual(server, tt.expectedResult) {
					t.Errorf("Servers.Get returned %+v, want %+v", server, tt.expectedResult)
				}
			}
		})
	}
}

func TestServersService_ListAll(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	page := 0
	mux.HandleFunc("/v0/servers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")

		w.Header().Set("Content-Type", "application/json")

		if page == 0 {
			testFormValues(t, r, values{})
			fmt.Fprint(w, `{
				"servers": [
					{"name": "server1", "version": "1.0.0"},
					{"name": "server2", "version": "2.0.0"}
				],
				"metadata": {"next_cursor": "page2"}
			}`)
			page++
		} else {
			testFormValues(t, r, values{"cursor": "page2"})
			fmt.Fprint(w, `{
				"servers": [
					{"name": "server3", "version": "3.0.0"}
				],
				"metadata": {}
			}`)
		}
	})

	ctx := context.Background()
	servers, err := client.Servers.ListAll(ctx, nil)
	if err != nil {
		t.Errorf("Servers.ListAll returned error: %v", err)
	}

	if len(servers) != 3 {
		t.Errorf("Expected 3 servers, got %d", len(servers))
	}

	expectedNames := []string{"server1", "server2", "server3"}
	for i, server := range servers {
		if server.Name != expectedNames[i] {
			t.Errorf("Expected server name %s, got %s", expectedNames[i], server.Name)
		}
	}
}

func TestServersService_GetByName(t *testing.T) {
	tests := []struct {
		name            string
		searchName      string
		expectedQuery   values
		responseBody    string
		expectedResults []registryv0.ServerJSON
		expectError     bool
	}{
		{
			name:       "single exact match found",
			searchName: "exact-name",
			expectedQuery: values{
				"search": "exact-name",
				"limit":  "100",
			},
			responseBody: `{
				"servers": [
					{"name": "exact-name", "version": "1.0.0"},
					{"name": "exact-name-plus", "version": "2.0.0"}
				],
				"metadata": {}
			}`,
			expectedResults: []registryv0.ServerJSON{
				{
					Name:    "exact-name",
					Version: "1.0.0",
				},
			},
			expectError: false,
		},
		{
			name:       "multiple versions of same server",
			searchName: "test-server",
			expectedQuery: values{
				"search": "test-server",
				"limit":  "100",
			},
			responseBody: `{
				"servers": [
					{"name": "test-server-alpha", "version": "1.0.0"},
					{"name": "test-server", "version": "2.0.0"},
					{"name": "test-server", "version": "1.5.0"},
					{"name": "test-server-beta", "version": "3.0.0"}
				],
				"metadata": {}
			}`,
			expectedResults: []registryv0.ServerJSON{
				{
					Name:    "test-server",
					Version: "2.0.0",
				},
				{
					Name:    "test-server",
					Version: "1.5.0",
				},
			},
			expectError: false,
		},
		{
			name:       "no servers found",
			searchName: "nonexistent",
			expectedQuery: values{
				"search": "nonexistent",
				"limit":  "100",
			},
			responseBody:    `{"servers": [], "metadata": {}}`,
			expectedResults: []registryv0.ServerJSON{},
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux, _, teardown := setup()
			defer teardown()

			mux.HandleFunc("/v0/servers", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, "GET")
				testFormValues(t, r, tt.expectedQuery)

				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, tt.responseBody)
			})

			ctx := context.Background()
			servers, err := client.Servers.GetByName(ctx, tt.searchName)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Servers.GetByName returned error: %v", err)
				}

				if len(servers) != len(tt.expectedResults) {
					t.Errorf("Expected %d servers, got %d", len(tt.expectedResults), len(servers))
				}

				for i, expectedServer := range tt.expectedResults {
					if i >= len(servers) {
						t.Errorf("Missing expected server at index %d", i)
						continue
					}
					if servers[i].Name != expectedServer.Name {
						t.Errorf("Expected server name %q at index %d, got %q", expectedServer.Name, i, servers[i].Name)
					}
					if servers[i].Version != expectedServer.Version {
						t.Errorf("Expected server version %q at index %d, got %q", expectedServer.Version, i, servers[i].Version)
					}
				}
			}
		})
	}
}

func TestServersService_GetByNameLatest(t *testing.T) {
	tests := []struct {
		name           string
		searchName     string
		expectedQuery  values
		responseBody   string
		expectedResult *registryv0.ServerJSON
		expectNil      bool
	}{
		{
			name:       "latest version found",
			searchName: "test-server",
			expectedQuery: values{
				"search":  "test-server",
				"version": "latest",
				"limit":   "100",
			},
			responseBody: `{
				"servers": [
					{"name": "test-server-alpha", "version": "2.0.0"},
					{"name": "test-server", "version": "3.0.0"},
					{"name": "test-server-beta", "version": "1.5.0"}
				],
				"metadata": {}
			}`,
			expectedResult: &registryv0.ServerJSON{
				Name:    "test-server",
				Version: "3.0.0",
			},
			expectNil: false,
		},
		{
			name:       "exact match among multiple similar names",
			searchName: "exact-name",
			expectedQuery: values{
				"search":  "exact-name",
				"version": "latest",
				"limit":   "100",
			},
			responseBody: `{
				"servers": [
					{"name": "exact-name", "version": "2.5.0"},
					{"name": "exact-name-plus", "version": "1.0.0"}
				],
				"metadata": {}
			}`,
			expectedResult: &registryv0.ServerJSON{
				Name:    "exact-name",
				Version: "2.5.0",
			},
			expectNil: false,
		},
		{
			name:       "server not found",
			searchName: "nonexistent",
			expectedQuery: values{
				"search":  "nonexistent",
				"version": "latest",
				"limit":   "100",
			},
			responseBody: `{"servers": [], "metadata": {}}`,
			expectNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux, _, teardown := setup()
			defer teardown()

			mux.HandleFunc("/v0/servers", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, "GET")
				testFormValues(t, r, tt.expectedQuery)

				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, tt.responseBody)
			})

			ctx := context.Background()
			server, err := client.Servers.GetByNameLatest(ctx, tt.searchName)

			if err != nil {
				t.Errorf("Servers.GetByNameLatest returned error: %v", err)
			}

			if tt.expectNil {
				if server != nil {
					t.Errorf("Expected nil server, got %+v", server)
				}
			} else {
				if server == nil {
					t.Error("Expected server but got nil")
				} else {
					if server.Name != tt.expectedResult.Name {
						t.Errorf("Expected server name %q, got %q", tt.expectedResult.Name, server.Name)
					}
					if server.Version != tt.expectedResult.Version {
						t.Errorf("Expected server version %q, got %q", tt.expectedResult.Version, server.Version)
					}
				}
			}
		})
	}
}

func TestServersService_GetByNameExactVersion(t *testing.T) {
	tests := []struct {
		name           string
		searchName     string
		searchVersion  string
		expectedQuery  values
		responseBody   string
		expectedResult *registryv0.ServerJSON
		expectNil      bool
	}{
		{
			name:          "exact version found",
			searchName:    "test-server",
			searchVersion: "2.0.0",
			expectedQuery: values{
				"search": "test-server",
				"limit":  "100",
			},
			responseBody: `{
				"servers": [
					{"name": "test-server-alpha", "version": "2.0.0"},
					{"name": "test-server", "version": "1.0.0"},
					{"name": "test-server", "version": "2.0.0"},
					{"name": "test-server-beta", "version": "1.5.0"}
				],
				"metadata": {}
			}`,
			expectedResult: &registryv0.ServerJSON{
				Name:    "test-server",
				Version: "2.0.0",
			},
			expectNil: false,
		},
		{
			name:          "version not found but server exists",
			searchName:    "test-server",
			searchVersion: "3.0.0",
			expectedQuery: values{
				"search": "test-server",
				"limit":  "100",
			},
			responseBody: `{
				"servers": [
					{"name": "test-server", "version": "1.0.0"},
					{"name": "test-server", "version": "2.0.0"}
				],
				"metadata": {}
			}`,
			expectNil: true,
		},
		{
			name:          "server name not found",
			searchName:    "nonexistent",
			searchVersion: "1.0.0",
			expectedQuery: values{
				"search": "nonexistent",
				"limit":  "100",
			},
			responseBody: `{"servers": [], "metadata": {}}`,
			expectNil:    true,
		},
		{
			name:          "exact match with similar server names",
			searchName:    "exact-name",
			searchVersion: "1.5.0",
			expectedQuery: values{
				"search": "exact-name",
				"limit":  "100",
			},
			responseBody: `{
				"servers": [
					{"name": "exact-name-plus", "version": "1.5.0"},
					{"name": "exact-name", "version": "1.0.0"},
					{"name": "exact-name", "version": "1.5.0"}
				],
				"metadata": {}
			}`,
			expectedResult: &registryv0.ServerJSON{
				Name:    "exact-name",
				Version: "1.5.0",
			},
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux, _, teardown := setup()
			defer teardown()

			mux.HandleFunc("/v0/servers", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, "GET")
				testFormValues(t, r, tt.expectedQuery)

				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, tt.responseBody)
			})

			ctx := context.Background()
			server, err := client.Servers.GetByNameExactVersion(ctx, tt.searchName, tt.searchVersion)

			if err != nil {
				t.Errorf("Servers.GetByNameExactVersion returned error: %v", err)
			}

			if tt.expectNil {
				if server != nil {
					t.Errorf("Expected nil server, got %+v", server)
				}
			} else {
				if server == nil {
					t.Error("Expected server but got nil")
				} else {
					if server.Name != tt.expectedResult.Name {
						t.Errorf("Expected server name %q, got %q", tt.expectedResult.Name, server.Name)
					}
					if server.Version != tt.expectedResult.Version {
						t.Errorf("Expected server version %q, got %q", tt.expectedResult.Version, server.Version)
					}
				}
			}
		})
	}
}

func TestServersService_GetByNameLatestActiveVersion(t *testing.T) {
	tests := []struct {
		name           string
		searchName     string
		expectedQuery  values
		responseBody   string
		expectedResult *registryv0.ServerJSON
		expectNil      bool
	}{
		{
			name:       "latest active version found",
			searchName: "test-server",
			expectedQuery: values{
				"search": "test-server",
				"limit":  "100",
			},
			responseBody: `{
				"servers": [
					{"name": "test-server", "version": "1.0.0", "status": "active"},
					{"name": "test-server", "version": "2.0.0", "status": "deprecated"},
					{"name": "test-server", "version": "1.5.0", "status": "active"}
				],
				"metadata": {}
			}`,
			expectedResult: &registryv0.ServerJSON{
				Name:    "test-server",
				Version: "1.5.0",
				Status:  "active",
			},
			expectNil: false,
		},
		{
			name:       "no active versions",
			searchName: "test-server",
			expectedQuery: values{
				"search": "test-server",
				"limit":  "100",
			},
			responseBody: `{
				"servers": [
					{"name": "test-server", "version": "1.0.0", "status": "deprecated"},
					{"name": "test-server", "version": "2.0.0", "status": "deleted"}
				],
				"metadata": {}
			}`,
			expectNil: true,
		},
		{
			name:       "server not found",
			searchName: "nonexistent",
			expectedQuery: values{
				"search": "nonexistent",
				"limit":  "100",
			},
			responseBody: `{"servers": [], "metadata": {}}`,
			expectNil:    true,
		},
		{
			name:       "skip invalid semantic versions",
			searchName: "test-server",
			expectedQuery: values{
				"search": "test-server",
				"limit":  "100",
			},
			responseBody: `{
				"servers": [
					{"name": "test-server", "version": "invalid", "status": "active"},
					{"name": "test-server", "version": "1.0.0", "status": "active"},
					{"name": "test-server", "version": "not-semver", "status": "active"}
				],
				"metadata": {}
			}`,
			expectedResult: &registryv0.ServerJSON{
				Name:    "test-server",
				Version: "1.0.0",
				Status:  "active",
			},
			expectNil: false,
		},
		{
			name:       "filter by exact name match",
			searchName: "exact-name",
			expectedQuery: values{
				"search": "exact-name",
				"limit":  "100",
			},
			responseBody: `{
				"servers": [
					{"name": "exact-name-plus", "version": "2.0.0", "status": "active"},
					{"name": "exact-name", "version": "1.0.0", "status": "active"}
				],
				"metadata": {}
			}`,
			expectedResult: &registryv0.ServerJSON{
				Name:    "exact-name",
				Version: "1.0.0",
				Status:  "active",
			},
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux, _, teardown := setup()
			defer teardown()

			mux.HandleFunc("/v0/servers", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, "GET")
				testFormValues(t, r, tt.expectedQuery)

				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, tt.responseBody)
			})

			ctx := context.Background()
			server, err := client.Servers.GetByNameLatestActiveVersion(ctx, tt.searchName)

			if err != nil {
				t.Errorf("Servers.GetByNameLatestActiveVersion returned error: %v", err)
			}

			if tt.expectNil {
				if server != nil {
					t.Errorf("Expected nil server, got %+v", server)
				}
			} else {
				if server == nil {
					t.Error("Expected server but got nil")
				} else {
					if server.Name != tt.expectedResult.Name {
						t.Errorf("Expected server name %q, got %q", tt.expectedResult.Name, server.Name)
					}
					if server.Version != tt.expectedResult.Version {
						t.Errorf("Expected server version %q, got %q", tt.expectedResult.Version, server.Version)
					}
					if server.Status != tt.expectedResult.Status {
						t.Errorf("Expected server status %q, got %q", tt.expectedResult.Status, server.Status)
					}
				}
			}
		})
	}
}

func TestServersService_ListUpdatedSince(t *testing.T) {
	// Create test timestamp
	testTime, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")

	tests := []struct {
		name          string
		since         time.Time
		responses     []string
		expectedCount int
		expectError   bool
	}{
		{
			name:  "single page with updated servers",
			since: testTime,
			responses: []string{
				`{
					"servers": [
						{
							"name": "test/server1",
							"version": "1.0.0",
							"status": "active",
							"meta": {
								"official": {
									"id": "server1",
									"updated_at": "2024-01-02T00:00:00Z"
								}
							}
						},
						{
							"name": "test/server2",
							"version": "1.1.0",
							"status": "active",
							"meta": {
								"official": {
									"id": "server2",
									"updated_at": "2024-01-03T00:00:00Z"
								}
							}
						}
					],
					"metadata": {}
				}`,
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:  "multiple pages with updated servers",
			since: testTime,
			responses: []string{
				`{
					"servers": [
						{
							"name": "test/server1",
							"version": "1.0.0",
							"status": "active",
							"meta": {
								"official": {
									"id": "server1",
									"updated_at": "2024-01-02T00:00:00Z"
								}
							}
						}
					],
					"metadata": {
						"next_cursor": "page2"
					}
				}`,
				`{
					"servers": [
						{
							"name": "test/server2",
							"version": "1.1.0",
							"status": "active",
							"meta": {
								"official": {
									"id": "server2",
									"updated_at": "2024-01-03T00:00:00Z"
								}
							}
						}
					],
					"metadata": {}
				}`,
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:  "no updated servers",
			since: testTime,
			responses: []string{
				`{
					"servers": [],
					"metadata": {}
				}`,
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:  "api error",
			since: testTime,
			responses: []string{
				`{"message": "Internal server error"}`,
			},
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux, _, teardown := setup()
			defer teardown()

			callCount := 0
			mux.HandleFunc("/v0/servers", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, "GET")

				// Verify that updated_since parameter is set correctly
				expectedValues := values{
					"updated_since": tt.since.Format(time.RFC3339),
					"limit":         "100",
				}
				if callCount > 0 {
					expectedValues["cursor"] = "page2"
				}
				testFormValues(t, r, expectedValues)

				if tt.expectError {
					w.WriteHeader(http.StatusInternalServerError)
				}

				if callCount < len(tt.responses) {
					fmt.Fprint(w, tt.responses[callCount])
				}
				callCount++
			})

			servers, err := client.Servers.ListUpdatedSince(context.Background(), tt.since)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Servers.ListUpdatedSince returned error: %v", err)
			}

			if len(servers) != tt.expectedCount {
				t.Errorf("Expected %d servers, got %d", tt.expectedCount, len(servers))
			}

			// Verify all returned servers have updated_at after the since timestamp
			for _, server := range servers {
				if server.Meta != nil && server.Meta.Official != nil && !server.Meta.Official.UpdatedAt.IsZero() {
					serverUpdatedAt := server.Meta.Official.UpdatedAt
					if serverUpdatedAt.Before(tt.since) {
						t.Errorf("Server %s updated_at %s is before since timestamp %s",
							server.Meta.Official.ID, serverUpdatedAt.Format(time.RFC3339), tt.since.Format(time.RFC3339))
					}
				}
			}
		})
	}
}

// Test helper functions

func setup() (client *Client, mux *http.ServeMux, serverURL string, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)

	client = NewClient(nil)
	url, _ := url.Parse(server.URL + "/")
	client.BaseURL = url

	return client, mux, server.URL, server.Close
}

func testMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

type values map[string]string

func testFormValues(t *testing.T, r *http.Request, values values) {
	t.Helper()
	want := url.Values{}
	for k, v := range values {
		want.Set(k, v)
	}

	r.ParseForm()
	if got := r.Form; !reflect.DeepEqual(got, want) {
		t.Errorf("Request parameters: %v, want %v", got, want)
	}
}
