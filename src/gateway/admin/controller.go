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
//go:generate ./controller.rb --model ProxyEndpoint --reflect
//go:generate ./controller.rb --model Job --reflect
//go:generate ./controller.rb --model RemoteEndpoint --account --api --check-delete
//go:generate ./controller.rb --model User --account --after-insert-hook --check-delete --transform-method c.sanitize --transform-type sanitizedUser
//go:generate ./controller.rb --model RemoteEndpointType
//go:generate ./controller.rb --model User --account --transform-method c.sanitize --transform-type sanitizedUser
//go:generate ./controller.rb --model ProxyEndpointSchema --reflect
//go:generate ./controller.rb --model ScratchPad --reflect
//go:generate ./controller.rb --model PushChannel --reflect
//go:generate ./controller.rb --model PushDevice --reflect
//go:generate ./controller.rb --model PushMessage --reflect
//go:generate ./controller.rb --model PushChannelMessage --reflect
//go:generate ./controller.rb --model Plan --allow-create=false --allow-update=false --allow-delete=false
//go:generate ./controller.rb --model Sample --account
//go:generate ./controller.rb --model Timer --reflect
//go:generate ./controller.rb --model JobTest --reflect
//go:generate ./controller.rb --model ProxyEndpointChannel --json Channel --reflect
//go:generate ./controller.rb --model CustomFunction --reflect --after-insert-hook
//go:generate ./controller.rb --model CustomFunctionFile --json File --reflect
//go:generate ./controller.rb --model CustomFunctionTest --json Test --reflect

// ResourceController defines what we expect a controller to do to route
// a RESTful resource
type ResourceController interface {
	List(w http.ResponseWriter, r *http.Request, db *apsql.DB) aphttp.Error
	Create(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error
	Show(w http.ResponseWriter, r *http.Request, db *apsql.DB) aphttp.Error
	Update(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error
	Delete(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error
}

type SingularResourceController interface {
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

func (c *BaseController) environmentID(r *http.Request) int64 {
	return environmentIDFromPath(r)
}
