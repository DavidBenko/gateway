package admin

/*******************************************************
*******************************************************
***                                                 ***
*** This is generated code. Do not edit directly.   ***
***                                                 ***
*******************************************************
*******************************************************/

import (
	aphttp "gateway/http"
	"gateway/model"
	"net/http"
)

func (c *AccountsController) deserializeInstance(r *http.Request) (*model.Account,
	error) {

	var wrapped struct {
		Account *model.Account `json:"account"`
	}
	if err := deserialize(&wrapped, r); err != nil {
		return nil, err
	}
	return wrapped.Account, nil
}

func (c *AccountsController) serializeInstance(instance *model.Account,
	w http.ResponseWriter) aphttp.Error {

	wrapped := struct {
		Account *model.Account `json:"account"`
	}{instance}
	return serialize(wrapped, w)
}

func (c *AccountsController) serializeCollection(collection []*model.Account,
	w http.ResponseWriter) aphttp.Error {

	wrapped := struct {
		Accounts []*model.Account `json:"accounts"`
	}{collection}
	return serialize(wrapped, w)
}
