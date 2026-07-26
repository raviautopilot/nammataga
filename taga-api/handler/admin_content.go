package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"taga-api/config"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ==================== DATA STRUCTURES ====================

// Event struct
type Event struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Date        string `json:"date"`
	Location    string `json:"location"`
	Description string `json:"description"`
	Attendees   int    `json:"attendees"`
	Status      string `json:"status"`
	ImageURL    string `json:"imageUrl,omitempty"`
}

// ResourceDocument struct
type ResourceDocument struct {
	Title       string `json:"title"`
	Year        string `json:"year"`
	URL         string `json:"url"`
	Subcategory string `json:"subcategory,omitempty"`
}

// ResourceCategory struct
type ResourceCategory struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Documents []ResourceDocument `json:"documents"`
}

// GalleryImage struct
type GalleryImage struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Event    string `json:"event"`
	ImageURL string `json:"imageUrl"`
	Date     string `json:"date"`
	Year     int    `json:"year"`
}

// ==================== EVENT MANAGEMENT ====================

// CreateEvent godoc
// @Summary Create a new event
// @Description Create a new event with optional image upload
// @Tags Admin Content
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param title formData string true "Event title"
// @Param date formData string true "Event date (YYYY-MM-DD HH:MM)"
// @Param location formData string false "Event location"
// @Param description formData string false "Event description"
// @Param status formData string false "Event status (upcoming/past/completed)"
// @Param attendees formData int false "Number of attendees"
// @Param image formData file false "Event image file"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/admin/events/create [post]
func CreateEvent(c *gin.Context) {
	title := c.PostForm("title")
	date := c.PostForm("date")
	location := c.PostForm("location")
	description := c.PostForm("description")
	status := c.PostForm("status")
	attendeesStr := c.PostForm("attendees")

	if title == "" || date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title and date are required"})
		return
	}

	if status == "" {
		status = "upcoming"
	}

	attendees := 0
	if attendeesStr != "" {
		attendees, _ = strconv.Atoi(attendeesStr)
	}

	eventsPath := filepath.Join("data", "events.json")
	data, err := os.ReadFile(eventsPath)
	if err != nil && !os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read events"})
		return
	}

	var events []Event
	if len(data) > 0 {
		if err := json.Unmarshal(data, &events); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse events"})
			return
		}
	}

	newEvent := Event{
		ID:          uuid.New().String(),
		Title:       title,
		Date:        date,
		Location:    location,
		Description: description,
		Attendees:   attendees,
		Status:      status,
	}

	// Handle image upload
	file, err := c.FormFile("image")
	if err == nil {
		year := strings.Split(date, "-")[0]
		imageDir := filepath.Join("data", "image", "eventImages", year)
		if err := os.MkdirAll(imageDir, 0755); err == nil {
			imageFilename := sanitizeFilename(title) + filepath.Ext(file.Filename)
			imagePath := filepath.Join(imageDir, imageFilename)
			if err := c.SaveUploadedFile(file, imagePath); err == nil {
				newEvent.ImageURL = fmt.Sprintf("/images/eventImages/%s/%s", year, imageFilename)
			}
		}
	}

	events = append(events, newEvent)

	updatedData, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal events"})
		return
	}

	if err := os.WriteFile(eventsPath, updatedData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Event created successfully",
		"event":   newEvent,
	})
}

// UpdateEvent godoc
// @Summary Update an existing event
// @Description Update an event by ID using multipart form data
// @Tags Admin Content
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param title formData string false "Event title"
// @Param date formData string false "Event date (YYYY-MM-DD HH:MM)"
// @Param location formData string false "Event location"
// @Param description formData string false "Event description"
// @Param status formData string false "Event status (upcoming/ongoing/completed/cancelled)"
// @Param image formData file false "Replacement event image"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/events/{id} [put]
func UpdateEvent(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event ID is required"})
		return
	}

	eventsPath := filepath.Join("data", "events.json")
	data, err := os.ReadFile(eventsPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read events"})
		return
	}

	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse events"})
		return
	}

	foundIdx := -1
	for i, event := range events {
		if event.ID == eventID {
			foundIdx = i
			break
		}
	}

	if foundIdx == -1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Only overwrite fields that are explicitly provided in the form
	if title := c.PostForm("title"); title != "" {
		events[foundIdx].Title = title
	}
	if date := c.PostForm("date"); date != "" {
		events[foundIdx].Date = date
	}
	if location := c.PostForm("location"); location != "" {
		events[foundIdx].Location = location
	}
	if description := c.PostForm("description"); description != "" {
		events[foundIdx].Description = description
	}
	if status := c.PostForm("status"); status != "" {
		events[foundIdx].Status = status
	}

	// Handle optional image replacement
	file, err := c.FormFile("image")
	if err == nil {
		year := strings.Split(events[foundIdx].Date, "-")[0]
		imageDir := filepath.Join("data", "image", "eventImages", year)
		if mkErr := os.MkdirAll(imageDir, 0755); mkErr == nil {
			imageFilename := fmt.Sprintf("event_%s%s", eventID, filepath.Ext(file.Filename))
			imagePath := filepath.Join(imageDir, imageFilename)
			if saveErr := c.SaveUploadedFile(file, imagePath); saveErr == nil {
				events[foundIdx].ImageURL = fmt.Sprintf("/images/eventImages/%s/%s", year, imageFilename)
			}
		}
	}

	updatedData, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal events"})
		return
	}

	if err := os.WriteFile(eventsPath, updatedData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Event updated successfully",
		"event":   events[foundIdx],
	})
}

// DeleteEvent godoc
// @Summary Delete an event
// @Description Delete an event by ID
// @Tags Admin Content
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/events/{id} [delete]
func DeleteEvent(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event ID is required"})
		return
	}

	eventsPath := filepath.Join("data", "events.json")
	data, err := os.ReadFile(eventsPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read events"})
		return
	}

	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse events"})
		return
	}

	found := false
	var deletedEvent Event
	for i, event := range events {
		if event.ID == eventID {
			deletedEvent = event
			events = append(events[:i], events[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Clean up image file if it exists
	if deletedEvent.ImageURL != "" {
		imagePath := strings.TrimPrefix(deletedEvent.ImageURL, "/")
		if _, err := os.Stat(imagePath); err == nil {
			os.Remove(imagePath)
		}
	}

	updatedData, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal events"})
		return
	}

	if err := os.WriteFile(eventsPath, updatedData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}

// ==================== RESOURCE MANAGEMENT ====================

// UploadResourceRequest
type UploadResourceRequest struct {
	CategoryID  string `form:"categoryId" binding:"required"`
	Title       string `form:"title" binding:"required"`
	Year        string `form:"year" binding:"required"`
	Subcategory string `form:"subcategory"`
}

// UploadResource godoc
// @Summary Upload a new resource document
// @Description Upload a PDF document to a specific resource category
// @Tags Admin Content
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param categoryId formData string true "Resource category ID"
// @Param title formData string true "Document title"
// @Param year formData string true "Year (e.g., 2025)"
// @Param subcategory formData string false "Subcategory (for Scheme G.Os: Central or State)"
// @Param file formData file true "PDF file to upload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/resources/upload [post]
func UploadResource(c *gin.Context) {
	var req UploadResourceRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	if !strings.HasSuffix(strings.ToLower(file.Filename), ".pdf") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF files are allowed"})
		return
	}

	targetDir, err := getResourceDirectory(req.CategoryID, req.Subcategory)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	safeFilename := sanitizeFilename(req.Title) + ".pdf"
	filePath := filepath.Join(targetDir, safeFilename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		config.Logger.Error("Failed to save resource file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	if err := addResourceToJSON(req.CategoryID, req.Title, req.Year, req.Subcategory, filePath); err != nil {
		config.Logger.Error("Failed to update resources.json", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update resources database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Resource uploaded successfully",
		"path":    filePath,
	})
}

// DeleteResource godoc
// @Summary Delete a resource document
// @Description Delete a resource document from a category
// @Tags Admin Content
// @Produce json
// @Security BearerAuth
// @Param categoryId path string true "Resource category ID"
// @Param documentTitle path string true "Document title to delete"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/resources/{categoryId}/{documentTitle} [delete]
func DeleteResource(c *gin.Context) {
	categoryID := c.Param("categoryId")
	documentTitle := c.Param("documentTitle")

	resourcesPath := filepath.Join("data", "resources.json")
	data, err := os.ReadFile(resourcesPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read resources"})
		return
	}

	var categories []ResourceCategory
	if err := json.Unmarshal(data, &categories); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse resources"})
		return
	}

	found := false
	for i, cat := range categories {
		if cat.ID == categoryID {
			for j, doc := range cat.Documents {
				if doc.Title == documentTitle {
					// Clean up the physical file if it exists
					if doc.URL != "" {
						filePath := strings.TrimPrefix(doc.URL, "/")
						if _, err := os.Stat(filePath); err == nil {
							os.Remove(filePath)
						}
					}
					categories[i].Documents = append(cat.Documents[:j], cat.Documents[j+1:]...)
					found = true
					break
				}
			}
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	updatedData, err := json.MarshalIndent(categories, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal data"})
		return
	}

	if err := os.WriteFile(resourcesPath, updatedData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write resources"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Resource deleted successfully"})
}

// ==================== GALLERY MANAGEMENT ====================

// UploadGalleryRequest
type UploadGalleryRequest struct {
	Title string `form:"title" binding:"required"`
	Event string `form:"event" binding:"required"`
	Date  string `form:"date" binding:"required"`
	Year  int    `form:"year" binding:"required"`
}

// UploadGalleryImage godoc
// @Summary Upload a gallery image
// @Description Add a new image to the photo gallery
// @Tags Admin Content
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param title formData string true "Image title"
// @Param event formData string true "Event name"
// @Param date formData string true "Date (YYYY-MM-DD)"
// @Param year formData int true "Year"
// @Param image formData file true "Image file (JPEG/PNG)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/gallery/upload [post]
func UploadGalleryImage(c *gin.Context) {
	var req UploadGalleryRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only JPEG and PNG images are allowed"})
		return
	}

	imageDir := filepath.Join("data", "image", "eventImages", fmt.Sprintf("%d", req.Year))
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		config.Logger.Error("Failed to create image directory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	imageFilename := sanitizeFilename(req.Title) + ext
	imagePath := filepath.Join(imageDir, imageFilename)
	if err := c.SaveUploadedFile(file, imagePath); err != nil {
		config.Logger.Error("Failed to save gallery image", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	galleryPath := filepath.Join("data", "gallery.json")
	data, err := os.ReadFile(galleryPath)
	if err != nil && !os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read gallery"})
		return
	}

	var gallery []GalleryImage
	if len(data) > 0 {
		if err := json.Unmarshal(data, &gallery); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse gallery"})
			return
		}
	}

	newImage := GalleryImage{
		ID:       uuid.New().String(),
		Title:    req.Title,
		Event:    req.Event,
		ImageURL: fmt.Sprintf("/images/eventImages/%d/%s", req.Year, imageFilename),
		Date:     req.Date,
		Year:     req.Year,
	}

	gallery = append(gallery, newImage)

	updatedData, err := json.MarshalIndent(gallery, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal gallery"})
		return
	}

	if err := os.WriteFile(galleryPath, updatedData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write gallery"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Gallery image uploaded successfully",
		"image":   newImage,
	})
}

// DeleteGalleryImage godoc
// @Summary Delete a gallery image
// @Description Remove an image from the gallery by ID
// @Tags Admin Content
// @Produce json
// @Security BearerAuth
// @Param id path string true "Gallery image ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/gallery/{id} [delete]
func DeleteGalleryImage(c *gin.Context) {
	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image ID is required"})
		return
	}

	galleryPath := filepath.Join("data", "gallery.json")
	data, err := os.ReadFile(galleryPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read gallery"})
		return
	}

	var gallery []GalleryImage
	if err := json.Unmarshal(data, &gallery); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse gallery"})
		return
	}

	found := false
	var deletedImage GalleryImage
	for i, img := range gallery {
		if img.ID == imageID {
			deletedImage = img
			gallery = append(gallery[:i], gallery[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// Clean up the physical image file if it exists
	if deletedImage.ImageURL != "" {
		imagePath := strings.TrimPrefix(deletedImage.ImageURL, "/")
		if _, err := os.Stat(imagePath); err == nil {
			os.Remove(imagePath)
		}
	}

	updatedData, err := json.MarshalIndent(gallery, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal gallery"})
		return
	}

	if err := os.WriteFile(galleryPath, updatedData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write gallery"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Gallery image deleted successfully"})
}

// ==================== HELPER FUNCTIONS ====================

func getResourceDirectory(categoryID, subcategory string) (string, error) {
	categoryMap := map[string]string{
		"establishment":   "data/docs/Establishment",
		"leave-forms":     "data/docs/Leave Forms & Other Applications",
		"miscellaneous":   "data/docs/Miscellaneous",
		"office-contacts": "data/docs/Office Address & Contacts",
		"pay-gos":         "data/docs/Pay Related G.Os",
		"scheme-gos":      "data/docs/Scheme G.Os/Scheme G.Os 2025-26",
		"taga-membership": "data/docs/TAGA Membership & TBF Application",
		"technical":       "data/docs/Technical",
		"links":           "data/docs",
	}

	dir, ok := categoryMap[categoryID]
	if !ok {
		return "", fmt.Errorf("invalid category ID: %s", categoryID)
	}

	if categoryID == "scheme-gos" && subcategory != "" {
		switch subcategory {
		case "Central":
			dir = filepath.Join(dir, "GOI Schemes 2025-26")
		case "State":
			dir = filepath.Join(dir, "States Schemes 2025-26")
		}
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	return dir, nil
}

func addResourceToJSON(categoryID, title, year, subcategory, filePath string) error {
	resourcesPath := filepath.Join("data", "resources.json")

	data, err := os.ReadFile(resourcesPath)
	if err != nil {
		return fmt.Errorf("failed to read resources.json: %w", err)
	}

	var categories []ResourceCategory
	if err := json.Unmarshal(data, &categories); err != nil {
		return fmt.Errorf("failed to parse resources.json: %w", err)
	}

	webPath := "/" + strings.ReplaceAll(filePath, "\\", "/")
	webPath = "/" + strings.TrimPrefix(strings.TrimPrefix(webPath, "/"), "data/")

	for i, cat := range categories {
		if cat.ID == categoryID {
			newDoc := ResourceDocument{
				Title:       title,
				Year:        year,
				URL:         webPath,
				Subcategory: subcategory,
			}
			categories[i].Documents = append(categories[i].Documents, newDoc)
			break
		}
	}

	updatedData, err := json.MarshalIndent(categories, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal resources: %w", err)
	}

	return os.WriteFile(resourcesPath, updatedData, 0644)
}

func sanitizeFilename(name string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " ", "'", ","}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}
