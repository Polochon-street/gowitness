package api

import (
	"encoding/json"
	"net/http"

	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
)

type tagListResponse struct {
	Value []string `json:"tags"`
}

// TagListHandler lists all tags
//
//	@Summary		Get all tags
//	@Description	Get all unique tags that have been applied to URLs.
//	@Tags			Tags
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	tagListResponse
//	@Router			/results/tags [get]
func (h *ApiHandler) TagListHandler(w http.ResponseWriter, r *http.Request) {
	var results = &tagListResponse{}

	if err := h.DB.Model(&models.Tag{}).Distinct("name").
		Find(&results.Value).Error; err != nil {

		log.Error("could not find distinct tags", "err", err)
		return
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
