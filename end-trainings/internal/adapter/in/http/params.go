package httpin

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func parseInt64Param(c *gin.Context, param string) (int64, error) {
	return strconv.ParseInt(c.Param(param), 10, 64)
}

func parseInt64Query(c *gin.Context, query string) (int64, error) {
	return strconv.ParseInt(c.Query(query), 10, 64)
}
