package api

import (
	"dras.cssllc.io/server/dras/dbcon"
	"encoding/json"
	"errors"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/swaggest/swgui"
	v3 "github.com/swaggest/swgui/v3"
	"gorm.io/gorm"
	"log"
	"net/http"
)

func generateOpenAPISpec(tables []dbconn.Table) *openapi3.T {
	swagger := openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:       "CRUD REST API",
			Description: "A REST API server with CRUD operations for a Postgres database.",
			Version:     "1.0",
		},
		Paths: openapi3.Paths{},
		Tags: openapi3.Tags{
			&openapi3.Tag{
				Name:        "OAS 3.0 Specification - YAML",
				Description: "Download the OAS specification in YAML format",
				ExternalDocs: &openapi3.ExternalDocs{
					Description: "YAML",
					URL:         "/spec/oas.yaml",
				},
			},
			&openapi3.Tag{
				Name:        "OAS 3.0 Specification - JSON",
				Description: "Download the OAS specification in JSON format",
				ExternalDocs: &openapi3.ExternalDocs{
					Description: "JSON",
					URL:         "/spec/oas.json",
				},
			},
		},
		Components: &openapi3.Components{
			Schemas: make(map[string]*openapi3.SchemaRef),
		},
	}

	for _, table := range tables {
		tableName := table.TableName

		// Get columns for the table
		columns, err := dbconn.GetColumns(tableName)
		if err != nil {
			panic("Error fetching columns for table " + tableName)
		}

		// Generate schema for the table
		schema := openapi3.NewObjectSchema()
		for _, column := range columns {
			schema.WithProperty(column.ColumnName, openapi3.NewStringSchema())
		}

		swagger.Components.Schemas[tableName] = &openapi3.SchemaRef{
			Value: schema,
		}

		pathItem := &openapi3.PathItem{}

		// Add the schema as a response for the GET request
		response := openapi3.NewResponse()
		response.WithDescription("An array of " + tableName + " entities.")
		response.WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
			Value: openapi3.NewArraySchema().WithItems(swagger.Components.Schemas[tableName].Value),
		}))

		operation := openapi3.NewOperation()
		operation.Description = "Retrieve all entities from the " + tableName + " table."
		operation.AddResponse(200, response)

		pathItem.Get = operation

		swagger.Paths["/"+tableName] = pathItem
	}

	return &swagger
}

var dbConn *gorm.DB

// SetupRouter - Set up router and routes
func SetupRouter(db *gorm.DB) *gin.Engine {

	dbConn = db

	// Swagger
	router := gin.Default()

	tables, err := dbconn.GetTables()
	if err != nil {
		log.Fatal("Error fetching tables from the database")
	}

	swaggerSpec := generateOpenAPISpec(tables)

	jsonBytes, _ := swaggerSpec.MarshalJSON()
	if err != nil {
		log.Fatalf("Error converting OpenAPI 3.0 specification to JSON: %v", err)
	}

	swguiHandler := v3.NewHandlerWithConfig(
		swgui.Config{
			BasePath: "/swagger-ui/",
			SettingsUI: map[string]string{
				"tryItOutEnabled": "true",
				"spec":            string(jsonBytes),
			},
		},
	)
	router.GET(swguiHandler.BasePath+"*any", gin.WrapH(swguiHandler))

	router.GET("/spec/oas.yaml", func(c *gin.Context) {
		c.Render(http.StatusOK, render.YAML{Data: swaggerSpec})
	})

	router.GET("/spec/oas.json", func(c *gin.Context) {
		c.Render(http.StatusOK, render.JSON{Data: swaggerSpec})
	})

	for _, table := range tables {
		tableGroup := router.Group("/" + table.TableName)
		tableGroup.GET("/", getAllEntities(table.TableName))
		tableGroup.GET("/:id", getEntity(table.TableName))
		//tableGroup.POST("/", createEntity(table.TableName))
		//tableGroup.PUT("/:id", updateEntity(table.TableName))
		//tableGroup.DELETE("/:id", deleteEntity(table.TableName))
	}

	return router
}

func getAllEntities(tableName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var results []map[string]interface{}
		dbConn.Table(tableName).Find(&results)
		jsonResult, err := json.Marshal(results)
		if err != nil {
			c.AbortWithStatus(404)
			_ = c.Error(err)
			return
		}
		c.JSON(200, string(jsonResult))
	}
}

func getEntity(tableName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var results []map[string]interface{}
		id := c.Params.ByName("id")
		dbConn.Table(tableName).Where(tableName+"_id = ?", id).Find(&results)
		jsonResult, err := json.Marshal(results[0])
		if err != nil {
			c.AbortWithStatus(404)
			_ = c.Error(err)
			return
		}
		c.JSON(200, string(jsonResult))

	}
}

func createEntity(tableName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var entity dbconn.Entity
		c.BindJSON(&entity)

		dbConn.Table(tableName).Create(&entity)
		c.JSON(201, entity)
	}
}

func updateEntity(tableName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Params.ByName("id")
		var entity dbconn.Entity
		err := dbConn.Table(tableName).First(&entity, id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.AbortWithStatus(404)
			c.Error(err)
			return
		}

		var updatedEntity dbconn.Entity
		c.BindJSON(&updatedEntity)

		dbConn.Table(tableName).Model(&entity).Updates(updatedEntity)
		c.JSON(200, entity)
	}
}

func deleteEntity(tableName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Params.ByName("id")
		var entity dbconn.Entity
		err := dbConn.Table(tableName).First(&entity, id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.AbortWithStatus(404)
			c.Error(err)
			return
		}
		dbConn.Table(tableName).Delete(&entity)
		c.Status(204)
	}
}
