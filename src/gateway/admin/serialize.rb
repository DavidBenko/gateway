#!/usr/bin/env ruby

require 'rubygems'
require 'active_support/inflector'

plural = ARGV.last
singular = plural.singularize
controller = "#{plural}Controller"

json_plural = plural.underscore
json_singular = singular.underscore

filename = "./#{json_plural}_gen.go"
output = File.open(filename, "w")

output.write <<-GOLANG

package admin

/*******************************************************
*******************************************************
***                                                 ***
*** This is generated code. Do not edit directly.   ***
***                                                 ***
*******************************************************
*******************************************************/

import (
  "gateway/model"
  aphttp "gateway/http"
)

func (c *#{controller}) deserializeInstance(r *http.Request) (*model.#{singular},
  error) {

  var wrapped struct {
    #{singular} *model.#{singular} `json:"account"`
  }
  if err := deserialize(&wrapped, r); err != nil {
    return nil, err
  }
  return wrapped.#{singular}, nil
}

func (c *#{controller}) serializeInstance(instance *model.#{singular},
  w http.ResponseWriter) aphttp.Error {

  wrapped := struct {
    #{singular} *model.#{singular} `json:"#{json_singular}"`
  }{instance}
  return serialize(wrapped, w)
}

func (c *#{controller}) serializeCollection(collection []*model.Account,
  w http.ResponseWriter) aphttp.Error {

  wrapped := struct {
    #{plural} []*model.#{singular} `json:"#{json_plural}"`
  }{collection}
  return serialize(wrapped, w)
}


GOLANG

output.close

`goimports -w ./#{filename}`
