package pagination

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Parse helper
func Parse(c *gin.Context) (page int, perPage int) {
	minPage := float64(1)
	minPerPage := float64(1)
	maxPerPage := float64(50)

	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ = strconv.Atoi(c.DefaultQuery("perPage", "25"))

	page = int(math.Max(float64(page), minPage))
	perPageFloored := math.Max(minPerPage, float64(perPage))
	perPage = int(math.Min(perPageFloored, maxPerPage))

	return page, perPage
}

// ToQuery takes a page and perPage and converts it to a limit and offset
func ToQuery(perPage int, page int) (limit int, offset int) {
	limit = perPage
	offset = int(math.Ceil(float64((page - 1) * limit)))

	return limit, offset
}
