package handler

import (
	"encoding/json"
	"net"
	"net/http"

	"shorturl/internal/config"
	"shorturl/internal/repository"
)

func InternalStats(statsProvider repository.StatsProvider, cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.TrustedSubnet == "" {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		_, subnet, err := net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		ipStr := r.Header.Get("X-Real-IP")
		if ipStr == "" {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		ip := net.ParseIP(ipStr)
		if ip == nil || !subnet.Contains(ip) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		urls, err := statsProvider.CountURLs()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		users, err := statsProvider.CountUsers()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resp := struct {
			URLs  int `json:"urls"`
			Users int `json:"users"`
		}{
			URLs:  urls,
			Users: users,
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}
