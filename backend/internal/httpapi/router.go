package httpapi

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"ao-regears/backend/internal/config"
	"ao-regears/backend/internal/store"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type API struct {
	cfg   config.Config
	store *store.Store
}

func New(cfg config.Config, st *store.Store) *gin.Engine {
	api := &API{cfg: cfg, store: st}
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.GET("/api/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	r.POST("/api/auth/login", api.login)
	r.POST("/api/auth/signup", api.signup)
	r.POST("/api/auth/logout", api.logout)

	protected := r.Group("/api", api.auth())
	protected.GET("/auth/me", api.me)
	protected.GET("/dashboard", api.dashboard)
	protected.GET("/builds", api.listBuilds)
	protected.POST("/builds", api.requireOfficer(), api.createBuild)
	protected.PATCH("/builds/:id", api.requireOfficer(), api.updateBuild)
	protected.DELETE("/builds/:id", api.requireOfficer(), api.deleteBuild)
	protected.GET("/regears", api.listRegears)
	protected.POST("/regears", api.createRegear)
	protected.PATCH("/regears/:id/status", api.updateRegearStatus)
	protected.GET("/inventory", api.listInventory)
	protected.POST("/inventory", api.requireOfficer(), api.upsertInventory)
	protected.PATCH("/inventory/:id", api.requireOfficer(), api.updateInventory)
	protected.DELETE("/inventory/:id", api.requireOfficer(), api.deleteInventory)
	protected.POST("/shopping-lists/generate", api.requireOfficer(), api.generateShoppingList)
	protected.GET("/shopping-lists/latest", api.latestShoppingList)
	protected.GET("/members/history", api.requireOfficer(), api.memberHistory)
	protected.PATCH("/members/:id/role", api.requireOfficer(), api.updateRole)
	protected.DELETE("/members/:id", api.requireOfficer(), api.deleteMember)
	return r
}

func (a *API) login(c *gin.Context) {
	var req struct {
		PlayerName string `json:"playerName"`
		Password   string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.PlayerName) == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}
	user, err := a.store.Login(req.PlayerName, req.Password)
	if err != nil {
		status := http.StatusInternalServerError
		if err == sql.ErrNoRows {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": "invalid username or password"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (a *API) signup(c *gin.Context) {
	var req struct {
		PlayerName string `json:"playerName"`
		Password   string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.PlayerName) == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}
	
	// Validate player name
	playerName := strings.TrimSpace(req.PlayerName)
	if len(playerName) < 2 || len(playerName) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player name must be between 2 and 50 characters"})
		return
	}
	
	// Validate password
	if len(req.Password) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 3 characters"})
		return
	}
	
	user, err := a.store.CreateUser(playerName, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (a *API) logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (a *API) auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		var user store.User
		var err error
		if token == "" {
			user, err = a.store.DefaultUser()
		} else {
			user, err = a.store.UserByToken(token)
		}
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user context unavailable"})
			c.Abort()
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

func (a *API) requireOfficer() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)
		if user.Role != "Officer" && user.Role != "Admin" && user.Role != "Owner" {
			c.JSON(http.StatusForbidden, gin.H{"error": "officer permission required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func currentUser(c *gin.Context) store.User {
	value, _ := c.Get("user")
	return value.(store.User)
}

func (a *API) me(c *gin.Context) {
	c.JSON(http.StatusOK, currentUser(c))
}

func (a *API) dashboard(c *gin.Context) {
	result, err := a.store.Dashboard(currentUser(c))
	respond(c, result, err)
}

func (a *API) listBuilds(c *gin.Context) {
	result, err := a.store.ListBuilds()
	respond(c, result, err)
}

func (a *API) createBuild(c *gin.Context) {
	var req store.Build
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := a.store.CreateBuild(currentUser(c), req)
	respond(c, result, err)
}

func (a *API) updateBuild(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req store.Build
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := a.store.UpdateBuild(currentUser(c), id, req)
	respond(c, result, err)
}

func (a *API) deleteBuild(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	respond(c, gin.H{"deleted": true}, a.store.DeleteBuild(currentUser(c), id))
}

func (a *API) listRegears(c *gin.Context) {
	result, err := a.store.ListRegears(currentUser(c))
	respond(c, result, err)
}

func (a *API) createRegear(c *gin.Context) {
	var req store.RegearRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := a.store.CreateRegear(currentUser(c), req)
	respond(c, result, err)
}

func (a *API) updateRegearStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req struct {
		Status         string `json:"status"`
		PickupLocation string `json:"pickupLocation"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := a.store.UpdateRegearStatus(currentUser(c), id, req.Status, req.PickupLocation)
	respond(c, result, err)
}

func (a *API) listInventory(c *gin.Context) {
	result, err := a.store.ListInventory()
	respond(c, result, err)
}

func (a *API) upsertInventory(c *gin.Context) {
	var req store.InventoryItem
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := a.store.UpsertInventory(currentUser(c), req)
	respond(c, result, err)
}

func (a *API) updateInventory(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req store.InventoryItem
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := a.store.UpdateInventory(currentUser(c), id, req)
	respond(c, result, err)
}

func (a *API) deleteInventory(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	respond(c, gin.H{"deleted": true}, a.store.DeleteInventory(currentUser(c), id))
}

func (a *API) generateShoppingList(c *gin.Context) {
	result, err := a.store.GenerateShoppingList(currentUser(c))
	respond(c, result, err)
}

func (a *API) latestShoppingList(c *gin.Context) {
	result, err := a.store.LatestShoppingList()
	respond(c, result, err)
}

func (a *API) memberHistory(c *gin.Context) {
	result, err := a.store.MemberHistory()
	respond(c, result, err)
}

func (a *API) updateRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	var req struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = a.store.UpdateUserRole(currentUser(c), id, req.Role)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (a *API) deleteMember(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	err = a.store.DeleteUser(currentUser(c), id)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}

func respond(c *gin.Context, value any, err error) {
	if err == nil {
		c.JSON(http.StatusOK, value)
		return
	}
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
