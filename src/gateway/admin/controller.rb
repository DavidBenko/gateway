#!/usr/bin/env ruby

require 'rubygems'
require 'active_support/inflector'
require 'optparse'
require 'erb'

singular = nil
transform_method = nil
transform_type = nil
account = false
api = false
custom_struct = false
check_delete = false
after_insert = false
after_update = false
after_delete = false
after_validate = false
before_insert = false
before_update = false
before_delete = false
before_validate = false
OptionParser.new do |opts|
  opts.banner = "Usage: example.rb [options]"

  opts.on("--model Model", "Name of model") do |value|
    singular = value
  end
  opts.on("--account", "Is model linked to Account?") do |value|
    account = value
  end
  opts.on("--api", "Is model linked to API?") do |value|
    api = value
  end
  opts.on("--check-delete", "Check if delete is possible first?") do |value|
    check_delete = value
  end
  opts.on("--transform-method Method", "Optional custom transform method") do |value|
    transform_method = value
  end
  opts.on("--transform-type Type", "Optional custom transform type") do |value|
    transform_type = value
  end
  opts.on("--after-insert-hook", "Does controller have an after insert hook?") do |value|
    after_insert = value
  end
  opts.on("--after-update-hook", "Does controller have an after update hook?") do |value|
    after_update = value
  end
  opts.on("--after-delete-hook", "Does controller have an after delete hook?") do |value|
    after_delete = value
  end
  opts.on("--before-insert-hook", "Does controller have a before insert hook?") do |value|
    before_insert = value
  end
  opts.on("--before-update-hook", "Does controller have a before update hook?") do |value|
    before_update = value
  end
  opts.on("--before-delete-hook", "Does controller have a before delete hook?") do |value|
    before_delete = value
  end
  opts.on("--before-validate-hook", "Does controller have a before validate hook?") do |value|
    before_validate = value
  end
  opts.on("--after-validate-hook", "Does controller have a after validate hook?") do |value|
    after_validate = value
  end

end.parse!

plural = singular.pluralize
controller = "#{plural}Controller"

local = singular.camelize(:lower)
local_plural = plural.camelize(:lower)

json_singular = singular.underscore
json_plural = json_singular.pluralize

pretty = singular.titleize.downcase

transform = !!transform_method

filename = "./#{json_plural}_gen.go"
output = File.open(filename, "w")

template = <<-ERB

package admin

/*****************************************************
 *****************************************************
 ***                                               ***
 *** This is generated code. Do not edit directly. ***
 ***                                               ***
 *****************************************************
 *****************************************************/

import (
  "errors"
  "gateway/config"
  aphttp "gateway/http"
  "gateway/model"
  apsql "gateway/sql"
  "log"
  "net/http"
)

// <%= controller %> manages <%= plural %>.
type <%= controller %> struct {
  BaseController
}

// List lists the <%= plural %>.
func (c *<%= controller %>) List(w http.ResponseWriter, r *http.Request,
  db *apsql.DB) aphttp.Error {

  <% if account && api %>
    <%= local_plural %>, err := model.All<%= plural %>ForAPIIDAndAccountID(db,
      c.apiID(r), c.accountID(r))
  <% elsif account %>
    <%= local_plural %>, err := model.All<%= plural %>ForAccountID(db, c.accountID(r))
  <% else %>
    <%= local_plural %>, err := model.All<%= plural %>(db)
  <% end %>

  if err != nil {
    log.Printf("%s Error listing <%= pretty %>: %v", config.System, err)
    return aphttp.DefaultServerError()
  }

  return c.serializeCollection(<%= local_plural %>, w)
}

// Create creates the <%= singular %>.
func (c *<%= controller %>) Create(w http.ResponseWriter, r *http.Request,
  tx *apsql.Tx) aphttp.Error {
  return c.insertOrUpdate(w, r, tx, true)
}

// Show shows the <%= singular %>.
func (c *<%= controller %>) Show(w http.ResponseWriter, r *http.Request,
  db *apsql.DB) aphttp.Error {

  id := instanceID(r)
  <% if account && api %>
    <%= local %>, err := model.Find<%= singular %>ForAPIIDAndAccountID(db,
      id, c.apiID(r), c.accountID(r))
  <% elsif account %>
    <%= local %>, err := model.Find<%= singular %>ForAccountID(db, id, c.accountID(r))
  <% else %>
    <%= local %>, err := model.Find<%= singular %>(db, id)
  <% end %>
  if err != nil {
    return c.notFound()
  }

  return c.serializeInstance(<%= local %>, w)
}

// Update updates the <%= singular %>.
func (c *<%= controller %>) Update(w http.ResponseWriter, r *http.Request,
  tx *apsql.Tx) aphttp.Error {

  return c.insertOrUpdate(w, r, tx, false)
}

// Delete deletes the <%= singular %>.
func (c *<%= controller %>) Delete(w http.ResponseWriter, r *http.Request,
  tx *apsql.Tx) aphttp.Error {

  id := instanceID(r)

  var err error
  <% if after_delete || before_delete %>
    db := tx.DB

    <% if account && api %>
      <%= local %>, err := model.Find<%= singular %>ForAPIIDAndAccountID(db,
        id, c.apiID(r), c.accountID(r))
    <% elsif account %>
      <%= local %>, err := model.Find<%= singular %>ForAccountID(db, id, c.accountID(r))
    <% else %>
      <%= local %>, err := model.Find<%= singular %>(db, id)
    <% end %>
    if err != nil {
      return c.notFound()
    }
  <% end %>

  <% if check_delete %>
    if err = model.CanDelete<%= singular %>(tx, id); err != nil {
      return aphttp.NewError(err, 400)
    }
  <% end %>

  <% if before_delete %>
    if err = c.BeforeDelete(<%= local %>, tx); err != nil {
      log.Printf("%s Error before delete: %v", config.System, err)
      return aphttp.DefaultServerError()
    }
  <% end %>

  <% if account && api %>
    err = model.Delete<%= singular %>ForAPIIDAndAccountID(tx,
      id, c.apiID(r), c.accountID(r), c.userID(r))
  <% elsif account %>
    err = model.Delete<%= singular %>ForAccountID(tx, id, c.accountID(r), c.userID(r))
  <% else %>
    err = model.Delete<%= singular %>(tx, id)
  <% end %>

  if err != nil {
    if err == apsql.ErrZeroRowsAffected {
      return c.notFound()
    }
    log.Printf("%s Error deleting <%= pretty %>: %v", config.System, err)
    return aphttp.NewServerError(err)
  }

  <% if after_delete %>
    if err := c.AfterDelete(<%= local %>, tx); err != nil {
      log.Printf("%s Error after delete: %v", config.System, err)
      return aphttp.DefaultServerError()
    }
  <% end %>

  w.WriteHeader(http.StatusOK)
  return nil
}

func (c *<%= controller %>) insertOrUpdate(w http.ResponseWriter, r *http.Request,
  tx *apsql.Tx, isInsert bool) aphttp.Error {

  <%= local %>, httpErr := c.deserializeInstance(r.Body)
  if httpErr != nil {
    return httpErr
  }
  <% if api %>
    <%= local %>.APIID = c.apiID(r)
  <% end %>
  <% if account %>
    <%= local %>.AccountID = c.accountID(r)
    <%= local %>.UserID = c.userID(r)
  <% end %>

  var method func(*apsql.Tx) error
  var desc string
  if isInsert {
    method = <%= local %>.Insert
    desc = "inserting"
  } else {
    <%= local %>.ID = instanceID(r)
    method = <%= local %>.Update
    desc = "updating"
  }

  <% if before_validate %>
  if err := c.BeforeValidate(<%= local %>, tx); err != nil {
    log.Printf("%s Error before validate: %v", config.System, err)
    return aphttp.DefaultServerError()
  }
  <% end %>

  validationErrors := <%= local %>.Validate()
  if !validationErrors.Empty() {
    return SerializableValidationErrors{validationErrors}
  }

  <% if after_validate %>
  if err := c.AfterValidate(<%= local %>, tx); err != nil {
    log.Printf("%s Error after validate: %v", config.System, err)
    return aphttp.DefaultServerError()
  }
  <% end %>

  <% if before_insert %>
  if isInsert {
    if err := c.BeforeInsert( <%= local %>, tx); err != nil {
      log.Printf("%s Error before insert: %v", config.System, err)
      return aphttp.DefaultServerError()
    }
  }
  <% end %>
  <% if before_update %>
  if !isInsert {
    if err := c.BeforeUpdate( <%= local %>, tx); err != nil {
      log.Printf("%s Error before update: %v", config.System, err)
      return aphttp.DefaultServerError()
    }
  }
  <% end %>

  if err := method(tx); err != nil {
    if err == apsql.ErrZeroRowsAffected {
      return c.notFound()
    }
    validationErrors = <%= local %>.ValidateFromDatabaseError(err)
    if !validationErrors.Empty() {
      return SerializableValidationErrors{validationErrors}
    }
    log.Printf("%s Error %s <%= pretty %>: %v", config.System, desc, err)
    return aphttp.NewServerError(err)
  }

  <% if after_insert %>
  if isInsert {
    if err := c.AfterInsert(<%= local %>, tx); err != nil {
      log.Printf("%s Error after insert: %v", config.System, err)
      return aphttp.DefaultServerError()
    }
  }
  <% end %>
  <% if after_update %>
  if !isInsert {
    if err := c.AfterUpdate(<%= local %>, tx); err != nil {
      log.Printf("%s Error after update: %v", config.System, err)
      return aphttp.DefaultServerError()
    }
  }
  <% end %>

  return c.serializeInstance(<%= local %>, w)
}

func (c *<%= controller %>) notFound() aphttp.Error {
  return aphttp.NewError(errors.New("No <%= pretty %> matches"), 404)
}

func (c *<%= controller %>) deserializeInstance(file io.Reader) (*model.<%= singular %>,
  aphttp.Error) {

  var wrapped struct {
    <%= singular %> *model.<%= singular %> `json:"<%= json_singular %>"`
  }
  if err := deserialize(&wrapped, file); err != nil {
    return nil, err
  }
  if wrapped.<%= singular %> == nil {
    return nil, aphttp.NewError(errors.New("Could not deserialize <%= singular %> from JSON."),
      http.StatusBadRequest)
  }
  return wrapped.<%= singular %>, nil
}

<% if transform %>
func (c *<%= controller %>) serializeInstance(instance *model.<%= singular %>,
  w http.ResponseWriter) aphttp.Error {

  wrapped := struct {
    <%= singular %> *<%= transform_type %> `json:"<%= json_singular %>"`
  }{<%= transform_method %>(instance)}
  return serialize(wrapped, w)
}

func (c *<%= controller %>) serializeCollection(collection []*model.<%= singular %>,
  w http.ResponseWriter) aphttp.Error {

  wrapped := struct {
    <%= plural %> []*<%= transform_type %> `json:"<%= json_plural %>"`
  }{[]*<%= transform_type %>{}}
  for _, instance := range collection {
    wrapped.<%= plural %> = append(wrapped.<%= plural %>, <%= transform_method %>(instance))
  }
  return serialize(wrapped, w)
}
<% else %>
func (c *<%= controller %>) serializeInstance(instance *model.<%= singular %>,
  w http.ResponseWriter) aphttp.Error {

  wrapped := struct {
    <%= singular %> *model.<%= singular %> `json:"<%= json_singular %>"`
  }{instance}
  return serialize(wrapped, w)
}

func (c *<%= controller %>) serializeCollection(collection []*model.<%= singular %>,
  w http.ResponseWriter) aphttp.Error {

  wrapped := struct {
    <%= plural %> []*model.<%= singular %> `json:"<%= json_plural %>"`
  }{collection}
  return serialize(wrapped, w)
}
<% end %>
ERB

output.write ERB.new(template).result
output.close

`goimports -w ./#{filename}`
