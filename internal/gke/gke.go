package gke

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	container "google.golang.org/api/container/v1beta1"
	pubsub "google.golang.org/api/pubsub/v1"
)

func Handler(l *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pm pubsub.PubsubMessage

		if err := json.NewDecoder(r.Body).Decode(&pm); err != nil {
			l.Error(fmt.Errorf("failed to decode request body: %w", err).Error())
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		requiredAttributes := []string{
			"project_id",
			"cluster_location",
			"cluster_name",
			"type_url",
			"payload",
		}
		for _, key := range requiredAttributes {
			if _, exists := pm.Attributes[key]; !exists {
				l.Error("missing attribute: " + key)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				break
			}
		}

		var event interface{}
		switch pm.Attributes["type_url"] {
		case "type.googleapis.com/google.container.v1beta1.SecurityBulletinEvent":
			event = &container.SecurityBulletinEvent{}
		case "type.googleapis.com/google.container.v1beta1.UpgradeAvailableEvent":
			event = &container.UpgradeAvailableEvent{}
		case "type.googleapis.com/google.container.v1beta1.UpgradeEvent":
			event = &container.UpgradeEvent{}
		case "type.googleapis.com/google.container.v1beta1.UpgradeInfoEvent":
			event = &container.UpgradeInfoEvent{}
		default:
			l.Error("unknown event type: " + pm.Attributes["type_url"])
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal([]byte(pm.Attributes["payload"]), event); err != nil {
			l.Error(fmt.Errorf("failed to unmarshal payload: %w", err).Error())
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		l.Info(
			pm.Data,
			"attributes", map[string]interface{}{
				"project_id":       pm.Attributes["project_id"],
				"cluster_location": pm.Attributes["cluster_location"],
				"cluster_name":     pm.Attributes["cluster_name"],
				"type_url":         pm.Attributes["type_url"],
				"payload":          event,
			},
		)

		w.WriteHeader(http.StatusOK)
	}
}
