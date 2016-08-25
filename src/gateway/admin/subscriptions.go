package admin

import (
	"errors"
	"gateway/config"
	aphttp "gateway/http"
	"gateway/mail"
	"gateway/model"
	apsql "gateway/sql"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/event"
)

type SubscriptionsController struct {
	BaseController
}

func RouteSubscriptions(controller *SubscriptionsController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	routes := map[string]http.Handler{
		"POST": write(db, controller.Subscription),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"POST", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(routes))
}

func (c *SubscriptionsController) Subscription(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error {
	request := stripe.Event{}
	if err := deserialize(&request, r.Body); err != nil {
		return err
	}

	// Go to Stripe to get the Event by ID to verify authenticity.
	event, err := event.Get(request.ID, nil)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}
	if account, err := model.FindAccountByStripeCustomer(tx.DB, event.Data.Obj["customer"].(string)); err == nil {
		user, err := model.FindAdminUserForAccountID(tx.DB, account.ID)
		if err != nil {
			return aphttp.NewError(err, http.StatusBadRequest)
		}
		if event.Type == "invoice.payment_succeeded" {
			err = mail.SendInvoicePaymentSucceededEmail(c.SMTP, c.ProxyServer, c.conf, user, true)
			if err != nil {
				return aphttp.NewError(err, http.StatusBadRequest)
			}
		} else if event.Type == "invoice.payment_failed" {
			account.SetStripePaymentRetryAttempt(tx, account.StripePaymentRetryAttempt+1)
			if account.StripePaymentRetryAttempt < c.conf.StripePaymentRetryAttempts {
				err = mail.SendInvoicePaymentFailedEmail(c.SMTP, c.ProxyServer, c.conf, user, true)
			} else {
				fallbackPlan, err := model.FindPlanByName(tx.DB, c.conf.StripeFallbackPlan)
				if err != nil {
					return aphttp.NewError(err, http.StatusBadRequest)
				}
				err = mail.SendInvoicePaymentFailedAndPlanDowngradedEmail(c.SMTP, c.ProxyServer, c.conf, user, true)
				account.PlanID.Int64 = fallbackPlan.ID
				err = account.Update(tx)
				if err != nil {
					return aphttp.NewError(err, http.StatusBadRequest)
				}
			}
			if err != nil {
				return aphttp.NewError(err, http.StatusBadRequest)
			}
		} else {
			return aphttp.NewError(errors.New("Unhandled event type."), http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
		return nil
	} else {
		return aphttp.NewError(err, http.StatusBadRequest)
	}
}
