/*
LaunchDarkly REST API

# Overview  ## Authentication  LaunchDarkly's REST API uses the HTTPS protocol with a minimum TLS version of 1.2.  All REST API resources are authenticated with either [personal or service access tokens](https://docs.launchdarkly.com/home/account-security/api-access-tokens), or session cookies. Other authentication mechanisms are not supported. You can manage personal access tokens on your [**Account settings**](https://app.launchdarkly.com/settings/tokens) page.  LaunchDarkly also has SDK keys, mobile keys, and client-side IDs that are used by our server-side SDKs, mobile SDKs, and JavaScript-based SDKs, respectively. **These keys cannot be used to access our REST API**. These keys are environment-specific, and can only perform read-only operations such as fetching feature flag settings.  | Auth mechanism                                                                                  | Allowed resources                                                                                     | Use cases                                          | | ----------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- | -------------------------------------------------- | | [Personal or service access tokens](https://docs.launchdarkly.com/home/account-security/api-access-tokens) | Can be customized on a per-token basis                                                                | Building scripts, custom integrations, data export. | | SDK keys                                                                                        | Can only access read-only resources specific to server-side SDKs. Restricted to a single environment. | Server-side SDKs                     | | Mobile keys                                                                                     | Can only access read-only resources specific to mobile SDKs, and only for flags marked available to mobile keys. Restricted to a single environment.           | Mobile SDKs                                        | | Client-side ID                                                                                  | Can only access read-only resources specific to JavaScript-based client-side SDKs, and only for flags marked available to client-side. Restricted to a single environment.           | Client-side JavaScript                             |  > #### Keep your access tokens and SDK keys private > > Access tokens should _never_ be exposed in untrusted contexts. Never put an access token in client-side JavaScript, or embed it in a mobile application. LaunchDarkly has special mobile keys that you can embed in mobile apps. If you accidentally expose an access token or SDK key, you can reset it from your [**Account settings**](https://app.launchdarkly.com/settings/tokens) page. > > The client-side ID is safe to embed in untrusted contexts. It's designed for use in client-side JavaScript.  ### Authentication using request header  The preferred way to authenticate with the API is by adding an `Authorization` header containing your access token to your requests. The value of the `Authorization` header must be your access token.  Manage personal access tokens from the [**Account settings**](https://app.launchdarkly.com/settings/tokens) page.  ### Authentication using session cookie  For testing purposes, you can make API calls directly from your web browser. If you are logged in to the LaunchDarkly application, the API will use your existing session to authenticate calls.  If you have a [role](https://docs.launchdarkly.com/home/team/built-in-roles) other than Admin, or have a [custom role](https://docs.launchdarkly.com/home/team/custom-roles) defined, you may not have permission to perform some API calls. You will receive a `401` response code in that case.  > ### Modifying the Origin header causes an error > > LaunchDarkly validates that the Origin header for any API request authenticated by a session cookie matches the expected Origin header. The expected Origin header is `https://app.launchdarkly.com`. > > If the Origin header does not match what's expected, LaunchDarkly returns an error. This error can prevent the LaunchDarkly app from working correctly. > > Any browser extension that intentionally changes the Origin header can cause this problem. For example, the `Allow-Control-Allow-Origin: *` Chrome extension changes the Origin header to `http://evil.com` and causes the app to fail. > > To prevent this error, do not modify your Origin header. > > LaunchDarkly does not require origin matching when authenticating with an access token, so this issue does not affect normal API usage.  ## Representations  All resources expect and return JSON response bodies. Error responses also send a JSON body. To learn more about the error format of the API, read [Errors](/#section/Overview/Errors).  In practice this means that you always get a response with a `Content-Type` header set to `application/json`.  In addition, request bodies for `PATCH`, `POST`, and `PUT` requests must be encoded as JSON with a `Content-Type` header set to `application/json`.  ### Summary and detailed representations  When you fetch a list of resources, the response includes only the most important attributes of each resource. This is a _summary representation_ of the resource. When you fetch an individual resource, such as a single feature flag, you receive a _detailed representation_ of the resource.  The best way to find a detailed representation is to follow links. Every summary representation includes a link to its detailed representation.  ### Expanding responses  Sometimes the detailed representation of a resource does not include all of the attributes of the resource by default. If this is the case, the request method will clearly document this and describe which attributes you can include in an expanded response.  To include the additional attributes, append the `expand` request parameter to your request and add a comma-separated list of the attributes to include. For example, when you append `?expand=members,roles` to the [Get team](/tag/Teams#operation/getTeam) endpoint, the expanded response includes both of these attributes.  ### Links and addressability  The best way to navigate the API is by following links. These are attributes in representations that link to other resources. The API always uses the same format for links:  - Links to other resources within the API are encapsulated in a `_links` object - If the resource has a corresponding link to HTML content on the site, it is stored in a special `_site` link  Each link has two attributes:  - An `href`, which contains the URL - A `type`, which describes the content type  For example, a feature resource might return the following:  ```json {   \"_links\": {     \"parent\": {       \"href\": \"/api/features\",       \"type\": \"application/json\"     },     \"self\": {       \"href\": \"/api/features/sort.order\",       \"type\": \"application/json\"     }   },   \"_site\": {     \"href\": \"/features/sort.order\",     \"type\": \"text/html\"   } } ```  From this, you can navigate to the parent collection of features by following the `parent` link, or navigate to the site page for the feature by following the `_site` link.  Collections are always represented as a JSON object with an `items` attribute containing an array of representations. Like all other representations, collections have `_links` defined at the top level.  Paginated collections include `first`, `last`, `next`, and `prev` links containing a URL with the respective set of elements in the collection.  ## Updates  Resources that accept partial updates use the `PATCH` verb. Most resources support the [JSON patch](/reference#updates-using-json-patch) format. Some resources also support the [JSON merge patch](/reference#updates-using-json-merge-patch) format, and some resources support the [semantic patch](/reference#updates-using-semantic-patch) format, which is a way to specify the modifications to perform as a set of executable instructions. Each resource supports optional [comments](/reference#updates-with-comments) that you can submit with updates. Comments appear in outgoing webhooks, the audit log, and other integrations.  When a resource supports both JSON patch and semantic patch, we document both in the request method. However, the specific request body fields and descriptions included in our documentation only match one type of patch or the other.  ### Updates using JSON patch  [JSON patch](https://datatracker.ietf.org/doc/html/rfc6902) is a way to specify the modifications to perform on a resource. JSON patch uses paths and a limited set of operations to describe how to transform the current state of the resource into a new state. JSON patch documents are always arrays, where each element contains an operation, a path to the field to update, and the new value.  For example, in this feature flag representation:  ```json {     \"name\": \"New recommendations engine\",     \"key\": \"engine.enable\",     \"description\": \"This is the description\",     ... } ``` You can change the feature flag's description with the following patch document:  ```json [{ \"op\": \"replace\", \"path\": \"/description\", \"value\": \"This is the new description\" }] ```  You can specify multiple modifications to perform in a single request. You can also test that certain preconditions are met before applying the patch:  ```json [   { \"op\": \"test\", \"path\": \"/version\", \"value\": 10 },   { \"op\": \"replace\", \"path\": \"/description\", \"value\": \"The new description\" } ] ```  The above patch request tests whether the feature flag's `version` is `10`, and if so, changes the feature flag's description.  Attributes that are not editable, such as a resource's `_links`, have names that start with an underscore.  ### Updates using JSON merge patch  [JSON merge patch](https://datatracker.ietf.org/doc/html/rfc7386) is another format for specifying the modifications to perform on a resource. JSON merge patch is less expressive than JSON patch. However, in many cases it is simpler to construct a merge patch document. For example, you can change a feature flag's description with the following merge patch document:  ```json {   \"description\": \"New flag description\" } ```  ### Updates using semantic patch  Some resources support the semantic patch format. A semantic patch is a way to specify the modifications to perform on a resource as a set of executable instructions.  Semantic patch allows you to be explicit about intent using precise, custom instructions. In many cases, you can define semantic patch instructions independently of the current state of the resource. This can be useful when defining a change that may be applied at a future date.  To make a semantic patch request, you must append `domain-model=launchdarkly.semanticpatch` to your `Content-Type` header.  Here's how:  ``` Content-Type: application/json; domain-model=launchdarkly.semanticpatch ```  If you call a semantic patch resource without this header, you will receive a `400` response because your semantic patch will be interpreted as a JSON patch.  The body of a semantic patch request takes the following properties:  * `comment` (string): (Optional) A description of the update. * `environmentKey` (string): (Required for some resources only) The environment key. * `instructions` (array): (Required) A list of actions the update should perform. Each action in the list must be an object with a `kind` property that indicates the instruction. If the instruction requires parameters, you must include those parameters as additional fields in the object. The documentation for each resource that supports semantic patch includes the available instructions and any additional parameters.  For example:  ```json {   \"comment\": \"optional comment\",   \"instructions\": [ {\"kind\": \"turnFlagOn\"} ] } ```  If any instruction in the patch encounters an error, the endpoint returns an error and will not change the resource. In general, each instruction silently does nothing if the resource is already in the state you request.  ### Updates with comments  You can submit optional comments with `PATCH` changes.  To submit a comment along with a JSON patch document, use the following format:  ```json {   \"comment\": \"This is a comment string\",   \"patch\": [{ \"op\": \"replace\", \"path\": \"/description\", \"value\": \"The new description\" }] } ```  To submit a comment along with a JSON merge patch document, use the following format:  ```json {   \"comment\": \"This is a comment string\",   \"merge\": { \"description\": \"New flag description\" } } ```  To submit a comment along with a semantic patch, use the following format:  ```json {   \"comment\": \"This is a comment string\",   \"instructions\": [ {\"kind\": \"turnFlagOn\"} ] } ```  ## Errors  The API always returns errors in a common format. Here's an example:  ```json {   \"code\": \"invalid_request\",   \"message\": \"A feature with that key already exists\",   \"id\": \"30ce6058-87da-11e4-b116-123b93f75cba\" } ```  The `code` indicates the general class of error. The `message` is a human-readable explanation of what went wrong. The `id` is a unique identifier. Use it when you're working with LaunchDarkly Support to debug a problem with a specific API call.  ### HTTP status error response codes  | Code | Definition        | Description                                                                                       | Possible Solution                                                | | ---- | ----------------- | ------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- | | 400  | Invalid request       | The request cannot be understood.                                    | Ensure JSON syntax in request body is correct.                   | | 401  | Invalid access token      | Requestor is unauthorized or does not have permission for this API call.                                                | Ensure your API access token is valid and has the appropriate permissions.                                     | | 403  | Forbidden         | Requestor does not have access to this resource.                                                | Ensure that the account member or access token has proper permissions set. | | 404  | Invalid resource identifier | The requested resource is not valid. | Ensure that the resource is correctly identified by ID or key. | | 405  | Method not allowed | The request method is not allowed on this resource. | Ensure that the HTTP verb is correct. | | 409  | Conflict          | The API request can not be completed because it conflicts with a concurrent API request. | Retry your request.                                              | | 422  | Unprocessable entity | The API request can not be completed because the update description can not be understood. | Ensure that the request body is correct for the type of patch you are using, either JSON patch or semantic patch. | 429  | Too many requests | Read [Rate limiting](/#section/Overview/Rate-limiting).                                               | Wait and try again later.                                        |  ## CORS  The LaunchDarkly API supports Cross Origin Resource Sharing (CORS) for AJAX requests from any origin. If an `Origin` header is given in a request, it will be echoed as an explicitly allowed origin. Otherwise the request returns a wildcard, `Access-Control-Allow-Origin: *`. For more information on CORS, read the [CORS W3C Recommendation](http://www.w3.org/TR/cors). Example CORS headers might look like:  ```http Access-Control-Allow-Headers: Accept, Content-Type, Content-Length, Accept-Encoding, Authorization Access-Control-Allow-Methods: OPTIONS, GET, DELETE, PATCH Access-Control-Allow-Origin: * Access-Control-Max-Age: 300 ```  You can make authenticated CORS calls just as you would make same-origin calls, using either [token or session-based authentication](/#section/Overview/Authentication). If you are using session authentication, you should set the `withCredentials` property for your `xhr` request to `true`. You should never expose your access tokens to untrusted entities.  ## Rate limiting  We use several rate limiting strategies to ensure the availability of our APIs. Rate-limited calls to our APIs return a `429` status code. Calls to our APIs include headers indicating the current rate limit status. The specific headers returned depend on the API route being called. The limits differ based on the route, authentication mechanism, and other factors. Routes that are not rate limited may not contain any of the headers described below.  > ### Rate limiting and SDKs > > LaunchDarkly SDKs are never rate limited and do not use the API endpoints defined here. LaunchDarkly uses a different set of approaches, including streaming/server-sent events and a global CDN, to ensure availability to the routes used by LaunchDarkly SDKs.  ### Global rate limits  Authenticated requests are subject to a global limit. This is the maximum number of calls that your account can make to the API per ten seconds. All service and personal access tokens on the account share this limit, so exceeding the limit with one access token will impact other tokens. Calls that are subject to global rate limits may return the headers below:  | Header name                    | Description                                                                      | | ------------------------------ | -------------------------------------------------------------------------------- | | `X-Ratelimit-Global-Remaining` | The maximum number of requests the account is permitted to make per ten seconds. | | `X-Ratelimit-Reset`            | The time at which the current rate limit window resets in epoch milliseconds.    |  We do not publicly document the specific number of calls that can be made globally. This limit may change, and we encourage clients to program against the specification, relying on the two headers defined above, rather than hardcoding to the current limit.  ### Route-level rate limits  Some authenticated routes have custom rate limits. These also reset every ten seconds. Any service or personal access tokens hitting the same route share this limit, so exceeding the limit with one access token may impact other tokens. Calls that are subject to route-level rate limits return the headers below:  | Header name                   | Description                                                                                           | | ----------------------------- | ----------------------------------------------------------------------------------------------------- | | `X-Ratelimit-Route-Remaining` | The maximum number of requests to the current route the account is permitted to make per ten seconds. | | `X-Ratelimit-Reset`           | The time at which the current rate limit window resets in epoch milliseconds.                         |  A _route_ represents a specific URL pattern and verb. For example, the [Delete environment](/tag/Environments#operation/deleteEnvironment) endpoint is considered a single route, and each call to delete an environment counts against your route-level rate limit for that route.  We do not publicly document the specific number of calls that an account can make to each endpoint per ten seconds. These limits may change, and we encourage clients to program against the specification, relying on the two headers defined above, rather than hardcoding to the current limits.  ### IP-based rate limiting  We also employ IP-based rate limiting on some API routes. If you hit an IP-based rate limit, your API response will include a `Retry-After` header indicating how long to wait before re-trying the call. Clients must wait at least `Retry-After` seconds before making additional calls to our API, and should employ jitter and backoff strategies to avoid triggering rate limits again.  ## OpenAPI (Swagger) and client libraries  We have a [complete OpenAPI (Swagger) specification](https://app.launchdarkly.com/api/v2/openapi.json) for our API.  We auto-generate multiple client libraries based on our OpenAPI specification. To learn more, visit the [collection of client libraries on GitHub](https://github.com/search?q=topic%3Alaunchdarkly-api+org%3Alaunchdarkly&type=Repositories). You can also use this specification to generate client libraries to interact with our REST API in your language of choice.  Our OpenAPI specification is supported by several API-based tools such as Postman and Insomnia. In many cases, you can directly import our specification to explore our APIs.  ## Method overriding  Some firewalls and HTTP clients restrict the use of verbs other than `GET` and `POST`. In those environments, our API endpoints that use `DELETE`, `PATCH`, and `PUT` verbs are inaccessible.  To avoid this issue, our API supports the `X-HTTP-Method-Override` header, allowing clients to \"tunnel\" `DELETE`, `PATCH`, and `PUT` requests using a `POST` request.  For example, to call a `PATCH` endpoint using a `POST` request, you can include `X-HTTP-Method-Override:PATCH` as a header.  ## Beta resources  We sometimes release new API resources in **beta** status before we release them with general availability.  Resources that are in beta are still undergoing testing and development. They may change without notice, including becoming backwards incompatible.  We try to promote resources into general availability as quickly as possible. This happens after sufficient testing and when we're satisfied that we no longer need to make backwards-incompatible changes.  We mark beta resources with a \"Beta\" callout in our documentation, pictured below:  > ### This feature is in beta > > To use this feature, pass in a header including the `LD-API-Version` key with value set to `beta`. Use this header with each call. To learn more, read [Beta resources](/#section/Overview/Beta-resources). > > Resources that are in beta are still undergoing testing and development. They may change without notice, including becoming backwards incompatible.  ### Using beta resources  To use a beta resource, you must include a header in the request. If you call a beta resource without this header, you receive a `403` response.  Use this header:  ``` LD-API-Version: beta ```  ## Federal environments  The version of LaunchDarkly that is available on domains controlled by the United States government is different from the version of LaunchDarkly available to the general public. If you are an employee or contractor for a United States federal agency and use LaunchDarkly in your work, you likely use the federal instance of LaunchDarkly.  If you are working in the federal instance of LaunchDarkly, the base URI for each request is `https://app.launchdarkly.us`. In the \"Try it\" sandbox for each request, click the request path to view the complete resource path for the federal environment.  To learn more, read [LaunchDarkly in federal environments](https://docs.launchdarkly.com/home/advanced/federal).  ## Versioning  We try hard to keep our REST API backwards compatible, but we occasionally have to make backwards-incompatible changes in the process of shipping new features. These breaking changes can cause unexpected behavior if you don't prepare for them accordingly.  Updates to our REST API include support for the latest features in LaunchDarkly. We also release a new version of our REST API every time we make a breaking change. We provide simultaneous support for multiple API versions so you can migrate from your current API version to a new version at your own pace.  ### Setting the API version per request  You can set the API version on a specific request by sending an `LD-API-Version` header, as shown in the example below:  ``` LD-API-Version: 20220603 ```  The header value is the version number of the API version you would like to request. The number for each version corresponds to the date the version was released in `yyyymmdd` format. In the example above the version `20220603` corresponds to June 03, 2022.  ### Setting the API version per access token  When you create an access token, you must specify a specific version of the API to use. This ensures that integrations using this token cannot be broken by version changes.  Tokens created before versioning was released have their version set to `20160426`, which is the version of the API that existed before the current versioning scheme, so that they continue working the same way they did before versioning.  If you would like to upgrade your integration to use a new API version, you can explicitly set the header described above.  > ### Best practice: Set the header for every client or integration > > We recommend that you set the API version header explicitly in any client or integration you build. > > Only rely on the access token API version during manual testing.  ### API version changelog  |<div style=\"width:75px\">Version</div> | Changes | End of life (EOL) |---|---|---| | `20220603` | <ul><li>Changed the [list projects](/tag/Projects#operation/getProjects) return value:<ul><li>Response is now paginated with a default limit of `20`.</li><li>Added support for filter and sort.</li><li>The project `environments` field is now expandable. This field is omitted by default.</li></ul></li><li>Changed the [get project](/tag/Projects#operation/getProject) return value:<ul><li>The `environments` field is now expandable. This field is omitted by default.</li></ul></li></ul> | Current | | `20210729` | <ul><li>Changed the [create approval request](/tag/Approvals#operation/postApprovalRequest) return value. It now returns HTTP Status Code `201` instead of `200`.</li><li> Changed the [get users](/tag/Users#operation/getUser) return value. It now returns a user record, not a user. </li><li>Added additional optional fields to environment, segments, flags, members, and segments, including the ability to create big segments. </li><li> Added default values for flag variations when new environments are created. </li><li>Added filtering and pagination for getting flags and members, including `limit`, `number`, `filter`, and `sort` query parameters. </li><li>Added endpoints for expiring user targets for flags and segments, scheduled changes, access tokens, Relay Proxy configuration, integrations and subscriptions, and approvals. </li></ul> | 2023-06-03 | | `20191212` | <ul><li>[List feature flags](/tag/Feature-flags#operation/getFeatureFlags) now defaults to sending summaries of feature flag configurations, equivalent to setting the query parameter `summary=true`. Summaries omit flag targeting rules and individual user targets from the payload. </li><li> Added endpoints for flags, flag status, projects, environments, audit logs, members, users, custom roles, segments, usage, streams, events, and data export. </li></ul> | 2022-07-29 | | `20160426` | <ul><li>Initial versioning of API. Tokens created before versioning have their version set to this.</li></ul> | 2020-12-12 | 

API version: 2.0
Contact: support@launchdarkly.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package ldapi

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)


// FeatureFlagsApiService FeatureFlagsApi service
type FeatureFlagsApiService service

type ApiCopyFeatureFlagRequest struct {
	ctx context.Context
	ApiService *FeatureFlagsApiService
	projectKey string
	featureFlagKey string
	flagCopyConfigPost *FlagCopyConfigPost
}

func (r ApiCopyFeatureFlagRequest) FlagCopyConfigPost(flagCopyConfigPost FlagCopyConfigPost) ApiCopyFeatureFlagRequest {
	r.flagCopyConfigPost = &flagCopyConfigPost
	return r
}

func (r ApiCopyFeatureFlagRequest) Execute() (*FeatureFlag, *http.Response, error) {
	return r.ApiService.CopyFeatureFlagExecute(r)
}

/*
CopyFeatureFlag Copy feature flag


> ### Copying flag settings is an Enterprise feature
>
> Copying flag settings is available to customers on an Enterprise plan. To learn more, [read about our pricing](https://launchdarkly.com/pricing/). To upgrade your plan, [contact Sales](https://launchdarkly.com/contact-sales/).

Copy flag settings from a source environment to a target environment.

By default, this operation copies the entire flag configuration. You can use the `includedActions` or `excludedActions` to specify that only part of the flag configuration is copied.

If you provide the optional `currentVersion` of a flag, this operation tests to ensure that the current flag version in the environment matches the version you've specified. The operation rejects attempts to copy flag settings if the environment's current version  of the flag does not match the version you've specified. You can use this to enforce optimistic locking on copy attempts.


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectKey The project key
 @param featureFlagKey The feature flag key. The key identifies the flag in your code.
 @return ApiCopyFeatureFlagRequest
*/
func (a *FeatureFlagsApiService) CopyFeatureFlag(ctx context.Context, projectKey string, featureFlagKey string) ApiCopyFeatureFlagRequest {
	return ApiCopyFeatureFlagRequest{
		ApiService: a,
		ctx: ctx,
		projectKey: projectKey,
		featureFlagKey: featureFlagKey,
	}
}

// Execute executes the request
//  @return FeatureFlag
func (a *FeatureFlagsApiService) CopyFeatureFlagExecute(r ApiCopyFeatureFlagRequest) (*FeatureFlag, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *FeatureFlag
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "FeatureFlagsApiService.CopyFeatureFlag")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v2/flags/{projectKey}/{featureFlagKey}/copy"
	localVarPath = strings.Replace(localVarPath, "{"+"projectKey"+"}", url.PathEscape(parameterToString(r.projectKey, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"featureFlagKey"+"}", url.PathEscape(parameterToString(r.featureFlagKey, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.flagCopyConfigPost == nil {
		return localVarReturnValue, nil, reportError("flagCopyConfigPost is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.flagCopyConfigPost
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKey"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		if localVarHTTPResponse.StatusCode == 400 {
			var v InvalidRequestErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 401 {
			var v UnauthorizedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 405 {
			var v MethodNotAllowedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 409 {
			var v StatusConflictErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 429 {
			var v RateLimitedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiDeleteFeatureFlagRequest struct {
	ctx context.Context
	ApiService *FeatureFlagsApiService
	projectKey string
	featureFlagKey string
}

func (r ApiDeleteFeatureFlagRequest) Execute() (*http.Response, error) {
	return r.ApiService.DeleteFeatureFlagExecute(r)
}

/*
DeleteFeatureFlag Delete feature flag

Delete a feature flag in all environments. Use with caution: only delete feature flags your application no longer uses.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectKey The project key
 @param featureFlagKey The feature flag key. The key identifies the flag in your code.
 @return ApiDeleteFeatureFlagRequest
*/
func (a *FeatureFlagsApiService) DeleteFeatureFlag(ctx context.Context, projectKey string, featureFlagKey string) ApiDeleteFeatureFlagRequest {
	return ApiDeleteFeatureFlagRequest{
		ApiService: a,
		ctx: ctx,
		projectKey: projectKey,
		featureFlagKey: featureFlagKey,
	}
}

// Execute executes the request
func (a *FeatureFlagsApiService) DeleteFeatureFlagExecute(r ApiDeleteFeatureFlagRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodDelete
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "FeatureFlagsApiService.DeleteFeatureFlag")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v2/flags/{projectKey}/{featureFlagKey}"
	localVarPath = strings.Replace(localVarPath, "{"+"projectKey"+"}", url.PathEscape(parameterToString(r.projectKey, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"featureFlagKey"+"}", url.PathEscape(parameterToString(r.featureFlagKey, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKey"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		if localVarHTTPResponse.StatusCode == 401 {
			var v UnauthorizedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 404 {
			var v NotFoundErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 409 {
			var v StatusConflictErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 429 {
			var v RateLimitedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarHTTPResponse, newErr
			}
			newErr.model = v
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiGetExpiringContextTargetsRequest struct {
	ctx context.Context
	ApiService *FeatureFlagsApiService
	projectKey string
	environmentKey string
	featureFlagKey string
}

func (r ApiGetExpiringContextTargetsRequest) Execute() (*ExpiringTargetGetResponse, *http.Response, error) {
	return r.ApiService.GetExpiringContextTargetsExecute(r)
}

/*
GetExpiringContextTargets Get expiring context targets for feature flag

Get a list of context targets on a feature flag that are scheduled for removal.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectKey The project key
 @param environmentKey The environment key
 @param featureFlagKey The feature flag key
 @return ApiGetExpiringContextTargetsRequest
*/
func (a *FeatureFlagsApiService) GetExpiringContextTargets(ctx context.Context, projectKey string, environmentKey string, featureFlagKey string) ApiGetExpiringContextTargetsRequest {
	return ApiGetExpiringContextTargetsRequest{
		ApiService: a,
		ctx: ctx,
		projectKey: projectKey,
		environmentKey: environmentKey,
		featureFlagKey: featureFlagKey,
	}
}

// Execute executes the request
//  @return ExpiringTargetGetResponse
func (a *FeatureFlagsApiService) GetExpiringContextTargetsExecute(r ApiGetExpiringContextTargetsRequest) (*ExpiringTargetGetResponse, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *ExpiringTargetGetResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "FeatureFlagsApiService.GetExpiringContextTargets")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v2/flags/{projectKey}/{featureFlagKey}/expiring-targets/{environmentKey}"
	localVarPath = strings.Replace(localVarPath, "{"+"projectKey"+"}", url.PathEscape(parameterToString(r.projectKey, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"environmentKey"+"}", url.PathEscape(parameterToString(r.environmentKey, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"featureFlagKey"+"}", url.PathEscape(parameterToString(r.featureFlagKey, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKey"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		if localVarHTTPResponse.StatusCode == 401 {
			var v UnauthorizedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 403 {
			var v ForbiddenErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 404 {
			var v NotFoundErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 429 {
			var v RateLimitedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiGetExpiringUserTargetsRequest struct {
	ctx context.Context
	ApiService *FeatureFlagsApiService
	projectKey string
	environmentKey string
	featureFlagKey string
}

func (r ApiGetExpiringUserTargetsRequest) Execute() (*ExpiringUserTargetGetResponse, *http.Response, error) {
	return r.ApiService.GetExpiringUserTargetsExecute(r)
}

/*
GetExpiringUserTargets Get expiring user targets for feature flag


> ### Contexts are now available
>
> After you have upgraded your LaunchDarkly SDK to use contexts instead of users, you should use [Get expiring context targets for feature flag](/tag/Feature-flags#operation/getExpiringContextTargets) instead of this endpoint. To learn more, read [Contexts](https://docs.launchdarkly.com/home/contexts).

Get a list of user targets on a feature flag that are scheduled for removal.


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectKey The project key
 @param environmentKey The environment key
 @param featureFlagKey The feature flag key
 @return ApiGetExpiringUserTargetsRequest
*/
func (a *FeatureFlagsApiService) GetExpiringUserTargets(ctx context.Context, projectKey string, environmentKey string, featureFlagKey string) ApiGetExpiringUserTargetsRequest {
	return ApiGetExpiringUserTargetsRequest{
		ApiService: a,
		ctx: ctx,
		projectKey: projectKey,
		environmentKey: environmentKey,
		featureFlagKey: featureFlagKey,
	}
}

// Execute executes the request
//  @return ExpiringUserTargetGetResponse
func (a *FeatureFlagsApiService) GetExpiringUserTargetsExecute(r ApiGetExpiringUserTargetsRequest) (*ExpiringUserTargetGetResponse, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *ExpiringUserTargetGetResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "FeatureFlagsApiService.GetExpiringUserTargets")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v2/flags/{projectKey}/{featureFlagKey}/expiring-user-targets/{environmentKey}"
	localVarPath = strings.Replace(localVarPath, "{"+"projectKey"+"}", url.PathEscape(parameterToString(r.projectKey, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"environmentKey"+"}", url.PathEscape(parameterToString(r.environmentKey, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"featureFlagKey"+"}", url.PathEscape(parameterToString(r.featureFlagKey, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKey"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		if localVarHTTPResponse.StatusCode == 401 {
			var v UnauthorizedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 403 {
			var v ForbiddenErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 404 {
			var v NotFoundErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 429 {
			var v RateLimitedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiGetFeatureFlagRequest struct {
	ctx context.Context
	ApiService *FeatureFlagsApiService
	projectKey string
	featureFlagKey string
	env *string
	expand *string
}

// Filter configurations by environment
func (r ApiGetFeatureFlagRequest) Env(env string) ApiGetFeatureFlagRequest {
	r.env = &env
	return r
}

// A comma-separated list of fields to expand in the response. Supported fields are explained above.
func (r ApiGetFeatureFlagRequest) Expand(expand string) ApiGetFeatureFlagRequest {
	r.expand = &expand
	return r
}

func (r ApiGetFeatureFlagRequest) Execute() (*FeatureFlag, *http.Response, error) {
	return r.ApiService.GetFeatureFlagExecute(r)
}

/*
GetFeatureFlag Get feature flag

Get a single feature flag by key. By default, this returns the configurations for all environments. You can filter environments with the `env` query parameter. For example, setting `env=production` restricts the returned configurations to just the `production` environment.

> #### Recommended use
>
> This endpoint can return a large amount of information. Specifying one or multiple environments with the `env` parameter can decrease response time and overall payload size. We recommend using this parameter to return only the environments relevant to your query.

### Expanding response

LaunchDarkly supports the `expand` query param to include additional fields in the response, with the following fields:

- `evaluation` includes evaluation information within returned environments, including which context kinds the flag has been evaluated for in the past 30 days 
- `migrationSettings` includes migration settings information within the flag and within returned environments. These settings are only included for migration flags, that is, where `purpose` is `migration`.

For example, `expand=evaluation` includes the `evaluation` field in the response.


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectKey The project key
 @param featureFlagKey The feature flag key
 @return ApiGetFeatureFlagRequest
*/
func (a *FeatureFlagsApiService) GetFeatureFlag(ctx context.Context, projectKey string, featureFlagKey string) ApiGetFeatureFlagRequest {
	return ApiGetFeatureFlagRequest{
		ApiService: a,
		ctx: ctx,
		projectKey: projectKey,
		featureFlagKey: featureFlagKey,
	}
}

// Execute executes the request
//  @return FeatureFlag
func (a *FeatureFlagsApiService) GetFeatureFlagExecute(r ApiGetFeatureFlagRequest) (*FeatureFlag, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *FeatureFlag
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "FeatureFlagsApiService.GetFeatureFlag")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v2/flags/{projectKey}/{featureFlagKey}"
	localVarPath = strings.Replace(localVarPath, "{"+"projectKey"+"}", url.PathEscape(parameterToString(r.projectKey, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"featureFlagKey"+"}", url.PathEscape(parameterToString(r.featureFlagKey, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.env != nil {
		localVarQueryParams.Add("env", parameterToString(*r.env, ""))
	}
	if r.expand != nil {
		localVarQueryParams.Add("expand", parameterToString(*r.expand, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKey"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		if localVarHTTPResponse.StatusCode == 401 {
			var v UnauthorizedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 403 {
			var v ForbiddenErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 404 {
			var v NotFoundErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 429 {
			var v RateLimitedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiGetFeatureFlagStatusRequest struct {
	ctx context.Context
	ApiService *FeatureFlagsApiService
	projectKey string
	environmentKey string
	featureFlagKey string
}

func (r ApiGetFeatureFlagStatusRequest) Execute() (*FlagStatusRep, *http.Response, error) {
	return r.ApiService.GetFeatureFlagStatusExecute(r)
}

/*
GetFeatureFlagStatus Get feature flag status

Get the status for a particular feature flag.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectKey The project key
 @param environmentKey The environment key
 @param featureFlagKey The feature flag key
 @return ApiGetFeatureFlagStatusRequest
*/
func (a *FeatureFlagsApiService) GetFeatureFlagStatus(ctx context.Context, projectKey string, environmentKey string, featureFlagKey string) ApiGetFeatureFlagStatusRequest {
	return ApiGetFeatureFlagStatusRequest{
		ApiService: a,
		ctx: ctx,
		projectKey: projectKey,
		environmentKey: environmentKey,
		featureFlagKey: featureFlagKey,
	}
}

// Execute executes the request
//  @return FlagStatusRep
func (a *FeatureFlagsApiService) GetFeatureFlagStatusExecute(r ApiGetFeatureFlagStatusRequest) (*FlagStatusRep, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *FlagStatusRep
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "FeatureFlagsApiService.GetFeatureFlagStatus")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v2/flag-statuses/{projectKey}/{environmentKey}/{featureFlagKey}"
	localVarPath = strings.Replace(localVarPath, "{"+"projectKey"+"}", url.PathEscape(parameterToString(r.projectKey, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"environmentKey"+"}", url.PathEscape(parameterToString(r.environmentKey, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"featureFlagKey"+"}", url.PathEscape(parameterToString(r.featureFlagKey, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKey"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		if localVarHTTPResponse.StatusCode == 401 {
			var v UnauthorizedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 403 {
			var v ForbiddenErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 404 {
			var v NotFoundErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 429 {
			var v RateLimitedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiGetFeatureFlagStatusAcrossEnvironmentsRequest struct {
	ctx context.Context
	ApiService *FeatureFlagsApiService
	projectKey string
	featureFlagKey string
	env *string
}

// Optional environment filter
func (r ApiGetFeatureFlagStatusAcrossEnvironmentsRequest) Env(env string) ApiGetFeatureFlagStatusAcrossEnvironmentsRequest {
	r.env = &env
	return r
}

func (r ApiGetFeatureFlagStatusAcrossEnvironmentsRequest) Execute() (*FeatureFlagStatusAcrossEnvironments, *http.Response, error) {
	return r.ApiService.GetFeatureFlagStatusAcrossEnvironmentsExecute(r)
}

/*
GetFeatureFlagStatusAcrossEnvironments Get flag status across environments

Get the status for a particular feature flag across environments.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectKey The project key
 @param featureFlagKey The feature flag key
 @return ApiGetFeatureFlagStatusAcrossEnvironmentsRequest
*/
func (a *FeatureFlagsApiService) GetFeatureFlagStatusAcrossEnvironments(ctx context.Context, projectKey string, featureFlagKey string) ApiGetFeatureFlagStatusAcrossEnvironmentsRequest {
	return ApiGetFeatureFlagStatusAcrossEnvironmentsRequest{
		ApiService: a,
		ctx: ctx,
		projectKey: projectKey,
		featureFlagKey: featureFlagKey,
	}
}

// Execute executes the request
//  @return FeatureFlagStatusAcrossEnvironments
func (a *FeatureFlagsApiService) GetFeatureFlagStatusAcrossEnvironmentsExecute(r ApiGetFeatureFlagStatusAcrossEnvironmentsRequest) (*FeatureFlagStatusAcrossEnvironments, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *FeatureFlagStatusAcrossEnvironments
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "FeatureFlagsApiService.GetFeatureFlagStatusAcrossEnvironments")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v2/flag-status/{projectKey}/{featureFlagKey}"
	localVarPath = strings.Replace(localVarPath, "{"+"projectKey"+"}", url.PathEscape(parameterToString(r.projectKey, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"featureFlagKey"+"}", url.PathEscape(parameterToString(r.featureFlagKey, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.env != nil {
		localVarQueryParams.Add("env", parameterToString(*r.env, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKey"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		if localVarHTTPResponse.StatusCode == 401 {
			var v UnauthorizedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 403 {
			var v ForbiddenErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 404 {
			var v NotFoundErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 429 {
			var v RateLimitedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiGetFeatureFlagStatusesRequest struct {
	ctx context.Context
	ApiService *FeatureFlagsApiService
	projectKey string
	environmentKey string
}

func (r ApiGetFeatureFlagStatusesRequest) Execute() (*FeatureFlagStatuses, *http.Response, error) {
	return r.ApiService.GetFeatureFlagStatusesExecute(r)
}

/*
GetFeatureFlagStatuses List feature flag statuses

Get a list of statuses for all feature flags. The status includes the last time the feature flag was requested, as well as a state, which is one of the following:

- `new`: You created the flag fewer than seven days ago and it has never been requested.
- `active`: LaunchDarkly is receiving requests for this flag, but there are either multiple variations configured, or it is toggled off, or there have been changes to configuration in the past seven days.
- `inactive`: You created the feature flag more than seven days ago, and hasn't been requested within the past seven days.
- `launched`: LaunchDarkly is receiving requests for this flag, it is toggled on, there is only one variation configured, and there have been no changes to configuration in the past seven days.

To learn more, read [Flag statuses](https://docs.launchdarkly.com/home/code/flag-status).


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectKey The project key
 @param environmentKey The environment key
 @return ApiGetFeatureFlagStatusesRequest
*/
func (a *FeatureFlagsApiService) GetFeatureFlagStatuses(ctx context.Context, projectKey string, environmentKey string) ApiGetFeatureFlagStatusesRequest {
	return ApiGetFeatureFlagStatusesRequest{
		ApiService: a,
		ctx: ctx,
		projectKey: projectKey,
		environmentKey: environmentKey,
	}
}

// Execute executes the request
//  @return FeatureFlagStatuses
func (a *FeatureFlagsApiService) GetFeatureFlagStatusesExecute(r ApiGetFeatureFlagStatusesRequest) (*FeatureFlagStatuses, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *FeatureFlagStatuses
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "FeatureFlagsApiService.GetFeatureFlagStatuses")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v2/flag-statuses/{projectKey}/{environmentKey}"
	localVarPath = strings.Replace(localVarPath, "{"+"projectKey"+"}", url.PathEscape(parameterToString(r.projectKey, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"environmentKey"+"}", url.PathEscape(parameterToString(r.environmentKey, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKey"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		if localVarHTTPResponse.StatusCode == 401 {
			var v UnauthorizedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 403 {
			var v ForbiddenErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 404 {
			var v NotFoundErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 429 {
			var v RateLimitedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiGetFeatureFlagsRequest struct {
	ctx context.Context
	ApiService *FeatureFlagsApiService
	projectKey string
	env *string
	tag *string
	limit *int64
	offset *int64
	archived *bool
	summary *bool
	filter *string
	sort *string
	compare *bool
	expand *string
}

// Filter configurations by environment
func (r ApiGetFeatureFlagsRequest) Env(env string) ApiGetFeatureFlagsRequest {
	r.env = &env
	return r
}

// Filter feature flags by tag
func (r ApiGetFeatureFlagsRequest) Tag(tag string) ApiGetFeatureFlagsRequest {
	r.tag = &tag
	return r
}

// The number of feature flags to return. Defaults to -1, which returns all flags
func (r ApiGetFeatureFlagsRequest) Limit(limit int64) ApiGetFeatureFlagsRequest {
	r.limit = &limit
	return r
}

// Where to start in the list. Use this with pagination. For example, an offset of 10 skips the first ten items and then returns the next items in the list, up to the query &#x60;limit&#x60;.
func (r ApiGetFeatureFlagsRequest) Offset(offset int64) ApiGetFeatureFlagsRequest {
	r.offset = &offset
	return r
}

// Deprecated, use &#x60;filter&#x3D;archived:true&#x60; instead. A boolean to filter the list to archived flags. When this is absent, only unarchived flags will be returned
// Deprecated
func (r ApiGetFeatureFlagsRequest) Archived(archived bool) ApiGetFeatureFlagsRequest {
	r.archived = &archived
	return r
}

// By default, flags do _not_ include their lists of prerequisites, targets, or rules for each environment. Set &#x60;summary&#x3D;0&#x60; to include these fields for each flag returned.
func (r ApiGetFeatureFlagsRequest) Summary(summary bool) ApiGetFeatureFlagsRequest {
	r.summary = &summary
	return r
}

// A comma-separated list of filters. Each filter is of the form field:value. Read the endpoint description for a full list of available filter fields.
func (r ApiGetFeatureFlagsRequest) Filter(filter string) ApiGetFeatureFlagsRequest {
	r.filter = &filter
	return r
}

// A comma-separated list of fields to sort by. Fields prefixed by a dash ( - ) sort in descending order. Read the endpoint description for a full list of available sort fields.
func (r ApiGetFeatureFlagsRequest) Sort(sort string) ApiGetFeatureFlagsRequest {
	r.sort = &sort
	return r
}

// A boolean to filter results by only flags that have differences between environments
func (r ApiGetFeatureFlagsRequest) Compare(compare bool) ApiGetFeatureFlagsRequest {
	r.compare = &compare
	return r
}

// A comma-separated list of fields to expand in the response. Supported fields are explained above.
func (r ApiGetFeatureFlagsRequest) Expand(expand string) ApiGetFeatureFlagsRequest {
	r.expand = &expand
	return r
}

func (r ApiGetFeatureFlagsRequest) Execute() (*FeatureFlags, *http.Response, error) {
	return r.ApiService.GetFeatureFlagsExecute(r)
}

/*
GetFeatureFlags List feature flags

Get a list of all feature flags in the given project. By default, each flag includes configurations for each environment. You can filter environments with the `env` query parameter. For example, setting `env=production` restricts the returned configurations to just your production environment. You can also filter feature flags by tag with the `tag` query parameter.

> #### Recommended use
>
> This endpoint can return a large amount of information. We recommend using some or all of these query parameters to decrease response time and overall payload size: `limit`, `env`, `query`, and `filter=creationDate`.

### Filtering flags

You can filter on certain fields using the `filter` query parameter. For example, setting `filter=query:dark-mode,tags:beta+test` matches flags with the string `dark-mode` in their key or name, ignoring case, which also have the tags `beta` and `test`.

The `filter` query parameter supports the following arguments:

| Filter argument       | Description | Example              |
|-----------------------|-------------|----------------------|
| `archived`              | A boolean value. It filters the list to archived flags. Setting the value to `true` returns only archived flags. When this is absent, only unarchived flags are returned. | `filter=archived:true` |
| `contextKindsEvaluated` | A `+`-separated list of context kind keys. It filters the list to flags which have been evaluated in the past 30 days for all of the context kinds in the list. | `filter=contextKindsEvaluated:user+application` |
| `contextKindTargeted`   | A string. It filters the list to flags that are targeting the given context kind key. | `filter=contextKindTargeted:user` |
| `codeReferences.max`    | An integer value. Use `0` to return flags that do not have code references. | `filter=codeReferences.max:0` |
| `codeReferences.min`    | An integer value. Use `1` to return flags that do have code references. | `filter=codeReferences.min:1` |
| `creationDate`          | An object with an optional `before` field whose value is Unix time in milliseconds. It filters the list to flags created before the date. | `filter=creationDate:{"before":1690527600000}` |
| `evaluated`             | An object that contains a key of `after` and a value in Unix time in milliseconds. It filters the list to all flags that have been evaluated since the time you specify, in the environment provided. This filter requires the `filterEnv` filter. | `filter=evaluation:{"after":1690527600000}` |
| `filterEnv`             | A string with the key of a valid environment. You must use this field for filters that are environment-specific. If there are multiple environment-specific filters, you only need to include this field once. | `filter=evaluated:{"after": 1590768455282},filterEnv:production,status:active` |
| `followerId`            | A valid member ID. It filters the list to flags that are being followed by this member. |  `filter=followerId:12ab3c45de678910abc12345` |
| `hasDataExport`         | A boolean value. It filters the list to flags that are exporting data in the specified environment. This includes flags that are exporting data from Experimentation. This filter requires the `filterEnv` filter. | `filter=hasDataExport:true,filterEnv:production` |
| `hasExperiment`         | A boolean value. It filters the list to flags that are used in an experiment. | `filter=hasExperiment:true` |
| `maintainerId`          | A valid member ID. It filters the list to flags that are maintained by this member. | `filter=maintainerId:12ab3c45de678910abc12345` |
| `maintainerTeamKey`     | A string. It filters the list to flags that are maintained by the team with this key. | `filter=maintainerTeamKey:example-team-key` |
| `query`                 | A string. It filters the list to flags that include the specified string in their key or name. It is not case sensitive. | `filter=query:example` |
| `sdkAvailability`       | A string, one of `client`, `mobile`, `anyClient`, `server`. Using `client` filters the list to flags whose client-side SDK availability is set to use the client-side ID. Using `mobile` filters to flags set to use the mobile key. Using `anyClient` filters to flags set to use either the client-side ID or the mobile key. Using `server` filters to flags set to use neither, that is, to flags only available in server-side SDKs.  | `filter=sdkAvailability:client` |
| `segmentTargeted`       | A string. It filters the list to flags that target the segment with this key. This filter requires the `filterEnv` filter. | `filter=segmentTargeted:example-segment-key,filterEnv:production` |
| `status`                | A string, either `new`, `inactive`, `active`, or `launched`. It filters the list to flags with the specified status in the specified environment. This filter requires the `filterEnv` filter. | `filter=status:active,filterEnv:production` |
| `tags`                  | A `+`-separated list of tags. It filters the list to flags that have all of the tags in the list. | `filter=tags:beta+test` |
| `type`                  | A string, either `temporary` or `permanent`. It filters the list to flags with the specified type. | `filter=type:permanent` |

The documented values for the `filter` query are prior to URL encoding. For example, the `+` in `filter=tags:beta+test` must be encoded to `%2B`.

By default, this endpoint returns all flags. You can page through the list with the `limit` parameter and by following the `first`, `prev`, `next`, and `last` links in the returned `_links` field. These links will not be present if the pages they refer to don't exist. For example, the `first` and `prev` links will be missing from the response on the first page.

### Sorting flags

You can sort flags based on the following fields:

- `creationDate` sorts by the creation date of the flag.
- `key` sorts by the key of the flag.
- `maintainerId` sorts by the flag maintainer.
- `name` sorts by flag name.
- `tags` sorts by tags.
- `targetingModifiedDate` sorts by the date that the flag's targeting rules were last modified in a given environment. It must be used with `env` parameter and it can not be combined with any other sort. If multiple `env` values are provided, it will perform sort using the first one. For example, `sort=-targetingModifiedDate&env=production&env=staging` returns results sorted by `targetingModifiedDate` for the `production` environment.
- `type` sorts by flag type

All fields are sorted in ascending order by default. To sort in descending order, prefix the field with a dash ( - ). For example, `sort=-name` sorts the response by flag name in descending order.

### Expanding response

LaunchDarkly supports the `expand` query param to include additional fields in the response, with the following fields:

- `codeReferences` includes code references for the feature flag
- `evaluation` includes evaluation information within returned environments, including which context kinds the flag has been evaluated for in the past 30 days
- `migrationSettings` includes migration settings information within the flag and within returned environments. These settings are only included for migration flags, that is, where `purpose` is `migration`.

For example, `expand=evaluation` includes the `evaluation` field in the response.

### Migration flags
For migration flags, the cohort information is included in the `rules` property of a flag's response, and default cohort information is included in the `fallthrough` property of a flag's response.
To learn more, read [Migration Flags](https://docs.launchdarkly.com/home/flag-types/migration-flags).


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectKey The project key
 @return ApiGetFeatureFlagsRequest
*/
func (a *FeatureFlagsApiService) GetFeatureFlags(ctx context.Context, projectKey string) ApiGetFeatureFlagsRequest {
	return ApiGetFeatureFlagsRequest{
		ApiService: a,
		ctx: ctx,
		projectKey: projectKey,
	}
}

// Execute executes the request
//  @return FeatureFlags
func (a *FeatureFlagsApiService) GetFeatureFlagsExecute(r ApiGetFeatureFlagsRequest) (*FeatureFlags, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *FeatureFlags
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "FeatureFlagsApiService.GetFeatureFlags")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v2/flags/{projectKey}"
	localVarPath = strings.Replace(localVarPath, "{"+"projectKey"+"}", url.PathEscape(parameterToString(r.projectKey, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.env != nil {
		localVarQueryParams.Add("env", parameterToString(*r.env, ""))
	}
	if r.tag != nil {
		localVarQueryParams.Add("tag", parameterToString(*r.tag, ""))
	}
	if r.limit != nil {
		localVarQueryParams.Add("limit", parameterToString(*r.limit, ""))
	}
	if r.offset != nil {
		localVarQueryParams.Add("offset", parameterToString(*r.offset, ""))
	}
	if r.archived != nil {
		localVarQueryParams.Add("archived", parameterToString(*r.archived, ""))
	}
	if r.summary != nil {
		localVarQueryParams.Add("summary", parameterToString(*r.summary, ""))
	}
	if r.filter != nil {
		localVarQueryParams.Add("filter", parameterToString(*r.filter, ""))
	}
	if r.sort != nil {
		localVarQueryParams.Add("sort", parameterToString(*r.sort, ""))
	}
	if r.compare != nil {
		localVarQueryParams.Add("compare", parameterToString(*r.compare, ""))
	}
	if r.expand != nil {
		localVarQueryParams.Add("expand", parameterToString(*r.expand, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKey"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		if localVarHTTPResponse.StatusCode == 400 {
			var v InvalidRequestErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 401 {
			var v UnauthorizedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 403 {
			var v ForbiddenErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 404 {
			var v NotFoundErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 429 {
			var v RateLimitedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiPatchExpiringTargetsRequest struct {
	ctx context.Context
	ApiService *FeatureFlagsApiService
	projectKey string
	environmentKey string
	featureFlagKey string
	patchFlagsRequest *PatchFlagsRequest
}

func (r ApiPatchExpiringTargetsRequest) PatchFlagsRequest(patchFlagsRequest PatchFlagsRequest) ApiPatchExpiringTargetsRequest {
	r.patchFlagsRequest = &patchFlagsRequest
	return r
}

func (r ApiPatchExpiringTargetsRequest) Execute() (*ExpiringTargetPatchResponse, *http.Response, error) {
	return r.ApiService.PatchExpiringTargetsExecute(r)
}

/*
PatchExpiringTargets Update expiring context targets on feature flag

Schedule a context for removal from individual targeting on a feature flag. The flag must already individually target the context.

You can add, update, or remove a scheduled removal date. You can only schedule a context for removal on a single variation per flag.

Updating an expiring target uses the semantic patch format. To make a semantic patch request, you must append `domain-model=launchdarkly.semanticpatch` to your `Content-Type` header. To learn more, read [Updates using semantic patch](/reference#updates-using-semantic-patch).

### Instructions

Semantic patch requests support the following `kind` instructions for updating expiring targets.

<details>
<summary>Click to expand instructions for <strong>updating expiring targets</strong></summary>

#### addExpiringTarget

Adds a date and time that LaunchDarkly will remove the context from the flag's individual targeting.

##### Parameters

* `value`: The time, in Unix milliseconds, when LaunchDarkly should remove the context from individual targeting for this flag
* `variationId`: ID of a variation on the flag
* `contextKey`: The context key for the context to remove from individual targeting
* `contextKind`: The kind of context represented by the `contextKey`

Here's an example:

```json
{
  "instructions": [{
    "kind": "addExpiringTarget",
    "value": 1754006460000,
    "variationId": "4254742c-71ae-411f-a992-43b18a51afe0",
    "contextKey": "user-key-123abc",
    "contextKind": "user"
  }]
}
```

#### updateExpiringTarget

Updates the date and time that LaunchDarkly will remove the context from the flag's individual targeting

##### Parameters

* `value`: The time, in Unix milliseconds, when LaunchDarkly should remove the context from individual targeting for this flag
* `variationId`: ID of a variation on the flag
* `contextKey`: The context key for the context to remove from individual targeting
* `contextKind`: The kind of context represented by the `contextKey`
* `version`: (Optional) The version of the expiring target to update. If included, update will fail if version doesn't match current version of the expiring target.

Here's an example:

```json
{
  "instructions": [{
    "kind": "updateExpiringTarget",
    "value": 1754006460000,
    "variationId": "4254742c-71ae-411f-a992-43b18a51afe0",
    "contextKey": "user-key-123abc",
    "contextKind": "user"
  }]
}
```

#### removeExpiringTarget

Removes the scheduled removal of the context from the flag's individual targeting. The context will remain part of the flag's individual targeting until you explicitly remove it, or until you schedule another removal.

##### Parameters

* `variationId`: ID of a variation on the flag
* `contextKey`: The context key for the context to remove from individual targeting
* `contextKind`: The kind of context represented by the `contextKey`

Here's an example:

```json
{
  "instructions": [{
    "kind": "removeExpiringTarget",
    "variationId": "4254742c-71ae-411f-a992-43b18a51afe0",
    "contextKey": "user-key-123abc",
    "contextKind": "user"
  }]
}
```

</details>


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectKey The project key
 @param environmentKey The environment key
 @param featureFlagKey The feature flag key
 @return ApiPatchExpiringTargetsRequest
*/
func (a *FeatureFlagsApiService) PatchExpiringTargets(ctx context.Context, projectKey string, environmentKey string, featureFlagKey string) ApiPatchExpiringTargetsRequest {
	return ApiPatchExpiringTargetsRequest{
		ApiService: a,
		ctx: ctx,
		projectKey: projectKey,
		environmentKey: environmentKey,
		featureFlagKey: featureFlagKey,
	}
}

// Execute executes the request
//  @return ExpiringTargetPatchResponse
func (a *FeatureFlagsApiService) PatchExpiringTargetsExecute(r ApiPatchExpiringTargetsRequest) (*ExpiringTargetPatchResponse, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPatch
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *ExpiringTargetPatchResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "FeatureFlagsApiService.PatchExpiringTargets")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v2/flags/{projectKey}/{featureFlagKey}/expiring-targets/{environmentKey}"
	localVarPath = strings.Replace(localVarPath, "{"+"projectKey"+"}", url.PathEscape(parameterToString(r.projectKey, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"environmentKey"+"}", url.PathEscape(parameterToString(r.environmentKey, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"featureFlagKey"+"}", url.PathEscape(parameterToString(r.featureFlagKey, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.patchFlagsRequest == nil {
		return localVarReturnValue, nil, reportError("patchFlagsRequest is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.patchFlagsRequest
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKey"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		if localVarHTTPResponse.StatusCode == 400 {
			var v InvalidRequestErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 401 {
			var v UnauthorizedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 403 {
			var v ForbiddenErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 404 {
			var v NotFoundErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 429 {
			var v RateLimitedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiPatchExpiringUserTargetsRequest struct {
	ctx context.Context
	ApiService *FeatureFlagsApiService
	projectKey string
	environmentKey string
	featureFlagKey string
	patchFlagsRequest *PatchFlagsRequest
}

func (r ApiPatchExpiringUserTargetsRequest) PatchFlagsRequest(patchFlagsRequest PatchFlagsRequest) ApiPatchExpiringUserTargetsRequest {
	r.patchFlagsRequest = &patchFlagsRequest
	return r
}

func (r ApiPatchExpiringUserTargetsRequest) Execute() (*ExpiringUserTargetPatchResponse, *http.Response, error) {
	return r.ApiService.PatchExpiringUserTargetsExecute(r)
}

/*
PatchExpiringUserTargets Update expiring user targets on feature flag

> ### Contexts are now available
>
> After you have upgraded your LaunchDarkly SDK to use contexts instead of users, you should use [Update expiring context targets on feature flag](/tag/Feature-flags#operation/patchExpiringTargets) instead of this endpoint. To learn more, read [Contexts](https://docs.launchdarkly.com/home/contexts).

Schedule a target for removal from individual targeting on a feature flag. The flag must already serve a variation to specific targets based on their key.

You can add, update, or remove a scheduled removal date. You can only schedule a target for removal on a single variation per flag.

Updating an expiring target uses the semantic patch format. To make a semantic patch request, you must append `domain-model=launchdarkly.semanticpatch` to your `Content-Type` header. To learn more, read [Updates using semantic patch](/reference#updates-using-semantic-patch).

### Instructions

Semantic patch requests support the following `kind` instructions for updating expiring user targets.

<details>
<summary>Click to expand instructions for <strong>updating expiring user targets</strong></summary>

#### addExpireUserTargetDate

Adds a date and time that LaunchDarkly will remove the user from the flag's individual targeting.

##### Parameters

* `value`: The time, in Unix milliseconds, when LaunchDarkly should remove the user from individual targeting for this flag
* `variationId`: ID of a variation on the flag
* `userKey`: The user key for the user to remove from individual targeting

#### updateExpireUserTargetDate

Updates the date and time that LaunchDarkly will remove the user from the flag's individual targeting.

##### Parameters

* `value`: The time, in Unix milliseconds, when LaunchDarkly should remove the user from individual targeting for this flag
* `variationId`: ID of a variation on the flag
* `userKey`: The user key for the user to remove from individual targeting
* `version`: (Optional) The version of the expiring user target to update. If included, update will fail if version doesn't match current version of the expiring user target.

#### removeExpireUserTargetDate

Removes the scheduled removal of the user from the flag's individual targeting. The user will remain part of the flag's individual targeting until you explicitly remove them, or until you schedule another removal.

##### Parameters

* `variationId`: ID of a variation on the flag
* `userKey`: The user key for the user to remove from individual targeting

</details>


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectKey The project key
 @param environmentKey The environment key
 @param featureFlagKey The feature flag key
 @return ApiPatchExpiringUserTargetsRequest
*/
func (a *FeatureFlagsApiService) PatchExpiringUserTargets(ctx context.Context, projectKey string, environmentKey string, featureFlagKey string) ApiPatchExpiringUserTargetsRequest {
	return ApiPatchExpiringUserTargetsRequest{
		ApiService: a,
		ctx: ctx,
		projectKey: projectKey,
		environmentKey: environmentKey,
		featureFlagKey: featureFlagKey,
	}
}

// Execute executes the request
//  @return ExpiringUserTargetPatchResponse
func (a *FeatureFlagsApiService) PatchExpiringUserTargetsExecute(r ApiPatchExpiringUserTargetsRequest) (*ExpiringUserTargetPatchResponse, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPatch
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *ExpiringUserTargetPatchResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "FeatureFlagsApiService.PatchExpiringUserTargets")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v2/flags/{projectKey}/{featureFlagKey}/expiring-user-targets/{environmentKey}"
	localVarPath = strings.Replace(localVarPath, "{"+"projectKey"+"}", url.PathEscape(parameterToString(r.projectKey, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"environmentKey"+"}", url.PathEscape(parameterToString(r.environmentKey, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"featureFlagKey"+"}", url.PathEscape(parameterToString(r.featureFlagKey, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.patchFlagsRequest == nil {
		return localVarReturnValue, nil, reportError("patchFlagsRequest is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.patchFlagsRequest
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKey"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		if localVarHTTPResponse.StatusCode == 400 {
			var v InvalidRequestErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 401 {
			var v UnauthorizedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 403 {
			var v ForbiddenErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 404 {
			var v NotFoundErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 429 {
			var v RateLimitedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiPatchFeatureFlagRequest struct {
	ctx context.Context
	ApiService *FeatureFlagsApiService
	projectKey string
	featureFlagKey string
	patchWithComment *PatchWithComment
}

func (r ApiPatchFeatureFlagRequest) PatchWithComment(patchWithComment PatchWithComment) ApiPatchFeatureFlagRequest {
	r.patchWithComment = &patchWithComment
	return r
}

func (r ApiPatchFeatureFlagRequest) Execute() (*FeatureFlag, *http.Response, error) {
	return r.ApiService.PatchFeatureFlagExecute(r)
}

/*
PatchFeatureFlag Update feature flag

Perform a partial update to a feature flag. The request body must be a valid semantic patch, JSON patch, or JSON merge patch. To learn more the different formats, read [Updates](/#section/Overview/Updates).

### Using semantic patches on a feature flag

To make a semantic patch request, you must append `domain-model=launchdarkly.semanticpatch` to your `Content-Type` header. To learn more, read [Updates using semantic patch](/reference#updates-using-semantic-patch).

The body of a semantic patch request for updating feature flags takes the following properties:

* `comment` (string): (Optional) A description of the update.
* `environmentKey` (string): (Required for some instructions only) The key of the LaunchDarkly environment.
* `instructions` (array): (Required) A list of actions the update should perform. Each action in the list must be an object with a `kind` property that indicates the instruction. If the action requires parameters, you must include those parameters as additional fields in the object. The body of a single semantic patch can contain many different instructions.

### Instructions

Semantic patch requests support the following `kind` instructions for updating feature flags.

<details>
<summary>Click to expand instructions for <strong>turning flags on and off</strong></summary>

These instructions require the `environmentKey` parameter.

#### turnFlagOff

Sets the flag's targeting state to **Off**.

Here's an example:

```json
{
  "environmentKey": "environment-key-123abc",
  "instructions": [ { "kind": "turnFlagOff" } ]
}
```

#### turnFlagOn

Sets the flag's targeting state to **On**.

Here's an example:

```json
{
  "environmentKey": "environment-key-123abc",
  "instructions": [ { "kind": "turnFlagOn" } ]
}
```

</details><br />

<details>
<summary>Click to expand instructions for <strong>working with targeting and variations</strong></summary>

These instructions require the `environmentKey` parameter.

Several of the instructions for working with targeting and variations require flag rule IDs, variation IDs, or clause IDs as parameters. Each of these are returned as part of the [Get feature flag](/tag/Feature-flags#operation/getFeatureFlag) response. The flag rule ID is the `_id` field of each element in the `rules` array within each environment listed in the `environments` object. The variation ID is the `_id` field in each element of the `variations` array. The clause ID is the `_id` field of each element of the `clauses` array within the `rules` array within each environment listed in the `environments` object.

#### addClauses

Adds the given clauses to the rule indicated by `ruleId`.

##### Parameters

- `ruleId`: ID of a rule in the flag.
- `clauses`: Array of clause objects, with `contextKind` (string), `attribute` (string), `op` (string), `negate` (boolean), and `values` (array of strings, numbers, or dates) properties. The `contextKind`, `attribute`, and `values` are case sensitive. The `op` must be lower-case.

Here's an example:

```json
{
	"environmentKey": "environment-key-123abc",
	"instructions": [{
		"kind": "addClauses",
		"ruleId": "a902ef4a-2faf-4eaf-88e1-ecc356708a29",
		"clauses": [{
			"contextKind": "user",
			"attribute": "country",
			"op": "in",
			"negate": false,
			"values": ["USA", "Canada"]
		}]
	}]
}
```

#### addPrerequisite

Adds the flag indicated by `key` with variation `variationId` as a prerequisite to the flag in the path parameter.

##### Parameters

- `key`: Flag key of the prerequisite flag.
- `variationId`: ID of a variation of the prerequisite flag.

Here's an example:

```json
{
	"environmentKey": "environment-key-123abc",
	"instructions": [{
		"kind": "addPrerequisite",
		"key": "example-prereq-flag-key",
		"variationId": "2f43f67c-3e4e-4945-a18a-26559378ca00"
	}]
}
```

#### addRule

Adds a new targeting rule to the flag. The rule may contain `clauses` and serve the variation that `variationId` indicates, or serve a percentage rollout that `rolloutWeights`, `rolloutBucketBy`, and `rolloutContextKind` indicate.

If you set `beforeRuleId`, this adds the new rule before the indicated rule. Otherwise, adds the new rule to the end of the list.

##### Parameters

- `clauses`: Array of clause objects, with `contextKind` (string), `attribute` (string), `op` (string), `negate` (boolean), and `values` (array of strings, numbers, or dates) properties. The `contextKind`, `attribute`, and `values` are case sensitive. The `op` must be lower-case.
- `beforeRuleId`: (Optional) ID of a flag rule.
- Either
  - `variationId`: ID of a variation of the flag.

  or

  - `rolloutWeights`: (Optional) Map of `variationId` to weight, in thousandths of a percent (0-100000).
  - `rolloutBucketBy`: (Optional) Context attribute available in the specified `rolloutContextKind`.
  - `rolloutContextKind`: (Optional) Context kind, defaults to `user`

Here's an example that uses a `variationId`:

```json
{
  "environmentKey": "environment-key-123abc",
  "instructions": [{
    "kind": "addRule",
    "variationId": "2f43f67c-3e4e-4945-a18a-26559378ca00",
    "clauses": [{
      "contextKind": "organization",
      "attribute": "located_in",
      "op": "in",
      "negate": false,
      "values": ["Sweden", "Norway"]
    }]
  }]
}
```

Here's an example that uses a percentage rollout:

```json
{
  "environmentKey": "environment-key-123abc",
  "instructions": [{
    "kind": "addRule",
    "clauses": [{
      "contextKind": "organization",
      "attribute": "located_in",
      "op": "in",
      "negate": false,
      "values": ["Sweden", "Norway"]
    }],
    "rolloutContextKind": "organization",
    "rolloutWeights": {
      "2f43f67c-3e4e-4945-a18a-26559378ca00": 15000, // serve 15% this variation
      "e5830889-1ec5-4b0c-9cc9-c48790090c43": 85000  // serve 85% this variation
    }
  }]
}
```

#### addTargets

Adds context keys to the individual context targets for the context kind that `contextKind` specifies and the variation that `variationId` specifies. Returns an error if this causes the flag to target the same context key in multiple variations.

##### Parameters

- `values`: List of context keys.
- `contextKind`: (Optional) Context kind to target, defaults to `user`
- `variationId`: ID of a variation on the flag.

Here's an example:

```json
{
	"environmentKey": "environment-key-123abc",
	"instructions": [{
		"kind": "addTargets",
		"values": ["context-key-123abc", "context-key-456def"],
		"variationId": "2f43f67c-3e4e-4945-a18a-26559378ca00"
	}]
}
```

#### addUserTargets

Adds user keys to the individual user targets for the variation that `variationId` specifies. Returns an error if this causes the flag to target the same user key in multiple variations. If you are working with contexts, use `addTargets` instead of this instruction.

##### Parameters

- `values`: List of user keys.
- `variationId`: ID of a variation on the flag.

Here's an example:

```json
{
	"environmentKey": "environment-key-123abc",
	"instructions": [{
		"kind": "addUserTargets",
		"values": ["user-key-123abc", "user-key-456def"],
		"variationId": "2f43f67c-3e4e-4945-a18a-26559378ca00"
	}]
}
```

#### addValuesToClause

Adds `values` to the values of the clause that `ruleId` and `clauseId` indicate. Does not update the context kind, attribute, or operator.

##### Parameters

- `ruleId`: ID of a rule in the flag.
- `clauseId`: ID of a clause in that rule.
- `values`: Array of strings, case sensitive.

Here's an example:

```json
{
	"environmentKey": "environment-key-123abc",
	"instructions": [{
		"kind": "addValuesToClause",
		"ruleId": "a902ef4a-2faf-4eaf-88e1-ecc356708a29",
		"clauseId": "10a58772-3121-400f-846b-b8a04e8944ed",
		"values": ["beta_testers"]
	}]
}
```

#### addVariation

Adds a variation to the flag.

##### Parameters

- `value`: The variation value.
- `name`: (Optional) The variation name.
- `description`: (Optional) A description for the variation.

Here's an example:

```json
{
	"instructions": [ { "kind": "addVariation", "value": 20, "name": "New variation" } ]
}
```

#### clearTargets

Removes all individual targets from the variation that `variationId` specifies. This includes both user and non-user targets.

##### Parameters

- `variationId`: ID of a variation on the flag.

Here's an example:

```json
{
  "environmentKey": "environment-key-123abc",
  "instructions": [ { "kind": "clearTargets", "variationId": "2f43f67c-3e4e-4945-a18a-26559378ca00" } ]
}
```

#### clearUserTargets

Removes all individual user targets from the variation that `variationId` specifies. If you are working with contexts, use `clearTargets` instead of this instruction.

##### Parameters

- `variationId`: ID of a variation on the flag.

Here's an example:

```json
{
  "environmentKey": "environment-key-123abc",
  "instructions": [ { "kind": "clearUserTargets", "variationId": "2f43f67c-3e4e-4945-a18a-26559378ca00" } ]
}
```

#### removeClauses

Removes the clauses specified by `clauseIds` from the rule indicated by `ruleId`.

##### Parameters

- `ruleId`: ID of a rule in the flag.
- `clauseIds`: Array of IDs of clauses in the rule.

Here's an example:

```json
{
	"environmentKey": "environment-key-123abc",
	"instructions": [{
		"kind": "removeClauses",
		"ruleId": "a902ef4a-2faf-4eaf-88e1-ecc356708a29",
		"clauseIds": ["10a58772-3121-400f-846b-b8a04e8944ed", "36a461dc-235e-4b08-97b9-73ce9365873e"]
	}]
}
```

#### removePrerequisite

Removes the prerequisite flag indicated by `key`. Does nothing if this prerequisite does not exist.

##### Parameters

- `key`: Flag key of an existing prerequisite flag.

Here's an example:

```json
{
  "environmentKey": "environment-key-123abc",
  "instructions": [ { "kind": "removePrerequisite", "key": "prereq-flag-key-123abc" } ]
}
```

#### removeRule

Removes the targeting rule specified by `ruleId`. Does nothing if the rule does not exist.

##### Parameters

- `ruleId`: ID of a rule in the flag.

Here's an example:

```json
{
  "environmentKey": "environment-key-123abc",
  "instructions": [ { "kind": "removeRule", "ruleId": "a902ef4a-2faf-4eaf-88e1-ecc356708a29" } ]
}
```

#### removeTargets

Removes context keys from the individual context targets for the context kind that `contextKind` specifies and the variation that `variationId` specifies. Does nothing if the flag does not target the context keys.

##### Parameters

- `values`: List of context keys.
- `contextKind`: (Optional) Context kind to target, defaults to `user`
- `variationId`: ID of a flag variation.

Here's an example:

```json
{
	"environmentKey": "environment-key-123abc",
	"instructions": [{
		"kind": "removeTargets",
		"values": ["context-key-123abc", "context-key-456def"],
		"variationId": "2f43f67c-3e4e-4945-a18a-26559378ca00"
	}]
}
```

#### removeUserTargets

Removes user keys from the individual user targets for the variation that `variationId` specifies. Does nothing if the flag does not target the user keys. If you are working with contexts, use `removeTargets` instead of this instruction.

##### Parameters

- `values`: List of user keys.
- `variationId`: ID of a flag variation.

Here's an example:

```json
{
	"environmentKey": "environment-key-123abc",
	"instructions": [{
		"kind": "removeUserTargets",
		"values": ["user-key-123abc", "user-key-456def"],
		"variationId": "2f43f67c-3e4e-4945-a18a-26559378ca00"
	}]
}
```

#### removeValuesFromClause

Removes `values` from the values of the clause indicated by `ruleId` and `clauseId`. Does not update the context kind, attribute, or operator.

##### Parameters

- `ruleId`: ID of a rule in the flag.
- `clauseId`: ID of a clause in that rule.
- `values`: Array of strings, case sensitive.

Here's an example:

```json
{
	"environmentKey": "environment-key-123abc",
	"instructions": [{
		"kind": "removeValuesFromClause",
		"ruleId": "a902ef4a-2faf-4eaf-88e1-ecc356708a29",
		"clauseId": "10a58772-3121-400f-846b-b8a04e8944ed",
		"values": ["beta_testers"]
	}]
}
```

#### removeVariation

Removes a variation from the flag.

##### Parameters

- `variationId`: ID of a variation of the flag to remove.

Here's an example:

```json
{
	"instructions": [ { "kind": "removeVariation", "variationId": "2f43f67c-3e4e-4945-a18a-26559378ca00" } ]
}
```

#### reorderRules

Rearranges the rules to match the order given in `ruleIds`. Returns an error if `ruleIds` does not match the current set of rules on the flag.

##### Parameters

- `ruleIds`: Array of IDs of all rules in the flag.

Here's an example:

```json
{
	"environmentKey": "environment-key-123abc",
	"instructions": [{
		"kind": "reorderRules",
		"ruleIds": ["a902ef4a-2faf-4eaf-88e1-ecc356708a29", "63c238d1-835d-435e-8f21-c8d5e40b2a3d"]
	}]
}
```

#### replacePrerequisites

Removes all existing prerequisites and replaces them with the list you provide.

##### Parameters

- `prerequisites`: A list of prerequisites. Each item in the list must include a flag `key` and `variationId`.

Here's an example:

```json
{
  "environmentKey": "environment-key-123abc",
  "instructions": [
    {
      "kind": "replacePrerequisites",
      "prerequisites": [
        {
          "key": "prereq-flag-key-123abc",
          "variationId": "10a58772-3121-400f-846b-b8a04e8944ed"
        },
        {
          "key": "another-prereq-flag-key-456def",
          "variationId": "e5830889-1ec5-4b0c-9cc9-c48790090c43"
        }
      ]
    }
  ]
}
```

#### replaceRules

Removes all targeting rules for the flag and replaces them with the list you provide.

##### Parameters

- `rules`: A list of rules.

Here's an example:

```json
{
  "environmentKey": "environment-key-123abc",
  "instructions": [
    {
      "kind": "replaceRules",
      "rules": [
        {
          "variationId": "2f43f67c-3e4e-4945-a18a-26559378ca00",
          "description": "My new rule",
          "clauses": [
            {
              "contextKind": "user",
              "attribute": "segmentMatch",
              "op": "segmentMatch",
              "values": ["test"]
            }
          ],
          "trackEvents": true
        }
      ]
    }
  ]
}
```

#### replaceTargets

Removes all existing targeting and replaces it with the list of targets you provide.

##### Parameters

- `targets`: A list of context targeting. Each item in the list includes an optional `contextKind` that defaults to `user`, a required `variationId`, and a required list of `values`.

Here's an example:

```json
{
  "environmentKey": "environment-key-123abc",
  "instructions": [
    {
      "kind": "replaceTargets",
      "targets": [
        {
          "contextKind": "user",
          "variationId": "2f43f67c-3e4e-4945-a18a-26559378ca00",
          "values": ["user-key-123abc"]
        },
        {
          "contextKind": "device",
          "variationId": "e5830889-1ec5-4b0c-9cc9-c48790090c43",
          "values": ["device-key-456def"]
        }
      ]
    }    
  ]
}
```

#### replaceUserTargets

Removes all existing user targeting and replaces it with the list of targets you provide. In the list of targets, you must include a target for each of the flag's variations. If you are working with contexts, use `replaceTargets` instead of this instruction.

##### Parameters

- `targets`: A list of user targeting. Each item in the list must include a `variationId` and a list of `values`.

Here's an example:

```json
{
  "environmentKey": "environment-key-123abc",
  "instructions": [
    {
      "kind": "replaceUserTargets",
      "targets": [
        {
          "variationId": "2f43f67c-3e4e-4945-a18a-26559378ca00",
          "values": ["user-key-123abc", "user-key-456def"]
        },
        {
          "variationId": "e5830889-1ec5-4b0c-9cc9-c48790090c43",
          "values": ["user-key-789ghi"]
        }
      ]
    }
  ]
}
```

#### updateClause

Replaces the clause indicated by `ruleId` and `clauseId` with `clause`.

##### Parameters

- `ruleId`: ID of a rule in the flag.
- `clauseId`: ID of a clause in that rule.
- `clause`: New `clause` object, with `contextKind` (string), `attribute` (string), `op` (string), `negate` (boolean), and `values` (array of strings, numbers, or dates) properties. The `contextKind`, `attribute`, and `values` are case sensitive. The `op` must be lower-case.

Here's an example:

```json
{
  "environmentKey": "environment-key-123abc",
  "instructions": [{
    "kind": "updateClause",
    "ruleId": "a902ef4a-2faf-4eaf-88e1-ecc356708a29",
    "clauseId": "10c7462a-2062-45ba-a8bb-dfb3de0f8af5",
    "clause": {
      "contextKind": "user",
      "attribute": "country",
      "op": "in",
      "negate": false,
      "values": ["Mexico", "Canada"]
    }
  }]
}
```

#### updateDefaultVariation

Updates the default on or off variation of the flag.

##### Parameters

- `onVariationValue`: (Optional) The value of the variation of the new on variation.
- `offVariationValue`: (Optional) The value of the variation of the new off variation

Here's an example:

```json
{
	"instructions": [ { "kind": "updateDefaultVariation", "OnVariationValue": true, "OffVariationValue": false } ]
}
```

#### updateFallthroughVariationOrRollout

Updates the default or "fallthrough" rule for the flag, which the flag serves when a context matches none of the targeting rules. The rule can serve either the variation that `variationId` indicates, or a percentage rollout that `rolloutWeights` and `rolloutBucketBy` indicate.

##### Parameters

- `variationId`: ID of a variation of the flag.

or

- `rolloutWeights`: Map of `variationId` to weight, in thousandths of a percent (0-100000).
- `rolloutBucketBy`: (Optional) Context attribute available in the specified `rolloutContextKind`.
- `rolloutContextKind`: (Optional) Context kind, defaults to `user`

Here's an example that uses a `variationId`:

```json
{
	"environmentKey": "environment-key-123abc",
	"instructions": [{
		"kind": "updateFallthroughVariationOrRollout",
		"variationId": "2f43f67c-3e4e-4945-a18a-26559378ca00"
	}]
}
```

Here's an example that uses a percentage rollout:

```json
{
	"environmentKey": "environment-key-123abc",
	"instructions": [{
		"kind": "updateFallthroughVariationOrRollout",
		"rolloutContextKind": "user",
		"rolloutWeights": {
			"2f43f67c-3e4e-4945-a18a-26559378ca00": 15000, // serve 15% this variation
			"e5830889-1ec5-4b0c-9cc9-c48790090c43": 85000  // serve 85% this variation
		}
	}]
}
```

#### updateOffVariation

Updates the default off variation to `variationId`. The flag serves the default off variation when the flag's targeting is **Off**.

##### Parameters

- `variationId`: ID of a variation of the flag.

Here's an example:

```json
{
  "environmentKey": "environment-key-123abc",
  "instructions": [ { "kind": "updateOffVariation", "variationId": "2f43f67c-3e4e-4945-a18a-26559378ca00" } ]
}
```

#### updatePrerequisite

Changes the prerequisite flag that `key` indicates to use the variation that `variationId` indicates. Returns an error if this prerequisite does not exist.

##### Parameters

- `key`: Flag key of an existing prerequisite flag.
- `variationId`: ID of a variation of the prerequisite flag.

Here's an example:

```json
{
	"environmentKey": "environment-key-123abc",
	"instructions": [{
		"kind": "updatePrerequisite",
		"key": "example-prereq-flag-key",
		"variationId": "2f43f67c-3e4e-4945-a18a-26559378ca00"
	}]
}
```

#### updateRuleDescription

Updates the description of the feature flag rule.

##### Parameters

- `description`: The new human-readable description for this rule.
- `ruleId`: The ID of the rule. You can retrieve this by making a GET request for the flag.

Here's an example:

```json
{
	"environmentKey": "environment-key-123abc",
	"instructions": [{
		"kind": "updateRuleDescription",
		"description": "New rule description",
		"ruleId": "a902ef4a-2faf-4eaf-88e1-ecc356708a29"
	}]
}
```

#### updateRuleTrackEvents

Updates whether or not LaunchDarkly tracks events for the feature flag associated with this rule.

##### Parameters

- `ruleId`: The ID of the rule. You can retrieve this by making a GET request for the flag.
- `trackEvents`: Whether or not events are tracked.

Here's an example:

```json
{
	"environmentKey": "environment-key-123abc",
	"instructions": [{
		"kind": "updateRuleTrackEvents",
		"ruleId": "a902ef4a-2faf-4eaf-88e1-ecc356708a29",
		"trackEvents": true
	}]
}
```

#### updateRuleVariationOrRollout

Updates what `ruleId` serves when its clauses evaluate to true. The rule can serve either the variation that `variationId` indicates, or a percent rollout that `rolloutWeights` and `rolloutBucketBy` indicate.

##### Parameters

- `ruleId`: ID of a rule in the flag.
- `variationId`: ID of a variation of the flag.

  or

- `rolloutWeights`: Map of `variationId` to weight, in thousandths of a percent (0-100000).
- `rolloutBucketBy`: (Optional) Context attribute available in the specified `rolloutContextKind`.
- `rolloutContextKind`: (Optional) Context kind, defaults to `user`

Here's an example:

```json
{
	"environmentKey": "environment-key-123abc",
	"instructions": [{
		"kind": "updateRuleVariationOrRollout",
		"ruleId": "a902ef4a-2faf-4eaf-88e1-ecc356708a29",
		"variationId": "2f43f67c-3e4e-4945-a18a-26559378ca00"
	}]
}
```

#### updateTrackEvents

Updates whether or not LaunchDarkly tracks events for the feature flag, for all rules.

##### Parameters

- `trackEvents`: Whether or not events are tracked.

Here's an example:

```json
{
  "environmentKey": "environment-key-123abc",
  "instructions": [ { "kind": "updateTrackEvents", "trackEvents": true } ]
}
```

#### updateTrackEventsFallthrough

Updates whether or not LaunchDarkly tracks events for the feature flag, for the default rule.

##### Parameters

- `trackEvents`: Whether or not events are tracked.

Here's an example:

```json
{
  "environmentKey": "environment-key-123abc",
  "instructions": [ { "kind": "updateTrackEventsFallthrough", "trackEvents": true } ]
}
```

#### updateVariation

Updates a variation of the flag.

##### Parameters

- `variationId`: The ID of the variation to update.
- `name`: (Optional) The updated variation name.
- `value`: (Optional) The updated variation value.
- `description`: (Optional) The updated variation description.

Here's an example:

```json
{
	"instructions": [ { "kind": "updateVariation", "variationId": "2f43f67c-3e4e-4945-a18a-26559378ca00", "value": 20 } ]
}
```

</details><br />

<details>
<summary>Click to expand instructions for <strong>updating flag settings</strong></summary>

These instructions do not require the `environmentKey` parameter. They make changes that apply to the flag across all environments.

#### addCustomProperties

Adds a new custom property to the feature flag. Custom properties are used to associate feature flags with LaunchDarkly integrations. For example, if you create an integration with an issue tracking service, you may want to associate a flag with a list of issues related to a feature's development.

##### Parameters

 - `key`: The custom property key.
 - `name`: The custom property name.
 - `values`: A list of the associated values for the custom property.

Here's an example:

```json
{
	"instructions": [{
		"kind": "addCustomProperties",
		"key": "example-custom-property",
		"name": "Example custom property",
		"values": ["value1", "value2"]
	}]
}
```

#### addTags

Adds tags to the feature flag.

##### Parameters

- `values`: A list of tags to add.

Here's an example:

```json
{
  "instructions": [ { "kind": "addTags", "values": ["tag1", "tag2"] } ]
}
```

#### makeFlagPermanent

Marks the feature flag as permanent. LaunchDarkly does not prompt you to remove permanent flags, even if one variation is rolled out to all your customers.

Here's an example:

```json
{
  "instructions": [ { "kind": "makeFlagPermanent" } ]
}
```

#### makeFlagTemporary

Marks the feature flag as temporary.

Here's an example:

```json
{
  "instructions": [ { "kind": "makeFlagTemporary" } ]
}
```

#### removeCustomProperties

Removes the associated values from a custom property. If all the associated values are removed, this instruction also removes the custom property.

##### Parameters

 - `key`: The custom property key.
 - `values`: A list of the associated values to remove from the custom property.

```json
{
	"instructions": [{
		"kind": "replaceCustomProperties",
		"key": "example-custom-property",
		"values": ["value1", "value2"]
	}]
}
```

#### removeMaintainer

Removes the flag's maintainer. To set a new maintainer, use the flag's **Settings** tab in the LaunchDarkly user interface.

Here's an example:

```json
{
  "instructions": [ { "kind": "removeMaintainer" } ]
}
```

#### removeTags

Removes tags from the feature flag.

##### Parameters

- `values`: A list of tags to remove.

Here's an example:

```json
{
  "instructions": [ { "kind": "removeTags", "values": ["tag1", "tag2"] } ]
}
```

#### replaceCustomProperties

Replaces the existing associated values for a custom property with the new values.

##### Parameters

 - `key`: The custom property key.
 - `name`: The custom property name.
 - `values`: A list of the new associated values for the custom property.

Here's an example:

```json
{
 "instructions": [{
   "kind": "replaceCustomProperties",
   "key": "example-custom-property",
   "name": "Example custom property",
   "values": ["value1", "value2"]
 }]
}
```

#### turnOffClientSideAvailability

Turns off client-side SDK availability for the flag. This is equivalent to unchecking the **SDKs using Mobile Key** and/or **SDKs using client-side ID** boxes for the flag. If you're using a client-side or mobile SDK, you must expose your feature flags in order for the client-side or mobile SDKs to evaluate them.

##### Parameters

- `value`: Use "usingMobileKey" to turn off availability for mobile SDKs. Use "usingEnvironmentId" to turn on availability for client-side SDKs.

Here's an example:

```json
{
  "instructions": [ { "kind": "turnOffClientSideAvailability", "value": "usingMobileKey" } ]
}
```

#### turnOnClientSideAvailability

Turns on client-side SDK availability for the flag. This is equivalent to unchecking the **SDKs using Mobile Key** and/or **SDKs using client-side ID** boxes for the flag. If you're using a client-side or mobile SDK, you must expose your feature flags in order for the client-side or mobile SDKs to evaluate them.

##### Parameters

- `value`: Use "usingMobileKey" to turn on availability for mobile SDKs. Use "usingEnvironmentId" to turn on availability for client-side SDKs.

Here's an example:

```json
{
  "instructions": [ { "kind": "turnOnClientSideAvailability", "value": "usingMobileKey" } ]
}
```

#### updateDescription

Updates the feature flag description.

##### Parameters

- `value`: The new description.

Here's an example:

```json
{
  "instructions": [ { "kind": "updateDescription", "value": "Updated flag description" } ]
}
```
#### updateMaintainerMember

Updates the maintainer of the flag to an existing member and removes the existing maintainer.

##### Parameters

- `value`: The ID of the member.

Here's an example:

```json
{
  "instructions": [ { "kind": "updateMaintainerMember", "value": "61e9b714fd47591727db558a" } ]
}
```

#### updateMaintainerTeam

Updates the maintainer of the flag to an existing team and removes the existing maintainer.

##### Parameters

- `value`: The key of the team.

Here's an example:

```json
{
  "instructions": [ { "kind": "updateMaintainerTeam", "value": "example-team-key" } ]
}
```

#### updateName

Updates the feature flag name.

##### Parameters

- `value`: The new name.

Here's an example:

```json
{
  "instructions": [ { "kind": "updateName", "value": "Updated flag name" } ]
}
```

</details><br />

<details>
<summary>Click to expand instructions for <strong>updating the flag lifecycle</strong></summary>

These instructions do not require the `environmentKey` parameter. They make changes that apply to the flag across all environments.

#### archiveFlag

Archives the feature flag. This retires it from LaunchDarkly without deleting it. You cannot archive a flag that is a prerequisite of other flags.

```json
{
  "instructions": [ { "kind": "archiveFlag" } ]
}
```

#### deleteFlag

Deletes the feature flag and its rules. You cannot restore a deleted flag. If this flag is requested again, the flag value defined in code will be returned for all contexts.

Here's an example:

```json
{
  "instructions": [ { "kind": "deleteFlag" } ]
}
```

#### deprecateFlag

Deprecates the feature flag. This hides it from the live flags list without archiving or deleting it.

Here's an example:

```json
{
  "instructions": [ { "kind": "deprecateFlag" } ]
}
```

#### restoreDeprecatedFlag

Restores the feature flag if it was previously deprecated.

Here's an example:

```json
{
  "instructions": [ { "kind": "restoreDeprecatedFlag" } ]
}
```

#### restoreFlag

Restores the feature flag if it was previously archived.

Here's an example:

```json
{
  "instructions": [ { "kind": "restoreFlag" } ]
}
```

</details>

### Using JSON patches on a feature flag

If you do not include the semantic patch header described above, you can use a [JSON patch](/reference#updates-using-json-patch) or [JSON merge patch](https://datatracker.ietf.org/doc/html/rfc7386) representation of the desired changes.

In the JSON patch representation, use a JSON pointer in the `path` element to describe what field to change. Use the [Get feature flag](/tag/Feature-flags#operation/getFeatureFlag) endpoint to find the field you want to update.

There are a few special cases to keep in mind when determining the value of the `path` element:

  * To add an individual target to a specific variation if the flag variation already has individual targets, the path for the JSON patch operation is:

  ```json
  [
    {
      "op": "add",
      "path": "/environments/devint/targets/0/values/-",
      "value": "TestClient10"
    }
  ]
  ```

  * To add an individual target to a specific variation if the flag variation does not already have individual targets, the path for the JSON patch operation is:

  ```json
  [
    {
      "op": "add",
      "path": "/environments/devint/targets/-",
      "value": { "variation": 0, "values": ["TestClient10"] }
    }
  ]
  ```

  * To add a flag to a release pipeline, the path for the JSON patch operation is:

  ```json
  [
    {
      "op": "add",
      "path": "/releasePipelineKey",
      "value": "example-release-pipeline-key"
    }
  ]
  ```

### Required approvals
If a request attempts to alter a flag configuration in an environment where approvals are required for the flag, the request will fail with a 405. Changes to the flag configuration in that environment will require creating an [approval request](/tag/Approvals) or a [workflow](/tag/Workflows).

### Conflicts
If a flag configuration change made through this endpoint would cause a pending scheduled change or approval request to fail, this endpoint will return a 400. You can ignore this check by adding an `ignoreConflicts` query parameter set to `true`.

### Migration flags
For migration flags, the cohort information is included in the `rules` property of a flag's response. You can update cohorts by updating `rules`. Default cohort information is included in the `fallthrough` property of a flag's response. You can update the default cohort by updating `fallthrough`.
When you update the rollout for a cohort or the default cohort through the API, provide a rollout instead of a single `variationId`.
To learn more, read [Migration Flags](https://docs.launchdarkly.com/home/flag-types/migration-flags).


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectKey The project key
 @param featureFlagKey The feature flag key. The key identifies the flag in your code.
 @return ApiPatchFeatureFlagRequest
*/
func (a *FeatureFlagsApiService) PatchFeatureFlag(ctx context.Context, projectKey string, featureFlagKey string) ApiPatchFeatureFlagRequest {
	return ApiPatchFeatureFlagRequest{
		ApiService: a,
		ctx: ctx,
		projectKey: projectKey,
		featureFlagKey: featureFlagKey,
	}
}

// Execute executes the request
//  @return FeatureFlag
func (a *FeatureFlagsApiService) PatchFeatureFlagExecute(r ApiPatchFeatureFlagRequest) (*FeatureFlag, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPatch
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *FeatureFlag
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "FeatureFlagsApiService.PatchFeatureFlag")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v2/flags/{projectKey}/{featureFlagKey}"
	localVarPath = strings.Replace(localVarPath, "{"+"projectKey"+"}", url.PathEscape(parameterToString(r.projectKey, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"featureFlagKey"+"}", url.PathEscape(parameterToString(r.featureFlagKey, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.patchWithComment == nil {
		return localVarReturnValue, nil, reportError("patchWithComment is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.patchWithComment
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKey"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		if localVarHTTPResponse.StatusCode == 400 {
			var v InvalidRequestErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 401 {
			var v UnauthorizedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 404 {
			var v NotFoundErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 405 {
			var v MethodNotAllowedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 409 {
			var v StatusConflictErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 429 {
			var v RateLimitedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiPostFeatureFlagRequest struct {
	ctx context.Context
	ApiService *FeatureFlagsApiService
	projectKey string
	featureFlagBody *FeatureFlagBody
	clone *string
}

func (r ApiPostFeatureFlagRequest) FeatureFlagBody(featureFlagBody FeatureFlagBody) ApiPostFeatureFlagRequest {
	r.featureFlagBody = &featureFlagBody
	return r
}

// The key of the feature flag to be cloned. The key identifies the flag in your code. For example, setting &#x60;clone&#x3D;flagKey&#x60; copies the full targeting configuration for all environments, including &#x60;on/off&#x60; state, from the original flag to the new flag.
func (r ApiPostFeatureFlagRequest) Clone(clone string) ApiPostFeatureFlagRequest {
	r.clone = &clone
	return r
}

func (r ApiPostFeatureFlagRequest) Execute() (*FeatureFlag, *http.Response, error) {
	return r.ApiService.PostFeatureFlagExecute(r)
}

/*
PostFeatureFlag Create a feature flag

Create a feature flag with the given name, key, and variations.

### Creating a migration flag

When you create a migration flag, the variations are pre-determined based on the number of stages in the migration.
To create a migration flag, omit the `variations` and `defaults` information. Instead, provide a `purpose` of `migration`, and `migrationSettings`. If you create a migration flag with six stages, `contextKind` is required. Otherwise, it should be omitted.

Here's an example:

```json
{
  "key": "flag-key-123",
  "purpose": "migration",
  "migrationSettings": {
    "stageCount": 6,
    "contextKind": "account"
  }
}
```

To learn more, read [Migration Flags](https://docs.launchdarkly.com/home/flag-types/migration-flags).


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectKey The project key
 @return ApiPostFeatureFlagRequest
*/
func (a *FeatureFlagsApiService) PostFeatureFlag(ctx context.Context, projectKey string) ApiPostFeatureFlagRequest {
	return ApiPostFeatureFlagRequest{
		ApiService: a,
		ctx: ctx,
		projectKey: projectKey,
	}
}

// Execute executes the request
//  @return FeatureFlag
func (a *FeatureFlagsApiService) PostFeatureFlagExecute(r ApiPostFeatureFlagRequest) (*FeatureFlag, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *FeatureFlag
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "FeatureFlagsApiService.PostFeatureFlag")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/v2/flags/{projectKey}"
	localVarPath = strings.Replace(localVarPath, "{"+"projectKey"+"}", url.PathEscape(parameterToString(r.projectKey, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.featureFlagBody == nil {
		return localVarReturnValue, nil, reportError("featureFlagBody is required and must be specified")
	}

	if r.clone != nil {
		localVarQueryParams.Add("clone", parameterToString(*r.clone, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.featureFlagBody
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKey"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		if localVarHTTPResponse.StatusCode == 400 {
			var v InvalidRequestErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 401 {
			var v UnauthorizedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 409 {
			var v StatusConflictErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 429 {
			var v RateLimitedErrorRep
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}
