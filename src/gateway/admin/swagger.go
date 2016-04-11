package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"gateway/config"
	aphttp "gateway/http"
	"gateway/model"
	apsql "gateway/sql"

	"github.com/gorilla/handlers"
)

type SwaggerController struct {
	matcher *HostMatcher
}

func RouteSwagger(controller *SwaggerController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	routes := map[string]http.Handler{
		"GET": readForHost(db, controller.Swagger),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(routes)).MatcherFunc(controller.matcher.isRouted)
}

func (s *SwaggerController) Swagger(w http.ResponseWriter, r *http.Request, db *apsql.DB, match *HostMatch) aphttp.Error {
	api, err := model.FindAPIForAccountIDForSwagger(db, match.APIID, match.AccountID)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	if !api.EnableSwagger {
		return aphttp.NewError(errors.New("Swagger is disabled"), http.StatusBadRequest)
	}

	endpoints := map[int64]*model.ProxyEndpoint{}
	for _, endpoint := range api.ProxyEndpoints {
		endpoints[endpoint.ID] = endpoint
	}

	schemas := map[int64]*model.ProxyEndpointSchema{}
	for _, schema := range api.ProxyEndpointSchemas {
		schemas[schema.ProxyEndpointID] = schema
	}

	paths := map[string]interface{}{}
	definitions := map[string]interface{}{}
	swagger := map[string]interface{}{
		"swagger": "2.0",
		"info": map[string]interface{}{
			"version":     "1.0.0",
			"title":       api.Name,
			"description": api.Description,
		},
		"paths":       paths,
		"definitions": definitions,
	}

	for _, endpoint := range api.ProxyEndpoints {
		routes, err := endpoint.GetRoutes()
		if err != nil {
			return aphttp.NewError(err, http.StatusBadRequest)
		}
		for _, route := range routes {
			path := map[string]interface{}{}
			if _path, ok := paths[route.Path]; ok {
				path = _path.(map[string]interface{})
			}
			for _, method := range route.Methods {
				parameters := []interface{}{}
				ok := map[string]interface{}{
					"description:": endpoint.Description,
				}
				routePath := []rune(route.Path)
				for i := 0; i < len(routePath); i++ {
					if routePath[i] == '{' {
						key := bytes.Buffer{}
						for i++; routePath[i] != '}'; i++ {
							if routePath[i] != ' ' && routePath[i] != '\t' {
								key.WriteRune(routePath[i])
							}
						}
						parameter := map[string]interface{}{
							"name":     key.String(),
							"in":       "path",
							"required": true,
							"type":     "string",
						}
						parameters = append(parameters, parameter)
					}
				}
				if sch := schemas[endpoint.ID]; sch != nil {
					if sch.RequestSchema != "" {
						schema := map[string]interface{}{}
						parameter := map[string]interface{}{
							"name":     sch.Name,
							"in":       "body",
							"required": true,
							"schema":   schema,
						}
						name := fmt.Sprintf("#/definitions/%vRequest%v", endpoint.Name, sch.Name)
						if sch.ResponseSameAsRequest {
							name = fmt.Sprintf("#/definitions/%v%v", endpoint.Name, sch.Name)
						}
						schema["$ref"] = name
						parameters = append(parameters, parameter)
					}
					if sch.ResponseSchema != "" {
						schema := map[string]interface{}{}
						name := fmt.Sprintf("#/definitions/%vResponse%v", endpoint.Name, sch.Name)
						if sch.ResponseSameAsRequest {
							name = fmt.Sprintf("#/definitions/%v%v", endpoint.Name, sch.Name)
						}
						schema["$ref"] = name
						ok["schema"] = schema
					}
				}
				path[strings.ToLower(method)] = map[string]interface{}{
					"description": endpoint.Description,
					"tags":        []interface{}{endpoint.Name},
					"parameters":  parameters,
					"responses": map[string]interface{}{
						"200": ok,
						"default": map[string]interface{}{
							"description": "error",
						},
					},
				}
			}
			paths[route.Path] = path
		}
	}

	for _, schema := range api.ProxyEndpointSchemas {
		if schema.RequestSchema != "" {
			requestSchema := map[string]interface{}{}
			err := json.Unmarshal([]byte(schema.RequestSchema), &requestSchema)
			if err != nil {
				return aphttp.NewError(err, http.StatusBadRequest)
			}
			name := fmt.Sprintf("%vRequest%v",
				endpoints[schema.ProxyEndpointID].Name,
				schema.Name)
			if schema.ResponseSameAsRequest {
				name = fmt.Sprintf("%v%v",
					endpoints[schema.ProxyEndpointID].Name,
					schema.Name)
			}
			delete(requestSchema, "$schema")
			definitions[name] = requestSchema
		}
		if schema.ResponseSchema != "" && !schema.ResponseSameAsRequest {
			responseSchema := map[string]interface{}{}
			err := json.Unmarshal([]byte(schema.ResponseSchema), &responseSchema)
			if err != nil {
				return aphttp.NewError(err, http.StatusBadRequest)
			}
			name := fmt.Sprintf("%vResponse%v",
				endpoints[schema.ProxyEndpointID].Name,
				schema.Name)
			delete(responseSchema, "$schema")
			definitions[name] = responseSchema
		}
	}

	body, err := json.MarshalIndent(&swagger, "", "    ")
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
	return nil
}
