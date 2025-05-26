package helper

import (
    "fmt"
    "math"
    "strings"
    "gorm.io/gorm"
    "my-project/internal/types"
)

// PaginationOptions contains common pagination parameters
type PaginationOptions struct {
    Page      int
    Limit     int
    SortBy    string
    SortOrder string
}

// GetPaginatedResults is a generic function to get paginated results from database
func GetPaginatedResults[T any](
    db *gorm.DB,
    options PaginationOptions,
    filters map[string]interface{},
    searchFields []string,
) (*types.PagedResponse[T], error) {
    var entities []T
    var count int64

    // Build base query with Debug mode
    query := db.Model(new(T))

    // Apply non-search filters first
    for field, value := range filters {
        if value != nil && field != "searchTerm" {
            query = query.Where(fmt.Sprintf("%s = ?", field), value)
        }
    }

    // Only apply search if searchTerm exists and is not nil
    if searchTerm, ok := filters["searchTerm"]; ok && searchTerm != nil {
        searchValue, ok := searchTerm.(*string)
        if ok && searchValue != nil && *searchValue != "" {
            var conditions []string
            var args []interface{}
            
            for _, field := range searchFields {
                conditions = append(conditions, fmt.Sprintf("%s LIKE ?", field))
                args = append(args, fmt.Sprintf("%%%s%%", *searchValue))
            }
            
            if len(conditions) > 0 {
                query = query.Where(strings.Join(conditions, " OR "), args...)
            }
        }
    }

    // Get total count
    if err := query.Count(&count).Error; err != nil {
        return nil, fmt.Errorf("error counting records: %w", err)
    }

    // Apply sorting and pagination
    offset := (options.Page - 1) * options.Limit
    query = query.Order(fmt.Sprintf("%s %s", options.SortBy, options.SortOrder))
    query = query.Offset(offset).Limit(options.Limit)

    // Execute final query
    if err := query.Find(&entities).Error; err != nil {
        return nil, fmt.Errorf("error fetching records: %w", err)
    }

    // Calculate total pages
    totalPages := int(math.Ceil(float64(count) / float64(options.Limit)))

    return &types.PagedResponse[T]{
        Meta: types.PaginationResponse{
            Page:      options.Page,
            Limit:     options.Limit,
            Count:     count,
            TotalPage: totalPages,
        },
        Data: entities,
    }, nil
}

// GetDefaultPaginationOptions returns pagination options with default values
func GetDefaultPaginationOptions(filters types.CommonFilters) PaginationOptions {
    options := PaginationOptions{
        Page:      1,
        Limit:     10,
        SortBy:    "created_at",
        SortOrder: "asc",
    }

    if filters.Page > 0 {
        options.Page = filters.Page
    }
    if filters.Limit > 0 {
        options.Limit = filters.Limit
    }
    if filters.SortBy != "" {
        options.SortBy = filters.SortBy
    }
    if filters.SortOrder != "" {
        options.SortOrder = filters.SortOrder
    }

    return options
}

// // Example usage in a handler
// func (h *Handler) GetAllItems(c *gin.Context) {
//     var filters types.CommonFilters
//     if err := c.ShouldBindQuery(&filters); err != nil {
//         response.ApiError(c, http.StatusBadRequest, "Invalid query parameters")
//         return
//     }

//     options := GetDefaultPaginationOptions(filters)
    
//     filterMap := map[string]interface{}{
//         "searchTerm": filters.SearchTerm,
//         "status":     "active",  // Additional filters as needed
//     }
    
//     searchFields := []string{"title", "description"}
    
//     result, err := GetPaginatedResults[YourModel](
//         h.db,
//         options,
//         filterMap,
//         searchFields,
//     )
//     if err != nil {
//         response.ApiError(c, http.StatusInternalServerError, "Failed to fetch items")
//         return
//     }

//     response.SendResponse(c, http.StatusOK, true, "Items fetched successfully", result, nil)
// }