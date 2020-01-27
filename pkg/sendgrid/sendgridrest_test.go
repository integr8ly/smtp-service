package sendgrid

import (
	"fmt"
	"testing"

	"github.com/sendgrid/rest"
	"github.com/sirupsen/logrus"
)

const (
	testApiKey = "testApiKey"
)

func newMockLogger() *logrus.Entry {
	return logrus.WithField("test", "test")
}

func TestBackendRESTClient_BuildRequest(t *testing.T) {
	type fields struct {
		apiHost string
		apiKey  string
		logger  *logrus.Entry
	}
	type args struct {
		endpoint string
		method   rest.Method
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   rest.Request
	}{
		{
			name: "provided args should be set correctly",
			fields: fields{
				apiHost: APIHost,
				apiKey:  testApiKey,
				logger:  newMockLogger(),
			},
			args: args{
				endpoint: APIRouteSubUsers,
				method:   rest.Post,
			},
			want: rest.Request{
				Method:  rest.Post,
				BaseURL: fmt.Sprintf("%s%s", APIHost, APIRouteSubUsers),
				Headers: map[string]string{
					"Authorization": fmt.Sprintf("Bearer %s", testApiKey),
				},
				QueryParams: map[string]string{},
				Body:        []byte{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BackendRESTClient{
				apiHost: tt.fields.apiHost,
				apiKey:  tt.fields.apiKey,
				logger:  tt.fields.logger,
			}
			got := b.BuildRequest(tt.args.endpoint, tt.args.method)
			if got.Method != tt.want.Method {
				t.Errorf("GetRequest() Method = %v, want %v", got.Method, tt.want.Method)
			}
			if got.BaseURL != tt.want.BaseURL {
				t.Errorf("GetRequest() BaseURL = %v, want %v", got.BaseURL, tt.want.BaseURL)
			}
			for key, value := range tt.want.Headers {
				if got.Headers[key] != tt.want.Headers[key] {
					t.Errorf("GetRequest() Headers[%s] = %v, want %v", key, value, tt.want.Headers[key])
				}
			}
		})
	}
}
