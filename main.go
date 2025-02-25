package main

import (
    "encoding/json"
    "io/ioutil"
    "math/rand"
    "net/http"
    "os"
    "path/filepath"
    "time"
	"io"
)

// ChatRequest represents the structure of the incoming request for chat
type ChatRequest struct {
    Context string `json:"context"`
}

// ChatResponse represents the structure of the response for chat
type ChatResponse struct {
    Response string `json:"response"`
}

// AssistantRequest represents the structure of the incoming request for assistant
type AssistantRequest struct {
    Title      string `json:"title"`
    RoleSetting string `json:"roleSetting"`
}

// RenameRequest represents the structure of the rename request
type RenameRequest struct {
    CurrentTitle string `json:"currentTitle"`
    NewTitle     string `json:"newTitle"`
}

// AssistantResponse represents the structure of the response for listing assistants
type AssistantResponse struct {
    Title  string `json:"title"`
    Avatar string `json:"avatar"`
}

// RoleSettingResponse represents the structure of the response for role settings
type RoleSettingResponse struct {
    Title       string `json:"title"`
    RoleSetting string `json:"roleSetting"`
}

// randomResponses holds a list of possible responses
var randomResponses = []string{
    "Hello! How can I assist you today?",
    "I'm here to help you with your queries.",
    "What would you like to know?",
    "Feel free to ask me anything!",
}

// CORS middleware to enable CORS
func enableCORS(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, DELETE, PUT, GET")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusNoContent)
        return
    }
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(w, r)

    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var chatRequest ChatRequest
    err := json.NewDecoder(r.Body).Decode(&chatRequest)
    if err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    rand.Seed(time.Now().UnixNano())
    response := randomResponses[rand.Intn(len(randomResponses))]

    chatResponse := ChatResponse{Response: response}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(chatResponse)
}

func createAssistantHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(w, r)

    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var assistantRequest AssistantRequest
    err := json.NewDecoder(r.Body).Decode(&assistantRequest)
    if err != nil || assistantRequest.Title == "" {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    // Create the assistant folder and subfolder
    assistantDir := filepath.Join("assistants", assistantRequest.Title)
    err = os.MkdirAll(filepath.Join(assistantDir, "KnowledgeBase"), os.ModePerm)
    if err != nil {
        http.Error(w, "Failed to create assistant directory", http.StatusInternalServerError)
        return
    }

    // Create the roleSetting.txt file
    roleSettingFile := filepath.Join(assistantDir, "roleSetting.txt")
    err = os.WriteFile(roleSettingFile, []byte(assistantRequest.RoleSetting), os.ModePerm)
    if err != nil {
        http.Error(w, "Failed to create roleSetting file", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"message": "Assistant created successfully"})
}

func deleteAssistantHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(w, r)

    if r.Method != http.MethodDelete {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    title := r.URL.Query().Get("title")
    if title == "" {
        http.Error(w, "Assistant title is required", http.StatusBadRequest)
        return
    }

    assistantDir := filepath.Join("assistants", title)
    err := os.RemoveAll(assistantDir)
    if err != nil {
        http.Error(w, "Failed to delete assistant directory", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Assistant deleted successfully"})
}

func updateAssistantHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(w, r)

    if r.Method != http.MethodPut {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var assistantRequest AssistantRequest
    err := json.NewDecoder(r.Body).Decode(&assistantRequest)
    if err != nil || assistantRequest.Title == "" {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    roleSettingFile := filepath.Join("assistants", assistantRequest.Title, "roleSetting.txt")
    err = os.WriteFile(roleSettingFile, []byte(assistantRequest.RoleSetting), os.ModePerm)
    if err != nil {
        http.Error(w, "Failed to update roleSetting file", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Role setting updated successfully"})
}

func renameAssistantHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(w, r)

    if r.Method != http.MethodPut {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var renameRequest RenameRequest
    err := json.NewDecoder(r.Body).Decode(&renameRequest)
    if err != nil || renameRequest.CurrentTitle == "" || renameRequest.NewTitle == "" {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    currentDir := filepath.Join("assistants", renameRequest.CurrentTitle)
    newDir := filepath.Join("assistants", renameRequest.NewTitle)

    err = os.Rename(currentDir, newDir)
    if err != nil {
        http.Error(w, "Failed to rename assistant directory", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Assistant renamed successfully"})
}

// List Assistants Handler
func listAssistantsHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(w, r)

    if r.Method != http.MethodGet {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    assistantsDir := "assistants"
    files, err := os.ReadDir(assistantsDir)
    if err != nil {
        http.Error(w, "Failed to read assistants directory", http.StatusInternalServerError)
        return
    }

    var assistants []AssistantResponse
    for _, file := range files {
        if file.IsDir() {
            assistants = append(assistants, AssistantResponse{
                Title:  file.Name(),
                Avatar: "🤖", // Static avatar for each assistant
            })
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(assistants)
}


// Get Role Setting Handler
func getRoleSettingHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(w, r)

    if r.Method != http.MethodGet {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    title := r.URL.Query().Get("title")
    if title == "" {
        http.Error(w, "Assistant title is required", http.StatusBadRequest)
        return
    }

    roleSettingFile := filepath.Join("assistants", title, "roleSetting.txt")
    content, err := ioutil.ReadFile(roleSettingFile)
    if err != nil {
        http.Error(w, "Failed to read roleSetting file", http.StatusInternalServerError)
        return
    }

    response := RoleSettingResponse{
        Title:       title,
        RoleSetting: string(content),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// Upload File Handler
func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(w, r)

    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    // Parse the assistant title from the request
    title := r.URL.Query().Get("title")
    if title == "" {
        http.Error(w, "Assistant title is required", http.StatusBadRequest)
        return
    }

    // Create the KnowledgeBase directory for the assistant if it doesn't exist
    knowledgeBaseDir := filepath.Join("assistants", title, "KnowledgeBase")
    err := os.MkdirAll(knowledgeBaseDir, os.ModePerm)
    if err != nil {
        http.Error(w, "Failed to create KnowledgeBase directory", http.StatusInternalServerError)
        return
    }

    // Get the uploaded file
    file, header, err := r.FormFile("file") // Correctly assign to three variables
    if err != nil {
        http.Error(w, "Failed to get file from request", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Create a destination file
    dst, err := os.Create(filepath.Join(knowledgeBaseDir, header.Filename)) // Use header.Filename for the file name
    if err != nil {
        http.Error(w, "Failed to create file", http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    // Copy the uploaded file to the destination
    if _, err := io.Copy(dst, file); err != nil {
        http.Error(w, "Failed to save file", http.StatusInternalServerError)
        return
    }

    // Respond with success message
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "File uploaded successfully"})
}
// DirectoryRequest represents the structure of the incoming request for directory creation
type DirectoryRequest struct {
    Name string `json:"name"`
}

// DirectoryResponse represents the structure of the response for directory creation
type DirectoryResponse struct {
    Message string `json:"message"`
}

// Create Directory Handler
func createDirectoryHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(w, r)

    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var dirRequest DirectoryRequest
    err := json.NewDecoder(r.Body).Decode(&dirRequest)
    if err != nil || dirRequest.Name == "" {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    // Create the directory in the root directory
    err = os.Mkdir(dirRequest.Name, os.ModePerm)
    if err != nil {
        http.Error(w, "Failed to create directory", http.StatusInternalServerError)
        return
    }

    // Respond with success message
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(DirectoryResponse{Message: "Directory created successfully"})
}

// ListDirectoriesResponse represents the structure of the response for listing directories
type ListDirectoriesResponse struct {
    Directories []string `json:"directories"`
}

// List Directories Handler
func listDirectoriesHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(w, r)

    if r.Method != http.MethodGet {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    // Read the root directory
    files, err := os.ReadDir(".") // Read the current directory (root)
    if err != nil {
        http.Error(w, "Failed to read root directory", http.StatusInternalServerError)
        return
    }

    var directories []string
    for _, file := range files {
        if file.IsDir() && file.Name() != "assistants" { // Exclude the "assistants" directory
            directories = append(directories, file.Name())
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ListDirectoriesResponse{Directories: directories})
}
// DeleteDirectoryRequest represents the structure of the incoming request for directory deletion
type DeleteDirectoryRequest struct {
    KnowledgeBaseName string `json:"knowledgeBaseName"`
}

// DeleteDirectoryResponse represents the structure of the response for directory deletion
type DeleteDirectoryResponse struct {
    Message string `json:"message"`
}

// Delete Directory Handler
func deleteDirectoryHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(w, r)

    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var deleteRequest DeleteDirectoryRequest
    err := json.NewDecoder(r.Body).Decode(&deleteRequest)
    if err != nil || deleteRequest.KnowledgeBaseName == "" {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    // Create the path to the directory to be deleted
    dirPath := filepath.Join(".", deleteRequest.KnowledgeBaseName)

    // Remove the directory
    err = os.RemoveAll(dirPath)
    if err != nil {
        http.Error(w, "Failed to delete directory", http.StatusInternalServerError)
        return
    }

    // Respond with success message
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(DeleteDirectoryResponse{Message: "Directory deleted successfully"})
}

// RenameDirectoryRequest represents the structure of the incoming request for directory renaming
type RenameDirectoryRequest struct {
    CurrentName string `json:"currentName"`
    NewName     string `json:"newName"`
}

// Rename Directory Handler
func renameDirectoryHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(w, r)

    if r.Method != http.MethodPut {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var renameRequest RenameDirectoryRequest
    err := json.NewDecoder(r.Body).Decode(&renameRequest)
    if err != nil || renameRequest.CurrentName == "" || renameRequest.NewName == "" {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    currentDir := filepath.Join(".", renameRequest.CurrentName)
    newDir := filepath.Join(".", renameRequest.NewName)

    err = os.Rename(currentDir, newDir)
    if err != nil {
        http.Error(w, "Failed to rename directory", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Directory renamed successfully"})
}
// FileInfoResponse represents the structure of the response for file information
type FileInfoResponse struct {
    Name         string `json:"name"`
    Type         string `json:"type"`
    CreationTime string `json:"creationTime"`
    UpdatedTime  string `json:"updatedTime"`
}

// ListFilesRequest represents the structure of the incoming request for listing files
type ListFilesRequest struct {
    KnowledgeBaseName string `json:"knowledgeBaseName"`
}

// ListFilesResponse represents the structure of the response for listing files
type ListFilesResponse struct {
    Files []FileInfoResponse `json:"files"`
}

// List Files Handler
func listFilesHandler(w http.ResponseWriter, r *http.Request) {
    enableCORS(w, r)

    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var listRequest ListFilesRequest
    err := json.NewDecoder(r.Body).Decode(&listRequest)
    if err != nil || listRequest.KnowledgeBaseName == "" {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    // Define the path to the specified knowledge base directory
    knowledgeBaseDir := filepath.Join(".", listRequest.KnowledgeBaseName)

    // Read the directory
    files, err := os.ReadDir(knowledgeBaseDir)
    if err != nil {
        http.Error(w, "Failed to read directory", http.StatusInternalServerError)
        return
    }

    var fileInfos []FileInfoResponse
    for _, file := range files {
        if !file.IsDir() { // Only process files
            fileInfo, err := file.Info()
            if err != nil {
                continue // Skip if we can't get file info
            }

            // Determine the file type based on the extension
            fileType := "unknown"
            switch filepath.Ext(file.Name()) {
            case ".jpg", ".jpeg", ".png", ".gif":
                fileType = "image"
            case ".mp4", ".mkv", ".avi":
                fileType = "video"
            case ".mp3", ".wav", ".aac":
                fileType = "audio"
            case ".pdf", ".doc", ".docx", ".txt":
                fileType = "document"
            }

            fileInfos = append(fileInfos, FileInfoResponse{
                Name:         file.Name(),
                Type:         fileType,
                CreationTime: fileInfo.ModTime().Format(time.RFC3339), // Use modification time as creation time
                UpdatedTime:  fileInfo.ModTime().Format(time.RFC3339),
            })
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ListFilesResponse{Files: fileInfos})
}
func main() {
    http.HandleFunc("/chat", chatHandler)
    http.HandleFunc("/createAssistant", createAssistantHandler)
    http.HandleFunc("/deleteAssistant", deleteAssistantHandler)
    http.HandleFunc("/updateAssistant", updateAssistantHandler)
    http.HandleFunc("/renameAssistant", renameAssistantHandler) // New endpoint for renaming
    http.HandleFunc("/listAssistants", listAssistantsHandler)   // New endpoint for listing assistants
    http.HandleFunc("/getRoleSetting", getRoleSettingHandler)     // New endpoint for getting role setting
	http.HandleFunc("/upload", uploadFileHandler) 
	http.HandleFunc("/create-knowledgebase", createDirectoryHandler) 
	http.HandleFunc("/list-knowledgebase", listDirectoriesHandler)  
	http.HandleFunc("/delete-knowledgebase", deleteDirectoryHandler)   
	http.HandleFunc("/rename-knowledgebase", renameDirectoryHandler)  
	http.HandleFunc("/list-files-knowledgebase", listFilesHandler)  
    http.ListenAndServe(":8080", nil)
}