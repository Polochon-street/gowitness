package api

import (
	"encoding/json"
	"net/http"

	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
)

type addTagRequest struct {
	ResultID uint   `json:"result_id"`
	TagName  string `json:"tag_name"`
}

type removeTagRequest struct {
	ResultID uint   `json:"result_id"`
	TagName  string `json:"tag_name"`
}

// AddTagHandler adds a tag to a result
//
//	@Summary		Add a tag to a result
//	@Description	Associates a tag with a URL result. Creates the tag if it doesn't exist.
//	@Tags			Tags
//	@Accept			json
//	@Produce		json
//	@Param			query	body		addTagRequest	true	"The result ID and tag name"
//	@Success		200		{string}	string			"ok"
//	@Router			/results/tag/add [post]
func (h *ApiHandler) AddTagHandler(w http.ResponseWriter, r *http.Request) {
	var request addTagRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error("failed to read json request", "err", err)
		http.Error(w, "Error reading JSON request", http.StatusInternalServerError)
		return
	}

	if request.TagName == "" {
		http.Error(w, "No tag name provided", http.StatusBadRequest)
		return
	}

	log.Info("adding tag to result", "result_id", request.ResultID, "tag", request.TagName)

	// Find or create the tag
	var tag models.Tag
	if err := h.DB.Where(models.Tag{Name: request.TagName}).FirstOrCreate(&tag).Error; err != nil {
		log.Error("failed to find or create tag", "err", err)
		http.Error(w, "Error processing tag", http.StatusInternalServerError)
		return
	}

	// Load the result
	var result models.Result
	if err := h.DB.First(&result, request.ResultID).Error; err != nil {
		log.Error("failed to find result", "err", err)
		http.Error(w, "Result not found", http.StatusNotFound)
		return
	}

	// Add the tag to the result (GORM will handle duplicate prevention via composite PK)
	if err := h.DB.Model(&result).Association("Tags").Append(&tag); err != nil {
		log.Error("failed to add tag to result", "err", err)
		http.Error(w, "Error adding tag to result", http.StatusInternalServerError)
		return
	}

	response := `ok`
	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

// RemoveTagHandler removes a tag from a result
//
//	@Summary		Remove a tag from a result
//	@Description	Removes the association between a tag and a URL result.
//	@Tags			Tags
//	@Accept			json
//	@Produce		json
//	@Param			query	body		removeTagRequest	true	"The result ID and tag name"
//	@Success		200		{string}	string				"ok"
//	@Router			/results/tag/remove [post]
func (h *ApiHandler) RemoveTagHandler(w http.ResponseWriter, r *http.Request) {
	var request removeTagRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error("failed to read json request", "err", err)
		http.Error(w, "Error reading JSON request", http.StatusInternalServerError)
		return
	}

	log.Info("removing tag from result", "result_id", request.ResultID, "tag", request.TagName)

	// Find the tag
	var tag models.Tag
	if err := h.DB.Where("name = ?", request.TagName).First(&tag).Error; err != nil {
		log.Error("tag not found", "err", err)
		http.Error(w, "Tag not found", http.StatusNotFound)
		return
	}

	// Load the result
	var result models.Result
	if err := h.DB.First(&result, request.ResultID).Error; err != nil {
		log.Error("failed to find result", "err", err)
		http.Error(w, "Result not found", http.StatusNotFound)
		return
	}

	// Remove the tag from the result
	if err := h.DB.Model(&result).Association("Tags").Delete(&tag); err != nil {
		log.Error("failed to remove tag from result", "err", err)
		http.Error(w, "Error removing tag from result", http.StatusInternalServerError)
		return
	}

	response := `ok`
	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
