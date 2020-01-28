package sendgrid

import (
	"reflect"
	"testing"
)

func Test_buildCreateSubUserBody(t *testing.T) {
	type args struct {
		username string
		email    string
		password string
		ips      []string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "correct format",
			args: args{
				username: "test",
				email:    "test",
				password: "test",
				ips:      []string{"127.0.0.1"},
			},
			want: []byte(`{"username":"test","email":"test","password":"test","ips":["127.0.0.1"]}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildCreateSubUserBody(tt.args.username, tt.args.email, tt.args.password, tt.args.ips)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildCreateSubUserBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildCreateSubUserBody() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildCreateApiKeyBody(t *testing.T) {
	type args struct {
		id     string
		scopes []string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "correct format",
			args: args{
				id:     "test",
				scopes: mockAPIScopes,
			},
			want: []byte(`{"name":"test","scopes":["test"]}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildCreateAPIKeyBody(tt.args.id, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildCreateAPIKeyBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildCreateAPIKeyBody() got = %v, want %v", got, tt.want)
			}
		})
	}
}
