package sendgrid

import (
	"fmt"
	"os"
	"strings"

	"github.com/integr8ly/smtp-service/pkg/smtpdetails"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-password/password"
	"github.com/sirupsen/logrus"
)

var (
	DefaultAPIKeyScopes = []string{"mail.send"}
)

var _ smtpdetails.Client = Client{}

type Client struct {
	sendgridClient              APIClient
	sendgridSubUserAPIKeyScopes []string
	passwordGenerator           smtpdetails.PasswordGenerator
	logger                      *logrus.Entry
}

//NewDefaultClient Create new client using API key from SENDGRID_API_KEY env var and the default SendGrid API host.
func NewDefaultClient(logger *logrus.Entry) (*Client, error) {
	passGen, err := password.NewGenerator(&password.GeneratorInput{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create default password generator")
	}
	sendgridRESTClient := NewBackendRESTClient(APIHost, os.Getenv(EnvApiKey), logger)
	sendgridClient := NewBackendAPIClient(sendgridRESTClient, logger)
	return NewClient(sendgridClient, DefaultAPIKeyScopes, passGen, logger.WithField(smtpdetails.LogFieldDetailProvider, ProviderName))
}

func NewClient(sendgridClient APIClient, apiKeyScopes []string, passGen smtpdetails.PasswordGenerator, logger *logrus.Entry) (*Client, error) {
	if sendgridClient == nil {
		return nil, errors.New("sendgridClient must be defined")
	}
	if len(apiKeyScopes) == 0 {
		return nil, errors.New("apiKeyScopes should be a non-empty list")
	}
	if passGen == nil {
		return nil, errors.New("passGen must be defined")
	}
	return &Client{
		sendgridClient:              sendgridClient,
		sendgridSubUserAPIKeyScopes: apiKeyScopes,
		passwordGenerator:           passGen,
		logger:                      logger,
	}, nil
}

func (c Client) Create(id string) (*smtpdetails.SMTPDetails, error) {
	// check if sub user exists
	c.logger.Infof("checking if sub user %s exists", id)
	subuser, err := c.sendgridClient.GetSubUserByUsername(id)
	if err != nil && !smtpdetails.IsNotExistError(err) {
		return nil, errors.Wrapf(err, "failed to check if sub user already exists")
	}
	// sub user doesn't exist, create it
	if subuser == nil {
		c.logger.Debugf("could not find existing user with username %s, creating it", id)
		// get an ip address from the sendgrid account to assign to the sub user
		ips, err := c.sendgridClient.ListIPAddresses()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list ip addresses")
		}
		if len(ips) < 1 {
			return nil, errors.New("no ip addresses found to assign to sub user")
		}
		ipAddr := ips[0]
		// if id isn't already an email, lazily convert it to one
		idEmail := id
		if !strings.Contains(id, "@") {
			idEmail = fmt.Sprintf("%s@email.com", id)
		}
		// handle password generation
		c.logger.Debugf("generating password for new sub user %s", id)
		password, err := c.passwordGenerator.Generate(10, 1, 1, false, true)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate password for sub user")
		}
		subuser, err = c.sendgridClient.CreateSubUser(id, idEmail, password, []string{ipAddr.IP})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create sub user")
		}
		c.logger.Infof("sub user created with details, username=%s email=%s password=%s", id, idEmail, password)
	} else {
		c.logger.Infof("sub user %s already exists, skipping creation", id)
	}
	// check if api key for sub user exists
	c.logger.Infof("checking if api key for sub user %s already exists", id)
	apiKeys, err := c.sendgridClient.GetAPIKeysForSubUser(id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to check if api key already exists")
	}
	var apiKey *APIKey
	for _, k := range apiKeys {
		if k.Name == id {
			apiKey = k
			break
		}
	}
	if apiKey != nil {
		return nil, &smtpdetails.AlreadyExistsError{Message: fmt.Sprintf("api key %s for sub user %s already exists", apiKey.Name, subuser.Username)}
	}
	// api key doesn't exist, create it
	c.logger.Infof("no api key found, creating api key for sub user %s", id)
	apiKey, err = c.sendgridClient.CreateAPIKeyForSubUser(subuser.Username, DefaultAPIKeyScopes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create api key for sub user")
	}
	return defaultConnectionDetails(apiKey.Name, apiKey.Key), nil
}

func (c Client) Get(id string) (*smtpdetails.SMTPDetails, error) {
	subuser, err := c.sendgridClient.GetSubUserByUsername(id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get user by username, %s", id)
	}
	c.logger.Debugf("found user with username %s, id=%d email=%s disabled=%t", subuser.Username, subuser.ID, subuser.Email, subuser.Disabled)
	apiKeys, err := c.sendgridClient.GetAPIKeysForSubUser(subuser.Username)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get api keys for sub user with username %s", subuser.Username)
	}
	if len(apiKeys) < 1 {
		return nil, errors.New(fmt.Sprintf("no api keys found for sub user %s", id))
	}
	var clusterApiKey *APIKey
	for _, k := range apiKeys {
		if k.Name == subuser.Username {
			clusterApiKey = k
			break
		}
	}
	if clusterApiKey == nil {
		return nil, &smtpdetails.NotExistError{Message: fmt.Sprintf("api key with id %s does not exist for sub user %s", subuser.Username, subuser.Username)}
	}
	return defaultConnectionDetails(clusterApiKey.Name, clusterApiKey.Key), nil
}

func (c Client) Delete(id string) error {
	c.logger.Debugf("checking if sub user %s exists", id)
	subuser, err := c.sendgridClient.GetSubUserByUsername(id)
	if err != nil {
		return errors.Wrapf(err, "failed to check if sub user exists")
	}
	if subuser.Username != id {
		return errors.New(fmt.Sprintf("found user does not have expected username, expected=%s found=%s", id, subuser.Username))
	}
	c.logger.Debugf("sub user %s exists, deleting it", subuser.Username)
	if err := c.sendgridClient.DeleteSubUser(subuser.Username); err != nil {
		return errors.Wrapf(err, "failed to delete sub user %s", id)
	}
	return nil
}

func defaultConnectionDetails(apiKeyID, apiKey string) *smtpdetails.SMTPDetails {
	return &smtpdetails.SMTPDetails{
		ID:       apiKeyID,
		Host:     ConnectionDetailsHost,
		Port:     ConnectionDetailsPort,
		TLS:      ConnectionDetailsTLS,
		Username: ConnectionDetailsUsername,
		Password: apiKey,
	}
}
