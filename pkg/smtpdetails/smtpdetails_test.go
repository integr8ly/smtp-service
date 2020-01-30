package smtpdetails

import (
	"reflect"
	"strconv"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "k8s.io/api/core/v1"
)

const (
	mockID       = "test"
	mockHost     = "smtp.test.com"
	mockPort     = 587
	mockTLS      = true
	mockUsername = "test"
	mockPassword = "test"
)

func newMockSMTPDetails() *SMTPDetails {
	return &SMTPDetails{
		ID:       mockID,
		Host:     mockHost,
		Port:     mockPort,
		TLS:      mockTLS,
		Username: mockUsername,
		Password: mockPassword,
	}
}

func TestConvertSMTPDetailsToSecret(t *testing.T) {
	type args struct {
		smtpDetails *SMTPDetails
		secretName  string
	}
	tests := []struct {
		name string
		args args
		want *apiv1.Secret
	}{
		{
			name: "successful convert",
			args: args{
				smtpDetails: newMockSMTPDetails(),
				secretName:  "testSec",
			},
			want: &apiv1.Secret{
				TypeMeta: v1.TypeMeta{
					Kind:       SecretGVKKind,
					APIVersion: SecretGVKVersion,
				},
				ObjectMeta: v1.ObjectMeta{
					Name: "testSec",
				},
				Data: map[string][]byte{
					SecretKeyPassword: []byte(mockPassword),
					SecretKeyUsername: []byte(mockUsername),
					SecretKeyTLS:      []byte(strconv.FormatBool(mockTLS)),
					SecretKeyPort:     []byte(strconv.Itoa(mockPort)),
					SecretKeyHost:     []byte(mockHost),
				},
				Type: apiv1.SecretTypeOpaque,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertSMTPDetailsToSecret(tt.args.smtpDetails, tt.args.secretName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertSMTPDetailsToSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}
