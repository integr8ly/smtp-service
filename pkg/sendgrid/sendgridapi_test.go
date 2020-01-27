package sendgrid

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/pkg/errors"

	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go"

	"github.com/sirupsen/logrus"
)

func newMockRESTClient(modifyFn func(c *RESTClientMock)) RESTClient {
	restClient := &RESTClientMock{
		BuildRequestFunc: func(endpoint string, method rest.Method) rest.Request {
			req := sendgrid.GetRequest("test", endpoint, APIHost)
			req.Method = method
			return req
		},
		InvokeRequestFunc: func(request rest.Request) (response *rest.Response, e error) {
			return &rest.Response{
				StatusCode: 200,
				Body:       "{}",
				Headers:    map[string][]string{},
			}, nil
		},
	}
	modifyFn(restClient)
	return restClient
}

var mockRESTClientInvalidJSON = newMockRESTClient(func(c *RESTClientMock) {
	c.InvokeRequestFunc = func(request rest.Request) (response *rest.Response, e error) {
		return &rest.Response{
			StatusCode: 200,
			Body:       "this is not json",
			Headers:    map[string][]string{},
		}, nil
	}
})

var mockRESTClientFailedInvoke = newMockRESTClient(func(c *RESTClientMock) {
	c.InvokeRequestFunc = func(request rest.Request) (response *rest.Response, e error) {
		return nil, errors.New("test")
	}
})

func TestBackendAPIClient_CreateAPIKeyForSubUser(t *testing.T) {
	type fields struct {
		restClient RESTClient
		logger     *logrus.Entry
	}
	type args struct {
		username string
		scopes   []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *APIKey
		wantErr bool
	}{
		{
			name: "successful create",
			fields: fields{
				restClient: newMockRESTClient(func(c *RESTClientMock) {
					c.InvokeRequestFunc = func(request rest.Request) (response *rest.Response, e error) {
						apiKeyJSON, err := json.Marshal(newMockAPIKey())
						if err != nil {
							panic(err)
						}
						return &rest.Response{
							StatusCode: 200,
							Body:       string(apiKeyJSON),
							Headers:    map[string][]string{},
						}, nil
					}
				}),
				logger: newMockLogger(),
			},
			args: args{
				username: "test",
				scopes:   mockAPIScopes,
			},
			want: newMockAPIKey(),
		},
		{
			name: "api request fails",
			fields: fields{
				restClient: mockRESTClientFailedInvoke,
				logger:     newMockLogger(),
			},
			args: args{
				username: "test",
				scopes:   mockAPIScopes,
			},
			wantErr: true,
		},
		{
			name: "api response invalid json",
			fields: fields{
				restClient: mockRESTClientInvalidJSON,
				logger:     newMockLogger(),
			},
			args: args{
				username: "test",
				scopes:   mockAPIScopes,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BackendAPIClient{
				restClient: tt.fields.restClient,
				logger:     tt.fields.logger,
			}
			got, err := c.CreateAPIKeyForSubUser(tt.args.username, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAPIKeyForSubUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateAPIKeyForSubUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackendAPIClient_CreateSubUser(t *testing.T) {
	type fields struct {
		restClient RESTClient
		logger     *logrus.Entry
	}
	type args struct {
		id       string
		email    string
		password string
		ips      []string
	}
	testArgs := args{
		id:       "test",
		email:    "test",
		password: "test",
		ips:      []string{"127.0.0.1"},
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *SubUser
		wantErr bool
	}{
		{
			name: "successful create",
			fields: fields{
				restClient: newMockRESTClient(func(c *RESTClientMock) {
					c.InvokeRequestFunc = func(request rest.Request) (response *rest.Response, e error) {
						subUserJSON, err := json.Marshal(newMockSubUser())
						if err != nil {
							panic(err)
						}
						return &rest.Response{
							StatusCode: 201,
							Body:       string(subUserJSON),
							Headers:    map[string][]string{},
						}, nil
					}
				}),
				logger: newMockLogger(),
			},
			args: testArgs,
			want: newMockSubUser(),
		},
		{
			name: "api request fails",
			fields: fields{
				restClient: mockRESTClientFailedInvoke,
				logger:     newMockLogger(),
			},
			args:    testArgs,
			wantErr: true,
		},
		{
			name: "api response invalid json",
			fields: fields{
				restClient: mockRESTClientInvalidJSON,
				logger:     newMockLogger(),
			},
			args:    testArgs,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BackendAPIClient{
				restClient: tt.fields.restClient,
				logger:     tt.fields.logger,
			}
			got, err := c.CreateSubUser(tt.args.id, tt.args.email, tt.args.password, tt.args.ips)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSubUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateSubUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackendAPIClient_DeleteSubUser(t *testing.T) {
	type fields struct {
		restClient RESTClient
		logger     *logrus.Entry
	}
	type args struct {
		username string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "successful delete",
			fields: fields{
				restClient: newMockRESTClient(func(c *RESTClientMock) {
					c.InvokeRequestFunc = func(request rest.Request) (response *rest.Response, e error) {
						return &rest.Response{
							StatusCode: 204,
							Body:       "{}",
							Headers:    map[string][]string{},
						}, nil
					}
				}),
				logger: newMockLogger(),
			},
			args: args{username: "test"},
		},
		{
			name: "username not defined",
			fields: fields{
				restClient: newMockRESTClient(func(c *RESTClientMock) {
					c.InvokeRequestFunc = func(request rest.Request) (response *rest.Response, e error) {
						return &rest.Response{
							StatusCode: 204,
							Body:       "{}",
							Headers:    map[string][]string{},
						}, nil
					}
				}),
				logger: newMockLogger(),
			},
			args:    args{username: ""},
			wantErr: true,
		},
		{
			name: "request fails",
			fields: fields{
				restClient: mockRESTClientFailedInvoke,
				logger:     newMockLogger(),
			},
			args:    args{username: "test"},
			wantErr: true,
		},
		{
			name: "request fails",
			fields: fields{
				restClient: mockRESTClientInvalidJSON,
				logger:     newMockLogger(),
			},
			args:    args{username: "test"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BackendAPIClient{
				restClient: tt.fields.restClient,
				logger:     tt.fields.logger,
			}
			if err := c.DeleteSubUser(tt.args.username); (err != nil) != tt.wantErr {
				t.Errorf("DeleteSubUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBackendAPIClient_GetAPIKeysForSubUser(t *testing.T) {
	type fields struct {
		restClient RESTClient
		logger     *logrus.Entry
	}
	type args struct {
		username string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*APIKey
		wantErr bool
	}{
		{
			name: "successful get",
			fields: fields{
				restClient: newMockRESTClient(func(c *RESTClientMock) {
					c.InvokeRequestFunc = func(request rest.Request) (response *rest.Response, e error) {
						apiKeyListResp := apiKeysListResponse{
							Result: []*APIKey{newMockAPIKey()},
						}
						apiKeyJSON, err := json.Marshal(apiKeyListResp)
						if err != nil {
							panic(err)
						}
						return &rest.Response{
							StatusCode: 200,
							Body:       string(apiKeyJSON),
							Headers:    map[string][]string{},
						}, nil
					}
				}),
				logger: newMockLogger(),
			},
			args: args{username: "test"},
			want: []*APIKey{newMockAPIKey()},
		},
		{
			name: "username is undefined",
			fields: fields{
				restClient: newMockRESTClient(func(c *RESTClientMock) {
					c.InvokeRequestFunc = func(request rest.Request) (response *rest.Response, e error) {
						apiKeyListResp := apiKeysListResponse{
							Result: []*APIKey{newMockAPIKey()},
						}
						apiKeyJSON, err := json.Marshal(apiKeyListResp)
						if err != nil {
							panic(err)
						}
						return &rest.Response{
							StatusCode: 200,
							Body:       string(apiKeyJSON),
							Headers:    map[string][]string{},
						}, nil
					}
				}),
				logger: newMockLogger(),
			},
			args:    args{username: ""},
			wantErr: true,
		},
		{
			name: "get request fails",
			fields: fields{
				restClient: mockRESTClientFailedInvoke,
				logger:     newMockLogger(),
			},
			args:    args{username: "test"},
			wantErr: true,
		},
		{
			name: "api response invalid json",
			fields: fields{
				restClient: mockRESTClientInvalidJSON,
				logger:     newMockLogger(),
			},
			args:    args{username: "test"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BackendAPIClient{
				restClient: tt.fields.restClient,
				logger:     tt.fields.logger,
			}
			got, err := c.GetAPIKeysForSubUser(tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAPIKeysForSubUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAPIKeysForSubUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackendAPIClient_ListIPAddresses(t *testing.T) {
	type fields struct {
		restClient RESTClient
		logger     *logrus.Entry
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*IPAddress
		wantErr bool
	}{
		{
			name: "successful list",
			fields: fields{
				restClient: newMockRESTClient(func(c *RESTClientMock) {
					c.InvokeRequestFunc = func(request rest.Request) (response *rest.Response, e error) {
						respJSON, err := json.Marshal([]*IPAddress{newMockIPAddress()})
						if err != nil {
							panic(err)
						}
						return &rest.Response{
							StatusCode: 200,
							Body:       string(respJSON),
							Headers:    map[string][]string{},
						}, nil
					}
				}),
				logger: newMockLogger(),
			},
			want: []*IPAddress{newMockIPAddress()},
		},
		{
			name: "get request fails",
			fields: fields{
				restClient: mockRESTClientFailedInvoke,
				logger:     newMockLogger(),
			},
			wantErr: true,
		},
		{
			name: "api response invalid json",
			fields: fields{
				restClient: mockRESTClientInvalidJSON,
				logger:     newMockLogger(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BackendAPIClient{
				restClient: tt.fields.restClient,
				logger:     tt.fields.logger,
			}
			got, err := c.ListIPAddresses()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListIPAddresses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListIPAddresses() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackendAPIClient_ListSubUsers(t *testing.T) {
	type fields struct {
		restClient RESTClient
		logger     *logrus.Entry
	}
	type args struct {
		query map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*SubUser
		wantErr bool
	}{
		{
			name: "successful list with undefined query",
			fields: fields{
				restClient: newMockRESTClient(func(c *RESTClientMock) {
					c.InvokeRequestFunc = func(request rest.Request) (response *rest.Response, e error) {
						respJSON, err := json.Marshal([]*SubUser{newMockSubUser()})
						if err != nil {
							panic(err)
						}
						return &rest.Response{
							StatusCode: 200,
							Body:       string(respJSON),
							Headers:    map[string][]string{},
						}, nil
					}
				}),
				logger: newMockLogger(),
			},
			args: args{query: nil},
			want: []*SubUser{newMockSubUser()},
		},
		{
			name: "get request fails",
			fields: fields{
				restClient: mockRESTClientFailedInvoke,
				logger:     newMockLogger(),
			},
			wantErr: true,
		},
		{
			name: "api response invalid json",
			fields: fields{
				restClient: mockRESTClientInvalidJSON,
				logger:     newMockLogger(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BackendAPIClient{
				restClient: tt.fields.restClient,
				logger:     tt.fields.logger,
			}
			got, err := c.ListSubUsers(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListSubUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListSubUsers() got = %v, want %v", got, tt.want)
			}
		})
	}
}
