package awsresolver

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

var config struct {
	ServiceEndpoints string `envconfig:"SERVICE_ENDPOINT_MAP"`
	AWSMockRegion    string `envconfig:"AWS_MOCK_REGION" default:"us-east-1"`
	AWSMockID        string `envconfig:"AWS_MOCK_ID" default:"test-mock-id" `
	AWSMockSecret    string `envconfig:"AWS_MOCK_SECRET" default:"test-mock-secret"`
}

// GetAWSSession gets either an actual aws session or a localstack session with an endpoint resolver
// depending on if any of the endpoints are specified
func GetAWSSession(awsConfig aws.Config) (*session.Session, error) {
	if err := envconfig.Process("", &config); err != nil {
		return nil, errors.Wrap(err, "invalid configuration")
	}
	if len(config.ServiceEndpoints) > 0 {
		serviceEndpointMap := make(map[string]string)
		endpoints := strings.Split(config.ServiceEndpoints, ",")
		for _, e := range endpoints {
			endpoint := strings.Split(e, "=")
			if len(endpoint) != 2 {
				return nil, fmt.Errorf("invalid config string %v", config.ServiceEndpoints)
			}
			serviceEndpointMap[endpoint[0]] = endpoint[1]
		}
		return getCustomAWSSession(serviceEndpointMap), nil
	}
	return getRegularAWSSession(awsConfig), nil
}

// getRegularAWSSession is called whenever we aren't running localstack in docker compose
func getRegularAWSSession(awsConfig aws.Config) *session.Session {
	return session.Must(session.NewSessionWithOptions(session.Options{
		Config: awsConfig,
	}))
}

// getCustomAWSSession is called whenever our custom endpoints are specified and we make a mock aws session.
func getCustomAWSSession(serviceEndpointMap map[string]string) *session.Session {
	// create the resolver so it knows what our endpoints are
	resolver := func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		if endpoint, ok := serviceEndpointMap[service]; ok {
			return endpoints.ResolvedEndpoint{
				URL:           endpoint,
				SigningRegion: config.AWSMockRegion,
			}, nil
		}

		return endpoints.DefaultResolver().EndpointFor(service, region, optFns...)
	}

	// return the session
	// disabling the ssl is on because localstack won't use it by default unless you tell it to
	// credentials have to be supplied but it doesn't matter what their values are
	// otherwise localstack will throw an error
	return session.Must(session.NewSession(&aws.Config{
		Region:                        &config.AWSMockRegion,
		EndpointResolver:              endpoints.ResolverFunc(resolver),
		CredentialsChainVerboseErrors: aws.Bool(true),
		S3ForcePathStyle:              aws.Bool(true),
		Credentials:                   credentials.NewStaticCredentials(config.AWSMockID, config.AWSMockSecret, ""),
	}))

}
