package validation

type AddResponseRequest struct {
	Question          string `json:"question" binding:"required"`                   
	SearchID          string `json:"searchId"`                            
	IsRelatedQuestion *bool  `json:"isRelatedQuestion"`                   
}