# AWS Endpoint Resolver Library

## Description

A minimalistic go package that helps you to test your AWS API calls in your go applications locally.


## How Does it work?

This is meant to be used in conjunction with [localstack](https://github.com/localstack/localstack) which has already implemented mocking most of the api calls in aws however it will theoretically work with any mock api that mocks all the aws api calls.

By simply setting some environment variables and then calling the `GetAWSSession()` function this will automatically create an aws session with locally resolved endpoints based on environment variables.

## How To Use

Before running your golang app simply set the following Environment Variables.

NOTE: This assumes that you have localstack or some other mock AWS API up and running somewhere either in docker-compose or locally on your machine.

* `SERVICE_ENDPOINT_MAP` - A string in the form of "serviceid1=endpoint1,serviceid2=endpoint2"
* `AWS_MOCK_REGION` DEFAULT: us-east-1 - The mock region to use 
* `AWS_MOCK_ID` DEFAULT: test-mock-id - Localstack does not verify this but only checks to make sure that it is provided 
* `AWS_MOCK_SECRET` DEFAULT: test-mock-secret - Localstack does not verify this but only checks to make sure that it is provided 


For a complete list of the service ids please see the [go aws-sdk endpoints](https://docs.aws.amazon.com/sdk-for-go/api/aws/endpoints/#pkg-constants) package.  




## Examples

### Running Locally

Specify Following Environment Variable

```
SERVICE_ENDPOINT_MAP=s3=http://<some-endpoint>:<some-port>
```

Then in your code the awsresolver will then map s3 to your custom endpoint

```
sess, err := awsresolver.GetAWSSession(aws.Config{})

// start creating services
s3Svc := s3.New(sess)
```



### docker-compose

```
localstack:
    image: localstack/localstack
    environment:
      - SERVICES: "s3"


my-app:
    image: my-app
    environment:
      SERVICE_ENDPOINT_MAP: "s3=http://localstack:<port>"      
```
