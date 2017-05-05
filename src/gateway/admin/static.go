package admin

import (
	"bytes"
	"fmt"
	"mime"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"text/template"
	"time"

	"gateway/config"
	"gateway/core"
	"gateway/logreport"
	"gateway/model"
	"gateway/version"

	"github.com/gorilla/mux"
	"github.com/stripe/stripe-go"
)

var pathRegex = regexp.MustCompile(`API_BASE_PATH_PLACEHOLDER`)
var slashPathRegex = regexp.MustCompile(`/API_BASE_PATH_PLACEHOLDER`)
var brokerHostRegex = regexp.MustCompile(`BROKER_PLACEHOLDER`)
var versionRegex = regexp.MustCompile(`%22VERSION`)
var showVersionRegex = regexp.MustCompile(`SHOW_VERSION`)
var commitRegex = regexp.MustCompile(`COMMIT`)
var devModeRegex = regexp.MustCompile(`DEV_MODE`)
var goosRegex = regexp.MustCompile(`GO_OS`)
var remoteEndpointTypesEnabledRegex = regexp.MustCompile(`REMOTE_ENDPOINT_TYPES_ENABLED`)
var registrationEnabledRegex = regexp.MustCompile(`REGISTRATION_ENABLED`)
var googleAnalyticsTrackingId = regexp.MustCompile(`GOOGLE_ANALYTICS_TRACKING_ID`)
var stripeEnabled = regexp.MustCompile(`ENABLE_PLAN_SUBSCRIPTIONS`)
var stripePublishableKey = regexp.MustCompile(`STRIPE_PUBLISHABLE_KEY`)
var adminApiHost = regexp.MustCompile(`ADMIN_API_HOST`)
var wsHeartbeatInterval = regexp.MustCompile(`WS_HEARTBEAT_INTERVAL`)
var uiPathRegex = regexp.MustCompile(`UI_BASE_PATH_PLACEHOLDER`)
var slashUiPathRegex = regexp.MustCompile(`/UI_BASE_PATH_PLACEHOLDER`)

// Normalize some mime types across OSes
var additionalMimeTypes = map[string]string{
	".svg": "image/svg+xml",
}

type assetResolver func(path string) ([]byte, error)

func init() {
	for k, v := range additionalMimeTypes {
		if err := mime.AddExtensionType(k, v); err != nil {
			logreport.Fatalf("Could not set mime type for %s", k)
		}
	}
}

func adminStaticFileHandler(conf config.ProxyAdmin) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := mux.Vars(r)["path"]

		if path == "" || path == "index.html" {
			serveIndex(w, r, conf)
			return
		}

		// Make JS request objects & functions available to the front-end so that
		// the front-end can introspect on those functions for autocomplete purposes
		if path == "ap-request.js" {
			serveAsset(w, r, "http/request.js", core.Asset)
			return
		}

		serveFile(w, r, path)
	}
}

func serveFile(w http.ResponseWriter, r *http.Request, path string) {
	serveAsset(w, r, path, Asset)
}

func serveAsset(w http.ResponseWriter, r *http.Request, path string, assetResolverFunc assetResolver) {
	data, err := assetResolverFunc(path)
	if err != nil || len(data) == 0 {
		http.NotFound(w, r)
		return
	}
	content := bytes.NewReader(data)
	http.ServeContent(w, r, path, time.Time{}, content)
}

func serveIndex(w http.ResponseWriter, r *http.Request, conf config.ProxyAdmin) {
	data, err := Asset("index.html.template")
	if err != nil {
		http.Error(w, "Could not find index template.", http.StatusInternalServerError)
		return
	}

	funcs := template.FuncMap{
		"interpolate": func(input string) string {
			pathReplacer := func(path string, prefix string) func(string) string {
				return func(string) string {
					if prefix == "" {
						return ""
					}
					return path
				}
			}

			// Handle API base path
			rightless := strings.TrimRight(conf.PathPrefix, "/")
			clean := strings.TrimLeft(rightless, "/")
			input = slashPathRegex.ReplaceAllStringFunc(input, pathReplacer(rightless, conf.PathPrefix))
			input = pathRegex.ReplaceAllStringFunc(input, pathReplacer(clean, conf.PathPrefix))

			// Handle UI base path
			rightless = strings.TrimRight(conf.UiPathPrefix, "/")
			clean = strings.TrimLeft(rightless, "/")
			input = slashUiPathRegex.ReplaceAllStringFunc(input, pathReplacer(rightless, conf.UiPathPrefix))
			input = uiPathRegex.ReplaceAllStringFunc(input, pathReplacer(clean, conf.UiPathPrefix))

			interpolatedValues := map[*regexp.Regexp]string{}

			// Version output
			interpolatedValues[showVersionRegex] = fmt.Sprintf("%t", conf.ShowVersion)
			interpolatedValues[versionRegex] = fmt.Sprintf("%%22%s", version.Name())
			interpolatedValues[commitRegex] = version.Commit()

			interpolatedValues[devModeRegex] = fmt.Sprintf("%t", conf.DevMode)
			interpolatedValues[goosRegex] = runtime.GOOS
			interpolatedValues[remoteEndpointTypesEnabledRegex] = remoteEndpointTypes()
			interpolatedValues[registrationEnabledRegex] = fmt.Sprintf("%t", conf.EnableRegistration)
			interpolatedValues[brokerHostRegex] = conf.BrokerWs
			interpolatedValues[googleAnalyticsTrackingId] = conf.GoogleAnalyticsTrackingId
			if stripe.Key != "" && conf.StripePublishableKey != "" {
				interpolatedValues[stripeEnabled] = "true"
				interpolatedValues[stripePublishableKey] = conf.StripePublishableKey
			} else {
				interpolatedValues[stripeEnabled] = "false"
				interpolatedValues[stripePublishableKey] = ""
			}
			interpolatedValues[adminApiHost] = conf.APIHost
			interpolatedValues[wsHeartbeatInterval] = fmt.Sprintf("%d", conf.WsHeartbeatInterval)

			for k, v := range interpolatedValues {
				input = k.ReplaceAllLiteralString(input, v)
			}

			return input
		},
		"analytics": func() string {
			return conf.GoogleAnalyticsTrackingId
		},
	}

	tmpl := template.New("index")
	if _, err = tmpl.Funcs(funcs).Parse(string(data)); err != nil {
		http.Error(w, "Could not parse index template.", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, nil)
	if err != nil {
		fmt.Fprintf(w, "\n\nError executing template: %v", err)
	}
}

func remoteEndpointTypes() string {
	tags := []string{}
	remoteEndpointTypes, _ := model.AllRemoteEndpointTypes(nil)
	for _, re := range remoteEndpointTypes {
		if re.Enabled {
			tags = append(tags, re.Value)
		}
	}
	return strings.Join(tags, ",")
}
