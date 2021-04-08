package v2

import "net/http"

// Controller is an interface to be able to mock the controllers
type Controller interface {
	ReadController
	CreateController
	DeleteController
	UpdateController
}

// ReadController is an interface for reading
type ReadController interface {
	Read(w http.ResponseWriter, r *http.Request)
}

// CreateController is an interface for creating
type CreateController interface {
	Create(w http.ResponseWriter, r *http.Request)
}

// DeleteController is an interface for deleting
type DeleteController interface {
	Delete(w http.ResponseWriter, r *http.Request)
}

// UpdateController is an interface for updating
type UpdateController interface {
	Update(w http.ResponseWriter, r *http.Request)
}