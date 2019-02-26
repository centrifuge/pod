/*
Package documents implements Centrifuge document models.

Models

The idea of having a model is to make the business functions of the document clearer and more readable. This also enables proper types and validations on the fields that require them. When an API call is received, the following list of transformations/steps needs to be executed on the request object.

1. Model conversion: this would ensure that proper types are created for each of the fields of the input document plus handling basic level validations that does not require business level understanding of the document. eg: telephone numbers, IDs

2. The converted model is updated using the existing document. After this there would be two versions of the document in the system old and the new

3. The two versions of the document are passed through a purpose specific validator chain that implements the following interface. (see chapter on validation)

Model Storage

A model objects must support storage in DB as a JSON serialized object. The rationale behind this is that JSON format enables easier inter-operability between systems that
would depend on database access such as BI (Although we do NOT recommend directly accessing the db eventually when we have proper APIs for accessing all data).

	// example of an implementation
	type InvoiceModel struct {
	// all invoice fields with proper types here
	}

	func (i *InvoiceModel) PackCoreDocument() *coredocumentpb.CoreDocument {
	panic("implement me")
	}

	func (i *InvoiceModel) UnpackCoreDocument(cd *coredocumentpb.CoreDocument) error {
	panic("implement me")
	}

	func (i *InvoiceModel) JSON() ([]byte, error) {
	panic("implement me")
	}

	func (i *InvoiceModel) Type() reflect.Type {
	return reflect.TypeOf(i)
	}


Model Package Hierarchy Specification

In the new package structure the package `documents` includes all model relevant implementations and interfaces.
The actual implementation can be found at the package `invoice` or `purchaseorder`.
The package `documents` should not include any of the actual implementations to avoid cycle dependencies.

Validation

Validations should be done depending on the situation. The below table is an example that shows a few examples of validations that should only depending on context of the validation.

	|                           | Verify  SigningRoot | Verify DocumentRoot | Check Invoice State Transition | Node must be in collaborators |
	|---------------------------|---------------------|---------------------|--------------------------------|-------------------------------|
	| Client Submission         |                     |                     | yes                            | yes                           |
	| Signature Request         | yes                 |                     |                                |                               |
	| Receive Anchored Document | yes                 | yes                 |                                | yes                           |
	| Store in DB               | yes                 | only if set         |                                | yes                           |


Validations are implemented in individual checks. These individual checks can be nested in different grouped set of validations. Validators all implement the Validator interface to allow simple nesting of checks.

There are three types of validators:

1. **Atomic Validators:** they validate an individual field or several fields and contain actual checking of certain values, checking of an anchor etc. Atomic validators are never split up. Anything an atomic validator validates would always be validated as one block.

2. **Internal Group Validations:** contain a collection of atomic validators but these are not grouped by purpose but elements that are reused in other validators (group or purpose validators)

3. **Purpose Validators:** these validators are public and used by the application logic to validate a document. They should never be used inside of other validators. Code performing validation should always call a purpose validator but never an internal group directly.


	type interface Validator {
	// Validate validates the updates to the model in newState. errors returned must always be centerrors
	Validate(oldState Model, newState Model) []error
	}


	// ValidatorGroup implements Validator for executing a set of validators.
	type ValidatorGroup struct {
	[]Validator validators
	}
	func (group *ValidatorGroup) Validate(oldState Model, newState Model) (validationErrors []error) {

	for v, i := range group.validators {
	validationErrors = append(validationErrors, v.Validate(oldState, newState))
	}
	}

	// An implementation of ValidatorGroup as an example.
	// Note that this is not a public validator but merely a private variable that can be reused in validators.
	var invoiceHeaderValidators = ValidatorGroup{
	validators: []Validator{
	Field1Validator,
	Field2Validator,
	Field3Validator
	}
	}

	// An implementation of a Validator that is used by other services
	// Note that this is public
	var ValidateOnSignatureRequest = ValidatorGroup{
	validators: []Validator{
	invoiceHeaderValidator, ...
	}
	}


Controllers, Services and Their Relationship with Models

1. Controllers

Controllers are generally the first contact point between an external request and the application logic. The implementations would just pass control over to a request specific `Service` layer call.

2. Services

Services in the CentNode must deal with only specific Model object plus its related objects. Eg: InvoiceService would only deal with InvoiceModel. Since in many cases a model object may need to be created based on some external input such as a coredocument, the following are some good base interfaces for a service to implement,

	// Implementation of deriving model objects
	type InvoiceService struct { }

	func (srv *InvoiceService) DeriveFromCoreDocument(cd *coredocument.CoreDocument) (*InvoiceModel, error) {
	panic("Implement me");
	}

	func (srv *InvoiceService)  DeriveWithCreateInvoiceInput(*CreateInvoiceInput) (*InvoiceModel, error) {
	panic("Implement me");
	}



Service Registry

To locate a service that could handle operations for a given CoreDocument object a service registry needs to be developed. This would use the `coreDocument.EmbeddedData.TypeUrl` to map the document to the relevant service.


	// in documents/registry.go

	func LocateService(cd *coredocument.CoreDocument) (ModelDeriver, error) {
	....
	}

Every service should be able to `register` itself at the `ServiceRegistry` if it implements the `ModelDeriver` interface.

	func (s *ServiceRegistry) Register(serviceID string, service ModelDeriver) error


The registry should be thread safe.


A Sample Flow for Handling Document Signature Requests

The following is an example modification of `Handler.RequestDocumentSignature` to show the usage of `Registry`, `Service` and `Model` interactions.

	func (srv *Handler) RequestDocumentSignature(ctx context.Context, sigReq *p2ppb.SignatureRequest) (*p2ppb.SignatureResponse, error) {
	service, err := registry.LocateService(sigReq.Document)
	if err != nil {
	return nil, err
	}

	model, err := service.DeriveWithCD(sigReq.Document)
	if err != nil {
	return nil, err
	}

	if p2pService, ok := service.(P2PSignatureRequestHandler); ok {
	sig, errors := p2pService.Sign(model)
	if len(errs) != 0 {
	return nil, centerrors.NewWithErrors(errs)
	}
	return &p2ppb.SignatureResponse{
	CentNodeVersion: version.GetVersion().String(),
	Signature:       sig,
	}, nil
	}

	return nil, someerrorThatcausedServiceMismatch

	}

*/
package documents
