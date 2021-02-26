package feathers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type requestRegistration struct {
	method  string
	service string
	id      string
	query   map[string]interface{}
}
type httpCaller struct {
	response chan<- interface{}
}

func (c *httpCaller) Callback(data interface{}) {
	c.response <- data
	close(c.response)
}

func (c *httpCaller) CallbackError(data error) {
	c.response <- data
	close(c.response)
}

type HttpProvider struct {
	server *http.ServeMux
	app    *App
}

func NewHttpProvider(app *App) *HttpProvider {
	provider := new(HttpProvider)
	provider.app = app
	return provider
}

func ConfigureHttpProvider(app *App, config map[string]interface{}) error {
	provider := NewHttpProvider(app)
	app.AddProvider("http", provider)
	return nil
}

func (h *HttpProvider) Listen(port int, serveMux *http.ServeMux) {
	h.server = serveMux
	serveMux.Handle("/", h)
	// fmt.Println("HTTP LISTENING")
}

func (p *HttpProvider) Publish(room string, event string, data interface{}, provider string) {
}

func (h *HttpProvider) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	// Todo imporove error handling
	serviceRequest, _ := RequestVars(*request)

	if _, ok := h.app.services[serviceRequest.service]; ok {
		chanResponse := make(chan interface{}, 0)
		caller := httpCaller{
			response: chanResponse,
		}

		switch request.Method {
		case "GET":
			var result interface{}
			if serviceRequest.id != "" {
				h.app.HandleRequest("http", Get, &caller, serviceRequest.service, make(map[string]interface{}), serviceRequest.id)
				result = <-chanResponse

			} else {
				h.app.HandleRequest("http", Find, &caller, serviceRequest.service, make(map[string]interface{}), serviceRequest.id)
				result = <-chanResponse
			}

			h.respond(response, result)
		case "POST":
			h.app.HandleRequest("http", Create, &caller, serviceRequest.service, make(map[string]interface{}), serviceRequest.id)
			result := <-chanResponse
			h.respond(response, result)

		case "PUT":
			h.app.HandleRequest("http", Update, &caller, serviceRequest.service, make(map[string]interface{}), serviceRequest.id)
			result := <-chanResponse
			h.respond(response, result)

		case "PATCH":
			h.app.HandleRequest("http", Patch, &caller, serviceRequest.service, make(map[string]interface{}), serviceRequest.id)
			result := <-chanResponse
			h.respond(response, result)
		case "DELETE":
			h.app.HandleRequest("http", Remove, &caller, serviceRequest.service, make(map[string]interface{}), serviceRequest.id)
			result := <-chanResponse
			h.respond(response, result)
		}
		return
	}
	http.StripPrefix("/", http.FileServer(http.Dir("./public/"))).ServeHTTP(response, request)
}

func (h *HttpProvider) respond(response http.ResponseWriter, data interface{}) {
	dataEnc, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err.Error())
		response.Write([]byte(err.Error()))
		return
	}
	// fmt.Printf("response: %#v", dataEnc)
	response.WriteHeader(responseCode(data))
	response.Write(dataEnc)
}

func responseCode(data interface{}) int {
	code := 200
	if err, ok := data.(fErr.FeathersError); ok {
		code = err.Code
	}
	return code
}

func RequestVars(request http.Request) (requestRegistration, error) {
	url, _ := url.Parse(request.RequestURI)
	var serviceName, id string
	pathParts := strings.Split(url.Path, "/")
	if len(pathParts) >= 2 {
		serviceName = pathParts[1]
	}
	if len(pathParts) >= 3 {
		id = pathParts[2]
	}
	return requestRegistration{
		method:  request.Method,
		id:      id,
		service: serviceName,
	}, nil
}
