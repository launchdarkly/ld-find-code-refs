/*
LaunchDarkly REST API

This documentation describes LaunchDarkly's REST API.  To access the complete OpenAPI spec directly, use [Get OpenAPI spec](https://launchdarkly.com/docs/api/other/get-openapi-spec).  ## Authentication  LaunchDarkly's REST API uses the HTTPS protocol with a minimum TLS version of 1.2.  All REST API resources are authenticated with either [personal or service access tokens](https://launchdarkly.com/docs/home/account/api), or session cookies. Other authentication mechanisms are not supported. You can manage personal access tokens on your [**Authorization**](https://app.launchdarkly.com/settings/authorization) page in the LaunchDarkly UI.  LaunchDarkly also has SDK keys, mobile keys, and client-side IDs that are used by our server-side SDKs, mobile SDKs, and JavaScript-based SDKs, respectively. **These keys cannot be used to access our REST API**. These keys are environment-specific, and can only perform read-only operations such as fetching feature flag settings.  | Auth mechanism                                                                                  | Allowed resources                                                                                     | Use cases                                          | | ----------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- | -------------------------------------------------- | | [Personal or service access tokens](https://launchdarkly.com/docs/home/account/api) | Can be customized on a per-token basis                                                                | Building scripts, custom integrations, data export. | | SDK keys                                                                                        | Can only access read-only resources specific to server-side SDKs. Restricted to a single environment. | Server-side SDKs                     | | Mobile keys                                                                                     | Can only access read-only resources specific to mobile SDKs, and only for flags marked available to mobile keys. Restricted to a single environment.           | Mobile SDKs                                        | | Client-side ID                                                                                  | Can only access read-only resources specific to JavaScript-based client-side SDKs, and only for flags marked available to client-side. Restricted to a single environment.           | Client-side JavaScript                             |  > #### Keep your access tokens and SDK keys private > > Access tokens should _never_ be exposed in untrusted contexts. Never put an access token in client-side JavaScript, or embed it in a mobile application. LaunchDarkly has special mobile keys that you can embed in mobile apps. If you accidentally expose an access token or SDK key, you can reset it from your [**Authorization**](https://app.launchdarkly.com/settings/authorization) page. > > The client-side ID is safe to embed in untrusted contexts. It's designed for use in client-side JavaScript.  ### Authentication using request header  The preferred way to authenticate with the API is by adding an `Authorization` header containing your access token to your requests. The value of the `Authorization` header must be your access token.  Manage personal access tokens from the [**Authorization**](https://app.launchdarkly.com/settings/authorization) page.  ### Authentication using session cookie  For testing purposes, you can make API calls directly from your web browser. If you are logged in to the LaunchDarkly application, the API will use your existing session to authenticate calls.  If you have a [role](https://launchdarkly.com/docs/home/account/built-in-roles) other than Admin, or have a [custom role](https://launchdarkly.com/docs/home/account/custom-roles) defined, you may not have permission to perform some API calls. You will receive a `401` response code in that case.  > ### Modifying the Origin header causes an error > > LaunchDarkly validates that the Origin header for any API request authenticated by a session cookie matches the expected Origin header. The expected Origin header is `https://app.launchdarkly.com`. > > If the Origin header does not match what's expected, LaunchDarkly returns an error. This error can prevent the LaunchDarkly app from working correctly. > > Any browser extension that intentionally changes the Origin header can cause this problem. For example, the `Allow-Control-Allow-Origin: *` Chrome extension changes the Origin header to `http://evil.com` and causes the app to fail. > > To prevent this error, do not modify your Origin header. > > LaunchDarkly does not require origin matching when authenticating with an access token, so this issue does not affect normal API usage.  ## Representations  All resources expect and return JSON response bodies. Error responses also send a JSON body. To learn more about the error format of the API, read [Errors](https://launchdarkly.com/docs/api#errors).  In practice this means that you always get a response with a `Content-Type` header set to `application/json`.  In addition, request bodies for `PATCH`, `POST`, and `PUT` requests must be encoded as JSON with a `Content-Type` header set to `application/json`.  ### Summary and detailed representations  When you fetch a list of resources, the response includes only the most important attributes of each resource. This is a _summary representation_ of the resource. When you fetch an individual resource, such as a single feature flag, you receive a _detailed representation_ of the resource.  The best way to find a detailed representation is to follow links. Every summary representation includes a link to its detailed representation.  ### Expanding responses  Sometimes the detailed representation of a resource does not include all of the attributes of the resource by default. If this is the case, the request method will clearly document this and describe which attributes you can include in an expanded response.  To include the additional attributes, append the `expand` request parameter to your request and add a comma-separated list of the attributes to include. For example, when you append `?expand=members,maintainers` to the [Get team](https://launchdarkly.com/docs/api/teams/get-team) endpoint, the expanded response includes both of these attributes.  ### Links and addressability  The best way to navigate the API is by following links. These are attributes in representations that link to other resources. The API always uses the same format for links:  - Links to other resources within the API are encapsulated in a `_links` object - If the resource has a corresponding link to HTML content on the site, it is stored in a special `_site` link  Each link has two attributes:  - An `href`, which contains the URL - A `type`, which describes the content type  For example, a feature resource might return the following:  ```json {   \"_links\": {     \"parent\": {       \"href\": \"/api/features\",       \"type\": \"application/json\"     },     \"self\": {       \"href\": \"/api/features/sort.order\",       \"type\": \"application/json\"     }   },   \"_site\": {     \"href\": \"/features/sort.order\",     \"type\": \"text/html\"   } } ```  From this, you can navigate to the parent collection of features by following the `parent` link, or navigate to the site page for the feature by following the `_site` link.  Collections are always represented as a JSON object with an `items` attribute containing an array of representations. Like all other representations, collections have `_links` defined at the top level.  Paginated collections include `first`, `last`, `next`, and `prev` links containing a URL with the respective set of elements in the collection.  ## Updates  Resources that accept partial updates use the `PATCH` verb. Most resources support the [JSON patch](https://launchdarkly.com/docs/api#updates-using-json-patch) format. Some resources also support the [JSON merge patch](https://launchdarkly.com/docs/api#updates-using-json-merge-patch) format, and some resources support the [semantic patch](https://launchdarkly.com/docs/api#updates-using-semantic-patch) format, which is a way to specify the modifications to perform as a set of executable instructions. Each resource supports optional [comments](https://launchdarkly.com/docs/api#updates-with-comments) that you can submit with updates. Comments appear in outgoing webhooks, the audit log, and other integrations.  When a resource supports both JSON patch and semantic patch, we document both in the request method. However, the specific request body fields and descriptions included in our documentation only match one type of patch or the other.  ### Updates using JSON patch  [JSON patch](https://datatracker.ietf.org/doc/html/rfc6902) is a way to specify the modifications to perform on a resource. JSON patch uses paths and a limited set of operations to describe how to transform the current state of the resource into a new state. JSON patch documents are always arrays, where each element contains an operation, a path to the field to update, and the new value.  For example, in this feature flag representation:  ```json {     \"name\": \"New recommendations engine\",     \"key\": \"engine.enable\",     \"description\": \"This is the description\",     ... } ``` You can change the feature flag's description with the following patch document:  ```json [{ \"op\": \"replace\", \"path\": \"/description\", \"value\": \"This is the new description\" }] ```  You can specify multiple modifications to perform in a single request. You can also test that certain preconditions are met before applying the patch:  ```json [   { \"op\": \"test\", \"path\": \"/version\", \"value\": 10 },   { \"op\": \"replace\", \"path\": \"/description\", \"value\": \"The new description\" } ] ```  The above patch request tests whether the feature flag's `version` is `10`, and if so, changes the feature flag's description.  Attributes that are not editable, such as a resource's `_links`, have names that start with an underscore.  ### Updates using JSON merge patch  [JSON merge patch](https://datatracker.ietf.org/doc/html/rfc7386) is another format for specifying the modifications to perform on a resource. JSON merge patch is less expressive than JSON patch. However, in many cases it is simpler to construct a merge patch document. For example, you can change a feature flag's description with the following merge patch document:  ```json {   \"description\": \"New flag description\" } ```  ### Updates using semantic patch  Some resources support the semantic patch format. A semantic patch is a way to specify the modifications to perform on a resource as a set of executable instructions.  Semantic patch allows you to be explicit about intent using precise, custom instructions. In many cases, you can define semantic patch instructions independently of the current state of the resource. This can be useful when defining a change that may be applied at a future date.  To make a semantic patch request, you must append `domain-model=launchdarkly.semanticpatch` to your `Content-Type` header.  Here's how:  ``` Content-Type: application/json; domain-model=launchdarkly.semanticpatch ```  If you call a semantic patch resource without this header, you will receive a `400` response because your semantic patch will be interpreted as a JSON patch.  The body of a semantic patch request takes the following properties:  * `comment` (string): (Optional) A description of the update. * `environmentKey` (string): (Required for some resources only) The environment key. * `instructions` (array): (Required) A list of actions the update should perform. Each action in the list must be an object with a `kind` property that indicates the instruction. If the instruction requires parameters, you must include those parameters as additional fields in the object. The documentation for each resource that supports semantic patch includes the available instructions and any additional parameters.  For example:  ```json {   \"comment\": \"optional comment\",   \"instructions\": [ {\"kind\": \"turnFlagOn\"} ] } ```  Semantic patches are not applied partially; either all of the instructions are applied or none of them are. If **any** instruction is invalid, the endpoint returns an error and will not change the resource. If all instructions are valid, the request succeeds and the resources are updated if necessary, or left unchanged if they are already in the state you request.  ### Updates with comments  You can submit optional comments with `PATCH` changes.  To submit a comment along with a JSON patch document, use the following format:  ```json {   \"comment\": \"This is a comment string\",   \"patch\": [{ \"op\": \"replace\", \"path\": \"/description\", \"value\": \"The new description\" }] } ```  To submit a comment along with a JSON merge patch document, use the following format:  ```json {   \"comment\": \"This is a comment string\",   \"merge\": { \"description\": \"New flag description\" } } ```  To submit a comment along with a semantic patch, use the following format:  ```json {   \"comment\": \"This is a comment string\",   \"instructions\": [ {\"kind\": \"turnFlagOn\"} ] } ```  ## Errors  The API always returns errors in a common format. Here's an example:  ```json {   \"code\": \"invalid_request\",   \"message\": \"A feature with that key already exists\",   \"id\": \"30ce6058-87da-11e4-b116-123b93f75cba\" } ```  The `code` indicates the general class of error. The `message` is a human-readable explanation of what went wrong. The `id` is a unique identifier. Use it when you're working with LaunchDarkly Support to debug a problem with a specific API call.  ### HTTP status error response codes  | Code | Definition        | Description                                                                                       | Possible Solution                                                | | ---- | ----------------- | ------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- | | 400  | Invalid request       | The request cannot be understood.                                    | Ensure JSON syntax in request body is correct.                   | | 401  | Invalid access token      | Requestor is unauthorized or does not have permission for this API call.                                                | Ensure your API access token is valid and has the appropriate permissions.                                     | | 403  | Forbidden         | Requestor does not have access to this resource.                                                | Ensure that the account member or access token has proper permissions set. | | 404  | Invalid resource identifier | The requested resource is not valid. | Ensure that the resource is correctly identified by ID or key. | | 405  | Method not allowed | The request method is not allowed on this resource. | Ensure that the HTTP verb is correct. | | 409  | Conflict          | The API request can not be completed because it conflicts with a concurrent API request. | Retry your request.                                              | | 422  | Unprocessable entity | The API request can not be completed because the update description can not be understood. | Ensure that the request body is correct for the type of patch you are using, either JSON patch or semantic patch. | 429  | Too many requests | Read [Rate limiting](https://launchdarkly.com/docs/api#rate-limiting).                                               | Wait and try again later.                                        |  ## CORS  The LaunchDarkly API supports Cross Origin Resource Sharing (CORS) for AJAX requests from any origin. If an `Origin` header is given in a request, it will be echoed as an explicitly allowed origin. Otherwise the request returns a wildcard, `Access-Control-Allow-Origin: *`. For more information on CORS, read the [CORS W3C Recommendation](http://www.w3.org/TR/cors). Example CORS headers might look like:  ```http Access-Control-Allow-Headers: Accept, Content-Type, Content-Length, Accept-Encoding, Authorization Access-Control-Allow-Methods: OPTIONS, GET, DELETE, PATCH Access-Control-Allow-Origin: * Access-Control-Max-Age: 300 ```  You can make authenticated CORS calls just as you would make same-origin calls, using either [token or session-based authentication](https://launchdarkly.com/docs/api#authentication). If you are using session authentication, you should set the `withCredentials` property for your `xhr` request to `true`. You should never expose your access tokens to untrusted entities.  ## Rate limiting  We use several rate limiting strategies to ensure the availability of our APIs. Rate-limited calls to our APIs return a `429` status code. Calls to our APIs include headers indicating the current rate limit status. The specific headers returned depend on the API route being called. The limits differ based on the route, authentication mechanism, and other factors. Routes that are not rate limited may not contain any of the headers described below.  > ### Rate limiting and SDKs > > LaunchDarkly SDKs are never rate limited and do not use the API endpoints defined here. LaunchDarkly uses a different set of approaches, including streaming/server-sent events and a global CDN, to ensure availability to the routes used by LaunchDarkly SDKs.  ### Global rate limits  Authenticated requests are subject to a global limit. This is the maximum number of calls that your account can make to the API per ten seconds. All service and personal access tokens on the account share this limit, so exceeding the limit with one access token will impact other tokens. Calls that are subject to global rate limits may return the headers below:  | Header name                    | Description                                                                      | | ------------------------------ | -------------------------------------------------------------------------------- | | `X-Ratelimit-Global-Remaining` | The maximum number of requests the account is permitted to make per ten seconds. | | `X-Ratelimit-Reset`            | The time at which the current rate limit window resets in epoch milliseconds.    |  We do not publicly document the specific number of calls that can be made globally. This limit may change, and we encourage clients to program against the specification, relying on the two headers defined above, rather than hardcoding to the current limit.  ### Route-level rate limits  Some authenticated routes have custom rate limits. These also reset every ten seconds. Any service or personal access tokens hitting the same route share this limit, so exceeding the limit with one access token may impact other tokens. Calls that are subject to route-level rate limits return the headers below:  | Header name                   | Description                                                                                           | | ----------------------------- | ----------------------------------------------------------------------------------------------------- | | `X-Ratelimit-Route-Remaining` | The maximum number of requests to the current route the account is permitted to make per ten seconds. | | `X-Ratelimit-Reset`           | The time at which the current rate limit window resets in epoch milliseconds.                         |  A _route_ represents a specific URL pattern and verb. For example, the [Delete environment](https://launchdarkly.com/docs/api/environments/delete-environment) endpoint is considered a single route, and each call to delete an environment counts against your route-level rate limit for that route.  We do not publicly document the specific number of calls that an account can make to each endpoint per ten seconds. These limits may change, and we encourage clients to program against the specification, relying on the two headers defined above, rather than hardcoding to the current limits.  ### IP-based rate limiting  We also employ IP-based rate limiting on some API routes. If you hit an IP-based rate limit, your API response will include a `Retry-After` header indicating how long to wait before re-trying the call. Clients must wait at least `Retry-After` seconds before making additional calls to our API, and should employ jitter and backoff strategies to avoid triggering rate limits again.  ## OpenAPI (Swagger) and client libraries  We have a [complete OpenAPI (Swagger) specification](https://app.launchdarkly.com/api/v2/openapi.json) for our API.  We auto-generate multiple client libraries based on our OpenAPI specification. To learn more, visit the [collection of client libraries on GitHub](https://github.com/search?q=topic%3Alaunchdarkly-api+org%3Alaunchdarkly&type=Repositories). You can also use this specification to generate client libraries to interact with our REST API in your language of choice.  Our OpenAPI specification is supported by several API-based tools such as Postman and Insomnia. In many cases, you can directly import our specification to explore our APIs.  ## Method overriding  Some firewalls and HTTP clients restrict the use of verbs other than `GET` and `POST`. In those environments, our API endpoints that use `DELETE`, `PATCH`, and `PUT` verbs are inaccessible.  To avoid this issue, our API supports the `X-HTTP-Method-Override` header, allowing clients to \"tunnel\" `DELETE`, `PATCH`, and `PUT` requests using a `POST` request.  For example, to call a `PATCH` endpoint using a `POST` request, you can include `X-HTTP-Method-Override:PATCH` as a header.  ## Beta resources  We sometimes release new API resources in **beta** status before we release them with general availability.  Resources that are in beta are still undergoing testing and development. They may change without notice, including becoming backwards incompatible.  We try to promote resources into general availability as quickly as possible. This happens after sufficient testing and when we're satisfied that we no longer need to make backwards-incompatible changes.  We mark beta resources with a \"Beta\" callout in our documentation, pictured below:  > ### This feature is in beta > > To use this feature, pass in a header including the `LD-API-Version` key with value set to `beta`. Use this header with each call. To learn more, read [Beta resources](https://launchdarkly.com/docs/api#beta-resources). > > Resources that are in beta are still undergoing testing and development. They may change without notice, including becoming backwards incompatible.  ### Using beta resources  To use a beta resource, you must include a header in the request. If you call a beta resource without this header, you receive a `403` response.  Use this header:  ``` LD-API-Version: beta ```  ## Federal environments  The version of LaunchDarkly that is available on domains controlled by the United States government is different from the version of LaunchDarkly available to the general public. If you are an employee or contractor for a United States federal agency and use LaunchDarkly in your work, you likely use the federal instance of LaunchDarkly.  If you are working in the federal instance of LaunchDarkly, the base URI for each request is `https://app.launchdarkly.us`.  To learn more, read [LaunchDarkly in federal environments](https://launchdarkly.com/docs/home/infrastructure/federal).  ## Versioning  We try hard to keep our REST API backwards compatible, but we occasionally have to make backwards-incompatible changes in the process of shipping new features. These breaking changes can cause unexpected behavior if you don't prepare for them accordingly.  Updates to our REST API include support for the latest features in LaunchDarkly. We also release a new version of our REST API every time we make a breaking change. We provide simultaneous support for multiple API versions so you can migrate from your current API version to a new version at your own pace.  ### Setting the API version per request  You can set the API version on a specific request by sending an `LD-API-Version` header, as shown in the example below:  ``` LD-API-Version: 20240415 ```  The header value is the version number of the API version you would like to request. The number for each version corresponds to the date the version was released in `yyyymmdd` format. In the example above the version `20240415` corresponds to April 15, 2024.  ### Setting the API version per access token  When you create an access token, you must specify a specific version of the API to use. This ensures that integrations using this token cannot be broken by version changes.  Tokens created before versioning was released have their version set to `20160426`, which is the version of the API that existed before the current versioning scheme, so that they continue working the same way they did before versioning.  If you would like to upgrade your integration to use a new API version, you can explicitly set the header described above.  > ### Best practice: Set the header for every client or integration > > We recommend that you set the API version header explicitly in any client or integration you build. > > Only rely on the access token API version during manual testing.  ### API version changelog  <table>   <tr>     <th>Version</th>     <th>Changes</th>     <th>End of life (EOL)</th>   </tr>   <tr>     <td>`20240415`</td>     <td>       <ul><li>Changed several endpoints from unpaginated to paginated. Use the `limit` and `offset` query parameters to page through the results.</li> <li>Changed the [list access tokens](https://launchdarkly.com/docs/api/access-tokens/get-tokens) endpoint: <ul><li>Response is now paginated with a default limit of `25`</li></ul></li> <li>Changed the [list account members](https://launchdarkly.com/docs/api/account-members/get-members) endpoint: <ul><li>The `accessCheck` filter is no longer available</li></ul></li> <li>Changed the [list custom roles](https://launchdarkly.com/docs/api/custom-roles/get-custom-roles) endpoint: <ul><li>Response is now paginated with a default limit of `20`</li></ul></li> <li>Changed the [list feature flags](https://launchdarkly.com/docs/api/feature-flags/get-feature-flags) endpoint: <ul><li>Response is now paginated with a default limit of `20`</li><li>The `environments` field is now only returned if the request is filtered by environment, using the `filterEnv` query parameter</li><li>The `followerId`, `hasDataExport`, `status`, `contextKindTargeted`, and `segmentTargeted` filters are no longer available</li><li>The `compare` query parameter is no longer available</li></ul></li> <li>Changed the [list segments](https://launchdarkly.com/docs/api/segments/get-segments) endpoint: <ul><li>Response is now paginated with a default limit of `20`</li></ul></li> <li>Changed the [list teams](https://launchdarkly.com/docs/api/teams/get-teams) endpoint: <ul><li>The `expand` parameter no longer supports including `projects` or `roles`</li><li>In paginated results, the maximum page size is now 100</li></ul></li> <li>Changed the [get workflows](https://launchdarkly.com/docs/api/workflows/get-workflows) endpoint: <ul><li>Response is now paginated with a default limit of `20`</li><li>The `_conflicts` field in the response is no longer available</li></ul></li> </ul>     </td>     <td>Current</td>   </tr>   <tr>     <td>`20220603`</td>     <td>       <ul><li>Changed the [list projects](https://launchdarkly.com/docs/api/projects/get-projects) return value:<ul><li>Response is now paginated with a default limit of `20`.</li><li>Added support for filter and sort.</li><li>The project `environments` field is now expandable. This field is omitted by default.</li></ul></li><li>Changed the [get project](https://launchdarkly.com/docs/api/projects/get-project) return value:<ul><li>The `environments` field is now expandable. This field is omitted by default.</li></ul></li></ul>     </td>     <td>2025-04-15</td>   </tr>   <tr>     <td>`20210729`</td>     <td>       <ul><li>Changed the [create approval request](https://launchdarkly.com/docs/api/approvals/post-approval-request) return value. It now returns HTTP Status Code `201` instead of `200`.</li><li> Changed the [get user](https://launchdarkly.com/docs/api/users/get-user) return value. It now returns a user record, not a user. </li><li>Added additional optional fields to environment, segments, flags, members, and segments, including the ability to create big segments. </li><li> Added default values for flag variations when new environments are created. </li><li>Added filtering and pagination for getting flags and members, including `limit`, `number`, `filter`, and `sort` query parameters. </li><li>Added endpoints for expiring user targets for flags and segments, scheduled changes, access tokens, Relay Proxy configuration, integrations and subscriptions, and approvals. </li></ul>     </td>     <td>2023-06-03</td>   </tr>   <tr>     <td>`20191212`</td>     <td>       <ul><li>[List feature flags](https://launchdarkly.com/docs/api/feature-flags/get-feature-flags) now defaults to sending summaries of feature flag configurations, equivalent to setting the query parameter `summary=true`. Summaries omit flag targeting rules and individual user targets from the payload. </li><li> Added endpoints for flags, flag status, projects, environments, audit logs, members, users, custom roles, segments, usage, streams, events, and data export. </li></ul>     </td>     <td>2022-07-29</td>   </tr>   <tr>     <td>`20160426`</td>     <td>       <ul><li>Initial versioning of API. Tokens created before versioning have their version set to this.</li></ul>     </td>     <td>2020-12-12</td>   </tr> </table>  To learn more about how EOL is determined, read LaunchDarkly's [End of Life (EOL) Policy](https://launchdarkly.com/policies/end-of-life-policy/). 

API version: 2.0
Contact: support@launchdarkly.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package ldapi

import (
	"encoding/json"
)

// TreatmentResultRep struct for TreatmentResultRep
type TreatmentResultRep struct {
	// The ID of the treatment
	TreatmentId *string `json:"treatmentId,omitempty"`
	// The name of the treatment
	TreatmentName *string `json:"treatmentName,omitempty"`
	// The average value of the variation in this sample. It doesn’t capture the uncertainty in the measurement, so it should not be the only measurement you use to make decisions.
	Mean *float32 `json:"mean,omitempty"`
	// The mean of the data, with no priors effecting the result.
	DataMean *float32 `json:"dataMean,omitempty"`
	// The standard deviation of the data, with no priors effecting the result.
	DataStdDev *float32 `json:"dataStdDev,omitempty"`
	CredibleInterval *CredibleIntervalRep `json:"credibleInterval,omitempty"`
	// The likelihood that this variation has the biggest effect on the primary metric. The variation with the highest probability is likely the best of the variations you're testing
	PBest *float32 `json:"pBest,omitempty"`
	// Estimates of the relative difference between this treatment's mean and the mean of each other treatment
	RelativeDifferences []RelativeDifferenceRep `json:"relativeDifferences,omitempty"`
	// The number of units exposed to this treatment that have event values, including those that are configured to default to 0
	Units *int64 `json:"units,omitempty"`
	// The number of units exposed to this treatment.
	Traffic *int64 `json:"traffic,omitempty"`
	// The sum of the event values for the units exposed to this treatment.
	EventValuesSum *float32 `json:"eventValuesSum,omitempty"`
	Distribution *Distribution `json:"distribution,omitempty"`
	// The outcome-covariate correlation
	Correlation *float32 `json:"correlation,omitempty"`
	// The ratio of the outcome SD to covariate SD
	StandardDeviationRatio *float32 `json:"standardDeviationRatio,omitempty"`
	// The imbalance between the covariate mean for the arm and the covariate mean for the experiment
	CovariateImbalance *float32 `json:"covariateImbalance,omitempty"`
	// The reduction in variance resulting from CUPED
	VarianceReduction *float32 `json:"varianceReduction,omitempty"`
	// The model used to calculate the results. Parameters specific to this model will be defined under the field under the same name
	Model *string `json:"model,omitempty"`
	BayesianNormal *BayesianNormalStatsRep `json:"bayesianNormal,omitempty"`
	BayesianBeta *BayesianBetaBinomialStatsRep `json:"bayesianBeta,omitempty"`
}

// NewTreatmentResultRep instantiates a new TreatmentResultRep object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewTreatmentResultRep() *TreatmentResultRep {
	this := TreatmentResultRep{}
	return &this
}

// NewTreatmentResultRepWithDefaults instantiates a new TreatmentResultRep object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewTreatmentResultRepWithDefaults() *TreatmentResultRep {
	this := TreatmentResultRep{}
	return &this
}

// GetTreatmentId returns the TreatmentId field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetTreatmentId() string {
	if o == nil || o.TreatmentId == nil {
		var ret string
		return ret
	}
	return *o.TreatmentId
}

// GetTreatmentIdOk returns a tuple with the TreatmentId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetTreatmentIdOk() (*string, bool) {
	if o == nil || o.TreatmentId == nil {
		return nil, false
	}
	return o.TreatmentId, true
}

// HasTreatmentId returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasTreatmentId() bool {
	if o != nil && o.TreatmentId != nil {
		return true
	}

	return false
}

// SetTreatmentId gets a reference to the given string and assigns it to the TreatmentId field.
func (o *TreatmentResultRep) SetTreatmentId(v string) {
	o.TreatmentId = &v
}

// GetTreatmentName returns the TreatmentName field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetTreatmentName() string {
	if o == nil || o.TreatmentName == nil {
		var ret string
		return ret
	}
	return *o.TreatmentName
}

// GetTreatmentNameOk returns a tuple with the TreatmentName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetTreatmentNameOk() (*string, bool) {
	if o == nil || o.TreatmentName == nil {
		return nil, false
	}
	return o.TreatmentName, true
}

// HasTreatmentName returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasTreatmentName() bool {
	if o != nil && o.TreatmentName != nil {
		return true
	}

	return false
}

// SetTreatmentName gets a reference to the given string and assigns it to the TreatmentName field.
func (o *TreatmentResultRep) SetTreatmentName(v string) {
	o.TreatmentName = &v
}

// GetMean returns the Mean field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetMean() float32 {
	if o == nil || o.Mean == nil {
		var ret float32
		return ret
	}
	return *o.Mean
}

// GetMeanOk returns a tuple with the Mean field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetMeanOk() (*float32, bool) {
	if o == nil || o.Mean == nil {
		return nil, false
	}
	return o.Mean, true
}

// HasMean returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasMean() bool {
	if o != nil && o.Mean != nil {
		return true
	}

	return false
}

// SetMean gets a reference to the given float32 and assigns it to the Mean field.
func (o *TreatmentResultRep) SetMean(v float32) {
	o.Mean = &v
}

// GetDataMean returns the DataMean field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetDataMean() float32 {
	if o == nil || o.DataMean == nil {
		var ret float32
		return ret
	}
	return *o.DataMean
}

// GetDataMeanOk returns a tuple with the DataMean field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetDataMeanOk() (*float32, bool) {
	if o == nil || o.DataMean == nil {
		return nil, false
	}
	return o.DataMean, true
}

// HasDataMean returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasDataMean() bool {
	if o != nil && o.DataMean != nil {
		return true
	}

	return false
}

// SetDataMean gets a reference to the given float32 and assigns it to the DataMean field.
func (o *TreatmentResultRep) SetDataMean(v float32) {
	o.DataMean = &v
}

// GetDataStdDev returns the DataStdDev field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetDataStdDev() float32 {
	if o == nil || o.DataStdDev == nil {
		var ret float32
		return ret
	}
	return *o.DataStdDev
}

// GetDataStdDevOk returns a tuple with the DataStdDev field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetDataStdDevOk() (*float32, bool) {
	if o == nil || o.DataStdDev == nil {
		return nil, false
	}
	return o.DataStdDev, true
}

// HasDataStdDev returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasDataStdDev() bool {
	if o != nil && o.DataStdDev != nil {
		return true
	}

	return false
}

// SetDataStdDev gets a reference to the given float32 and assigns it to the DataStdDev field.
func (o *TreatmentResultRep) SetDataStdDev(v float32) {
	o.DataStdDev = &v
}

// GetCredibleInterval returns the CredibleInterval field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetCredibleInterval() CredibleIntervalRep {
	if o == nil || o.CredibleInterval == nil {
		var ret CredibleIntervalRep
		return ret
	}
	return *o.CredibleInterval
}

// GetCredibleIntervalOk returns a tuple with the CredibleInterval field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetCredibleIntervalOk() (*CredibleIntervalRep, bool) {
	if o == nil || o.CredibleInterval == nil {
		return nil, false
	}
	return o.CredibleInterval, true
}

// HasCredibleInterval returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasCredibleInterval() bool {
	if o != nil && o.CredibleInterval != nil {
		return true
	}

	return false
}

// SetCredibleInterval gets a reference to the given CredibleIntervalRep and assigns it to the CredibleInterval field.
func (o *TreatmentResultRep) SetCredibleInterval(v CredibleIntervalRep) {
	o.CredibleInterval = &v
}

// GetPBest returns the PBest field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetPBest() float32 {
	if o == nil || o.PBest == nil {
		var ret float32
		return ret
	}
	return *o.PBest
}

// GetPBestOk returns a tuple with the PBest field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetPBestOk() (*float32, bool) {
	if o == nil || o.PBest == nil {
		return nil, false
	}
	return o.PBest, true
}

// HasPBest returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasPBest() bool {
	if o != nil && o.PBest != nil {
		return true
	}

	return false
}

// SetPBest gets a reference to the given float32 and assigns it to the PBest field.
func (o *TreatmentResultRep) SetPBest(v float32) {
	o.PBest = &v
}

// GetRelativeDifferences returns the RelativeDifferences field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetRelativeDifferences() []RelativeDifferenceRep {
	if o == nil || o.RelativeDifferences == nil {
		var ret []RelativeDifferenceRep
		return ret
	}
	return o.RelativeDifferences
}

// GetRelativeDifferencesOk returns a tuple with the RelativeDifferences field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetRelativeDifferencesOk() ([]RelativeDifferenceRep, bool) {
	if o == nil || o.RelativeDifferences == nil {
		return nil, false
	}
	return o.RelativeDifferences, true
}

// HasRelativeDifferences returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasRelativeDifferences() bool {
	if o != nil && o.RelativeDifferences != nil {
		return true
	}

	return false
}

// SetRelativeDifferences gets a reference to the given []RelativeDifferenceRep and assigns it to the RelativeDifferences field.
func (o *TreatmentResultRep) SetRelativeDifferences(v []RelativeDifferenceRep) {
	o.RelativeDifferences = v
}

// GetUnits returns the Units field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetUnits() int64 {
	if o == nil || o.Units == nil {
		var ret int64
		return ret
	}
	return *o.Units
}

// GetUnitsOk returns a tuple with the Units field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetUnitsOk() (*int64, bool) {
	if o == nil || o.Units == nil {
		return nil, false
	}
	return o.Units, true
}

// HasUnits returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasUnits() bool {
	if o != nil && o.Units != nil {
		return true
	}

	return false
}

// SetUnits gets a reference to the given int64 and assigns it to the Units field.
func (o *TreatmentResultRep) SetUnits(v int64) {
	o.Units = &v
}

// GetTraffic returns the Traffic field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetTraffic() int64 {
	if o == nil || o.Traffic == nil {
		var ret int64
		return ret
	}
	return *o.Traffic
}

// GetTrafficOk returns a tuple with the Traffic field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetTrafficOk() (*int64, bool) {
	if o == nil || o.Traffic == nil {
		return nil, false
	}
	return o.Traffic, true
}

// HasTraffic returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasTraffic() bool {
	if o != nil && o.Traffic != nil {
		return true
	}

	return false
}

// SetTraffic gets a reference to the given int64 and assigns it to the Traffic field.
func (o *TreatmentResultRep) SetTraffic(v int64) {
	o.Traffic = &v
}

// GetEventValuesSum returns the EventValuesSum field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetEventValuesSum() float32 {
	if o == nil || o.EventValuesSum == nil {
		var ret float32
		return ret
	}
	return *o.EventValuesSum
}

// GetEventValuesSumOk returns a tuple with the EventValuesSum field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetEventValuesSumOk() (*float32, bool) {
	if o == nil || o.EventValuesSum == nil {
		return nil, false
	}
	return o.EventValuesSum, true
}

// HasEventValuesSum returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasEventValuesSum() bool {
	if o != nil && o.EventValuesSum != nil {
		return true
	}

	return false
}

// SetEventValuesSum gets a reference to the given float32 and assigns it to the EventValuesSum field.
func (o *TreatmentResultRep) SetEventValuesSum(v float32) {
	o.EventValuesSum = &v
}

// GetDistribution returns the Distribution field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetDistribution() Distribution {
	if o == nil || o.Distribution == nil {
		var ret Distribution
		return ret
	}
	return *o.Distribution
}

// GetDistributionOk returns a tuple with the Distribution field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetDistributionOk() (*Distribution, bool) {
	if o == nil || o.Distribution == nil {
		return nil, false
	}
	return o.Distribution, true
}

// HasDistribution returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasDistribution() bool {
	if o != nil && o.Distribution != nil {
		return true
	}

	return false
}

// SetDistribution gets a reference to the given Distribution and assigns it to the Distribution field.
func (o *TreatmentResultRep) SetDistribution(v Distribution) {
	o.Distribution = &v
}

// GetCorrelation returns the Correlation field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetCorrelation() float32 {
	if o == nil || o.Correlation == nil {
		var ret float32
		return ret
	}
	return *o.Correlation
}

// GetCorrelationOk returns a tuple with the Correlation field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetCorrelationOk() (*float32, bool) {
	if o == nil || o.Correlation == nil {
		return nil, false
	}
	return o.Correlation, true
}

// HasCorrelation returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasCorrelation() bool {
	if o != nil && o.Correlation != nil {
		return true
	}

	return false
}

// SetCorrelation gets a reference to the given float32 and assigns it to the Correlation field.
func (o *TreatmentResultRep) SetCorrelation(v float32) {
	o.Correlation = &v
}

// GetStandardDeviationRatio returns the StandardDeviationRatio field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetStandardDeviationRatio() float32 {
	if o == nil || o.StandardDeviationRatio == nil {
		var ret float32
		return ret
	}
	return *o.StandardDeviationRatio
}

// GetStandardDeviationRatioOk returns a tuple with the StandardDeviationRatio field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetStandardDeviationRatioOk() (*float32, bool) {
	if o == nil || o.StandardDeviationRatio == nil {
		return nil, false
	}
	return o.StandardDeviationRatio, true
}

// HasStandardDeviationRatio returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasStandardDeviationRatio() bool {
	if o != nil && o.StandardDeviationRatio != nil {
		return true
	}

	return false
}

// SetStandardDeviationRatio gets a reference to the given float32 and assigns it to the StandardDeviationRatio field.
func (o *TreatmentResultRep) SetStandardDeviationRatio(v float32) {
	o.StandardDeviationRatio = &v
}

// GetCovariateImbalance returns the CovariateImbalance field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetCovariateImbalance() float32 {
	if o == nil || o.CovariateImbalance == nil {
		var ret float32
		return ret
	}
	return *o.CovariateImbalance
}

// GetCovariateImbalanceOk returns a tuple with the CovariateImbalance field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetCovariateImbalanceOk() (*float32, bool) {
	if o == nil || o.CovariateImbalance == nil {
		return nil, false
	}
	return o.CovariateImbalance, true
}

// HasCovariateImbalance returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasCovariateImbalance() bool {
	if o != nil && o.CovariateImbalance != nil {
		return true
	}

	return false
}

// SetCovariateImbalance gets a reference to the given float32 and assigns it to the CovariateImbalance field.
func (o *TreatmentResultRep) SetCovariateImbalance(v float32) {
	o.CovariateImbalance = &v
}

// GetVarianceReduction returns the VarianceReduction field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetVarianceReduction() float32 {
	if o == nil || o.VarianceReduction == nil {
		var ret float32
		return ret
	}
	return *o.VarianceReduction
}

// GetVarianceReductionOk returns a tuple with the VarianceReduction field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetVarianceReductionOk() (*float32, bool) {
	if o == nil || o.VarianceReduction == nil {
		return nil, false
	}
	return o.VarianceReduction, true
}

// HasVarianceReduction returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasVarianceReduction() bool {
	if o != nil && o.VarianceReduction != nil {
		return true
	}

	return false
}

// SetVarianceReduction gets a reference to the given float32 and assigns it to the VarianceReduction field.
func (o *TreatmentResultRep) SetVarianceReduction(v float32) {
	o.VarianceReduction = &v
}

// GetModel returns the Model field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetModel() string {
	if o == nil || o.Model == nil {
		var ret string
		return ret
	}
	return *o.Model
}

// GetModelOk returns a tuple with the Model field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetModelOk() (*string, bool) {
	if o == nil || o.Model == nil {
		return nil, false
	}
	return o.Model, true
}

// HasModel returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasModel() bool {
	if o != nil && o.Model != nil {
		return true
	}

	return false
}

// SetModel gets a reference to the given string and assigns it to the Model field.
func (o *TreatmentResultRep) SetModel(v string) {
	o.Model = &v
}

// GetBayesianNormal returns the BayesianNormal field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetBayesianNormal() BayesianNormalStatsRep {
	if o == nil || o.BayesianNormal == nil {
		var ret BayesianNormalStatsRep
		return ret
	}
	return *o.BayesianNormal
}

// GetBayesianNormalOk returns a tuple with the BayesianNormal field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetBayesianNormalOk() (*BayesianNormalStatsRep, bool) {
	if o == nil || o.BayesianNormal == nil {
		return nil, false
	}
	return o.BayesianNormal, true
}

// HasBayesianNormal returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasBayesianNormal() bool {
	if o != nil && o.BayesianNormal != nil {
		return true
	}

	return false
}

// SetBayesianNormal gets a reference to the given BayesianNormalStatsRep and assigns it to the BayesianNormal field.
func (o *TreatmentResultRep) SetBayesianNormal(v BayesianNormalStatsRep) {
	o.BayesianNormal = &v
}

// GetBayesianBeta returns the BayesianBeta field value if set, zero value otherwise.
func (o *TreatmentResultRep) GetBayesianBeta() BayesianBetaBinomialStatsRep {
	if o == nil || o.BayesianBeta == nil {
		var ret BayesianBetaBinomialStatsRep
		return ret
	}
	return *o.BayesianBeta
}

// GetBayesianBetaOk returns a tuple with the BayesianBeta field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TreatmentResultRep) GetBayesianBetaOk() (*BayesianBetaBinomialStatsRep, bool) {
	if o == nil || o.BayesianBeta == nil {
		return nil, false
	}
	return o.BayesianBeta, true
}

// HasBayesianBeta returns a boolean if a field has been set.
func (o *TreatmentResultRep) HasBayesianBeta() bool {
	if o != nil && o.BayesianBeta != nil {
		return true
	}

	return false
}

// SetBayesianBeta gets a reference to the given BayesianBetaBinomialStatsRep and assigns it to the BayesianBeta field.
func (o *TreatmentResultRep) SetBayesianBeta(v BayesianBetaBinomialStatsRep) {
	o.BayesianBeta = &v
}

func (o TreatmentResultRep) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.TreatmentId != nil {
		toSerialize["treatmentId"] = o.TreatmentId
	}
	if o.TreatmentName != nil {
		toSerialize["treatmentName"] = o.TreatmentName
	}
	if o.Mean != nil {
		toSerialize["mean"] = o.Mean
	}
	if o.DataMean != nil {
		toSerialize["dataMean"] = o.DataMean
	}
	if o.DataStdDev != nil {
		toSerialize["dataStdDev"] = o.DataStdDev
	}
	if o.CredibleInterval != nil {
		toSerialize["credibleInterval"] = o.CredibleInterval
	}
	if o.PBest != nil {
		toSerialize["pBest"] = o.PBest
	}
	if o.RelativeDifferences != nil {
		toSerialize["relativeDifferences"] = o.RelativeDifferences
	}
	if o.Units != nil {
		toSerialize["units"] = o.Units
	}
	if o.Traffic != nil {
		toSerialize["traffic"] = o.Traffic
	}
	if o.EventValuesSum != nil {
		toSerialize["eventValuesSum"] = o.EventValuesSum
	}
	if o.Distribution != nil {
		toSerialize["distribution"] = o.Distribution
	}
	if o.Correlation != nil {
		toSerialize["correlation"] = o.Correlation
	}
	if o.StandardDeviationRatio != nil {
		toSerialize["standardDeviationRatio"] = o.StandardDeviationRatio
	}
	if o.CovariateImbalance != nil {
		toSerialize["covariateImbalance"] = o.CovariateImbalance
	}
	if o.VarianceReduction != nil {
		toSerialize["varianceReduction"] = o.VarianceReduction
	}
	if o.Model != nil {
		toSerialize["model"] = o.Model
	}
	if o.BayesianNormal != nil {
		toSerialize["bayesianNormal"] = o.BayesianNormal
	}
	if o.BayesianBeta != nil {
		toSerialize["bayesianBeta"] = o.BayesianBeta
	}
	return json.Marshal(toSerialize)
}

type NullableTreatmentResultRep struct {
	value *TreatmentResultRep
	isSet bool
}

func (v NullableTreatmentResultRep) Get() *TreatmentResultRep {
	return v.value
}

func (v *NullableTreatmentResultRep) Set(val *TreatmentResultRep) {
	v.value = val
	v.isSet = true
}

func (v NullableTreatmentResultRep) IsSet() bool {
	return v.isSet
}

func (v *NullableTreatmentResultRep) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableTreatmentResultRep(val *TreatmentResultRep) *NullableTreatmentResultRep {
	return &NullableTreatmentResultRep{value: val, isSet: true}
}

func (v NullableTreatmentResultRep) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableTreatmentResultRep) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


