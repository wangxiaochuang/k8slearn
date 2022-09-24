package common

type RouteContainer interface {
	RootPath() string
	PathParameters() []Parameter
	Routes() []Route
}

// Route is a logical endpoint of a service.
type Route interface {
	// Method defines the HTTP Method.
	Method() string
	// Path defines the route's endpoint.
	Path() string
	// OperationName defines a machine-readable ID for the route.
	OperationName() string
	// Parameters defines the list of accepted parameters.
	Parameters() []Parameter
	// Description is a human-readable route description.
	Description() string
	// Consumes defines the consumed content-types.
	Consumes() []string
	// Produces defines the produced content-types.
	Produces() []string
	// Metadata allows adding extensions to the generated spec.
	Metadata() map[string]interface{}
	// RequestPayloadSample defines an example request payload. Can return nil.
	RequestPayloadSample() interface{}
	// ResponsePayloadSample defines an example response payload. Can return nil.
	ResponsePayloadSample() interface{}
	// StatusCodeResponses defines a mapping of HTTP Status Codes to the specific response(s).
	// Multiple responses with the same HTTP Status Code are acceptable.
	StatusCodeResponses() []StatusCodeResponse
}

type StatusCodeResponse interface {
	Code() int
	Message() string
	Model() interface{}
}

type Parameter interface {
	Name() string
	Description() string
	Required() bool
	Kind() ParameterKind
	DataType() string
	AllowMultiple() bool
}

type ParameterKind int

const (
	PathParameterKind = ParameterKind(iota)
	QueryParameterKind
	BodyParameterKind
	HeaderParameterKind
	FormParameterKind
	UnknownParameterKind
)
