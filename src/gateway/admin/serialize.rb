#!/usr/bin/env ruby

require 'rubygems'
require 'active_support/inflector'

singular = ARGV.first
plural = singular.pluralize
controller = "#{plural}Controller"

json_singular = singular.underscore
json_plural = json_singular.pluralize

transform = ARGV.count > 1
transform_method = transform ? ARGV[1] : ""
transform_type = transform ? ARGV[2] : ""

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
    #{singular} *model.#{singular} `json:"#{json_singular}"`
  }
  if err := deserialize(&wrapped, r); err != nil {
    return nil, err
  }
  return wrapped.#{singular}, nil
}
GOLANG

if transform
  output.write <<-GOLANG
func (c *#{controller}) serializeInstance(instance *model.#{singular},
  w http.ResponseWriter) aphttp.Error {

  wrapped := struct {
    #{singular} *#{transform_type} `json:"#{json_singular}"`
  }{#{transform_method}(instance)}
  return serialize(wrapped, w)
}

func (c *#{controller}) serializeCollection(collection []*model.#{singular},
  w http.ResponseWriter) aphttp.Error {

  wrapped := struct {
    #{plural} []*#{transform_type} `json:"#{json_plural}"`
  }{}
  for _, instance := range collection {
    wrapped.#{plural} = append(wrapped.#{plural}, #{transform_method}(instance))
  }
  return serialize(wrapped, w)
}
GOLANG
else
  output.write <<-GOLANG
func (c *#{controller}) serializeInstance(instance *model.#{singular},
  w http.ResponseWriter) aphttp.Error {

  wrapped := struct {
    #{singular} *model.#{singular} `json:"#{json_singular}"`
  }{instance}
  return serialize(wrapped, w)
}

func (c *#{controller}) serializeCollection(collection []*model.#{singular},
  w http.ResponseWriter) aphttp.Error {

  wrapped := struct {
    #{plural} []*model.#{singular} `json:"#{json_plural}"`
  }{collection}
  return serialize(wrapped, w)
}
GOLANG
end

output.close

`goimports -w ./#{filename}`
