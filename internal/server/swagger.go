package server

import "github.com/gin-gonic/gin"

// RegisterSwagger mounts API documentation.
//
// Handler comments throughout the codebase use swaggo annotations. Generate the
// spec with:
//
//	cd backend && swag init -g cmd/api/main.go -o docs/swagger
//
// then uncomment the gin-swagger wiring below. Until generated, /docs returns a hint.
func RegisterSwagger(r *gin.Engine) {
	r.GET("/swagger", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Run `swag init -g cmd/api/main.go -o docs/swagger` then enable gin-swagger in swagger.go",
			"design":  "See docs/API.md for the full REST design.",
		})
	})

	// After generating docs, replace the handler above with:
	//
	//   import (
	//       _ "github.com/collegeassess/backend/docs/swagger"
	//       swaggerFiles "github.com/swaggo/files"
	//       ginSwagger "github.com/swaggo/gin-swagger"
	//   )
	//   r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
