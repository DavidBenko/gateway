package admin

import (
	"errors"
	"net/http"

	"gateway/config"
	aperrors "gateway/errors"
	aphttp "gateway/http"
	"gateway/mail"
	"gateway/model"
	apsql "gateway/sql"

	"github.com/gorilla/handlers"
	"github.com/stripe/stripe-go"
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
	PlanID               int64  `json:"plan_id"`
	StripeToken          string `json:"stripe_token"`
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
	if stripe.Key != "" {
		if registration.PlanID < 1 {
			verrors.Add("plan_id", "must not be blank")
			return SerializableValidationErrors{verrors}
		}
		plan, err := model.FindPlan(tx.DB, registration.PlanID)
		if err != nil {
			verrors.Add("plan_id", "not found")
		}
		if plan.Price > 0 && registration.StripeToken == "" {
			verrors.Add("stripe_token", "must not be blank")
		}
	}
	if !verrors.Empty() {
		return SerializableValidationErrors{verrors}
	}

	name := registration.Email
	if registration.Organization != "" {
		name = registration.Organization
	}

	if user, err := model.FindUserByEmail(tx.DB, registration.Email); err == nil {
		if user.Confirmed {
			err = mail.SendWelcomeEmail(c.SMTP, c.ProxyServer, c.conf, user, true, true)
			if err != nil {
				return aphttp.NewError(err, http.StatusBadRequest)
			}
		} else if user.HasConfirmToken() {
			err = mail.SendConfirmEmail(c.SMTP, c.ProxyServer, c.conf, user, tx, true)
			if err != nil {
				return aphttp.NewError(err, http.StatusBadRequest)
			}
		}

		w.WriteHeader(http.StatusOK)
		return nil
	}

	account := &model.Account{Name: name, PlanID: apsql.MakeNullInt64(registration.PlanID), StripeToken: registration.StripeToken}
	err := account.Insert(tx)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	user := &model.User{
		AccountID:   account.ID,
		Name:        registration.Email,
		Email:       registration.Email,
		NewPassword: registration.Password,
	}
	err = user.Insert(tx)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	err = mail.SendConfirmEmail(c.SMTP, c.ProxyServer, c.conf, user, tx, true)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

type ConfirmationController struct {
	BaseController
}

func RouteConfirmation(controller *ConfirmationController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	routes := map[string]http.Handler{
		"GET": write(db, controller.Confirmation),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(routes))
}

func (c *ConfirmationController) Confirmation(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error {
	err := r.ParseForm()
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	if len(r.Form["token"]) != 1 {
		return aphttp.NewError(errors.New("token is required"), http.StatusBadRequest)
	}

	token := r.Form["token"][0]
	user, err := model.ValidateUserToken(tx, token, true)
	if err != nil {
		http.Redirect(w, r, c.conf.PathPrefix, http.StatusFound)
		return nil
	}

	user.Confirmed = true
	err = user.Update(tx)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	err = mail.SendWelcomeEmail(c.SMTP, c.ProxyServer, c.conf, user, false, true)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	http.Redirect(w, r, c.conf.PathPrefix, http.StatusFound)
	return nil
}
