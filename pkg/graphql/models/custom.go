package models

import (
	"github.com/kfsoftware/statuspage/pkg/db"
	"time"
)

type StatusPage struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Namespace      string `json:"namespace"`
	Title          string `json:"title"`
	StatusPageItem db.StatusPage
}

type HTTPCheck struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Namespace   string     `json:"namespace"`
	Frecuency   string     `json:"frecuency"`
	URL         string     `json:"url"`
	Status      string     `json:"status"`
	LatestCheck *time.Time `json:"latestCheck"`
	Message     string     `json:"message"`
	ErrorMsg    string     `json:"errorMsg"`
}

func (HTTPCheck) IsCheck() {}

type ICMPCheck struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Namespace   string     `json:"namespace"`
	Frecuency   string     `json:"frecuency"`
	Address     string     `json:"address"`
	Status      string     `json:"status"`
	LatestCheck *time.Time `json:"latestCheck"`
	Message     string     `json:"message"`
	ErrorMsg    string     `json:"errorMsg"`
}

func (ICMPCheck) IsCheck() {}

type TCPCheck struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Namespace   string     `json:"namespace"`
	Frecuency   string     `json:"frecuency"`
	Address     string     `json:"address"`
	Status      string     `json:"status"`
	LatestCheck *time.Time `json:"latestCheck"`
	Message     string     `json:"message"`
	ErrorMsg    string     `json:"errorMsg"`
}

func (TCPCheck) IsCheck() {}

type TLSCheck struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Namespace   string     `json:"namespace"`
	Frecuency   string     `json:"frecuency"`
	Address     string     `json:"address"`
	Status      string     `json:"status"`
	LatestCheck *time.Time `json:"latestCheck"`
	Message     string     `json:"message"`
	ErrorMsg    string     `json:"errorMsg"`
}

func (TLSCheck) IsCheck() {}
