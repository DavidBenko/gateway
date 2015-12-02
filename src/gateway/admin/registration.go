package admin

import (
	"net/http"

	"gateway/config"
	aperrors "gateway/errors"
	aphttp "gateway/http"
	"gateway/model"
	apsql "gateway/sql"

	"github.com/gorilla/handlers"
)

type RegistrationController struct {
	BaseController
}

type Registration struct {
	Name                 string `json:"name"`
	Email                string `json:"email"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
	Organization         string `json:"organization"`
}

func RouteRegistration(controller *RegistrationController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	routes := map[string]http.Handler{
		"POST": write(db, controller.Registration),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"POST", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(routes))
}

func (c *RegistrationController) Registration(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error {
	request := struct {
		Registration Registration `json:"registration"`
	}{}
	if err := deserialize(&request, r.Body); err != nil {
		return err
	}

	registration, verrors := request.Registration, make(aperrors.Errors)
	if registration.Email == "" {
		verrors.Add("email", "must not be blank")
	}
	if registration.Password == "" {
		verrors.Add("password", "must not be blank")
	}
	if registration.Password != registration.PasswordConfirmation {
		verrors.Add("password_confirmation", "must match password")
	}
	if !verrors.Empty() {
		return SerializableValidationErrors{verrors}
	}

	name := registration.Email
	if registration.Organization != "" {
		name = registration.Organization
	}
	account := &model.Account{Name: name}
	err := account.Insert(tx)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}
	return nil
}
