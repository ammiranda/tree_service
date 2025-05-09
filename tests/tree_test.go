package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"theary_test/database"
	"theary_test/handlers"
	"theary_test/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	api := r.Group("/api")
	{
		api.GET("/tree", handlers.GetTree)
		api.POST("/tree", handlers.CreateNode)
	}
	return r
}

func TestGetTreeNotFound(t *testing.T) {
	// Setup
	setupTestDB(t)
	defer cleanupTestDB(t)

	r := setupTestRouter()

	// Test empty database
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/tree", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "tree not found", response["error"])
}

func TestGetTree(t *testing.T) {
	// Setup
	setupTestDB(t)
	defer cleanupTestDB(t)

	// Create initial root node
	db := database.GetDB()
	_, err := db.Exec("INSERT INTO nodes (label, parent_id) VALUES (?, ?)", "root", nil)
	assert.NoError(t, err)

	r := setupTestRouter()

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/tree", nil)
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Node
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, "root", response[0].Label)
}

func TestCreateNode(t *testing.T) {
	// Setup
	setupTestDB(t)
	defer cleanupTestDB(t)

	// Create initial root node
	db := database.GetDB()
	result, err := db.Exec("INSERT INTO nodes (label, parent_id) VALUES (?, ?)", "root", nil)
	assert.NoError(t, err)
	rootID, err := result.LastInsertId()
	assert.NoError(t, err)

	r := setupTestRouter()

	tests := []struct {
		name       string
		payload    models.CreateNodeRequest
		wantStatus int
		wantError  string
	}{
		{
			name: "valid request",
			payload: models.CreateNodeRequest{
				Label:    "child",
				ParentID: rootID,
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "empty label",
			payload: models.CreateNodeRequest{
				Label:    "",
				ParentID: rootID,
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "non-existent parent",
			payload: models.CreateNodeRequest{
				Label:    "child",
				ParentID: 999,
			},
			wantStatus: http.StatusNotFound,
			wantError:  "parent node not found",
		},
		{
			name: "label too long",
			payload: models.CreateNodeRequest{
				Label:    string(make([]byte, 101)),
				ParentID: rootID,
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, _ := json.Marshal(tt.payload)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/tree", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantError, response["error"])
			}

			if tt.wantStatus == http.StatusCreated {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.payload.Label, response["label"])
				assert.Equal(t, float64(tt.payload.ParentID), response["parentId"])
			}
		})
	}
}

func TestTreeStructure(t *testing.T) {
	// Setup
	setupTestDB(t)
	defer cleanupTestDB(t)

	// Create initial root node
	db := database.GetDB()
	result, err := db.Exec("INSERT INTO nodes (label, parent_id) VALUES (?, ?)", "root", nil)
	assert.NoError(t, err)
	rootID, err := result.LastInsertId()
	assert.NoError(t, err)

	// Create child node
	result, err = db.Exec("INSERT INTO nodes (label, parent_id) VALUES (?, ?)", "child", rootID)
	assert.NoError(t, err)
	childID, err := result.LastInsertId()
	assert.NoError(t, err)

	// Create grandchild node
	_, err = db.Exec("INSERT INTO nodes (label, parent_id) VALUES (?, ?)", "grandchild", childID)
	assert.NoError(t, err)

	r := setupTestRouter()

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/tree", nil)
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Node
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, "root", response[0].Label)
	assert.Len(t, response[0].Children, 1)
	assert.Equal(t, "child", response[0].Children[0].Label)
	assert.Len(t, response[0].Children[0].Children, 1)
	assert.Equal(t, "grandchild", response[0].Children[0].Children[0].Label)
}
