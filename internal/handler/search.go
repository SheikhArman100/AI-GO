package handler

import (
	"fmt"
	"my-project/internal/database"
	"my-project/internal/helper"
	"my-project/internal/model"
	"my-project/internal/response"
	"my-project/internal/types"
	"my-project/internal/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	db database.Service
}

func NewSearchHandler(db database.Service) *SearchHandler {
	return &SearchHandler{db: db}
}

func (h *SearchHandler) CreateResponse(c *gin.Context) {
	// get user info
	userInfo, err := helper.GetUserInfoFromContext(c)
	if err != nil {
		response.ApiError(c, http.StatusUnauthorized, err.Error())
		return
	}

	// get validated data
	req, err := helper.GetValidatedFromContext[validation.AddResponseRequest](c)
	if err != nil {
		response.ApiError(c, http.StatusInternalServerError, err.Error())
		return
	}
	fmt.Println("Request data:", req)

	if req.IsRelatedQuestion != nil && *req.IsRelatedQuestion && req.SearchID == "" {
		response.ApiError(c, http.StatusBadRequest, "searchId is required when isRelatedQuestion is true")
		return
	}

	if req.SearchID == "" {
		fmt.Println("Creating new search for question:", req.Question)
		// Get AI response for new search
		aiResponse, err := helper.GetAiResponse(helper.AiRequestPayload{
			Question: req.Question,
			History:  []helper.HistoryItem{},
		})

		if err != nil {
			fmt.Println("Error getting AI response:", err)
			aiResponse = &helper.AiResponse{} // Provide a fallback to avoid nil dereference
			// response.ApiError(c, http.StatusInternalServerError, err.Error())
			// return
		}

		// create new search with response
		newSearch := model.Search{
			Title:  req.Question,
			Ip:     c.ClientIP(),
			UserID: userInfo.ID,
			Responses: []model.Response{
				{
					Question:          req.Question,
					Details:           aiResponse.AnswerDetails,
					RelatedQuestions:  aiResponse.RelatedQuestions,
					Images:            aiResponse.Images,
					Charts:            helper.ConvertToMapSlice(aiResponse.Charts),
					IsRelatedQuestion: false,
				},
			},
		}

		// Set fallback values
		if newSearch.Responses[0].Details == "" {
			newSearch.Responses[0].Details = "An error occurred. Try again later."
		}
		if newSearch.Responses[0].RelatedQuestions == nil {
			newSearch.Responses[0].RelatedQuestions = []string{}
		}
		if newSearch.Responses[0].Images == nil {
			newSearch.Responses[0].Images = []string{}
		}
		if newSearch.Responses[0].Charts == nil {
			newSearch.Responses[0].Charts = []map[string]interface{}{}
		}

		if err := h.db.DB().Create(&newSearch).Error; err != nil {
			response.ApiError(c, http.StatusInternalServerError, "Failed to create search")
			return
		}

		response.SendResponse(c, http.StatusCreated, true, "Response created successfully", newSearch, nil)
		return
	} else {
		var existingSearch model.Search
		if err := h.db.DB().First(&existingSearch, "id = ?", req.SearchID).Error; err != nil {
			response.ApiError(c, http.StatusNotFound, "Search not found")
			return
		}

		// Validate user ownership
		if existingSearch.UserID != userInfo.ID {
			response.ApiError(c, http.StatusForbidden, "You are not authorized to access this search")
			return
		}

		// Get previous responses for AI history
		var responses []model.Response
		if err := h.db.DB().Where("search_id = ?", existingSearch.ID).Find(&responses).Error; err != nil {
			response.ApiError(c, http.StatusInternalServerError, "Failed to fetch responses")
			return
		}

		// Create history items from previous responses
		history := make([]helper.HistoryItem, len(responses))
		for i, resp := range responses {
			history[i] = helper.HistoryItem{
				Question:  resp.Question,
				Answer:    resp.Details,
				Timestamp: existingSearch.CreatedAt.Unix(),
			}
		}

		// Get AI response with history
		aiResponse, err := helper.GetAiResponse(helper.AiRequestPayload{
			Question: req.Question,
			History:  history,
		})
		if err != nil {
			fmt.Println("Error getting AI response:", err)
			aiResponse = &helper.AiResponse{} // Provide a fallback to avoid nil dereference
			// response.ApiError(c, http.StatusInternalServerError, err.Error())
			// return
		}

		// Create new response
		newResponse := model.Response{
			SearchID:          existingSearch.ID,
			Question:          req.Question,
			Details:           aiResponse.AnswerDetails,
			RelatedQuestions:  aiResponse.RelatedQuestions,
			Images:            aiResponse.Images,
			Charts:            helper.ConvertToMapSlice(aiResponse.Charts),
			IsRelatedQuestion: req.IsRelatedQuestion != nil && *req.IsRelatedQuestion,
		}

		// Set fallback values
		if newResponse.Details == "" {
			newResponse.Details = "An error occurred. Try again later."
		}
		if newResponse.RelatedQuestions == nil {
			newResponse.RelatedQuestions = []string{}
		}
		if newResponse.Images == nil {
			newResponse.Images = []string{}
		}
		if newResponse.Charts == nil {
			newResponse.Charts = []map[string]interface{}{}
		}

		if err := h.db.DB().Create(&newResponse).Error; err != nil {
			response.ApiError(c, http.StatusInternalServerError, "Failed to create response")
			return
		}

		response.SendResponse(c, http.StatusCreated, true, "Response created successfully", newResponse, nil)

	}
}

func (h *SearchHandler) GetAllSearches(c *gin.Context) {
	// Get user info
	userInfo, err := helper.GetUserInfoFromContext(c)
	if err != nil {
		response.ApiError(c, http.StatusUnauthorized, err.Error())
		return
	}

	// Bind query parameters to filters
	var filters struct {
		types.CommonFilters
        Ip    *string `form:"ip"`
        Title *string `form:"title"`
	}
	if err := c.ShouldBindQuery(&filters); err != nil {
		response.ApiError(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	options := helper.PaginationOptions{
		Page:      filters.Page,
		Limit:     filters.Limit,
		SortBy:    filters.SortBy,
		SortOrder: filters.SortOrder,
	}

	// Build filter map with IP and Title
	filterMap := map[string]interface{}{
		"user_id":    userInfo.ID,
		"searchTerm": filters.SearchTerm,
	}

	// Add IP filter if provided
	if filters.Ip != nil && *filters.Ip != "" {
		filterMap["ip"] = filters.Ip
	}

	// Add Title filter if provided
	if filters.Title != nil && *filters.Title != "" {
		filterMap["title"] = filters.Title
	}
    //searchable fields
	searchFields := []string{"title"}

	result, err := helper.GetPaginatedResults[model.Search](
		h.db.DB(),
		options,
		filterMap,
		searchFields,
	)

	if err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to fetch searches")
		return
	}

	response.SendResponse(c, http.StatusOK, true, "Searches fetched successfully", result, nil)
}

func (h *SearchHandler) GetSearchByID(c *gin.Context) {
	// Get user info
	userInfo, err := helper.GetUserInfoFromContext(c)
	if err != nil {
		response.ApiError(c, http.StatusUnauthorized, err.Error())
		return
	}

	searchID := c.Param("searchId")
	if searchID == "" {
		response.ApiError(c, http.StatusBadRequest, "searchId is required")
		return
	}

	var search model.Search
	if err := h.db.DB().Preload("Responses").First(&search, "id = ?", searchID).Error; err != nil {
		response.ApiError(c, http.StatusNotFound, "Search not found")
		return
	}

	// Validate user ownership
	if search.UserID != userInfo.ID {
		response.ApiError(c, http.StatusForbidden, "You are not authorized to access this search")
		return
	}

	response.SendResponse(c, http.StatusOK, true, "Search fetched successfully", search, nil)
}
