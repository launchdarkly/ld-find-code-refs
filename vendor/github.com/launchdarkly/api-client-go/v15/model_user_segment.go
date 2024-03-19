/*
LaunchDarkly REST API

# Overview  ## Authentication  LaunchDarkly's REST API uses the HTTPS protocol with a minimum TLS version of 1.2.  All REST API resources are authenticated with either [personal or service access tokens](https://docs.launchdarkly.com/home/account-security/api-access-tokens), or session cookies. Other authentication mechanisms are not supported. You can manage personal access tokens on your [**Account settings**](https://app.launchdarkly.com/settings/tokens) page.  LaunchDarkly also has SDK keys, mobile keys, and client-side IDs that are used by our server-side SDKs, mobile SDKs, and JavaScript-based SDKs, respectively. **These keys cannot be used to access our REST API**. These keys are environment-specific, and can only perform read-only operations such as fetching feature flag settings.  | Auth mechanism                                                                                  | Allowed resources                                                                                     | Use cases                                          | | ----------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- | -------------------------------------------------- | | [Personal or service access tokens](https://docs.launchdarkly.com/home/account-security/api-access-tokens) | Can be customized on a per-token basis                                                                | Building scripts, custom integrations, data export. | | SDK keys                                                                                        | Can only access read-only resources specific to server-side SDKs. Restricted to a single environment. | Server-side SDKs                     | | Mobile keys                                                                                     | Can only access read-only resources specific to mobile SDKs, and only for flags marked available to mobile keys. Restricted to a single environment.           | Mobile SDKs                                        | | Client-side ID                                                                                  | Can only access read-only resources specific to JavaScript-based client-side SDKs, and only for flags marked available to client-side. Restricted to a single environment.           | Client-side JavaScript                             |  > #### Keep your access tokens and SDK keys private > > Access tokens should _never_ be exposed in untrusted contexts. Never put an access token in client-side JavaScript, or embed it in a mobile application. LaunchDarkly has special mobile keys that you can embed in mobile apps. If you accidentally expose an access token or SDK key, you can reset it from your [**Account settings**](https://app.launchdarkly.com/settings/tokens) page. > > The client-side ID is safe to embed in untrusted contexts. It's designed for use in client-side JavaScript.  ### Authentication using request header  The preferred way to authenticate with the API is by adding an `Authorization` header containing your access token to your requests. The value of the `Authorization` header must be your access token.  Manage personal access tokens from the [**Account settings**](https://app.launchdarkly.com/settings/tokens) page.  ### Authentication using session cookie  For testing purposes, you can make API calls directly from your web browser. If you are logged in to the LaunchDarkly application, the API will use your existing session to authenticate calls.  If you have a [role](https://docs.launchdarkly.com/home/team/built-in-roles) other than Admin, or have a [custom role](https://docs.launchdarkly.com/home/team/custom-roles) defined, you may not have permission to perform some API calls. You will receive a `401` response code in that case.  > ### Modifying the Origin header causes an error > > LaunchDarkly validates that the Origin header for any API request authenticated by a session cookie matches the expected Origin header. The expected Origin header is `https://app.launchdarkly.com`. > > If the Origin header does not match what's expected, LaunchDarkly returns an error. This error can prevent the LaunchDarkly app from working correctly. > > Any browser extension that intentionally changes the Origin header can cause this problem. For example, the `Allow-Control-Allow-Origin: *` Chrome extension changes the Origin header to `http://evil.com` and causes the app to fail. > > To prevent this error, do not modify your Origin header. > > LaunchDarkly does not require origin matching when authenticating with an access token, so this issue does not affect normal API usage.  ## Representations  All resources expect and return JSON response bodies. Error responses also send a JSON body. To learn more about the error format of the API, read [Errors](/#section/Overview/Errors).  In practice this means that you always get a response with a `Content-Type` header set to `application/json`.  In addition, request bodies for `PATCH`, `POST`, and `PUT` requests must be encoded as JSON with a `Content-Type` header set to `application/json`.  ### Summary and detailed representations  When you fetch a list of resources, the response includes only the most important attributes of each resource. This is a _summary representation_ of the resource. When you fetch an individual resource, such as a single feature flag, you receive a _detailed representation_ of the resource.  The best way to find a detailed representation is to follow links. Every summary representation includes a link to its detailed representation.  ### Expanding responses  Sometimes the detailed representation of a resource does not include all of the attributes of the resource by default. If this is the case, the request method will clearly document this and describe which attributes you can include in an expanded response.  To include the additional attributes, append the `expand` request parameter to your request and add a comma-separated list of the attributes to include. For example, when you append `?expand=members,roles` to the [Get team](/tag/Teams#operation/getTeam) endpoint, the expanded response includes both of these attributes.  ### Links and addressability  The best way to navigate the API is by following links. These are attributes in representations that link to other resources. The API always uses the same format for links:  - Links to other resources within the API are encapsulated in a `_links` object - If the resource has a corresponding link to HTML content on the site, it is stored in a special `_site` link  Each link has two attributes:  - An `href`, which contains the URL - A `type`, which describes the content type  For example, a feature resource might return the following:  ```json {   \"_links\": {     \"parent\": {       \"href\": \"/api/features\",       \"type\": \"application/json\"     },     \"self\": {       \"href\": \"/api/features/sort.order\",       \"type\": \"application/json\"     }   },   \"_site\": {     \"href\": \"/features/sort.order\",     \"type\": \"text/html\"   } } ```  From this, you can navigate to the parent collection of features by following the `parent` link, or navigate to the site page for the feature by following the `_site` link.  Collections are always represented as a JSON object with an `items` attribute containing an array of representations. Like all other representations, collections have `_links` defined at the top level.  Paginated collections include `first`, `last`, `next`, and `prev` links containing a URL with the respective set of elements in the collection.  ## Updates  Resources that accept partial updates use the `PATCH` verb. Most resources support the [JSON patch](/reference#updates-using-json-patch) format. Some resources also support the [JSON merge patch](/reference#updates-using-json-merge-patch) format, and some resources support the [semantic patch](/reference#updates-using-semantic-patch) format, which is a way to specify the modifications to perform as a set of executable instructions. Each resource supports optional [comments](/reference#updates-with-comments) that you can submit with updates. Comments appear in outgoing webhooks, the audit log, and other integrations.  When a resource supports both JSON patch and semantic patch, we document both in the request method. However, the specific request body fields and descriptions included in our documentation only match one type of patch or the other.  ### Updates using JSON patch  [JSON patch](https://datatracker.ietf.org/doc/html/rfc6902) is a way to specify the modifications to perform on a resource. JSON patch uses paths and a limited set of operations to describe how to transform the current state of the resource into a new state. JSON patch documents are always arrays, where each element contains an operation, a path to the field to update, and the new value.  For example, in this feature flag representation:  ```json {     \"name\": \"New recommendations engine\",     \"key\": \"engine.enable\",     \"description\": \"This is the description\",     ... } ``` You can change the feature flag's description with the following patch document:  ```json [{ \"op\": \"replace\", \"path\": \"/description\", \"value\": \"This is the new description\" }] ```  You can specify multiple modifications to perform in a single request. You can also test that certain preconditions are met before applying the patch:  ```json [   { \"op\": \"test\", \"path\": \"/version\", \"value\": 10 },   { \"op\": \"replace\", \"path\": \"/description\", \"value\": \"The new description\" } ] ```  The above patch request tests whether the feature flag's `version` is `10`, and if so, changes the feature flag's description.  Attributes that are not editable, such as a resource's `_links`, have names that start with an underscore.  ### Updates using JSON merge patch  [JSON merge patch](https://datatracker.ietf.org/doc/html/rfc7386) is another format for specifying the modifications to perform on a resource. JSON merge patch is less expressive than JSON patch. However, in many cases it is simpler to construct a merge patch document. For example, you can change a feature flag's description with the following merge patch document:  ```json {   \"description\": \"New flag description\" } ```  ### Updates using semantic patch  Some resources support the semantic patch format. A semantic patch is a way to specify the modifications to perform on a resource as a set of executable instructions.  Semantic patch allows you to be explicit about intent using precise, custom instructions. In many cases, you can define semantic patch instructions independently of the current state of the resource. This can be useful when defining a change that may be applied at a future date.  To make a semantic patch request, you must append `domain-model=launchdarkly.semanticpatch` to your `Content-Type` header.  Here's how:  ``` Content-Type: application/json; domain-model=launchdarkly.semanticpatch ```  If you call a semantic patch resource without this header, you will receive a `400` response because your semantic patch will be interpreted as a JSON patch.  The body of a semantic patch request takes the following properties:  * `comment` (string): (Optional) A description of the update. * `environmentKey` (string): (Required for some resources only) The environment key. * `instructions` (array): (Required) A list of actions the update should perform. Each action in the list must be an object with a `kind` property that indicates the instruction. If the instruction requires parameters, you must include those parameters as additional fields in the object. The documentation for each resource that supports semantic patch includes the available instructions and any additional parameters.  For example:  ```json {   \"comment\": \"optional comment\",   \"instructions\": [ {\"kind\": \"turnFlagOn\"} ] } ```  If any instruction in the patch encounters an error, the endpoint returns an error and will not change the resource. In general, each instruction silently does nothing if the resource is already in the state you request.  ### Updates with comments  You can submit optional comments with `PATCH` changes.  To submit a comment along with a JSON patch document, use the following format:  ```json {   \"comment\": \"This is a comment string\",   \"patch\": [{ \"op\": \"replace\", \"path\": \"/description\", \"value\": \"The new description\" }] } ```  To submit a comment along with a JSON merge patch document, use the following format:  ```json {   \"comment\": \"This is a comment string\",   \"merge\": { \"description\": \"New flag description\" } } ```  To submit a comment along with a semantic patch, use the following format:  ```json {   \"comment\": \"This is a comment string\",   \"instructions\": [ {\"kind\": \"turnFlagOn\"} ] } ```  ## Errors  The API always returns errors in a common format. Here's an example:  ```json {   \"code\": \"invalid_request\",   \"message\": \"A feature with that key already exists\",   \"id\": \"30ce6058-87da-11e4-b116-123b93f75cba\" } ```  The `code` indicates the general class of error. The `message` is a human-readable explanation of what went wrong. The `id` is a unique identifier. Use it when you're working with LaunchDarkly Support to debug a problem with a specific API call.  ### HTTP status error response codes  | Code | Definition        | Description                                                                                       | Possible Solution                                                | | ---- | ----------------- | ------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- | | 400  | Invalid request       | The request cannot be understood.                                    | Ensure JSON syntax in request body is correct.                   | | 401  | Invalid access token      | Requestor is unauthorized or does not have permission for this API call.                                                | Ensure your API access token is valid and has the appropriate permissions.                                     | | 403  | Forbidden         | Requestor does not have access to this resource.                                                | Ensure that the account member or access token has proper permissions set. | | 404  | Invalid resource identifier | The requested resource is not valid. | Ensure that the resource is correctly identified by ID or key. | | 405  | Method not allowed | The request method is not allowed on this resource. | Ensure that the HTTP verb is correct. | | 409  | Conflict          | The API request can not be completed because it conflicts with a concurrent API request. | Retry your request.                                              | | 422  | Unprocessable entity | The API request can not be completed because the update description can not be understood. | Ensure that the request body is correct for the type of patch you are using, either JSON patch or semantic patch. | 429  | Too many requests | Read [Rate limiting](/#section/Overview/Rate-limiting).                                               | Wait and try again later.                                        |  ## CORS  The LaunchDarkly API supports Cross Origin Resource Sharing (CORS) for AJAX requests from any origin. If an `Origin` header is given in a request, it will be echoed as an explicitly allowed origin. Otherwise the request returns a wildcard, `Access-Control-Allow-Origin: *`. For more information on CORS, read the [CORS W3C Recommendation](http://www.w3.org/TR/cors). Example CORS headers might look like:  ```http Access-Control-Allow-Headers: Accept, Content-Type, Content-Length, Accept-Encoding, Authorization Access-Control-Allow-Methods: OPTIONS, GET, DELETE, PATCH Access-Control-Allow-Origin: * Access-Control-Max-Age: 300 ```  You can make authenticated CORS calls just as you would make same-origin calls, using either [token or session-based authentication](/#section/Overview/Authentication). If you are using session authentication, you should set the `withCredentials` property for your `xhr` request to `true`. You should never expose your access tokens to untrusted entities.  ## Rate limiting  We use several rate limiting strategies to ensure the availability of our APIs. Rate-limited calls to our APIs return a `429` status code. Calls to our APIs include headers indicating the current rate limit status. The specific headers returned depend on the API route being called. The limits differ based on the route, authentication mechanism, and other factors. Routes that are not rate limited may not contain any of the headers described below.  > ### Rate limiting and SDKs > > LaunchDarkly SDKs are never rate limited and do not use the API endpoints defined here. LaunchDarkly uses a different set of approaches, including streaming/server-sent events and a global CDN, to ensure availability to the routes used by LaunchDarkly SDKs.  ### Global rate limits  Authenticated requests are subject to a global limit. This is the maximum number of calls that your account can make to the API per ten seconds. All service and personal access tokens on the account share this limit, so exceeding the limit with one access token will impact other tokens. Calls that are subject to global rate limits may return the headers below:  | Header name                    | Description                                                                      | | ------------------------------ | -------------------------------------------------------------------------------- | | `X-Ratelimit-Global-Remaining` | The maximum number of requests the account is permitted to make per ten seconds. | | `X-Ratelimit-Reset`            | The time at which the current rate limit window resets in epoch milliseconds.    |  We do not publicly document the specific number of calls that can be made globally. This limit may change, and we encourage clients to program against the specification, relying on the two headers defined above, rather than hardcoding to the current limit.  ### Route-level rate limits  Some authenticated routes have custom rate limits. These also reset every ten seconds. Any service or personal access tokens hitting the same route share this limit, so exceeding the limit with one access token may impact other tokens. Calls that are subject to route-level rate limits return the headers below:  | Header name                   | Description                                                                                           | | ----------------------------- | ----------------------------------------------------------------------------------------------------- | | `X-Ratelimit-Route-Remaining` | The maximum number of requests to the current route the account is permitted to make per ten seconds. | | `X-Ratelimit-Reset`           | The time at which the current rate limit window resets in epoch milliseconds.                         |  A _route_ represents a specific URL pattern and verb. For example, the [Delete environment](/tag/Environments#operation/deleteEnvironment) endpoint is considered a single route, and each call to delete an environment counts against your route-level rate limit for that route.  We do not publicly document the specific number of calls that an account can make to each endpoint per ten seconds. These limits may change, and we encourage clients to program against the specification, relying on the two headers defined above, rather than hardcoding to the current limits.  ### IP-based rate limiting  We also employ IP-based rate limiting on some API routes. If you hit an IP-based rate limit, your API response will include a `Retry-After` header indicating how long to wait before re-trying the call. Clients must wait at least `Retry-After` seconds before making additional calls to our API, and should employ jitter and backoff strategies to avoid triggering rate limits again.  ## OpenAPI (Swagger) and client libraries  We have a [complete OpenAPI (Swagger) specification](https://app.launchdarkly.com/api/v2/openapi.json) for our API.  We auto-generate multiple client libraries based on our OpenAPI specification. To learn more, visit the [collection of client libraries on GitHub](https://github.com/search?q=topic%3Alaunchdarkly-api+org%3Alaunchdarkly&type=Repositories). You can also use this specification to generate client libraries to interact with our REST API in your language of choice.  Our OpenAPI specification is supported by several API-based tools such as Postman and Insomnia. In many cases, you can directly import our specification to explore our APIs.  ## Method overriding  Some firewalls and HTTP clients restrict the use of verbs other than `GET` and `POST`. In those environments, our API endpoints that use `DELETE`, `PATCH`, and `PUT` verbs are inaccessible.  To avoid this issue, our API supports the `X-HTTP-Method-Override` header, allowing clients to \"tunnel\" `DELETE`, `PATCH`, and `PUT` requests using a `POST` request.  For example, to call a `PATCH` endpoint using a `POST` request, you can include `X-HTTP-Method-Override:PATCH` as a header.  ## Beta resources  We sometimes release new API resources in **beta** status before we release them with general availability.  Resources that are in beta are still undergoing testing and development. They may change without notice, including becoming backwards incompatible.  We try to promote resources into general availability as quickly as possible. This happens after sufficient testing and when we're satisfied that we no longer need to make backwards-incompatible changes.  We mark beta resources with a \"Beta\" callout in our documentation, pictured below:  > ### This feature is in beta > > To use this feature, pass in a header including the `LD-API-Version` key with value set to `beta`. Use this header with each call. To learn more, read [Beta resources](/#section/Overview/Beta-resources). > > Resources that are in beta are still undergoing testing and development. They may change without notice, including becoming backwards incompatible.  ### Using beta resources  To use a beta resource, you must include a header in the request. If you call a beta resource without this header, you receive a `403` response.  Use this header:  ``` LD-API-Version: beta ```  ## Federal environments  The version of LaunchDarkly that is available on domains controlled by the United States government is different from the version of LaunchDarkly available to the general public. If you are an employee or contractor for a United States federal agency and use LaunchDarkly in your work, you likely use the federal instance of LaunchDarkly.  If you are working in the federal instance of LaunchDarkly, the base URI for each request is `https://app.launchdarkly.us`. In the \"Try it\" sandbox for each request, click the request path to view the complete resource path for the federal environment.  To learn more, read [LaunchDarkly in federal environments](https://docs.launchdarkly.com/home/advanced/federal).  ## Versioning  We try hard to keep our REST API backwards compatible, but we occasionally have to make backwards-incompatible changes in the process of shipping new features. These breaking changes can cause unexpected behavior if you don't prepare for them accordingly.  Updates to our REST API include support for the latest features in LaunchDarkly. We also release a new version of our REST API every time we make a breaking change. We provide simultaneous support for multiple API versions so you can migrate from your current API version to a new version at your own pace.  ### Setting the API version per request  You can set the API version on a specific request by sending an `LD-API-Version` header, as shown in the example below:  ``` LD-API-Version: 20220603 ```  The header value is the version number of the API version you would like to request. The number for each version corresponds to the date the version was released in `yyyymmdd` format. In the example above the version `20220603` corresponds to June 03, 2022.  ### Setting the API version per access token  When you create an access token, you must specify a specific version of the API to use. This ensures that integrations using this token cannot be broken by version changes.  Tokens created before versioning was released have their version set to `20160426`, which is the version of the API that existed before the current versioning scheme, so that they continue working the same way they did before versioning.  If you would like to upgrade your integration to use a new API version, you can explicitly set the header described above.  > ### Best practice: Set the header for every client or integration > > We recommend that you set the API version header explicitly in any client or integration you build. > > Only rely on the access token API version during manual testing.  ### API version changelog  |<div style=\"width:75px\">Version</div> | Changes | End of life (EOL) |---|---|---| | `20220603` | <ul><li>Changed the [list projects](/tag/Projects#operation/getProjects) return value:<ul><li>Response is now paginated with a default limit of `20`.</li><li>Added support for filter and sort.</li><li>The project `environments` field is now expandable. This field is omitted by default.</li></ul></li><li>Changed the [get project](/tag/Projects#operation/getProject) return value:<ul><li>The `environments` field is now expandable. This field is omitted by default.</li></ul></li></ul> | Current | | `20210729` | <ul><li>Changed the [create approval request](/tag/Approvals#operation/postApprovalRequest) return value. It now returns HTTP Status Code `201` instead of `200`.</li><li> Changed the [get users](/tag/Users#operation/getUser) return value. It now returns a user record, not a user. </li><li>Added additional optional fields to environment, segments, flags, members, and segments, including the ability to create big segments. </li><li> Added default values for flag variations when new environments are created. </li><li>Added filtering and pagination for getting flags and members, including `limit`, `number`, `filter`, and `sort` query parameters. </li><li>Added endpoints for expiring user targets for flags and segments, scheduled changes, access tokens, Relay Proxy configuration, integrations and subscriptions, and approvals. </li></ul> | 2023-06-03 | | `20191212` | <ul><li>[List feature flags](/tag/Feature-flags#operation/getFeatureFlags) now defaults to sending summaries of feature flag configurations, equivalent to setting the query parameter `summary=true`. Summaries omit flag targeting rules and individual user targets from the payload. </li><li> Added endpoints for flags, flag status, projects, environments, audit logs, members, users, custom roles, segments, usage, streams, events, and data export. </li></ul> | 2022-07-29 | | `20160426` | <ul><li>Initial versioning of API. Tokens created before versioning have their version set to this.</li></ul> | 2020-12-12 | 

API version: 2.0
Contact: support@launchdarkly.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package ldapi

import (
	"encoding/json"
)

// UserSegment struct for UserSegment
type UserSegment struct {
	// A human-friendly name for the segment.
	Name string `json:"name"`
	// A description of the segment's purpose. Defaults to <code>null</code> and is omitted in the response if not provided.
	Description *string `json:"description,omitempty"`
	// Tags for the segment. Defaults to an empty array.
	Tags []string `json:"tags"`
	CreationDate int64 `json:"creationDate"`
	LastModifiedDate int64 `json:"lastModifiedDate"`
	// A unique key used to reference the segment
	Key string `json:"key"`
	// An array of keys for included targets. Included individual targets are always segment members, regardless of segment rules. For list-based segments over 15,000 entries, also called big segments, this array is either empty or omitted.
	Included []string `json:"included,omitempty"`
	// An array of keys for excluded targets. Segment rules bypass individual excluded targets, so they will never be included based on rules. Excluded targets may still be included explicitly. This value is omitted for list-based segments over 15,000 entries, also called big segments.
	Excluded []string `json:"excluded,omitempty"`
	IncludedContexts []SegmentTarget `json:"includedContexts,omitempty"`
	ExcludedContexts []SegmentTarget `json:"excludedContexts,omitempty"`
	// The location and content type of related resources
	Links map[string]Link `json:"_links"`
	// An array of the targeting rules for this segment.
	Rules []UserSegmentRule `json:"rules"`
	// Version of the segment
	Version int32 `json:"version"`
	// Whether the segment has been deleted
	Deleted bool `json:"deleted"`
	Access *Access `json:"_access,omitempty"`
	// A list of flags targeting this segment. Only included when getting a single segment, using the <code>getSegment</code> endpoint.
	Flags []FlagListingRep `json:"_flags,omitempty"`
	// Whether this is a standard segment (<code>false</code>) or a big segment (<code>true</code>). Standard segments include rule-based segments and smaller list-based segments. Big segments include larger list-based segments and synced segments. If omitted, the segment is a standard segment.
	Unbounded *bool `json:"unbounded,omitempty"`
	// For big segments, the targeted context kind.
	UnboundedContextKind *string `json:"unboundedContextKind,omitempty"`
	// For big segments, how many times this segment has been created.
	Generation int32 `json:"generation"`
	UnboundedMetadata *SegmentMetadata `json:"_unboundedMetadata,omitempty"`
	// The external data store backing this segment. Only applies to synced segments.
	External *string `json:"_external,omitempty"`
	// The URL for the external data store backing this segment. Only applies to synced segments.
	ExternalLink *string `json:"_externalLink,omitempty"`
	// Whether an import is currently in progress for the specified segment. Only applies to big segments.
	ImportInProgress *bool `json:"_importInProgress,omitempty"`
}

// NewUserSegment instantiates a new UserSegment object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUserSegment(name string, tags []string, creationDate int64, lastModifiedDate int64, key string, links map[string]Link, rules []UserSegmentRule, version int32, deleted bool, generation int32) *UserSegment {
	this := UserSegment{}
	this.Name = name
	this.Tags = tags
	this.CreationDate = creationDate
	this.LastModifiedDate = lastModifiedDate
	this.Key = key
	this.Links = links
	this.Rules = rules
	this.Version = version
	this.Deleted = deleted
	this.Generation = generation
	return &this
}

// NewUserSegmentWithDefaults instantiates a new UserSegment object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUserSegmentWithDefaults() *UserSegment {
	this := UserSegment{}
	return &this
}

// GetName returns the Name field value
func (o *UserSegment) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *UserSegment) SetName(v string) {
	o.Name = v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *UserSegment) GetDescription() string {
	if o == nil || o.Description == nil {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetDescriptionOk() (*string, bool) {
	if o == nil || o.Description == nil {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *UserSegment) HasDescription() bool {
	if o != nil && o.Description != nil {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *UserSegment) SetDescription(v string) {
	o.Description = &v
}

// GetTags returns the Tags field value
func (o *UserSegment) GetTags() []string {
	if o == nil {
		var ret []string
		return ret
	}

	return o.Tags
}

// GetTagsOk returns a tuple with the Tags field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetTagsOk() ([]string, bool) {
	if o == nil {
		return nil, false
	}
	return o.Tags, true
}

// SetTags sets field value
func (o *UserSegment) SetTags(v []string) {
	o.Tags = v
}

// GetCreationDate returns the CreationDate field value
func (o *UserSegment) GetCreationDate() int64 {
	if o == nil {
		var ret int64
		return ret
	}

	return o.CreationDate
}

// GetCreationDateOk returns a tuple with the CreationDate field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetCreationDateOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.CreationDate, true
}

// SetCreationDate sets field value
func (o *UserSegment) SetCreationDate(v int64) {
	o.CreationDate = v
}

// GetLastModifiedDate returns the LastModifiedDate field value
func (o *UserSegment) GetLastModifiedDate() int64 {
	if o == nil {
		var ret int64
		return ret
	}

	return o.LastModifiedDate
}

// GetLastModifiedDateOk returns a tuple with the LastModifiedDate field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetLastModifiedDateOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.LastModifiedDate, true
}

// SetLastModifiedDate sets field value
func (o *UserSegment) SetLastModifiedDate(v int64) {
	o.LastModifiedDate = v
}

// GetKey returns the Key field value
func (o *UserSegment) GetKey() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Key
}

// GetKeyOk returns a tuple with the Key field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetKeyOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Key, true
}

// SetKey sets field value
func (o *UserSegment) SetKey(v string) {
	o.Key = v
}

// GetIncluded returns the Included field value if set, zero value otherwise.
func (o *UserSegment) GetIncluded() []string {
	if o == nil || o.Included == nil {
		var ret []string
		return ret
	}
	return o.Included
}

// GetIncludedOk returns a tuple with the Included field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetIncludedOk() ([]string, bool) {
	if o == nil || o.Included == nil {
		return nil, false
	}
	return o.Included, true
}

// HasIncluded returns a boolean if a field has been set.
func (o *UserSegment) HasIncluded() bool {
	if o != nil && o.Included != nil {
		return true
	}

	return false
}

// SetIncluded gets a reference to the given []string and assigns it to the Included field.
func (o *UserSegment) SetIncluded(v []string) {
	o.Included = v
}

// GetExcluded returns the Excluded field value if set, zero value otherwise.
func (o *UserSegment) GetExcluded() []string {
	if o == nil || o.Excluded == nil {
		var ret []string
		return ret
	}
	return o.Excluded
}

// GetExcludedOk returns a tuple with the Excluded field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetExcludedOk() ([]string, bool) {
	if o == nil || o.Excluded == nil {
		return nil, false
	}
	return o.Excluded, true
}

// HasExcluded returns a boolean if a field has been set.
func (o *UserSegment) HasExcluded() bool {
	if o != nil && o.Excluded != nil {
		return true
	}

	return false
}

// SetExcluded gets a reference to the given []string and assigns it to the Excluded field.
func (o *UserSegment) SetExcluded(v []string) {
	o.Excluded = v
}

// GetIncludedContexts returns the IncludedContexts field value if set, zero value otherwise.
func (o *UserSegment) GetIncludedContexts() []SegmentTarget {
	if o == nil || o.IncludedContexts == nil {
		var ret []SegmentTarget
		return ret
	}
	return o.IncludedContexts
}

// GetIncludedContextsOk returns a tuple with the IncludedContexts field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetIncludedContextsOk() ([]SegmentTarget, bool) {
	if o == nil || o.IncludedContexts == nil {
		return nil, false
	}
	return o.IncludedContexts, true
}

// HasIncludedContexts returns a boolean if a field has been set.
func (o *UserSegment) HasIncludedContexts() bool {
	if o != nil && o.IncludedContexts != nil {
		return true
	}

	return false
}

// SetIncludedContexts gets a reference to the given []SegmentTarget and assigns it to the IncludedContexts field.
func (o *UserSegment) SetIncludedContexts(v []SegmentTarget) {
	o.IncludedContexts = v
}

// GetExcludedContexts returns the ExcludedContexts field value if set, zero value otherwise.
func (o *UserSegment) GetExcludedContexts() []SegmentTarget {
	if o == nil || o.ExcludedContexts == nil {
		var ret []SegmentTarget
		return ret
	}
	return o.ExcludedContexts
}

// GetExcludedContextsOk returns a tuple with the ExcludedContexts field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetExcludedContextsOk() ([]SegmentTarget, bool) {
	if o == nil || o.ExcludedContexts == nil {
		return nil, false
	}
	return o.ExcludedContexts, true
}

// HasExcludedContexts returns a boolean if a field has been set.
func (o *UserSegment) HasExcludedContexts() bool {
	if o != nil && o.ExcludedContexts != nil {
		return true
	}

	return false
}

// SetExcludedContexts gets a reference to the given []SegmentTarget and assigns it to the ExcludedContexts field.
func (o *UserSegment) SetExcludedContexts(v []SegmentTarget) {
	o.ExcludedContexts = v
}

// GetLinks returns the Links field value
func (o *UserSegment) GetLinks() map[string]Link {
	if o == nil {
		var ret map[string]Link
		return ret
	}

	return o.Links
}

// GetLinksOk returns a tuple with the Links field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetLinksOk() (*map[string]Link, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Links, true
}

// SetLinks sets field value
func (o *UserSegment) SetLinks(v map[string]Link) {
	o.Links = v
}

// GetRules returns the Rules field value
func (o *UserSegment) GetRules() []UserSegmentRule {
	if o == nil {
		var ret []UserSegmentRule
		return ret
	}

	return o.Rules
}

// GetRulesOk returns a tuple with the Rules field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetRulesOk() ([]UserSegmentRule, bool) {
	if o == nil {
		return nil, false
	}
	return o.Rules, true
}

// SetRules sets field value
func (o *UserSegment) SetRules(v []UserSegmentRule) {
	o.Rules = v
}

// GetVersion returns the Version field value
func (o *UserSegment) GetVersion() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Version
}

// GetVersionOk returns a tuple with the Version field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetVersionOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Version, true
}

// SetVersion sets field value
func (o *UserSegment) SetVersion(v int32) {
	o.Version = v
}

// GetDeleted returns the Deleted field value
func (o *UserSegment) GetDeleted() bool {
	if o == nil {
		var ret bool
		return ret
	}

	return o.Deleted
}

// GetDeletedOk returns a tuple with the Deleted field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetDeletedOk() (*bool, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Deleted, true
}

// SetDeleted sets field value
func (o *UserSegment) SetDeleted(v bool) {
	o.Deleted = v
}

// GetAccess returns the Access field value if set, zero value otherwise.
func (o *UserSegment) GetAccess() Access {
	if o == nil || o.Access == nil {
		var ret Access
		return ret
	}
	return *o.Access
}

// GetAccessOk returns a tuple with the Access field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetAccessOk() (*Access, bool) {
	if o == nil || o.Access == nil {
		return nil, false
	}
	return o.Access, true
}

// HasAccess returns a boolean if a field has been set.
func (o *UserSegment) HasAccess() bool {
	if o != nil && o.Access != nil {
		return true
	}

	return false
}

// SetAccess gets a reference to the given Access and assigns it to the Access field.
func (o *UserSegment) SetAccess(v Access) {
	o.Access = &v
}

// GetFlags returns the Flags field value if set, zero value otherwise.
func (o *UserSegment) GetFlags() []FlagListingRep {
	if o == nil || o.Flags == nil {
		var ret []FlagListingRep
		return ret
	}
	return o.Flags
}

// GetFlagsOk returns a tuple with the Flags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetFlagsOk() ([]FlagListingRep, bool) {
	if o == nil || o.Flags == nil {
		return nil, false
	}
	return o.Flags, true
}

// HasFlags returns a boolean if a field has been set.
func (o *UserSegment) HasFlags() bool {
	if o != nil && o.Flags != nil {
		return true
	}

	return false
}

// SetFlags gets a reference to the given []FlagListingRep and assigns it to the Flags field.
func (o *UserSegment) SetFlags(v []FlagListingRep) {
	o.Flags = v
}

// GetUnbounded returns the Unbounded field value if set, zero value otherwise.
func (o *UserSegment) GetUnbounded() bool {
	if o == nil || o.Unbounded == nil {
		var ret bool
		return ret
	}
	return *o.Unbounded
}

// GetUnboundedOk returns a tuple with the Unbounded field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetUnboundedOk() (*bool, bool) {
	if o == nil || o.Unbounded == nil {
		return nil, false
	}
	return o.Unbounded, true
}

// HasUnbounded returns a boolean if a field has been set.
func (o *UserSegment) HasUnbounded() bool {
	if o != nil && o.Unbounded != nil {
		return true
	}

	return false
}

// SetUnbounded gets a reference to the given bool and assigns it to the Unbounded field.
func (o *UserSegment) SetUnbounded(v bool) {
	o.Unbounded = &v
}

// GetUnboundedContextKind returns the UnboundedContextKind field value if set, zero value otherwise.
func (o *UserSegment) GetUnboundedContextKind() string {
	if o == nil || o.UnboundedContextKind == nil {
		var ret string
		return ret
	}
	return *o.UnboundedContextKind
}

// GetUnboundedContextKindOk returns a tuple with the UnboundedContextKind field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetUnboundedContextKindOk() (*string, bool) {
	if o == nil || o.UnboundedContextKind == nil {
		return nil, false
	}
	return o.UnboundedContextKind, true
}

// HasUnboundedContextKind returns a boolean if a field has been set.
func (o *UserSegment) HasUnboundedContextKind() bool {
	if o != nil && o.UnboundedContextKind != nil {
		return true
	}

	return false
}

// SetUnboundedContextKind gets a reference to the given string and assigns it to the UnboundedContextKind field.
func (o *UserSegment) SetUnboundedContextKind(v string) {
	o.UnboundedContextKind = &v
}

// GetGeneration returns the Generation field value
func (o *UserSegment) GetGeneration() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Generation
}

// GetGenerationOk returns a tuple with the Generation field value
// and a boolean to check if the value has been set.
func (o *UserSegment) GetGenerationOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Generation, true
}

// SetGeneration sets field value
func (o *UserSegment) SetGeneration(v int32) {
	o.Generation = v
}

// GetUnboundedMetadata returns the UnboundedMetadata field value if set, zero value otherwise.
func (o *UserSegment) GetUnboundedMetadata() SegmentMetadata {
	if o == nil || o.UnboundedMetadata == nil {
		var ret SegmentMetadata
		return ret
	}
	return *o.UnboundedMetadata
}

// GetUnboundedMetadataOk returns a tuple with the UnboundedMetadata field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetUnboundedMetadataOk() (*SegmentMetadata, bool) {
	if o == nil || o.UnboundedMetadata == nil {
		return nil, false
	}
	return o.UnboundedMetadata, true
}

// HasUnboundedMetadata returns a boolean if a field has been set.
func (o *UserSegment) HasUnboundedMetadata() bool {
	if o != nil && o.UnboundedMetadata != nil {
		return true
	}

	return false
}

// SetUnboundedMetadata gets a reference to the given SegmentMetadata and assigns it to the UnboundedMetadata field.
func (o *UserSegment) SetUnboundedMetadata(v SegmentMetadata) {
	o.UnboundedMetadata = &v
}

// GetExternal returns the External field value if set, zero value otherwise.
func (o *UserSegment) GetExternal() string {
	if o == nil || o.External == nil {
		var ret string
		return ret
	}
	return *o.External
}

// GetExternalOk returns a tuple with the External field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetExternalOk() (*string, bool) {
	if o == nil || o.External == nil {
		return nil, false
	}
	return o.External, true
}

// HasExternal returns a boolean if a field has been set.
func (o *UserSegment) HasExternal() bool {
	if o != nil && o.External != nil {
		return true
	}

	return false
}

// SetExternal gets a reference to the given string and assigns it to the External field.
func (o *UserSegment) SetExternal(v string) {
	o.External = &v
}

// GetExternalLink returns the ExternalLink field value if set, zero value otherwise.
func (o *UserSegment) GetExternalLink() string {
	if o == nil || o.ExternalLink == nil {
		var ret string
		return ret
	}
	return *o.ExternalLink
}

// GetExternalLinkOk returns a tuple with the ExternalLink field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetExternalLinkOk() (*string, bool) {
	if o == nil || o.ExternalLink == nil {
		return nil, false
	}
	return o.ExternalLink, true
}

// HasExternalLink returns a boolean if a field has been set.
func (o *UserSegment) HasExternalLink() bool {
	if o != nil && o.ExternalLink != nil {
		return true
	}

	return false
}

// SetExternalLink gets a reference to the given string and assigns it to the ExternalLink field.
func (o *UserSegment) SetExternalLink(v string) {
	o.ExternalLink = &v
}

// GetImportInProgress returns the ImportInProgress field value if set, zero value otherwise.
func (o *UserSegment) GetImportInProgress() bool {
	if o == nil || o.ImportInProgress == nil {
		var ret bool
		return ret
	}
	return *o.ImportInProgress
}

// GetImportInProgressOk returns a tuple with the ImportInProgress field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UserSegment) GetImportInProgressOk() (*bool, bool) {
	if o == nil || o.ImportInProgress == nil {
		return nil, false
	}
	return o.ImportInProgress, true
}

// HasImportInProgress returns a boolean if a field has been set.
func (o *UserSegment) HasImportInProgress() bool {
	if o != nil && o.ImportInProgress != nil {
		return true
	}

	return false
}

// SetImportInProgress gets a reference to the given bool and assigns it to the ImportInProgress field.
func (o *UserSegment) SetImportInProgress(v bool) {
	o.ImportInProgress = &v
}

func (o UserSegment) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["name"] = o.Name
	}
	if o.Description != nil {
		toSerialize["description"] = o.Description
	}
	if true {
		toSerialize["tags"] = o.Tags
	}
	if true {
		toSerialize["creationDate"] = o.CreationDate
	}
	if true {
		toSerialize["lastModifiedDate"] = o.LastModifiedDate
	}
	if true {
		toSerialize["key"] = o.Key
	}
	if o.Included != nil {
		toSerialize["included"] = o.Included
	}
	if o.Excluded != nil {
		toSerialize["excluded"] = o.Excluded
	}
	if o.IncludedContexts != nil {
		toSerialize["includedContexts"] = o.IncludedContexts
	}
	if o.ExcludedContexts != nil {
		toSerialize["excludedContexts"] = o.ExcludedContexts
	}
	if true {
		toSerialize["_links"] = o.Links
	}
	if true {
		toSerialize["rules"] = o.Rules
	}
	if true {
		toSerialize["version"] = o.Version
	}
	if true {
		toSerialize["deleted"] = o.Deleted
	}
	if o.Access != nil {
		toSerialize["_access"] = o.Access
	}
	if o.Flags != nil {
		toSerialize["_flags"] = o.Flags
	}
	if o.Unbounded != nil {
		toSerialize["unbounded"] = o.Unbounded
	}
	if o.UnboundedContextKind != nil {
		toSerialize["unboundedContextKind"] = o.UnboundedContextKind
	}
	if true {
		toSerialize["generation"] = o.Generation
	}
	if o.UnboundedMetadata != nil {
		toSerialize["_unboundedMetadata"] = o.UnboundedMetadata
	}
	if o.External != nil {
		toSerialize["_external"] = o.External
	}
	if o.ExternalLink != nil {
		toSerialize["_externalLink"] = o.ExternalLink
	}
	if o.ImportInProgress != nil {
		toSerialize["_importInProgress"] = o.ImportInProgress
	}
	return json.Marshal(toSerialize)
}

type NullableUserSegment struct {
	value *UserSegment
	isSet bool
}

func (v NullableUserSegment) Get() *UserSegment {
	return v.value
}

func (v *NullableUserSegment) Set(val *UserSegment) {
	v.value = val
	v.isSet = true
}

func (v NullableUserSegment) IsSet() bool {
	return v.isSet
}

func (v *NullableUserSegment) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUserSegment(val *UserSegment) *NullableUserSegment {
	return &NullableUserSegment{value: val, isSet: true}
}

func (v NullableUserSegment) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUserSegment) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


