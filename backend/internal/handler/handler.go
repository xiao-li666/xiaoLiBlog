package handler

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"blogapp/backend/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	DB            *gorm.DB
	RDB           *redis.Client
	JWTKey        []byte
	UploadDir     string
	summaryAI     *summaryGenerator
	notifications *notificationHub
	verification  *verificationCodeStore
	mailer        *verificationMailer
}

type notificationHub struct {
	mu   sync.Mutex
	subs map[chan []byte]struct{}
}

func newNotificationHub() *notificationHub {
	return &notificationHub{subs: make(map[chan []byte]struct{})}
}

type verificationCodeEntry struct {
	Code      string
	ExpiresAt time.Time
}

type verificationCodeStore struct {
	mu      sync.Mutex
	entries map[string]verificationCodeEntry
}

func newVerificationCodeStore() *verificationCodeStore {
	return &verificationCodeStore{entries: make(map[string]verificationCodeEntry)}
}

func (s *verificationCodeStore) issue(email, purpose, code string, ttl time.Duration) {
	key := verificationCodeKey(email, purpose)
	s.mu.Lock()
	s.entries[key] = verificationCodeEntry{Code: code, ExpiresAt: time.Now().Add(ttl)}
	s.mu.Unlock()
}

func (s *verificationCodeStore) verify(email, purpose, code string) bool {
	key := verificationCodeKey(email, purpose)
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[key]
	if !ok {
		return false
	}
	if now.After(entry.ExpiresAt) {
		delete(s.entries, key)
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(entry.Code), strings.TrimSpace(code)) {
		return false
	}
	delete(s.entries, key)
	return true
}

func verificationCodeKey(email, purpose string) string {
	return strings.ToLower(strings.TrimSpace(email)) + "|" + strings.ToLower(strings.TrimSpace(purpose))
}

func (h *notificationHub) subscribe() chan []byte {
	ch := make(chan []byte, 8)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *notificationHub) unsubscribe(ch chan []byte) {
	h.mu.Lock()
	if _, ok := h.subs[ch]; ok {
		delete(h.subs, ch)
		close(ch)
	}
	h.mu.Unlock()
}

func (h *notificationHub) publish(payload []byte) {
	h.mu.Lock()
	subs := make([]chan []byte, 0, len(h.subs))
	for ch := range h.subs {
		subs = append(subs, ch)
	}
	h.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- payload:
		default:
		}
	}
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
	Email            string `json:"email"`
	Password         string `json:"password"`
	VerificationCode string `json:"verificationCode"`
}

type RegisterInput struct {
	Name             string `json:"name"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	ConfirmPassword  string `json:"confirmPassword"`
	VerificationCode string `json:"verificationCode"`
}

type VerificationCodeInput struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
}

type UpdateMeInput struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

type CommentInput struct {
	Body string `json:"body"`
}

type CategoryInput struct {
	Name string `json:"name"`
}

type ExternalLinkInput struct {
	Name        string `json:"name"`
	Platform    string `json:"platform"`
	Description string `json:"description"`
	LinkURL     string `json:"linkUrl"`
	QRCodeURL   string `json:"qrCodeUrl"`
	SortOrder   int    `json:"sortOrder"`
	IsActive    *bool  `json:"isActive"`
}

type ArticleSummaryInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type SummaryAIConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}

type summaryGenerator struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewSummaryGenerator(cfg SummaryAIConfig) *summaryGenerator {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	modelName := strings.TrimSpace(cfg.Model)
	if modelName == "" {
		modelName = "deepseek-v4-flash"
	}
	return &summaryGenerator{
		apiKey:  strings.TrimSpace(cfg.APIKey),
		baseURL: baseURL,
		model:   modelName,
		client:  &http.Client{Timeout: timeout},
	}
}

func New(db *gorm.DB, rdb *redis.Client, jwtKey []byte, uploadDir string, mailer *verificationMailer, summaryAI *summaryGenerator) *Handler {
	return &Handler{
		DB:            db,
		RDB:           rdb,
		JWTKey:        jwtKey,
		UploadDir:     uploadDir,
		summaryAI:     summaryAI,
		notifications: newNotificationHub(),
		verification:  newVerificationCodeStore(),
		mailer:        mailer,
	}
}

func (g *summaryGenerator) Generate(ctx context.Context, title, content string) (string, error) {
	if g == nil {
		return "", errors.New("summary generator is not configured")
	}
	payload := map[string]any{
		"model": g.model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是个人博客的摘要生成器。请根据文章标题和正文生成一段适合列表展示的中文摘要。要求：只输出摘要正文，不要编号、引号、前后说明；长度控制在60到120个中文字符之间；保留专业术语；不要提到“本文”“文章”“摘要”等字样。",
			},
			{
				"role":    "user",
				"content": fmt.Sprintf("标题：%s\n正文：\n%s", strings.TrimSpace(title), strings.TrimSpace(content)),
			},
		},
		"temperature": 0.2,
		"max_tokens":  180,
		"stream":      false,
		"thinking": map[string]any{
			"type": "disabled",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("deepseek api error: %s", strings.TrimSpace(string(raw)))
	}

	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return "", err
	}
	if len(decoded.Choices) == 0 {
		return "", errors.New("deepseek returned no summary")
	}
	summary := normalizeSummary(decoded.Choices[0].Message.Content)
	if summary == "" {
		return "", errors.New("deepseek returned empty summary")
	}
	return truncateRunes(summary, 140), nil
}

func normalizeSummary(text string) string {
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	text = strings.TrimSpace(text)
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	text = strings.Trim(text, `"'“”‘’`)
	return text
}

func truncateRunes(text string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	if limit <= 1 {
		return string(runes[:limit])
	}
	return string(runes[:limit-1]) + "…"
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

func (h *Handler) SendVerificationCode(c *gin.Context) {
	var input VerificationCodeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err)
		return
	}
	email := strings.ToLower(strings.TrimSpace(input.Email))
	purpose := strings.ToLower(strings.TrimSpace(input.Purpose))
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}
	if purpose != "login" && purpose != "register" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "purpose must be login or register"})
		return
	}
	code := randomVerificationCode()
	if h.mailer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "smtp is not configured"})
		return
	}
	if err := h.mailer.SendVerificationCode(email, purpose, code); err != nil {
		internalError(c, err)
		return
	}
	if h.verification != nil {
		h.verification.issue(email, purpose, code, 5*time.Minute)
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":        true,
		"expiresIn": 300,
	})
}

func (h *Handler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err)
		return
	}
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if input.Name == "" || email == "" || strings.TrimSpace(input.Password) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, email and password are required"})
		return
	}
	if input.Password != input.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "passwords do not match"})
		return
	}
	if len(input.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 6 characters"})
		return
	}
	var existingCount int64
	if err := h.DB.Model(&model.User{}).Where("email = ?", email).Count(&existingCount).Error; err == nil && existingCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already registered"})
		return
	}
	if h.verification == nil || !h.verification.verify(email, "register", input.VerificationCode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "verification code is invalid or expired"})
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
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
	}
	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already registered"})
		return
	}
	h.createNotification("register", "新用户注册", fmt.Sprintf("%s 刚刚注册了账号", user.Name), &user.ID, nil, nil)
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
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" || strings.TrimSpace(input.Password) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email, password and verification code are required"})
		return
	}
	var user model.User
	if err := h.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if h.verification == nil || !h.verification.verify(email, "login", input.VerificationCode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "verification code is invalid or expired"})
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
	type categoryItem struct {
		ID           uint   `json:"id"`
		Name         string `json:"name"`
		Slug         string `json:"slug"`
		ArticleCount int64  `json:"articleCount" gorm:"column:article_count"`
	}

	var items []categoryItem
	if err := h.DB.Model(&model.Category{}).
		Select("categories.id, categories.name, categories.slug, COUNT(articles.id) AS article_count").
		Joins("LEFT JOIN articles ON articles.category_id = categories.id AND articles.status = ?", "published").
		Group("categories.id, categories.name, categories.slug").
		Order("categories.name asc").
		Scan(&items).Error; err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) ListTags(c *gin.Context) {
	type tagRow struct {
		Tags string
	}
	var rows []tagRow
	if err := h.DB.Model(&model.Article{}).
		Select("tags").
		Where("status = ? AND tags <> ''", "published").
		Scan(&rows).Error; err != nil {
		internalError(c, err)
		return
	}

	counts := make(map[string]int64)
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		seen := make(map[string]struct{})
		for _, tag := range splitTags(row.Tags) {
			if _, ok := seen[tag]; ok {
				continue
			}
			seen[tag] = struct{}{}
			counts[tag]++
		}
	}

	for name, count := range counts {
		items = append(items, gin.H{"name": name, "articleCount": count})
	}
	sort.Slice(items, func(i, j int) bool {
		left := items[i]["articleCount"].(int64)
		right := items[j]["articleCount"].(int64)
		if left != right {
			return left > right
		}
		return items[i]["name"].(string) < items[j]["name"].(string)
	})

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) GenerateArticleSummary(c *gin.Context) {
	var input ArticleSummaryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err)
		return
	}
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}
	if h.summaryAI == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "deepseek summary generator is not configured"})
		return
	}
	summary, err := h.summaryAI.Generate(c.Request.Context(), title, content)
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"summary": summary})
}

func (h *Handler) ListExternalLinks(c *gin.Context) {
	var rows []model.ExternalLink
	if err := h.DB.Where("is_active = ?", true).Order("sort_order asc, id desc").Find(&rows).Error; err != nil {
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
	h.createNotification(
		"comment",
		"新评论",
		fmt.Sprintf("%s 评论了《%s》", user.Name, article.Title),
		&user.ID,
		&articleID,
		&comment.ID,
	)
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
		notificationType := kind
		if kind == "favorite" || kind == "like" {
			h.createNotification(
				notificationType,
				map[string]string{"like": "新点赞", "favorite": "新收藏"}[kind],
				fmt.Sprintf("%s 对《%s》进行了%s", user.Name, article.Title, map[string]string{"like": "点赞", "favorite": "收藏"}[kind]),
				&user.ID,
				&articleID,
				nil,
			)
		}
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

func (h *Handler) AdminNotifications(c *gin.Context) {
	var rows []model.Notification
	if err := h.DB.Preload("User").Preload("Article").Preload("Comment").Order("id desc").Limit(50).Find(&rows).Error; err != nil {
		internalError(c, err)
		return
	}
	counts := map[string]int64{}
	var unreadTotal int64
	_ = h.DB.Model(&model.Notification{}).Where("is_read = ?", false).Count(&unreadTotal).Error
	types := []string{"register", "comment", "like", "favorite"}
	for _, typ := range types {
		var count int64
		_ = h.DB.Model(&model.Notification{}).Where("is_read = ? AND type = ?", false, typ).Count(&count).Error
		counts[typ] = count
	}
	c.JSON(http.StatusOK, gin.H{
		"items":       notificationViews(rows),
		"unreadTotal": unreadTotal,
		"counts":      counts,
	})
}

func (h *Handler) AdminNotificationStream(c *gin.Context) {
	if h.notifications == nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	ch := h.notifications.subscribe()
	defer h.notifications.unsubscribe(ch)

	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()

	writeEvent := func(data string) {
		_, _ = c.Writer.Write([]byte("data: " + data + "\n\n"))
		flusher.Flush()
	}

	writeEvent("connected")

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-heartbeat.C:
			writeEvent("ping")
		case data, ok := <-ch:
			if !ok {
				return
			}
			writeEvent(string(data))
		}
	}
}

func (h *Handler) MarkNotificationsRead(c *gin.Context) {
	var payload struct {
		Types      []string `json:"types"`
		ArticleIDs []uint   `json:"articleIds"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		badRequest(c, err)
		return
	}
	tx := h.DB.Model(&model.Notification{}).Where("is_read = ?", false)
	if len(payload.Types) > 0 {
		tx = tx.Where("type IN ?", payload.Types)
	}
	if len(payload.ArticleIDs) > 0 {
		tx = tx.Where("article_id IN ?", payload.ArticleIDs)
	}
	if err := tx.Update("is_read", true).Error; err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) DeleteNotifications(c *gin.Context) {
	if err := h.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Notification{}).Error; err != nil {
		internalError(c, err)
		return
	}
	if h.notifications != nil {
		h.notifications.publish([]byte("cleared"))
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
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

func (h *Handler) AdminListExternalLinks(c *gin.Context) {
	var rows []model.ExternalLink
	if err := h.DB.Order("sort_order asc, id desc").Find(&rows).Error; err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}

func (h *Handler) UpsertExternalLink(c *gin.Context) {
	var input ExternalLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, err)
		return
	}
	name := strings.TrimSpace(input.Name)
	platform := strings.TrimSpace(input.Platform)
	if name == "" || platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform and name are required"})
		return
	}
	active := true
	if input.IsActive != nil {
		active = *input.IsActive
	}

	id := parseUint(c.Param("id"))
	data := map[string]any{
		"name":        name,
		"platform":    platform,
		"description": strings.TrimSpace(input.Description),
		"link_url":    strings.TrimSpace(input.LinkURL),
		"qr_code_url": strings.TrimSpace(input.QRCodeURL),
		"sort_order":  input.SortOrder,
		"is_active":   active,
	}
	if c.Param("id") == "" {
		item := model.ExternalLink{
			Name:        name,
			Platform:    platform,
			Description: strings.TrimSpace(input.Description),
			LinkURL:     strings.TrimSpace(input.LinkURL),
			QRCodeURL:   strings.TrimSpace(input.QRCodeURL),
			SortOrder:   input.SortOrder,
			IsActive:    active,
		}
		if err := h.DB.Create(&item).Error; err != nil {
			internalError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"item": item})
		return
	}

	res := h.DB.Model(&model.ExternalLink{}).Where("id = ?", id).Updates(data)
	if res.Error != nil {
		internalError(c, res.Error)
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "external link not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) DeleteExternalLink(c *gin.Context) {
	id := parseUint(c.Param("id"))
	res := h.DB.Delete(&model.ExternalLink{}, id)
	if res.Error != nil {
		internalError(c, res.Error)
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "external link not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		tokenString := ""
		if auth != "" && strings.HasPrefix(auth, "Bearer ") {
			tokenString = strings.TrimPrefix(auth, "Bearer ")
		} else {
			tokenString = strings.TrimSpace(c.Query("token"))
		}
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
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
	tokenString := ""
	if auth != "" && strings.HasPrefix(auth, "Bearer ") {
		tokenString = strings.TrimPrefix(auth, "Bearer ")
	} else {
		tokenString = strings.TrimSpace(c.Query("token"))
	}
	if tokenString == "" {
		return false
	}
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

func randomVerificationCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
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

func notificationView(n model.Notification) gin.H {
	var user any
	if n.User.ID != 0 {
		user = userView(n.User)
	}
	var article any
	if n.Article.ID != 0 {
		article = articleView(n.Article)
	}
	var comment any
	if n.Comment.ID != 0 {
		comment = gin.H{
			"id":        n.Comment.ID,
			"articleId": n.Comment.ArticleID,
			"body":      n.Comment.Body,
			"status":    n.Comment.Status,
			"createdAt": n.Comment.CreatedAt,
		}
	}
	return gin.H{
		"id":        n.ID,
		"type":      n.Type,
		"title":     n.Title,
		"content":   n.Content,
		"isRead":    n.IsRead,
		"userId":    n.UserID,
		"articleId": n.ArticleID,
		"commentId": n.CommentID,
		"user":      user,
		"article":   article,
		"comment":   comment,
		"createdAt": n.CreatedAt,
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
		var article any
		if row.Article.ID != 0 {
			article = articleView(row.Article)
		}
		out = append(out, gin.H{
			"id":        row.ID,
			"articleId": row.ArticleID,
			"article":   article,
			"user":      userView(row.User),
			"body":      row.Body,
			"status":    row.Status,
			"createdAt": row.CreatedAt,
		})
	}
	return out
}

func notificationViews(rows []model.Notification) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, notificationView(row))
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

func (h *Handler) createNotification(kind, title, content string, userID, articleID, commentID *uint) {
	if h.DB == nil {
		return
	}
	notification := model.Notification{
		Type:      kind,
		Title:     title,
		Content:   content,
		UserID:    userID,
		ArticleID: articleID,
		CommentID: commentID,
	}
	if err := h.DB.Create(&notification).Error; err == nil && h.notifications != nil {
		h.notifications.publish([]byte(kind))
	}
}

func badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func internalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
