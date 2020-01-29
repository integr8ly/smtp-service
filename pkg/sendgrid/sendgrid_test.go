package sendgrid

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/integr8ly/smtp-service/pkg/smtpdetails"

	"github.com/sirupsen/logrus"
)

var (
	mockAPIClient   = defaultTestAPIClient()
	mockAPIScopes   = []string{"test"}
	mockPasswordGen = newMockPasswordGenerator()
)

func newMockSubUser() *SubUser {
	return &SubUser{
		ID:       0,
		Username: "test",
		Email:    "test@email.com",
		Disabled: false,
	}
}

func newMockAPIKey() *APIKey {
	return &APIKey{
		ID:     "test",
		Key:    "test",
		Name:   "test",
		Scopes: mockAPIScopes,
	}
}

func newMockIPAddress() *IPAddress {
	return &IPAddress{
		IP:        "127.0.0.1",
		Warmup:    false,
		StartDate: 0,
		SubUsers:  nil,
		RDNS:      "",
		Pools:     []string{"test"},
	}
}

func newMockSMTPDetails() *smtpdetails.SMTPDetails {
	return defaultConnectionDetails("test", "test")
}

func newMockAPIClient(modifyFn func(c *APIClientMock)) APIClient {
	apiClient := &APIClientMock{
		CreateAPIKeyForSubUserFunc: func(username string, scopes []string) (key *APIKey, e error) {
			return newMockAPIKey(), nil
		},
		CreateSubUserFunc: func(id string, email string, password string, ips []string) (user *SubUser, e error) {
			return newMockSubUser(), nil
		},
		DeleteSubUserFunc: func(username string) error {
			return nil
		},
		GetAPIKeysForSubUserFunc: func(username string) (keys []*APIKey, e error) {
			return []*APIKey{newMockAPIKey()}, nil
		},
		GetSubUserByUsernameFunc: func(username string) (user *SubUser, e error) {
			return newMockSubUser(), nil
		},
		ListIPAddressesFunc: func() (addresses []*IPAddress, e error) {
			return []*IPAddress{newMockIPAddress()}, nil
		},
		ListSubUsersFunc: func(query map[string]string) (users []*SubUser, e error) {
			return []*SubUser{newMockSubUser()}, nil
		},
	}
	modifyFn(apiClient)
	return apiClient
}

func newMockPasswordGenerator() smtpdetails.PasswordGenerator {
	return &smtpdetails.PasswordGeneratorMock{
		GenerateFunc: func(length int, numDigits int, numSymbols int, noUpper bool, allowRepeat bool) (s string, e error) {
			return "test", nil
		},
	}
}

func testAPIClient(apiRespCode int, apiRespBody string) APIClient {
	return &APIClientMock{
		CreateAPIKeyForSubUserFunc: nil,
		CreateSubUserFunc:          nil,
		DeleteSubUserFunc:          nil,
		GetAPIKeysForSubUserFunc:   nil,
		GetSubUserByUsernameFunc:   nil,
		ListIPAddressesFunc:        nil,
		ListSubUsersFunc:           nil,
	}
}

func defaultTestAPIClient() APIClient {
	return testAPIClient(200, "")
}

func TestNewClient(t *testing.T) {
	type args struct {
		sendgridClient APIClient
		apiKeyScopes   []string
		passGen        smtpdetails.PasswordGenerator
		logger         *logrus.Entry
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr bool
	}{
		{
			name: "undefined sendgrid client should cause error",
			args: args{
				sendgridClient: nil,
				apiKeyScopes:   mockAPIScopes,
				passGen:        mockPasswordGen,
				logger:         newMockLogger(),
			},
			wantErr: true,
		},
		{
			name: "undefined password generator should cause error",
			args: args{
				sendgridClient: defaultTestAPIClient(),
				apiKeyScopes:   mockAPIScopes,
				passGen:        nil,
				logger:         newMockLogger(),
			},
			wantErr: true,
		},
		{
			name: "empty scopes list should cause error",
			args: args{
				sendgridClient: defaultTestAPIClient(),
				apiKeyScopes:   []string{},
				passGen:        mockPasswordGen,
				logger:         newMockLogger(),
			},
			wantErr: true,
		},
		{
			name: "successful creation",
			args: args{
				sendgridClient: mockAPIClient,
				apiKeyScopes:   mockAPIScopes,
				passGen:        mockPasswordGen,
				logger:         newMockLogger(),
			},
			want: &Client{
				sendgridClient:              mockAPIClient,
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           mockPasswordGen,
				logger:                      newMockLogger(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.args.sendgridClient, tt.args.apiKeyScopes, tt.args.passGen, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				fmt.Print(got.logger, tt.want.logger)
				t.Errorf("NewClient() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Delete(t *testing.T) {
	type fields struct {
		sendgridClient              APIClient
		sendgridSubUserAPIKeyScopes []string
		passwordGenerator           smtpdetails.PasswordGenerator
		logger                      *logrus.Entry
	}
	type args struct {
		id string
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
				sendgridClient:              newMockAPIClient(func(c *APIClientMock) {}),
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           mockPasswordGen,
				logger:                      newMockLogger(),
			},
			args: args{id: "test"},
		},
		{
			name: "getting user fails",
			fields: fields{
				sendgridClient: newMockAPIClient(func(c *APIClientMock) {
					c.GetSubUserByUsernameFunc = func(username string) (user *SubUser, e error) {
						return nil, errors.New("test")
					}
				}),
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           mockPasswordGen,
				logger:                      newMockLogger(),
			},
			args: args{
				id: "test",
			},
			wantErr: true,
		},
		{
			name: "deleting user fails",
			fields: fields{
				sendgridClient: newMockAPIClient(func(c *APIClientMock) {
					c.DeleteSubUserFunc = func(username string) error {
						return errors.New("test")
					}
				}),
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           mockPasswordGen,
				logger:                      newMockLogger(),
			},
			args: args{
				id: "test",
			},
			wantErr: true,
		},
		{
			name: "existing user has incorrect username",
			fields: fields{
				sendgridClient:              newMockAPIClient(func(c *APIClientMock) {}),
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           mockPasswordGen,
				logger:                      newMockLogger(),
			},
			args:    args{id: "notTest"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				sendgridClient:              tt.fields.sendgridClient,
				sendgridSubUserAPIKeyScopes: tt.fields.sendgridSubUserAPIKeyScopes,
				passwordGenerator:           tt.fields.passwordGenerator,
				logger:                      tt.fields.logger,
			}
			if err := c.Delete(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_Get(t *testing.T) {
	type fields struct {
		sendgridClient              APIClient
		sendgridSubUserAPIKeyScopes []string
		passwordGenerator           smtpdetails.PasswordGenerator
		logger                      *logrus.Entry
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *smtpdetails.SMTPDetails
		wantErr bool
	}{
		{
			name: "successful get",
			fields: fields{
				sendgridClient:              newMockAPIClient(func(c *APIClientMock) {}),
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           mockPasswordGen,
				logger:                      newMockLogger(),
			},
			args: args{id: "test"},
			want: newMockSMTPDetails(),
		},
		{
			name: "getting sub user fails",
			fields: fields{
				sendgridClient: newMockAPIClient(func(c *APIClientMock) {
					c.GetSubUserByUsernameFunc = func(username string) (user *SubUser, e error) {
						return nil, errors.New("test")
					}
				}),
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           mockPasswordGen,
				logger:                      newMockLogger(),
			},
			args:    args{id: "test"},
			wantErr: true,
		},
		{
			name: "getting api key fails",
			fields: fields{
				sendgridClient: newMockAPIClient(func(c *APIClientMock) {
					c.GetAPIKeysForSubUserFunc = func(username string) (keys []*APIKey, e error) {
						return nil, errors.New("test")
					}
				}),
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           mockPasswordGen,
				logger:                      newMockLogger(),
			},
			args:    args{id: "test"},
			wantErr: true,
		},
		{
			name: "getting api keys returns empty list",
			fields: fields{
				sendgridClient: newMockAPIClient(func(c *APIClientMock) {
					c.GetAPIKeysForSubUserFunc = func(username string) (keys []*APIKey, e error) {
						return []*APIKey{}, nil
					}
				}),
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           mockPasswordGen,
				logger:                      newMockLogger(),
			},
			args:    args{id: "test"},
			wantErr: true,
		},
		{
			name: "retrieved api keys does not contain expected keys",
			fields: fields{
				sendgridClient: newMockAPIClient(func(c *APIClientMock) {
					c.GetAPIKeysForSubUserFunc = func(username string) (keys []*APIKey, e error) {
						return []*APIKey{{
							ID:     "notTest",
							Key:    "notTest",
							Name:   "notTest",
							Scopes: []string{},
						}}, nil
					}
				}),
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           mockPasswordGen,
				logger:                      newMockLogger(),
			},
			args:    args{id: "test"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				sendgridClient:              tt.fields.sendgridClient,
				sendgridSubUserAPIKeyScopes: tt.fields.sendgridSubUserAPIKeyScopes,
				passwordGenerator:           tt.fields.passwordGenerator,
				logger:                      tt.fields.logger,
			}
			got, err := c.Get(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Create(t *testing.T) {
	type fields struct {
		sendgridClient              APIClient
		sendgridSubUserAPIKeyScopes []string
		passwordGenerator           smtpdetails.PasswordGenerator
		logger                      *logrus.Entry
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *smtpdetails.SMTPDetails
		wantErr bool
	}{
		{
			name: "successful create when sub user does not exist",
			fields: fields{
				sendgridClient: newMockAPIClient(func(c *APIClientMock) {
					c.GetSubUserByUsernameFunc = func(username string) (user *SubUser, e error) {
						return nil, &NotExistError{Message: ""}
					}
					c.GetAPIKeysForSubUserFunc = func(username string) (keys []*APIKey, e error) {
						return []*APIKey{}, nil
					}
				}),
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           newMockPasswordGenerator(),
				logger:                      newMockLogger(),
			},
			args: args{id: "test"},
			want: newMockSMTPDetails(),
		},
		{
			name: "successful create when sub user does exist",
			fields: fields{
				sendgridClient: newMockAPIClient(func(c *APIClientMock) {
					c.GetAPIKeysForSubUserFunc = func(username string) (keys []*APIKey, e error) {
						return []*APIKey{}, nil
					}
				}),
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           newMockPasswordGenerator(),
				logger:                      newMockLogger(),
			},
			args: args{id: "test"},
			want: newMockSMTPDetails(),
		},
		{
			name: "getting sub user fails",
			fields: fields{
				sendgridClient: newMockAPIClient(func(c *APIClientMock) {
					c.GetSubUserByUsernameFunc = func(username string) (user *SubUser, e error) {
						return nil, errors.New("test")
					}
				}),
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           newMockPasswordGenerator(),
				logger:                      newMockLogger(),
			},
			args:    args{id: "test"},
			wantErr: true,
		},
		{
			name: "listing ip addresses fails",
			fields: fields{
				sendgridClient: newMockAPIClient(func(c *APIClientMock) {
					c.GetSubUserByUsernameFunc = func(username string) (user *SubUser, e error) {
						return nil, &smtpdetails.NotExistError{Message: ""}
					}
					c.ListIPAddressesFunc = func() (addresses []*IPAddress, e error) {
						return nil, errors.New("test")
					}
				}),
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           newMockPasswordGenerator(),
				logger:                      newMockLogger(),
			},
			args:    args{id: "test"},
			wantErr: true,
		},
		{
			name: "creating sub user fails",
			fields: fields{
				sendgridClient: newMockAPIClient(func(c *APIClientMock) {
					c.GetSubUserByUsernameFunc = func(username string) (user *SubUser, e error) {
						return nil, &smtpdetails.NotExistError{Message: ""}
					}
					c.ListIPAddressesFunc = func() (addresses []*IPAddress, e error) {
						return nil, errors.New("test")
					}
				}),
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           newMockPasswordGenerator(),
				logger:                      newMockLogger(),
			},
			args:    args{id: "test"},
			wantErr: true,
		},
		{
			name: "getting api keys for sub user fails",
			fields: fields{
				sendgridClient: newMockAPIClient(func(c *APIClientMock) {
					c.GetAPIKeysForSubUserFunc = func(username string) (keys []*APIKey, e error) {
						return nil, errors.New("")
					}
				}),
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           newMockPasswordGenerator(),
				logger:                      newMockLogger(),
			},
			args:    args{id: "test"},
			wantErr: true,
		},
		{
			name: "api key already exists",
			fields: fields{
				sendgridClient:              newMockAPIClient(func(c *APIClientMock) {}),
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           newMockPasswordGenerator(),
				logger:                      newMockLogger(),
			},
			args:    args{id: "test"},
			wantErr: true,
		},
		{
			name: "creating api key fails",
			fields: fields{
				sendgridClient: newMockAPIClient(func(c *APIClientMock) {
					c.CreateAPIKeyForSubUserFunc = func(username string, scopes []string) (key *APIKey, e error) {
						return nil, errors.New("test")
					}
				}),
				sendgridSubUserAPIKeyScopes: mockAPIScopes,
				passwordGenerator:           newMockPasswordGenerator(),
				logger:                      newMockLogger(),
			},
			args:    args{id: "test"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				sendgridClient:              tt.fields.sendgridClient,
				sendgridSubUserAPIKeyScopes: tt.fields.sendgridSubUserAPIKeyScopes,
				passwordGenerator:           tt.fields.passwordGenerator,
				logger:                      tt.fields.logger,
			}
			got, err := c.Create(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() got = %v, want %v", got, tt.want)
			}
		})
	}
}
