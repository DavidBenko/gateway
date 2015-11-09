package admin

import (
	"net/http"

	"gateway/config"
	aphttp "gateway/http"
	apsql "gateway/sql"
)

//go:generate ./controller.rb --model Account
//go:generate ./controller.rb --model API --account --before-validate-hook --after-insert-hook --after-find-hook --after-update-hook --transform-method c.addBaseURL --transform-type enhancedAPI
//go:generate ./controller.rb --model EndpointGroup --account --api
//go:generate ./controller.rb --model Environment --account --api --check-delete
//go:generate ./controller.rb --model Host --account --api
//go:generate ./controller.rb --model Library --account --api
//go:generate ./controller.rb --model SharedComponent --account --api
//go:generate ./controller.rb --model ProxyEndpoint --account --api --before-validate-hook
//go:generate ./controller.rb --model RemoteEndpoint --account --api --check-delete
//go:generate ./controller.rb --model User --account --after-insert-hook --check-delete --transform-method c.sanitize --transform-type sanitizedUser
//go:generate ./controller.rb --model RemoteEndpointType
//go:generate ./controller.rb --model User --account --transform-method c.sanitize --transform-type sanitizedUser
//go:generate ./controller.rb --model ProxyEndpointSchema --reflect
//go:generate ./controller.rb --model ScratchPad --reflect

// ResourceController defines what we expect a controller to do to route
// a RESTful resource
type ResourceController interface {
	List(w http.ResponseWriter, r *http.Request, db *apsql.DB) aphttp.Error
	Create(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error
	Show(w http.ResponseWriter, r *http.Request, db *apsql.DB) aphttp.Error
	Update(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error
	Delete(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error
}

type BaseController struct {
	conf      config.ProxyAdmin
	accountID func(r *http.Request) int64
	userID    func(r *http.Request) int64
	auth      aphttp.AuthType
	config.SMTP
	config.ProxyServer
}

func (c *BaseController) apiID(r *http.Request) int64 {
	return apiIDFromPath(r)
}

func (c *BaseController) proxyEndpointID(r *http.Request) int64 {
	return endpointIDFromPath(r)
}

func (c *BaseController) collectionID(r *http.Request) int64 {
	return collectionIDFromPath(r)
}
