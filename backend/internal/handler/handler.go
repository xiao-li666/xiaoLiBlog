package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"blogapp/backend/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	DB        *gorm.DB
	RDB       *redis.Client
	JWTKey    []byte
	UploadDir string
}

type Claims struct {
	UserID uint   `json:"userId"`
	Role   string `json:"role"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type ArticleInput struct {
	Title      string `json:"title"`
	Summary    string `json:"summary"`
	CoverURL   string `json:"coverUrl"`
	Content    string `json:"content"`
	Tags       string `json:"tags"`
	Status     string `json:"status"`
	CategoryID uint   `json:"categoryId"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateMeInput struct {
	Name     string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

type CommentInput struct {
	Body string `json:"body"`
}

type CategoryInput struct {
	Name string `json:"name"`
}

func New(db *gorm.DB, rdb *redis.Client, jwtKey []byte, uploadDir string) *Handler {
	return &Handler{
		DB:        db,
		RDB:       rdb,
		JWTKey:    jwtKey,
		UploadDir: uploadDir,
	}
}

func (h *Handler) AdminStats(c *gin.Context) {
	var articleTotal int64
	var publishedTotal int64
	var draftTotal int64
	var categoryTotal int64
	var commentTotal int64
	var userTotal int64
	var likeTotal int64
	var favoriteTotal int64

	_ = h.DB.Model(&model.Article{}).Count(&articleTotal).Error
	_ = h.DB.Model(&model.Article{}).Where("status = ?", "published").Count(&publishedTotal).Error
	_ = h.DB.Model(&model.Article{}).Where("status = ?", "draft").Count(&draftTotal).Error
	_ = h.DB.Model(&model.Category{}).Count(&categoryTotal).Error
	_ = h.DB.Model(&model.Comment{}).Count(&commentTotal).Error
	_ = h.DB.Model(&model.User{}).Count(&userTotal).Error
	_ = h.DB.Model(&model.Reaction{}).Where("type = ?", "like").Count(&likeTotal).Error
	_ = h.DB.Model(&model.Reaction{}).Where("type = ?", "favorite").Count(&favoriteTotal).Error

	c.JSON(http.StatusOK, gin.H{
		"articles": gin.H{
			"total":     articleTotal,
			"published": publishedTotal,
			"draft":     draftTotal,
		},
		"categories": categoryTotal,
		"comments":   commentTotal,
		"users":      userTotal,
		"reactions": gin.H{
			"likes":     likeTotal,
			"favorites": favoriteTotal,
		},
	})
}

func (h *Handler) UploadFile(c *gin.Context) {
	kind := strings.ToLower(strings.TrimSpace(c.PostForm("kind")))
	file, err := c.FormFile("file")
	if err != nil {
		badRequest(c, err)
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if kind == "markdown" || ext == ".md" || ext == ".markdown" {
		if file.Size > 2*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "markdown file is too large"})
			return
		}
		src, err := file.Open()
		if err != nil {
			internalError(c, err)
			return
		}
		defer src.Close()
		content, err := io.ReadAll(src)
		if err != nil {
			internalError(c, err)
			return
		}
		title := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))
		c.JSON(http.StatusOK, gin.H{
			"kind":     "markdown",
			"title":    title,
			"content":  string(content),
			"filename": file.Filename,
		})
		return
	}

	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file is too large"})
		return
	}

	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "only png, jpg, jpeg, gif, webp, md are supported"})
		return
	}

	name := fmt.Sprintf("%s%s", randomSlug(), ext)
	dstPath := filepath.Join(h.UploadDir, name)
	if err := c.SaveUploadedFile(file, dstPath); err != nil {
		internalError(c, err)
		return
	}

	scheme := "http"
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	url := fmt.Sprintf("%s://%s/uploads/%s", scheme, c.Request.Host, name)
	c.JSON(http.StatusOK, gin.H{
		"kind":     "image",
		"url":      url,
		"filename": file.Filename,
	})
}

func (h *Handler) UploadAvatar(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		badRequest(c, err)
		return
	}
	if file.Size > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar file is too large"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "only png, jpg, jpeg, gif, webp are supported"})
		return
	}

	name := fmt.Sprintf("avatar-%s%s", randomSlug(), ext)
	dstPath := filepath.Join(h.UploadDir, name)
	if err := c.SaveUploadedFile(file, dstPath); err != nil {
		internalError(c, err)
		return
	}

	scheme := "http"
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	url := fmt.Sprintf("%s://%s/uploads/%s", scheme, c.Request.Host, name)
	c.JSON(http.StatusOK, gin.H{
		"kind":     "avatar",
		"url":      url,
		"filename": file.Filename,
	})
}

func (h *Handler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err)
		return
	}
	if len(input.Password) < 6 || input.Email == "" || input.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, email and password are required"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		internalError(c, err)
		return
	}

	role := "user"
	var userCount int64
	if err := h.DB.Model(&model.User{}).Count(&userCount).Error; err == nil && userCount == 0 {
		role = "admin"
	}
	user := model.User{
		Name:         input.Name,
		Email:        strings.ToLower(strings.TrimSpace(input.Email)),
		PasswordHash: string(hash),
		Role:         role,
	}
	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already registered"})
		return
	}
	token, err := h.issueToken(user)
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": token, "user": userView(user)})
}

func (h *Handler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err)
		return
	}
	var user model.User
	if err := h.DB.Where("email = ?", strings.ToLower(strings.TrimSpace(input.Email))).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	token, err := h.issueToken(user)
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": userView(user)})
}

func (h *Handler) Me(c *gin.Context) {
	user := currentUser(c)
	c.JSON(http.StatusOK, gin.H{"user": userView(user)})
}

func (h *Handler) UpdateMe(c *gin.Context) {
	user := currentUser(c)
	var input UpdateMeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err)
		return
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	updates := map[string]any{
		"name":       name,
		"avatar_url": strings.TrimSpace(input.AvatarURL),
	}
	if err := h.DB.Model(&model.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		internalError(c, err)
		return
	}
	var updated model.User
	if err := h.DB.First(&updated, user.ID).Error; err != nil {
		internalError(c, err)
		return
	}
	c.Set("user", updated)
	c.JSON(http.StatusOK, gin.H{"user": userView(updated)})
}

func (h *Handler) ListArticles(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	category := strings.TrimSpace(c.Query("category"))
	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("pageSize"), 10)
	if pageSize > 50 {
		pageSize = 50
	}

	cacheKey := fmt.Sprintf("articles:list:%s:%s:%d:%d", query, category, page, pageSize)
	if h.RDB != nil {
		if body, err := h.RDB.Get(c.Request.Context(), cacheKey).Bytes(); err == nil {
			c.Data(http.StatusOK, "application/json", body)
			return
		}
	}

	var rows []model.Article
	tx := h.DB.Model(&model.Article{}).
		Preload("Category").
		Preload("Author").
		Where("status = ?", "published")
	if query != "" {
		pattern := "%" + query + "%"
		tx = tx.Where("title LIKE ? OR summary LIKE ? OR content LIKE ? OR tags LIKE ?", pattern, pattern, pattern, pattern)
	}
	if category != "" {
		tx = tx.Joins("JOIN categories ON categories.id = articles.category_id").Where("categories.slug = ? OR categories.name = ?", category, category)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		internalError(c, err)
		return
	}
	if err := tx.Order("published_at desc, id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		internalError(c, err)
		return
	}

	resp := gin.H{
		"items":    articlesView(rows),
		"page":     page,
		"pageSize": pageSize,
		"total":    total,
		"pages":    int(math.Ceil(float64(total) / float64(pageSize))),
	}
	if h.RDB != nil {
		if data, err := json.Marshal(resp); err == nil {
			_ = h.RDB.Set(c.Request.Context(), cacheKey, data, 45*time.Second).Err()
		}
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) PopularArticles(c *gin.Context) {
	var rows []model.Article
	if err := h.DB.Preload("Category").Preload("Author").Where("status = ?", "published").Order("likes_count desc, favorites_count desc, published_at desc").Limit(6).Find(&rows).Error; err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": articlesView(rows)})
}

func (h *Handler) GetArticleBySlug(c *gin.Context) {
	slug := c.Param("slug")
	var article model.Article
	if err := h.DB.Preload("Category").Preload("Author").Where("slug = ?", slug).First(&article).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	if article.Status != "published" && !h.allowDraft(c) {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	var comments []model.Comment
	_ = h.DB.Preload("User").Where("article_id = ? AND status = ?", article.ID, "published").Order("id desc").Find(&comments).Error
	c.JSON(http.StatusOK, gin.H{
		"article":  articleView(article),
		"comments": commentViews(comments),
		"related":  h.relatedArticles(article),
	})
}

func (h *Handler) relatedArticles(article model.Article) []map[string]any {
	if article.CategoryID == nil {
		return []map[string]any{}
	}
	var rows []model.Article
	_ = h.DB.Preload("Category").Preload("Author").
		Where("status = ? AND category_id = ? AND id <> ?", "published", *article.CategoryID, article.ID).
		Order("published_at desc").Limit(4).Find(&rows).Error
	return articlesView(rows)
}

func (h *Handler) ListComments(c *gin.Context) {
	articleID := parseUint(c.Param("id"))
	var comments []model.Comment
	if err := h.DB.Preload("User").Where("article_id = ? AND status = ?", articleID, "published").Order("id desc").Find(&comments).Error; err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": commentViews(comments)})
}

func (h *Handler) ListCategories(c *gin.Context) {
	var rows []model.Category
	if err := h.DB.Order("name asc").Find(&rows).Error; err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}

func (h *Handler) CreateComment(c *gin.Context) {
	user := currentUser(c)
	articleID := parseUint(c.Param("id"))
	var article model.Article
	if err := h.DB.First(&article, articleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	if article.Status != "published" && user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "article is not published"})
		return
	}

	var input CommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err)
		return
	}
	if strings.TrimSpace(input.Body) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment body is required"})
		return
	}
	comment := model.Comment{ArticleID: articleID, UserID: user.ID, Body: strings.TrimSpace(input.Body), Status: "published"}
	if err := h.DB.Create(&comment).Error; err != nil {
		internalError(c, err)
		return
	}
	_ = h.invalidateArticleCaches(c)
	c.JSON(http.StatusCreated, gin.H{"comment": comment})
}

func (h *Handler) ToggleReaction(c *gin.Context) {
	user := currentUser(c)
	articleID := parseUint(c.Param("id"))
	kind := c.Param("type")
	if kind != "like" && kind != "favorite" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be like or favorite"})
		return
	}

	var article model.Article
	if err := h.DB.First(&article, articleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	if article.Status != "published" && user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "article is not published"})
		return
	}

	var existing model.Reaction
	err := h.DB.Where("user_id = ? AND article_id = ? AND type = ?", user.ID, articleID, kind).First(&existing).Error
	toggledOn := false
	if errors.Is(err, gorm.ErrRecordNotFound) {
		reaction := model.Reaction{UserID: user.ID, ArticleID: articleID, Type: kind}
		if err := h.DB.Create(&reaction).Error; err != nil {
			internalError(c, err)
			return
		}
		toggledOn = true
		h.adjustCounters(articleID, kind, 1)
	} else if err == nil {
		_ = h.DB.Delete(&existing).Error
		h.adjustCounters(articleID, kind, -1)
	} else {
		internalError(c, err)
		return
	}
	_ = h.invalidateArticleCaches(c)
	c.JSON(http.StatusOK, gin.H{"on": toggledOn})
}

func (h *Handler) Library(c *gin.Context) {
	user := currentUser(c)
	var likes []model.Reaction
	var favorites []model.Reaction
	_ = h.DB.Where("user_id = ? AND type = ?", user.ID, "like").Find(&likes).Error
	_ = h.DB.Where("user_id = ? AND type = ?", user.ID, "favorite").Find(&favorites).Error
	c.JSON(http.StatusOK, gin.H{
		"likes":     extractArticleIDs(likes),
		"favorites": extractArticleIDs(favorites),
	})
}

func (h *Handler) AdminListArticles(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	status := strings.TrimSpace(c.DefaultQuery("status", ""))
	var rows []model.Article
	tx := h.DB.Preload("Category").Preload("Author").Model(&model.Article{})
	if q != "" {
		pattern := "%" + q + "%"
		tx = tx.Where("title LIKE ? OR summary LIKE ? OR content LIKE ? OR tags LIKE ?", pattern, pattern, pattern, pattern)
	}
	if status != "" {
		tx = tx.Where("status = ?", status)
	}
	if err := tx.Order("id desc").Find(&rows).Error; err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": articlesView(rows)})
}

func (h *Handler) UpsertArticle(c *gin.Context) {
	var input ArticleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err)
		return
	}
	user := currentUser(c)
	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
		return
	}
	if input.Status == "" {
		input.Status = "draft"
	}
	if input.Title == "" || input.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title and content are required"})
		return
	}

	slug := uniqueSlug(h.DB, input.Title, parseUint(c.Param("id")))
	now := time.Now()
	if c.Param("id") == "" {
		article := model.Article{
			Title:      input.Title,
			Slug:       slug,
			Summary:    input.Summary,
			CoverURL:   input.CoverURL,
			Content:    input.Content,
			Tags:       input.Tags,
			Status:     input.Status,
			CategoryID: categoryIDPtr(input.CategoryID),
			AuthorID:   user.ID,
		}
		if input.Status == "published" {
			article.PublishedAt = &now
		}
		if err := h.DB.Create(&article).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create article"})
			return
		}
		_ = h.invalidateArticleCaches(c)
		c.JSON(http.StatusCreated, gin.H{"article": article})
		return
	}

	id := parseUint(c.Param("id"))
	var article model.Article
	if err := h.DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	updates := map[string]any{
		"title":       input.Title,
		"slug":        slug,
		"summary":     input.Summary,
		"cover_url":   input.CoverURL,
		"content":     input.Content,
		"tags":        input.Tags,
		"status":      input.Status,
		"category_id": categoryIDPtr(input.CategoryID),
	}
	if input.Status == "published" && article.PublishedAt == nil {
		updates["published_at"] = now
	}
	if input.Status != "published" {
		updates["published_at"] = nil
	}
	if err := h.DB.Model(&model.Article{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		internalError(c, err)
		return
	}
	_ = h.invalidateArticleCaches(c)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) DeleteArticle(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("article_id = ?", id).Delete(&model.Reaction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("article_id = ?", id).Delete(&model.Comment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Article{}, id).Error
	}); err != nil {
		internalError(c, err)
		return
	}
	_ = h.invalidateArticleCaches(c)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) AdminListComments(c *gin.Context) {
	var rows []model.Comment
	if err := h.DB.Preload("User").Preload("Article").Order("id desc").Find(&rows).Error; err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": commentViews(rows)})
}

func (h *Handler) ModerateComment(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var payload struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		badRequest(c, err)
		return
	}
	if payload.Status != "published" && payload.Status != "hidden" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be published or hidden"})
		return
	}
	if err := h.DB.Model(&model.Comment{}).Where("id = ?", id).Update("status", payload.Status).Error; err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) ListUsers(c *gin.Context) {
	var rows []model.User
	if err := h.DB.Order("id desc").Find(&rows).Error; err != nil {
		internalError(c, err)
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, u := range rows {
		out = append(out, userView(u))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *Handler) AdminListCategories(c *gin.Context) {
	h.ListCategories(c)
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var input CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err)
		return
	}
	if strings.TrimSpace(input.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	category := model.Category{Name: strings.TrimSpace(input.Name), Slug: slugify(input.Name)}
	if err := h.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category already exists"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"category": category})
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var input CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err)
		return
	}
	if strings.TrimSpace(input.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	updates := map[string]any{"name": strings.TrimSpace(input.Name), "slug": slugify(input.Name)}
	if err := h.DB.Model(&model.Category{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var count int64
	_ = h.DB.Model(&model.Article{}).Where("category_id = ?", id).Count(&count).Error
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category in use"})
		return
	}
	if err := h.DB.Delete(&model.Category{}, id).Error; err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tokenString := strings.TrimPrefix(auth, "Bearer ")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			return h.JWTKey, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		var user model.User
		if err := h.DB.First(&user, claims.UserID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

func (h *Handler) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if currentUser(c).Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin only"})
			return
		}
		c.Next()
	}
}

func (h *Handler) issueToken(user model.User) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.JWTKey)
}

func (h *Handler) adjustCounters(articleID uint, kind string, delta int64) {
	column := "likes_count"
	if kind == "favorite" {
		column = "favorites_count"
	}
	_ = h.DB.Model(&model.Article{}).Where("id = ?", articleID).UpdateColumn(column, gorm.Expr(fmt.Sprintf("%s + ?", column), delta)).Error
}

func (h *Handler) invalidateArticleCaches(c *gin.Context) error {
	if h.RDB == nil {
		return nil
	}
	keys, err := h.RDB.Keys(c.Request.Context(), "articles:list:*").Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return h.RDB.Del(c.Request.Context(), keys...).Err()
	}
	return nil
}

func (h *Handler) allowDraft(c *gin.Context) bool {
	auth := c.GetHeader("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return false
	}
	tokenString := strings.TrimPrefix(auth, "Bearer ")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return h.JWTKey, nil
	})
	if err != nil || !token.Valid {
		return false
	}
	var user model.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil {
		return false
	}
	return user.Role == "admin"
}

func parseInt(v string, fallback int) int {
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

func parseUint(v string) uint {
	n, _ := strconv.ParseUint(v, 10, 64)
	return uint(n)
}

func categoryIDPtr(id uint) *uint {
	if id == 0 {
		return nil
	}
	return &id
}

func categoryIDValue(id *uint) uint {
	if id == nil {
		return 0
	}
	return *id
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '-' || r == '_' || r == '/' || r == '\\':
			if !lastDash && b.Len() > 0 {
				b.WriteRune('-')
				lastDash = true
			}
		default:
			if r > 127 {
				b.WriteRune(r)
				lastDash = false
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return randomSlug()
	}
	return out
}

func uniqueSlug(db *gorm.DB, title string, excludeID uint) string {
	base := slugify(title)
	slug := base
	for i := 1; ; i++ {
		var count int64
		query := db.Model(&model.Article{}).Where("slug = ?", slug)
		if excludeID != 0 {
			query = query.Where("id <> ?", excludeID)
		}
		_ = query.Count(&count).Error
		if count == 0 {
			return slug
		}
		slug = fmt.Sprintf("%s-%d", base, i+1)
	}
}

func randomSlug() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	sum := sha256.Sum256(b[:])
	return hex.EncodeToString(sum[:8])
}

func currentUser(c *gin.Context) model.User {
	v, _ := c.Get("user")
	return v.(model.User)
}

func userView(u model.User) gin.H {
	return gin.H{"id": u.ID, "name": u.Name, "email": u.Email, "role": u.Role, "avatarUrl": u.AvatarURL, "createdAt": u.CreatedAt}
}

func articleView(a model.Article) gin.H {
	return gin.H{
		"id":             a.ID,
		"title":          a.Title,
		"slug":           a.Slug,
		"summary":        a.Summary,
		"coverUrl":       a.CoverURL,
		"content":        a.Content,
		"tags":           splitTags(a.Tags),
		"status":         a.Status,
		"publishedAt":    a.PublishedAt,
		"categoryId":     categoryIDValue(a.CategoryID),
		"category":       a.Category,
		"author":         userView(a.Author),
		"likesCount":     a.LikesCount,
		"favoritesCount": a.FavoritesCount,
	}
}

func articlesView(rows []model.Article) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, articleView(row))
	}
	return out
}

func commentViews(rows []model.Comment) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, gin.H{
			"id":        row.ID,
			"articleId": row.ArticleID,
			"user":      userView(row.User),
			"body":      row.Body,
			"status":    row.Status,
			"createdAt": row.CreatedAt,
		})
	}
	return out
}

func splitTags(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func extractArticleIDs(rows []model.Reaction) []uint {
	out := make([]uint, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ArticleID)
	}
	return out
}

func badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func internalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
