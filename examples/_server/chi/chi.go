// A simpler standalone example for xsd validation and http
package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/terminalstatic/go-xsd-validate"
	"io/ioutil"
	"log"
	"net/http"
)

var xsdHandler *xsdvalidate.XsdHandler

func parseBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(w, fmt.Sprintf("%s<error><![CDATA[%s]]></error>", xml.Header, err))
			return
		}

		ctx := context.WithValue(r.Context(), "body", body)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func validateBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body := (ctx.Value("body").([]byte))
		err := xsdHandler.ValidateMem(body, xsdvalidate.ParsErrVerbose)
		if err != nil {
			fmt.Fprintf(w, fmt.Sprintf("%s<error><![CDATA[%s]]></error>", xml.Header, err))
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	addr := ":9999"
	xsdvalidate.Init()
	defer xsdvalidate.Cleanup()
	var err error
	xsdHandler, err = xsdvalidate.NewXsdHandlerUrl("address.xsd", xsdvalidate.ParsErrDefault)
	defer xsdHandler.Free()
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(parseBody)
	r.Use(validateBody)

	r.Route("/", func(r chi.Router) {
		r.Post("/address", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "application/xml; charset=utf-8")

			ctx := r.Context()
			body := (ctx.Value("body").([]byte))

			fmt.Fprintf(w, fmt.Sprintf("%s<request-content><![CDATA[%s]]></request-content><no-error>No errors</no-error>", xml.Header, body))
		})

	})

	/*	var err error
		xsdHandler, err = xsdvalidate.NewXsdHandlerUrl("address.xsd", xsdvalidate.ParsErrDefault)
		defer xsdHandler.Free()
		if err != nil {
			panic(err)
		}
		http.HandleFunc("/address", handler)*/
	fmt.Printf("Starting http server on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
