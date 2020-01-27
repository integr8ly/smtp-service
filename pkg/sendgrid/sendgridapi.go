package sendgrid

import (
	"encoding/json"
	"fmt"

	"github.com/integr8ly/smtp-service/pkg/smtpdetails"

	"github.com/pkg/errors"
	"github.com/sendgrid/rest"
	"github.com/sirupsen/logrus"
)

var _ APIClient = &BackendAPIClient{}

//go:generate moq -out sendgridapi_moq.go . APIClient
type APIClient interface {
	// ip addresses
	ListIPAddresses() ([]*IPAddress, error)
	// api keys
	GetAPIKeysForSubUser(username string) ([]*APIKey, error)
	CreateAPIKeyForSubUser(username string, scopes []string) (*APIKey, error)
	// sub users
	CreateSubUser(id, email, password string, ips []string) (*SubUser, error)
	DeleteSubUser(username string) error
	ListSubUsers(query map[string]string) ([]*SubUser, error)
	GetSubUserByUsername(username string) (*SubUser, error)
}

//apiKeysListResponse A fix for the irregular api keys list response, with format { "results": [] }
type apiKeysListResponse struct {
	Result []*APIKey `json:"result"`
}

//BackendAPIClient Light wrapper around the default SendGrid library to allow for mocking
type BackendAPIClient struct {
	restClient RESTClient
	logger     *logrus.Entry
}

//NewBackendClient Create a new BackendAPIClient with default logger labels
func NewBackendAPIClient(restClient RESTClient, logger *logrus.Entry) *BackendAPIClient {
	return &BackendAPIClient{restClient: restClient, logger: logger.WithField(LogFieldAPIClient, ProviderName)}
}

func (c *BackendAPIClient) ListIPAddresses() ([]*IPAddress, error) {
	listReq := c.restClient.BuildRequest(APIRouteIPAddresses, rest.Get)
	listResp, err := c.restClient.InvokeRequest(listReq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list ip addresses")
	}
	var ips []*IPAddress
	if err = json.Unmarshal([]byte(listResp.Body), &ips); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal ip address response, content=%s", listResp.Body)
	}
	return ips, nil
}

func (c *BackendAPIClient) GetAPIKeysForSubUser(username string) ([]*APIKey, error) {
	if username == "" {
		return nil, errors.New("username must be a non-empty string")
	}
	listReq := c.restClient.BuildRequest(APIRouteAPIKeys, rest.Get)
	listReq.Headers[HeaderOnBehalfOf] = username
	listResp, err := c.restClient.InvokeRequest(listReq)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list api keys for user %s", username)
	}
	var apiKeysResp apiKeysListResponse
	if err := json.Unmarshal([]byte(listResp.Body), &apiKeysResp); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal api keys response, content=%s", listResp.Body)
	}
	return apiKeysResp.Result, nil
}

func (c *BackendAPIClient) CreateAPIKeyForSubUser(username string, scopes []string) (*APIKey, error) {
	if username == "" {
		return nil, errors.New("username must be a non-empty string")
	}
	createReq := c.restClient.BuildRequest(APIRouteAPIKeys, rest.Post)
	createReq.Headers[HeaderOnBehalfOf] = username
	createBody, err := buildCreateApiKeyBody(username, scopes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create api key request body")
	}
	createReq.Body = createBody
	createResp, err := c.restClient.InvokeRequest(createReq)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create api key for user %s", username)
	}
	var apiKey *APIKey
	if err = json.Unmarshal([]byte(createResp.Body), &apiKey); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal api key response, content=%s", createResp.Body)
	}
	return apiKey, nil
}

func (c *BackendAPIClient) CreateSubUser(id, email, password string, ips []string) (*SubUser, error) {
	createReq := c.restClient.BuildRequest(APIRouteSubUsers, rest.Post)
	createReqBody, err := buildCreateSubUserBody(id, email, password, ips)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create sub user request body")
	}
	createReq.Body = createReqBody
	createResp, err := c.restClient.InvokeRequest(createReq)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create sub user %s", id)
	}
	if createResp.StatusCode != 201 {
		return nil, errors.New(fmt.Sprintf("non-201 status code returned, code=%d body=%s", createResp.StatusCode, createResp.Body))
	}
	var subuser *SubUser
	if err = json.Unmarshal([]byte(createResp.Body), &subuser); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal sub user request, content=%s", createResp.Body)
	}
	c.logger.Debug(createResp.Body, createResp.StatusCode)
	return subuser, nil
}

func (c *BackendAPIClient) DeleteSubUser(username string) error {
	if username == "" {
		return errors.New("username must be a non-empty string")
	}
	deleteReq := c.restClient.BuildRequest(fmt.Sprintf("%s/%s", APIRouteSubUsers, username), rest.Delete)
	deleteResp, err := c.restClient.InvokeRequest(deleteReq)
	if err != nil {
		return errors.Wrapf(err, "failed to delete sub user %s", username)
	}
	if deleteResp.StatusCode != 204 {
		return errors.New(fmt.Sprintf("non-204 status code returned, code=%d body=%s", deleteResp.StatusCode, deleteResp.Body))
	}
	return nil
}

func (c *BackendAPIClient) ListSubUsers(query map[string]string) ([]*SubUser, error) {
	if query == nil {
		query = map[string]string{}
	}
	listReq := c.restClient.BuildRequest(APIRouteSubUsers, rest.Get)
	listReq.QueryParams = query
	listResp, err := c.restClient.InvokeRequest(listReq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list sub users")
	}
	var subusers []*SubUser
	if err = json.Unmarshal([]byte(listResp.Body), &subusers); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal sub users, content=%s", listResp.Body)
	}
	return subusers, nil
}

func (c *BackendAPIClient) GetSubUserByUsername(username string) (*SubUser, error) {
	if username == "" {
		return nil, errors.New("username must be a non-empty string")
	}
	subusers, err := c.ListSubUsers(map[string]string{"username": username})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list sub users with username %s", username)
	}
	if len(subusers) != 1 {
		return nil, &smtpdetails.NotExistError{Message: fmt.Sprintf("should be exactly one sub user with username %s, found %d", username, len(subusers))}
	}
	return subusers[0], nil
}
