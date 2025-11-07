package main

import (
	"net/http"
)

type RulesExecutor interface {
	CheckRequest(*http.Request) bool
	CheckResponse(http.ResponseWriter) bool
	SetRule(*http.Request) error
	NewResponseWriter(w http.ResponseWriter) http.ResponseWriter
	Restrict() (int, error)
	Send() (int, error)
}

type Firewall struct {
	ru RulesExecutor
}

func NewFirewall(ru RulesExecutor) *Firewall {
	return &Firewall{
		ru: ru,
	}
}

func (f *Firewall) CheckRequest(r *http.Request) bool {
	return f.ru.CheckRequest(r)
}

func (f *Firewall) CheckResponse(w http.ResponseWriter) bool {
	return f.ru.CheckResponse(w)
}

func (f *Firewall) SetRule(r *http.Request) error {
	return f.ru.SetRule(r)
}

func (f *Firewall) NewResponseWriter(w http.ResponseWriter) http.ResponseWriter {
	return f.ru.NewResponseWriter(w)
}

func (f *Firewall) Restrict() {
	f.ru.Restrict()
}

func (f *Firewall) Send() {
	f.ru.Send()
}

func (f *Firewall) Wrap(next http.Handler) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			f.SetRule(r)

			fw := f.NewResponseWriter(w)

			if !f.CheckRequest(r) {
				f.Restrict()
				return
			}

			next.ServeHTTP(fw, r)
			if !f.CheckResponse(fw) {
				f.Restrict()
				return
			}

			f.Send()
		},
	)
}
