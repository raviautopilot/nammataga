package handler

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"taga-api/config"
	"taga-api/model"
	"taga-api/service"
	"taga-api/service/audit"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

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

type ResourceDocument struct {
	Title       string `json:"title"`
	Year        string `json:"year"`
	URL         string `json:"url"`
	Subcategory string `json:"subcategory,omitempty"`
}

type ResourceCategory struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Documents []ResourceDocument `json:"documents"`
}

type AnnouncementRequest struct {
	Title    string `json:"title" binding:"required"`
	Message  string `json:"message" binding:"required"`
	Priority string `json:"priority"`
	SendTo   string `json:"sendTo"`
	District string `json:"district,omitempty"`
}

type AnnouncementResponse struct {
	Message        string `json:"message"`
	Recipients     int    `json:"recipients"`
	SendTo         string `json:"send_to"`
	AnnouncementID string `json:"announcement_id"`
}

type AdminSubscriptionData struct {
	PaymentID        string
	OrderID          string
	Amount           int
	CustomerEmail    string
	SubscriptionID   string
	SubscriptionName string
	MemberName       string
	MemberTagaID     string
	MemberEmail      string
	PaymentType      string
}

type AdminRoomBookingData struct {
	BookingID     string
	PaymentID     string
	OrderID       string
	Amount        int
	CustomerEmail string
	CustomerPhone string
	RoomName      string
	RoomNumber    string
	BedCount      int
	CheckInDate   string
	CheckOutDate  string
	BookerName    string
	BookerTagaID  string
	BookerPhone   string
	BookingFor    string
	GuestDetails  string
	PaymentType   string
}

// CreateEvent godoc
// @Summary Create a new event
// @Tags Admin Content
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param title formData string true "Event title"
// @Param date formData string true "Event date (e.g., 2025-05-15 10:00)"
// @Param location formData string false "Event location"
// @Param description formData string false "Event description"
// @Param status formData string false "Event status (upcoming, completed, etc.)"
// @Param attendees formData integer false "Number of attendees"
// @Param image formData file false "Event banner image file"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/events/create [post]
func CreateEvent(c *gin.Context) {
	title := c.PostForm("title")
	date := c.PostForm("date")
	location := c.PostForm("location")
	description := c.PostForm("description")
	status := c.PostForm("status")
	attendeesStr := c.PostForm("attendees")

	if title == "" || date == "" {
		respondError(c, http.StatusBadRequest, "Title and date are required")
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
	var events []Event
	if err == nil {
		_ = json.Unmarshal(data, &events)
	}

	imageURL := ""
	file, err := c.FormFile("image")
	if err == nil {
		uploadDir := filepath.Join("data", "image", "events")
		_ = os.MkdirAll(uploadDir, 0755)
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
		dst := filepath.Join(uploadDir, filename)
		if err := c.SaveUploadedFile(file, dst); err == nil {
			imageURL = "/api/images/events/" + filename
		}
	}

	newEvent := Event{
		ID:          fmt.Sprintf("%d", time.Now().Unix()),
		Title:       title,
		Date:        date,
		Location:    location,
		Description: description,
		Attendees:   attendees,
		Status:      status,
		ImageURL:    imageURL,
	}

	events = append(events, newEvent)
	updatedData, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save event")
		return
	}

	if err := os.WriteFile(eventsPath, updatedData, 0644); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to write event file")
		return
	}

	// Audit event creation
	_ = audit.Log(c, "admin", getAdminUsername(c),
		audit.ActionCreate, audit.ModuleEvent,
		"event", newEvent.ID,
		fmt.Sprintf("Admin created event '%s' (ID: %s)", newEvent.Title, newEvent.ID),
		nil, audit.Sanitize(newEvent))

	respondOK(c, gin.H{
		"message": "Event created successfully",
		"event":   newEvent,
	})
}

// UpdateEvent godoc
// @Summary Update an existing event
// @Tags Admin Content
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param event body Event true "Updated event details"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/events/{id} [put]
func UpdateEvent(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	eventsPath := filepath.Join("data", "events.json")
	data, err := os.ReadFile(eventsPath)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to read events")
		return
	}

	var events []Event
	_ = json.Unmarshal(data, &events)

	found := false
	var oldState Event
	var newState Event
	for i, e := range events {
		if e.ID == id {
			oldState = e
			if title, ok := req["title"].(string); ok {
				events[i].Title = title
			}
			if date, ok := req["date"].(string); ok {
				events[i].Date = date
			}
			if location, ok := req["location"].(string); ok {
				events[i].Location = location
			}
			if description, ok := req["description"].(string); ok {
				events[i].Description = description
			}
			if status, ok := req["status"].(string); ok {
				events[i].Status = status
			}
			if attendees, ok := req["attendees"].(float64); ok {
				events[i].Attendees = int(attendees)
			}
			newState = events[i]
			found = true
			break
		}
	}

	if !found {
		respondError(c, http.StatusNotFound, "Event not found")
		return
	}

	updatedData, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save event")
		return
	}

	if err := os.WriteFile(eventsPath, updatedData, 0644); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to write event file")
		return
	}

	// Audit event update
	_ = audit.Log(c, "admin", getAdminUsername(c),
		audit.ActionUpdate, audit.ModuleEvent,
		"event", id,
		fmt.Sprintf("Admin updated event %s (ID: %s)", oldState.Title, id),
		audit.Sanitize(oldState), audit.Sanitize(newState))

	respondMessage(c, "Event updated successfully")
}

// DeleteEvent godoc
// @Summary Delete an event
// @Tags Admin Content
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/events/{id} [delete]
func DeleteEvent(c *gin.Context) {
	id := c.Param("id")
	eventsPath := filepath.Join("data", "events.json")
	data, err := os.ReadFile(eventsPath)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to read events")
		return
	}

	var events []Event
	_ = json.Unmarshal(data, &events)

	found := false
	var deletedEvent Event
	var filtered []Event
	for _, e := range events {
		if e.ID == id {
			found = true
			deletedEvent = e
			continue
		}
		filtered = append(filtered, e)
	}

	if !found {
		respondError(c, http.StatusNotFound, "Event not found")
		return
	}

	updatedData, _ := json.MarshalIndent(filtered, "", "  ")
	_ = os.WriteFile(eventsPath, updatedData, 0644)

	// Audit event deletion
	_ = audit.Log(c, "admin", getAdminUsername(c),
		audit.ActionDelete, audit.ModuleEvent,
		"event", id,
		fmt.Sprintf("Admin deleted event '%s' (ID: %s)", deletedEvent.Title, id),
		audit.Sanitize(deletedEvent), nil)

	respondMessage(c, "Event deleted successfully")
}

// UploadResource godoc
// @Summary Upload resource document
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
// @Router /api/admin/resources/upload [post]
func UploadResource(c *gin.Context) {
	categoryID := c.PostForm("categoryId")
	title := c.PostForm("title")
	year := c.PostForm("year")
	subcategory := c.PostForm("subcategory")

	if categoryID == "" || title == "" || year == "" {
		respondError(c, http.StatusBadRequest, "categoryId, title, and year are required")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, "PDF file is required")
		return
	}

	uploadDir := filepath.Join("data", "docs")
	_ = os.MkdirAll(uploadDir, 0755)
	filename := fmt.Sprintf("%s_%s", categoryID, filepath.Base(file.Filename))
	dst := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save PDF file")
		return
	}

	docURL := "/docs/" + filename
	resourcesPath := filepath.Join("data", "resources.json")
	data, _ := os.ReadFile(resourcesPath)

	var categories []ResourceCategory
	_ = json.Unmarshal(data, &categories)

	found := false
	newDoc := ResourceDocument{
		Title:       title,
		Year:        year,
		URL:         docURL,
		Subcategory: subcategory,
	}

	for i, cat := range categories {
		if cat.ID == categoryID {
			categories[i].Documents = append(categories[i].Documents, newDoc)
			found = true
			break
		}
	}

	if !found {
		respondError(c, http.StatusNotFound, "Resource category not found")
		return
	}

	updatedData, _ := json.MarshalIndent(categories, "", "  ")
	_ = os.WriteFile(resourcesPath, updatedData, 0644)

	// Audit resource creation
	_ = audit.Log(c, "admin", getAdminUsername(c),
		audit.ActionCreate, audit.ModuleResource,
		"resource", title,
		fmt.Sprintf("Admin uploaded resource document '%s' (category: %s)", title, categoryID),
		nil, audit.Sanitize(newDoc))

	respondOK(c, gin.H{
		"message":  "Resource uploaded successfully",
		"document": newDoc,
	})
}

// DeleteResource godoc
// @Summary Delete resource document
// @Tags Admin Content
// @Produce json
// @Security BearerAuth
// @Param categoryId path string true "Resource category ID"
// @Param documentTitle path string true "Document title to delete"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/resources/{categoryId}/{documentTitle} [delete]
func DeleteResource(c *gin.Context) {
	categoryID := c.Param("categoryId")
	documentTitle := c.Param("documentTitle")

	resourcesPath := filepath.Join("data", "resources.json")
	data, err := os.ReadFile(resourcesPath)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to read resources")
		return
	}

	var categories []ResourceCategory
	_ = json.Unmarshal(data, &categories)

	found := false
	var deletedDoc ResourceDocument
	for i, cat := range categories {
		if cat.ID == categoryID {
			var filteredDocs []ResourceDocument
			for _, doc := range cat.Documents {
				if strings.EqualFold(doc.Title, documentTitle) {
					found = true
					deletedDoc = doc
					continue
				}
				filteredDocs = append(filteredDocs, doc)
			}
			categories[i].Documents = filteredDocs
			break
		}
	}

	if !found {
		if categoryID == "links" {
			if deleted, err := deleteExternalLinkFromCSV(c, documentTitle); err == nil && deleted {
				respondMessage(c, "External link deleted successfully")
				return
			}
		}
		respondError(c, http.StatusNotFound, "Resource document not found")
		return
	}

	updatedData, _ := json.MarshalIndent(categories, "", "  ")
	_ = os.WriteFile(resourcesPath, updatedData, 0644)

	// Audit resource deletion
	_ = audit.Log(c, "admin", getAdminUsername(c),
		audit.ActionDelete, audit.ModuleResource,
		"resource", documentTitle,
		fmt.Sprintf("Admin deleted resource document '%s' from category %s", documentTitle, categoryID),
		audit.Sanitize(deletedDoc), nil)

	respondMessage(c, "Resource document deleted successfully")
}

var externalLinksMutex sync.Mutex

type AddExternalLinkRequest struct {
	Title string `json:"title" binding:"required"`
	URL   string `json:"url" binding:"required"`
}

type DeleteExternalLinkRequest struct {
	Title string `json:"title"`
}

// AddExternalLink godoc
// @Summary Add a new external link to the External Links CSV
// @Tags Admin Content
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param link body AddExternalLinkRequest true "External link details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/resources/external-links [post]
func AddExternalLink(c *gin.Context) {
	var req AddExternalLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Title and URL are required")
		return
	}

	title := strings.TrimSpace(req.Title)
	linkURL := strings.TrimSpace(req.URL)

	if title == "" || linkURL == "" {
		respondError(c, http.StatusBadRequest, "Title and URL cannot be empty")
		return
	}

	if !strings.HasPrefix(linkURL, "http://") && !strings.HasPrefix(linkURL, "https://") {
		linkURL = "https://" + linkURL
	}

	externalLinksMutex.Lock()
	defer externalLinksMutex.Unlock()

	filePath := filepath.Join("data", "docs", "External Links.csv")
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to create directory")
		return
	}

	var records [][]string
	if fileBytes, err := os.ReadFile(filePath); err == nil {
		reader := csv.NewReader(bytes.NewReader(fileBytes))
		reader.FieldsPerRecord = -1
		records, _ = reader.ReadAll()
	}

	if len(records) == 0 {
		records = append(records, []string{"Title", "URL"})
	}

	// Check for duplicate title
	for i, row := range records {
		if i == 0 {
			continue // skip header
		}
		if len(row) > 0 && strings.EqualFold(strings.TrimSpace(row[0]), title) {
			respondError(c, http.StatusConflict, fmt.Sprintf("External link with title '%s' already exists", title))
			return
		}
	}

	records = append(records, []string{title, linkURL})

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.WriteAll(records); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to write CSV")
		return
	}
	writer.Flush()

	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save CSV file")
		return
	}

	newLink := gin.H{"title": title, "url": linkURL}

	// Audit resource creation
	_ = audit.Log(c, "admin", getAdminUsername(c),
		audit.ActionCreate, audit.ModuleResource,
		"external_link", title,
		fmt.Sprintf("Admin added external link '%s' (%s)", title, linkURL),
		nil, audit.Sanitize(newLink))

	respondOK(c, gin.H{
		"message": "External link added successfully",
		"link":    newLink,
	})
}

// DeleteExternalLink godoc
// @Summary Delete an external link from the External Links CSV
// @Tags Admin Content
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param title path string false "Link title to delete"
// @Param title query string false "Link title to delete (query param)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/resources/external-links/{title} [delete]
func DeleteExternalLink(c *gin.Context) {
	title := c.Param("title")
	if title == "" {
		title = c.Query("title")
	}
	if title == "" {
		var req DeleteExternalLinkRequest
		if err := c.ShouldBindJSON(&req); err == nil {
			title = req.Title
		}
	}

	title = strings.TrimSpace(title)
	if title == "" {
		respondError(c, http.StatusBadRequest, "Title is required")
		return
	}

	deleted, err := deleteExternalLinkFromCSV(c, title)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to delete external link: "+err.Error())
		return
	}
	if !deleted {
		respondError(c, http.StatusNotFound, fmt.Sprintf("External link '%s' not found", title))
		return
	}

	respondMessage(c, "External link deleted successfully")
}

func deleteExternalLinkFromCSV(c *gin.Context, title string) (bool, error) {
	title = strings.TrimSpace(title)
	if unescaped, err := url.QueryUnescape(title); err == nil {
		title = strings.TrimSpace(unescaped)
	}
	if title == "" {
		return false, fmt.Errorf("empty title")
	}

	externalLinksMutex.Lock()
	defer externalLinksMutex.Unlock()

	filePath := filepath.Join("data", "docs", "External Links.csv")
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	reader := csv.NewReader(bytes.NewReader(fileBytes))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return false, err
	}

	found := false
	var deletedURL string
	var updatedRecords [][]string

	for i, row := range records {
		if i == 0 {
			updatedRecords = append(updatedRecords, row)
			continue
		}

		if len(row) > 0 && strings.EqualFold(strings.TrimSpace(row[0]), title) {
			found = true
			if len(row) > 1 {
				deletedURL = strings.TrimSpace(row[1])
			}
			continue
		}

		updatedRecords = append(updatedRecords, row)
	}

	if !found {
		return false, nil
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.WriteAll(updatedRecords); err != nil {
		return false, err
	}
	writer.Flush()

	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return false, err
	}

	// Audit external link deletion
	_ = audit.Log(c, "admin", getAdminUsername(c),
		audit.ActionDelete, audit.ModuleResource,
		"external_link", title,
		fmt.Sprintf("Admin deleted external link '%s'", title),
		audit.Sanitize(gin.H{"title": title, "url": deletedURL}), nil)

	return true, nil
}

// UploadGalleryImage godoc
// @Summary Upload gallery image
// @Tags Admin Content
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param title formData string true "Image title"
// @Param event formData string false "Event name"
// @Param date formData string false "Event date (YYYY-MM-DD)"
// @Param year formData integer true "Year (e.g., 2025)"
// @Param image formData file true "Gallery image file"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/gallery/upload [post]
func UploadGalleryImage(c *gin.Context) {
	title := c.PostForm("title")
	event := c.PostForm("event")
	date := c.PostForm("date")
	yearStr := c.PostForm("year")

	if title == "" {
		respondError(c, http.StatusBadRequest, "Title is required")
		return
	}

	year := time.Now().Year()
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	file, err := c.FormFile("image")
	if err != nil {
		respondError(c, http.StatusBadRequest, "Image file is required")
		return
	}

	uploadDir := filepath.Join("data", "image", "gallery")
	_ = os.MkdirAll(uploadDir, 0755)
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
	dst := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save image")
		return
	}

	imageURL := "/api/images/gallery/" + filename
	galleryPath := filepath.Join("data", "gallery.json")
	data, _ := os.ReadFile(galleryPath)

	var images []model.GalleryImage
	_ = json.Unmarshal(data, &images)

	newImg := model.GalleryImage{
		ID:       fmt.Sprintf("%d", time.Now().Unix()),
		Title:    title,
		Event:    event,
		ImageURL: imageURL,
		Date:     date,
		Year:     year,
	}

	images = append(images, newImg)
	updatedData, _ := json.MarshalIndent(images, "", "  ")
	_ = os.WriteFile(galleryPath, updatedData, 0644)

	// Audit gallery upload
	_ = audit.Log(c, "admin", getAdminUsername(c),
		audit.ActionCreate, audit.ModuleGallery,
		"gallery", newImg.ID,
		fmt.Sprintf("Admin uploaded gallery image '%s' (ID: %s)", newImg.Title, newImg.ID),
		nil, audit.Sanitize(newImg))

	respondOK(c, gin.H{
		"message": "Gallery image uploaded successfully",
		"image":   newImg,
	})
}

// DeleteGalleryImage godoc
// @Summary Delete gallery image
// @Tags Admin Content
// @Produce json
// @Security BearerAuth
// @Param id path string true "Gallery image ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/gallery/{id} [delete]
func DeleteGalleryImage(c *gin.Context) {
	id := c.Param("id")
	galleryPath := filepath.Join("data", "gallery.json")
	data, err := os.ReadFile(galleryPath)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to read gallery")
		return
	}

	var images []model.GalleryImage
	_ = json.Unmarshal(data, &images)

	found := false
	var deletedImg model.GalleryImage
	var filtered []model.GalleryImage
	for _, img := range images {
		if img.ID == id {
			found = true
			deletedImg = img
			continue
		}
		filtered = append(filtered, img)
	}

	if !found {
		respondError(c, http.StatusNotFound, "Gallery image not found")
		return
	}

	updatedData, _ := json.MarshalIndent(filtered, "", "  ")
	_ = os.WriteFile(galleryPath, updatedData, 0644)

	// Audit gallery deletion
	_ = audit.Log(c, "admin", getAdminUsername(c),
		audit.ActionDelete, audit.ModuleGallery,
		"gallery", id,
		fmt.Sprintf("Admin deleted gallery image '%s' (ID: %s)", deletedImg.Title, id),
		audit.Sanitize(deletedImg), nil)

	respondMessage(c, "Gallery image deleted successfully")
}

// SendAnnouncement godoc
// @Summary Send announcement to members
// @Tags Admin Announcements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param announcement body AnnouncementRequest true "Announcement details"
// @Success 200 {object} AnnouncementResponse
// @Router /api/admin/announcements/send [post]
func SendAnnouncement(c *gin.Context) {
	var req AnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if req.Priority == "" {
		req.Priority = "normal"
	}
	if req.SendTo == "" {
		req.SendTo = "all"
	}

	adminEmail, exists := c.Get("username")
	if !exists {
		adminEmail = config.GetConfig().AdminEmail
	}

	members, err := readExistingMembers()
	if err != nil {
		config.Logger.Error("Failed to read members", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "Failed to read members")
		return
	}

	if len(members) == 0 {
		respondOK(c, gin.H{
			"message":    "No members found in the system",
			"recipients": 0,
			"send_to":    req.SendTo,
		})
		return
	}

	var recipients []map[string]interface{}
	sendToLower := strings.ToLower(strings.TrimSpace(req.SendTo))

	for _, member := range members {
		email, ok := member["emailId"].(string)
		if !ok || email == "" {
			continue
		}

		memberID, _ := member["id"].(string)
		workingDistrict := getString(member, "working_district")
		paymentStatus := getString(member, "payment_status")
		include := false

		switch sendToLower {
		case "all", "all members":
			include = true
		case "paid", "paid members", "paid members only":
			if strings.EqualFold(paymentStatus, "paid") {
				include = true
			}
		case "unpaid", "unpaid members":
			if strings.EqualFold(paymentStatus, "unpaid") {
				include = true
			}
		case "district", "district members":
			if req.District != "" && strings.EqualFold(workingDistrict, req.District) {
				include = true
			}
		default:
			include = true
		}

		if include {
			recipients = append(recipients, map[string]interface{}{
				"email":            email,
				"id":               memberID,
				"working_district": workingDistrict,
				"payment_status":   paymentStatus,
			})
		}
	}

	if len(recipients) == 0 {
		respondOK(c, gin.H{
			"message":    fmt.Sprintf("No recipients found for filter: %s", req.SendTo),
			"recipients": 0,
			"send_to":    req.SendTo,
		})
		return
	}

	announcement := model.Notification{
		ID:         uuid.New().String(),
		Title:      req.Title,
		Message:    req.Message,
		Priority:   req.Priority,
		SendTo:     req.SendTo,
		SentBy:     adminEmail.(string),
		SentAt:     time.Now(),
		Recipients: len(recipients),
		District:   req.District,
	}

	_ = saveAnnouncementToFile(announcement)

	for _, recipient := range recipients {
		memberNotification := model.MemberNotification{
			ID:             uuid.New().String(),
			MemberID:       recipient["id"].(string),
			MemberEmail:    recipient["email"].(string),
			NotificationID: announcement.ID,
			Title:          req.Title,
			Message:        req.Message,
			Priority:       req.Priority,
			IsRead:         false,
			CreatedAt:      time.Now(),
		}
		_ = saveMemberNotificationToFile(memberNotification)
	}

	go func() {
		for _, recipient := range recipients {
			subject := formatAnnouncementSubject(req.Title, req.Priority)
			body := buildAnnouncementEmailContent(req.Title, req.Message, req.Priority)
			_ = sendEmail(recipient["email"].(string), subject, body)
		}
	}()

	// Audit announcement send
	_ = audit.Log(c, "admin", getAdminUsername(c),
		audit.ActionCreate, audit.ModuleResource,
		"announcement", announcement.ID,
		fmt.Sprintf("Admin sent announcement '%s' to %d recipients via %s", announcement.Title, len(recipients), announcement.SendTo),
		nil, audit.Sanitize(announcement))

	respondOK(c, AnnouncementResponse{
		Message:        fmt.Sprintf("Announcement sent successfully to %d recipients", len(recipients)),
		Recipients:     len(recipients),
		SendTo:         req.SendTo,
		AnnouncementID: announcement.ID,
	})
}

// HandleSendRenewalReminders godoc
// @Summary Manually trigger renewal reminders
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Router /api/admin/send-renewal-reminders [post]
func HandleSendRenewalReminders(c *gin.Context) {
	if err := service.SendRemindersIfDue(); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit manual renewal reminder trigger
	_ = audit.Log(c, "admin", getAdminUsername(c),
		audit.ActionUpdate, audit.ModuleMember,
		"reminders", "renewal",
		"Admin manually triggered renewal reminders email process",
		nil, nil)

	respondMessage(c, "Reminders processed")
}

// File Storage Helpers for Notifications
func getAnnouncementsFilePath() string {
	dir := "data/announcements"
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "announcements.json")
}

func getMemberNotificationsFilePath() string {
	dir := "data/notifications"
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "member_notifications.json")
}

func saveAnnouncementToFile(announcement model.Notification) error {
	filePath := getAnnouncementsFilePath()
	var announcements []model.Notification
	data, err := os.ReadFile(filePath)
	if err == nil && len(data) > 0 {
		_ = json.Unmarshal(data, &announcements)
	}

	announcements = append([]model.Notification{announcement}, announcements...)
	updatedData, err := json.MarshalIndent(announcements, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, updatedData, 0644)
}

func saveMemberNotificationToFile(notification model.MemberNotification) error {
	filePath := getMemberNotificationsFilePath()
	var notifications []model.MemberNotification
	data, err := os.ReadFile(filePath)
	if err == nil && len(data) > 0 {
		_ = json.Unmarshal(data, &notifications)
	}

	notifications = append(notifications, notification)
	updatedData, err := json.MarshalIndent(notifications, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, updatedData, 0644)
}

func formatAnnouncementSubject(title, priority string) string {
	switch priority {
	case "urgent":
		return "🚨 URGENT: " + title
	case "high":
		return "⚠️ IMPORTANT: " + title
	default:
		return "📢 " + title
	}
}

func buildAnnouncementEmailContent(title, message, priority string) string {
	var badgeText, headerGradient string
	switch priority {
	case "urgent":
		badgeText = "URGENT ANNOUNCEMENT"
		headerGradient = "linear-gradient(135deg, #dc2626 0%, #991b1b 100%)"
	case "high":
		badgeText = "IMPORTANT NOTICE"
		headerGradient = "linear-gradient(135deg, #d97706 0%, #b45309 100%)"
	default:
		badgeText = "ANNOUNCEMENT"
		headerGradient = "linear-gradient(135deg, #065f46 0%, #047857 100%)"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #1f2937; background-color: #f3f4f6; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06); border: 1px solid #e5e7eb;">
        
        <!-- Header -->
        <div style="background: %s; color: white; padding: 32px 24px; text-align: center;">
            <div style="display: inline-block; padding: 6px 14px; border-radius: 20px; font-size: 12px; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; background: rgba(255, 255, 255, 0.2); margin-bottom: 12px;">
                %s
            </div>
            <h1 style="margin: 0; font-size: 24px; font-weight: 800;">TAGA Official Broadcast</h1>
        </div>

        <!-- Body -->
        <div style="padding: 32px 24px;">
            <h2 style="margin: 0 0 16px 0; font-size: 20px; font-weight: 700; color: #0f172a; line-height: 1.4;">%s</h2>
            <div style="font-size: 15px; color: #374151; line-height: 1.7; white-space: pre-line; background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 20px;">
%s
            </div>
        </div>

        <!-- Footer -->
        <div style="background: #f8fafc; padding: 24px; text-align: center; border-top: 1px solid #e2e8f0;">
            <p style="margin: 0 0 6px 0; font-size: 13px; font-weight: 600; color: #475569;">Tamil Nadu Agricultural Graduates Association</p>
            <p style="margin: 0; font-size: 12px; color: #94a3b8;">TAGA Towers, Chennai &bull; &copy; 2026 TAGA. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, html.EscapeString(title), headerGradient, badgeText, html.EscapeString(title), html.EscapeString(message))
}

func sendAdminSubscriptionEmail(data AdminSubscriptionData) error {
	cfg := config.GetConfig()
	adminEmail := cfg.AdminEmail
	if adminEmail == "" {
		return nil
	}

	amountInRupees := float64(data.Amount) / 100
	subject := fmt.Sprintf("💰 New Subscription Payment: %s - ₹%.2f", data.SubscriptionName, amountInRupees)
	body := buildSubscriptionEmailBody(data)
	return sendEmail(adminEmail, subject, body)
}

func sendAdminRoomBookingEmail(data AdminRoomBookingData) error {
	cfg := config.GetConfig()
	adminEmail := cfg.AdminEmail
	if adminEmail == "" {
		return nil
	}

	amountInRupees := float64(data.Amount) / 100
	subject := fmt.Sprintf("🏨 New Room Booking: %s - ₹%.2f", data.RoomName, amountInRupees)
	body := buildRoomBookingEmailBody(data)
	return sendEmail(adminEmail, subject, body)
}
