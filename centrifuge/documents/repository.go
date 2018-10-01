package documents

// Loader interface can be implemented by any type that handles document retrieval
type Loader interface {
	// GetKey will prepare the the identifier key from ID
	GetKey(id []byte) (key []byte)

	// GetByID finds the doc with identifier and marshals it into message
	LoadByID(id []byte, model Model) error
}

// Checker interface can be implemented by any type that handles if document exists
type Checker interface {
	// Exists checks for document existence
	// True if exists else false
	Exists(id []byte) bool
}

// Creator interface can be implemented by any type that handles document creation
type Creator interface {
	// Create stores the initial document
	// If document exist, it errors out
	Create(id []byte, model Model) error
}

// Updater interface can be implemented by any type that handles document update
type Updater interface {
	// Update updates the already stored document
	// errors out when document is missing
	Update(id []byte, model Model) error
}

// Repository should be implemented by any type that wants to store a document in key-value storage
type Repository interface {
	Checker
	Loader
	Creator
	Updater
}
