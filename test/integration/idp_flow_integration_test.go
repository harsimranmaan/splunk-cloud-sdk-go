/*
 * Copyright 2019 Splunk, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"): you may
 * not use this file except in compliance with the License. You may obtain
 * a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations
 * under the License.
 */

package integration

import (
	"os"

	"github.com/splunk/splunk-cloud-sdk-go/v2/sdk"
	"github.com/splunk/splunk-cloud-sdk-go/v2/services"

	"testing"

	"github.com/splunk/splunk-cloud-sdk-go/v2/idp"
	"github.com/splunk/splunk-cloud-sdk-go/v2/services/identity"
	"github.com/splunk/splunk-cloud-sdk-go/v2/services/ingest"
	testutils "github.com/splunk/splunk-cloud-sdk-go/v2/test/utils"
	"github.com/splunk/splunk-cloud-sdk-go/v2/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RefreshToken - RefreshToken to refresh the bearer token if expired
var RefreshToken = os.Getenv("REFRESH_TOKEN_UPDATED_VERSION")

var IdpHost = os.Getenv("IDP_HOST")

// NativeClientID - App Registry Client Id for SDK Native App
var NativeClientID = os.Getenv("REFRESH_TOKEN_CLIENT_ID")

// NativeAppRedirectURI is one of the redirect uris configured for the app
const NativeAppRedirectURI = "https://login.splunkbeta.com"

// BackendClientID - App Registry Client id for client credentials flow
var BackendClientID = os.Getenv("BACKEND_CLIENT_ID")

// BackendClientSecret - App Registry Client secret for client credentials flow
var BackendClientSecret = os.Getenv("BACKEND_CLIENT_SECRET")

// BackendServiceScope - scope for obtaining access token for client credentials flow
const BackendServiceScope = ""

// TestUsername corresponds to the test user for integration testing
var TestUsername = os.Getenv("TEST_USERNAME")

// TestPassword corresponds to the test user's password for integration testing
var TestPassword = os.Getenv("TEST_PASSWORD")

type retryTokenRetriever struct {
	TR idp.TokenRetriever
	n  int
}

func (r *retryTokenRetriever) GetTokenContext() (*idp.Context, error) {
	r.n++
	// Return a bad access token the first time for testing 401 retry logic
	if r.n == 1 {
		return &idp.Context{AccessToken: testutils.ExpiredAuthenticationToken}, nil
	}
	// For subsequent requests get the real token using the real TokenRetriever
	return r.TR.GetTokenContext()
}

type badTokenRetriever struct {
	N int
}

func (r *badTokenRetriever) GetTokenContext() (*idp.Context, error) {
	r.N++
	// Return a bad access token every time
	return &idp.Context{AccessToken: testutils.ExpiredAuthenticationToken}, nil
}

// TestIntegrationRefreshTokenInitWorkflow tests initializing the client with a TokenRetriever impleme
func TestIntegrationRefreshTokenInitWorkflow(t *testing.T) {
	// get a new refresh token
	tr := idp.NewPKCERetriever(NativeClientID, NativeAppRedirectURI, idp.DefaultRefreshScope, TestUsername, TestPassword, IdpHost)
	ctx, err := tr.GetTokenContext()
	require.Emptyf(t, err, "Error validating using access token generated from PKCE flow: %s", err)
	require.NotNil(t, ctx)

	tr_refresh := idp.NewRefreshTokenRetriever(NativeClientID, idp.DefaultRefreshScope, ctx.RefreshToken, IdpHost)
	client, err := sdk.NewClient(&services.Config{
		TokenRetriever: tr_refresh,
		Host:           testutils.TestSplunkCloudHost,
		Tenant:         testutils.TestTenant,
		Timeout:        testutils.TestTimeOut,
	})
	require.Emptyf(t, err, "Error initializing client: %s", err)

	input := identity.ValidateTokenQueryParams{Include: []string{"principal", "tenant"}}
	info, err := client.IdentityService.ValidateToken(&input)
	assert.Emptyf(t, err, "Error validating using access token generated from refresh token: %s", err)
	assert.NotNil(t, info)
}

// TestIntegrationRefreshTokenRetryWorkflow tests ingesting event with invalid access token then retrying after obtaining new access token with refresh token
func TestIntegrationRefreshTokenRetryWorkflow(t *testing.T) {
	// get a new refresh token
	tr := idp.NewPKCERetriever(NativeClientID, NativeAppRedirectURI, idp.DefaultRefreshScope, TestUsername, TestPassword, IdpHost)
	ctx, err := tr.GetTokenContext()
	require.Emptyf(t, err, "Error validating using access token generated from PKCE flow: %s", err)
	require.NotNil(t, ctx)

	tr_refresh := &retryTokenRetriever{TR: idp.NewRefreshTokenRetriever(NativeClientID, idp.DefaultRefreshScope, ctx.RefreshToken, IdpHost)}
	client, err := sdk.NewClient(&services.Config{
		TokenRetriever: tr_refresh,
		Host:           testutils.TestSplunkCloudHost,
		Tenant:         testutils.TestTenant,
		Timeout:        testutils.TestTimeOut,
	})
	require.Emptyf(t, err, "Error initializing client: %s", err)

	sourcetype := "sourcetype:refreshtokentest"
	source := "manual-events"
	host := client.GetURL("").RequestURI()
	body := make(map[string]interface{})
	body["event"] = "refreshtokentest"
	timeValue := int64(1529945001)
	attributes := make(map[string]interface{})
	attributes1 := make(map[string]interface{})
	attributes1["testKey"] = "testValue"
	attributes["testkey2"] = attributes1

	testIngestEvent := ingest.Event{
		Host:       &host,
		Body:       body,
		Sourcetype: &sourcetype,
		Source:     &source,
		Timestamp:  &timeValue,
		Attributes: attributes}

	_, err = client.IngestService.PostEvents([]ingest.Event{testIngestEvent})
	assert.Emptyf(t, err, "Error ingesting test event using refresh token: %s", err)
}

// TestIntegrationClientCredentialsInitWorkflow tests initializing the client with a TokenRetriever impleme
func TestIntegrationClientCredentialsInitWorkflow(t *testing.T) {
	tr := idp.NewClientCredentialsRetriever(BackendClientID, BackendClientSecret, BackendServiceScope, IdpHost)
	client, err := sdk.NewClient(&services.Config{
		TokenRetriever: tr,
		Host:           testutils.TestSplunkCloudHost,
		Tenant:         testutils.TestTenant,
		Timeout:        testutils.TestTimeOut,
	})
	require.Emptyf(t, err, "Error initializing client: %s", err)

	input := identity.ValidateTokenQueryParams{Include: []string{"principal", "tenant"}}
	_, err = client.IdentityService.ValidateToken(&input)
	assert.Emptyf(t, err, "Error validating using access token generated from client credentials: %s", err)
}

// TestIntegrationClientCredentialsRetryWorkflow tests ingesting event with invalid access token then retrying after obtaining new access token with client credentials flow
func TestIntegrationClientCredentialsRetryWorkflow(t *testing.T) {
	tr := &retryTokenRetriever{TR: idp.NewClientCredentialsRetriever(BackendClientID, BackendClientSecret, BackendServiceScope, IdpHost)}
	client, err := sdk.NewClient(&services.Config{
		TokenRetriever: tr,
		Host:           testutils.TestSplunkCloudHost,
		Tenant:         testutils.TestTenant,
		Timeout:        testutils.TestTimeOut,
	})
	require.Emptyf(t, err, "Error initializing client: %s", err)

	// Make sure the backend client id has been added to the tenant as an admin (authz is needed for ingest), errs are ignored - if either fails (e.g. for 405 duplicate) we are probably still OK
	_, _ = getClient(t).IdentityService.AddMember(identity.AddMemberBody{Name: BackendClientID})
	_, _ = getClient(t).IdentityService.AddGroupMember("tenant.admins", identity.AddGroupMemberBody{Name: BackendClientID})

	sourcetype := "sourcetype:clientcredentialstest"
	source := "manual-events"
	host := client.GetURL("").RequestURI()
	body := make(map[string]interface{})
	body["event"] = "clientcredentialstest"
	timeValue := int64(1529945001)
	attributes := make(map[string]interface{})
	attributes1 := make(map[string]interface{})
	attributes1["testKey"] = "testValue"
	attributes["testkey2"] = attributes1

	testIngestEvent := ingest.Event{
		Host:       &host,
		Body:       body,
		Sourcetype: &sourcetype,
		Source:     &source,
		Timestamp:  &timeValue,
		Attributes: attributes}

	_, err = client.IngestService.PostEvents([]ingest.Event{testIngestEvent})
	assert.Emptyf(t, err, "Error ingesting test event using client credentials flow error: %s", err)
}

// TestIntegrationPKCEInitWorkflow tests initializing the client with a TokenRetriever which obtains a new access token with PKCE flow
func TestIntegrationPKCEInitWorkflow(t *testing.T) {
	tr := idp.NewPKCERetriever(NativeClientID, NativeAppRedirectURI, idp.DefaultOIDCScopes, TestUsername, TestPassword, IdpHost)
	client, err := sdk.NewClient(&services.Config{
		TokenRetriever: tr,
		Host:           testutils.TestSplunkCloudHost,
		Tenant:         testutils.TestTenant,
		Timeout:        testutils.TestTimeOut,
	})
	require.Emptyf(t, err, "Error initializing client: %s", err)

	input := identity.ValidateTokenQueryParams{Include: []string{"principal", "tenant"}}
	_, err = client.IdentityService.ValidateToken(&input)
	assert.Emptyf(t, err, "Error validating using access token generated from PKCE flow: %s", err)
}

// TestIntegrationPKCERetryWorkflow tests ingesting event with invalid access token then retrying after obtaining new access token with PKCE flow
func TestIntegrationPKCERetryWorkflow(t *testing.T) {
	tr := &retryTokenRetriever{TR: idp.NewPKCERetriever(NativeClientID, NativeAppRedirectURI, idp.DefaultOIDCScopes, TestUsername, TestPassword, IdpHost)}

	client, err := sdk.NewClient(&services.Config{
		TokenRetriever: tr,
		Host:           testutils.TestSplunkCloudHost,
		Tenant:         testutils.TestTenant,
		Timeout:        testutils.TestTimeOut,
	})
	require.Emptyf(t, err, "Error initializing client: %s", err)

	sourcetype := "sourcetype:clientcredentialstest"
	source := "manual-events"
	host := client.GetURL("").RequestURI()
	body := make(map[string]interface{})
	body["event"] = "clientcredentialstest"
	timeValue := int64(1529945001)
	attributes := make(map[string]interface{})
	attributes1 := make(map[string]interface{})
	attributes1["testKey"] = "testValue"
	attributes["testkey2"] = attributes1

	testIngestEvent := ingest.Event{
		Host:       &host,
		Body:       body,
		Sourcetype: &sourcetype,
		Source:     &source,
		Timestamp:  &timeValue,
		Attributes: attributes}

	_, err = client.IngestService.PostEvents([]ingest.Event{testIngestEvent})
	assert.Emptyf(t, err, "Error ingesting test event using PKCE flow error: %s", err)
}

// TestBadTokenRetryWorkflow tests to make sure that a 401 is returned to the end user when a bad token is retrieved and requests are re-tried exactly once
func TestBadTokenRetryWorkflow(t *testing.T) {
	tr := &badTokenRetriever{}

	client, err := sdk.NewClient(&services.Config{
		TokenRetriever: tr,
		Host:           testutils.TestSplunkCloudHost,
		Tenant:         testutils.TestTenant,
		Timeout:        testutils.TestTimeOut,
	})
	require.Emptyf(t, err, "Error initializing client: %s", err)

	sourcetype := "sourcetype:badtokentest"
	source := "manual-events"
	host := client.GetURL("").RequestURI()
	body := make(map[string]interface{})
	body["event"] = "badtokentest"
	timeValue := int64(1529945001)
	attributes := make(map[string]interface{})
	attributes1 := make(map[string]interface{})
	attributes1["testKey"] = "testValue"
	attributes["testkey2"] = attributes1

	testIngestEvent := ingest.Event{
		Host:       &host,
		Body:       body,
		Sourcetype: &sourcetype,
		Source:     &source,
		Timestamp:  &timeValue,
		Attributes: attributes}

	_, err = client.IngestService.PostEvents([]ingest.Event{testIngestEvent})
	assert.Equal(t, tr.N, 2, "Expected exactly two calls to TokenRetriever.GetTokenContext(): 1) at client initialization and 2) after 401 is encountered when client.IngestService.CreateEvent is called")
	require.NotNil(t, err)
	httpErr, ok := err.(*util.HTTPError)
	require.True(t, ok, "Expected err to be util.HTTPError")
	assert.True(t, httpErr.HTTPStatusCode == 401, "Expected error code 401 for multiple attempts with expired access tokens")
}
