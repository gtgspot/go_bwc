package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// EvidenceStatus represents the current state of evidence
type EvidenceStatus string

const (
	StatusCollected  EvidenceStatus = "COLLECTED"
	StatusProcessing EvidenceStatus = "PROCESSING"
	StatusAnalyzed   EvidenceStatus = "ANALYZED"
	StatusArchived   EvidenceStatus = "ARCHIVED"
	StatusDeleted    EvidenceStatus = "DELETED"
)

// Evidence represents a body-worn camera video file
type Evidence struct {
	ID              string         `json:"id"`
	CaseNumber      string         `json:"case_number"`
	OfficerID       string         `json:"officer_id"`
	OfficerName     string         `json:"officer_name"`
	Timestamp       time.Time      `json:"timestamp"`
	Duration        int            `json:"duration_seconds"`
	Location        string         `json:"location"`
	FilePath        string         `json:"file_path"`
	FileHash        string         `json:"file_hash"`
	FileSize        int64          `json:"file_size"`
	Status          EvidenceStatus `json:"status"`
	Tags            []string       `json:"tags"`
	Notes           string         `json:"notes"`
	ChainOfCustody  []CustodyEntry `json:"chain_of_custody"`
	CreatedAt       time.Time      `json:"created_at"`
	LastModified    time.Time      `json:"last_modified"`
	IntegrityChecks []IntegrityCheck `json:"integrity_checks"`
}

// CustodyEntry represents a chain of custody record
type CustodyEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	FromOfficer  string    `json:"from_officer"`
	ToOfficer    string    `json:"to_officer"`
	Action       string    `json:"action"`
	Purpose      string    `json:"purpose"`
	VerifiedHash string    `json:"verified_hash"`
}

// IntegrityCheck represents a file integrity verification
type IntegrityCheck struct {
	Timestamp  time.Time `json:"timestamp"`
	CheckedBy  string    `json:"checked_by"`
	HashValue  string    `json:"hash_value"`
	IsValid    bool      `json:"is_valid"`
	Notes      string    `json:"notes"`
}

// AuditLog represents system activity logging
type AuditLog struct {
	Timestamp  time.Time `json:"timestamp"`
	UserID     string    `json:"user_id"`
	Action     string    `json:"action"`
	EvidenceID string    `json:"evidence_id"`
	Details    string    `json:"details"`
	IPAddress  string    `json:"ip_address"`
}

// BWCSystem is the main forensic body-worn camera management system
type BWCSystem struct {
	evidenceDB    map[string]*Evidence
	auditLogs     []AuditLog
	storagePath   string
	mu            sync.RWMutex
	auditMu       sync.Mutex
}

// NewBWCSystem creates a new forensic BWC system instance
func NewBWCSystem(storagePath string) (*BWCSystem, error) {
	if err := os.MkdirAll(storagePath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &BWCSystem{
		evidenceDB:  make(map[string]*Evidence),
		auditLogs:   make([]AuditLog, 0),
		storagePath: storagePath,
	}, nil
}

// IngestEvidence ingests a new body-worn camera video file into the system
func (bwc *BWCSystem) IngestEvidence(filePath, caseNumber, officerID, officerName, location string, tags []string) (*Evidence, error) {
	bwc.mu.Lock()
	defer bwc.mu.Unlock()

	// Verify file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// Calculate file hash for integrity
	hash, err := calculateFileHash(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate file hash: %w", err)
	}

	// Generate unique evidence ID
	evidenceID := generateEvidenceID(caseNumber, officerID)

	// Copy file to secure storage
	destPath := filepath.Join(bwc.storagePath, evidenceID+filepath.Ext(filePath))
	if err := copyFile(filePath, destPath); err != nil {
		return nil, fmt.Errorf("failed to copy file to secure storage: %w", err)
	}

	// Create evidence record
	evidence := &Evidence{
		ID:          evidenceID,
		CaseNumber:  caseNumber,
		OfficerID:   officerID,
		OfficerName: officerName,
		Timestamp:   time.Now(),
		Location:    location,
		FilePath:    destPath,
		FileHash:    hash,
		FileSize:    fileInfo.Size(),
		Status:      StatusCollected,
		Tags:        tags,
		ChainOfCustody: []CustodyEntry{
			{
				Timestamp:    time.Now(),
				FromOfficer:  "SYSTEM",
				ToOfficer:    officerID,
				Action:       "INGESTED",
				Purpose:      "Initial evidence collection",
				VerifiedHash: hash,
			},
		},
		CreatedAt:    time.Now(),
		LastModified: time.Now(),
		IntegrityChecks: []IntegrityCheck{
			{
				Timestamp:  time.Now(),
				CheckedBy:  "SYSTEM",
				HashValue:  hash,
				IsValid:    true,
				Notes:      "Initial integrity check",
			},
		},
	}

	bwc.evidenceDB[evidenceID] = evidence

	// Log audit trail
	bwc.logAudit(officerID, "INGEST_EVIDENCE", evidenceID, 
		fmt.Sprintf("Evidence ingested from case %s", caseNumber), "")

	return evidence, nil
}

// VerifyIntegrity verifies the integrity of evidence by comparing file hash
func (bwc *BWCSystem) VerifyIntegrity(evidenceID, checkedBy string) (bool, error) {
	bwc.mu.Lock()
	defer bwc.mu.Unlock()

	evidence, exists := bwc.evidenceDB[evidenceID]
	if !exists {
		return false, errors.New("evidence not found")
	}

	// Calculate current file hash
	currentHash, err := calculateFileHash(evidence.FilePath)
	if err != nil {
		return false, fmt.Errorf("failed to calculate file hash: %w", err)
	}

	isValid := currentHash == evidence.FileHash

	// Record integrity check
	check := IntegrityCheck{
		Timestamp:  time.Now(),
		CheckedBy:  checkedBy,
		HashValue:  currentHash,
		IsValid:    isValid,
		Notes:      "",
	}

	if !isValid {
		check.Notes = "ALERT: File hash mismatch detected - possible tampering"
	}

	evidence.IntegrityChecks = append(evidence.IntegrityChecks, check)
	evidence.LastModified = time.Now()

	// Log audit trail
	status := "PASSED"
	if !isValid {
		status = "FAILED"
	}
	bwc.logAudit(checkedBy, "VERIFY_INTEGRITY", evidenceID,
		fmt.Sprintf("Integrity check %s", status), "")

	return isValid, nil
}

// TransferCustody transfers evidence custody from one officer to another
func (bwc *BWCSystem) TransferCustody(evidenceID, fromOfficer, toOfficer, purpose string) error {
	bwc.mu.Lock()
	defer bwc.mu.Unlock()

	evidence, exists := bwc.evidenceDB[evidenceID]
	if !exists {
		return errors.New("evidence not found")
	}

	// Verify integrity before transfer
	currentHash, err := calculateFileHash(evidence.FilePath)
	if err != nil {
		return fmt.Errorf("failed to verify integrity during transfer: %w", err)
	}

	if currentHash != evidence.FileHash {
		return errors.New("integrity check failed - cannot transfer compromised evidence")
	}

	// Record custody transfer
	entry := CustodyEntry{
		Timestamp:    time.Now(),
		FromOfficer:  fromOfficer,
		ToOfficer:    toOfficer,
		Action:       "TRANSFERRED",
		Purpose:      purpose,
		VerifiedHash: currentHash,
	}

	evidence.ChainOfCustody = append(evidence.ChainOfCustody, entry)
	evidence.LastModified = time.Now()

	// Log audit trail
	bwc.logAudit(fromOfficer, "TRANSFER_CUSTODY", evidenceID,
		fmt.Sprintf("Transferred to %s - %s", toOfficer, purpose), "")

	return nil
}

// UpdateStatus updates the status of evidence
func (bwc *BWCSystem) UpdateStatus(evidenceID, officerID string, newStatus EvidenceStatus, notes string) error {
	bwc.mu.Lock()
	defer bwc.mu.Unlock()

	evidence, exists := bwc.evidenceDB[evidenceID]
	if !exists {
		return errors.New("evidence not found")
	}

	oldStatus := evidence.Status
	evidence.Status = newStatus
	evidence.Notes = notes
	evidence.LastModified = time.Now()

	// Log audit trail
	bwc.logAudit(officerID, "UPDATE_STATUS", evidenceID,
		fmt.Sprintf("Status changed from %s to %s", oldStatus, newStatus), "")

	return nil
}

// SearchEvidence searches for evidence by various criteria
func (bwc *BWCSystem) SearchEvidence(caseNumber, officerID string, status EvidenceStatus) []*Evidence {
	bwc.mu.RLock()
	defer bwc.mu.RUnlock()

	results := make([]*Evidence, 0)

	for _, evidence := range bwc.evidenceDB {
		match := true

		if caseNumber != "" && evidence.CaseNumber != caseNumber {
			match = false
		}
		if officerID != "" && evidence.OfficerID != officerID {
			match = false
		}
		if status != "" && evidence.Status != status {
			match = false
		}

		if match {
			results = append(results, evidence)
		}
	}

	return results
}

// GetEvidence retrieves evidence by ID
func (bwc *BWCSystem) GetEvidence(evidenceID string) (*Evidence, error) {
	bwc.mu.RLock()
	defer bwc.mu.RUnlock()

	evidence, exists := bwc.evidenceDB[evidenceID]
	if !exists {
		return nil, errors.New("evidence not found")
	}

	return evidence, nil
}

// GetChainOfCustody retrieves the complete chain of custody for evidence
func (bwc *BWCSystem) GetChainOfCustody(evidenceID string) ([]CustodyEntry, error) {
	bwc.mu.RLock()
	defer bwc.mu.RUnlock()

	evidence, exists := bwc.evidenceDB[evidenceID]
	if !exists {
		return nil, errors.New("evidence not found")
	}

	return evidence.ChainOfCustody, nil
}

// ExportEvidence exports evidence record to JSON
func (bwc *BWCSystem) ExportEvidence(evidenceID, exportPath string) error {
	bwc.mu.RLock()
	defer bwc.mu.RUnlock()

	evidence, exists := bwc.evidenceDB[evidenceID]
	if !exists {
		return errors.New("evidence not found")
	}

	data, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal evidence: %w", err)
	}

	if err := os.WriteFile(exportPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	return nil
}

// GetAuditLogs retrieves audit logs for a specific evidence or user
func (bwc *BWCSystem) GetAuditLogs(evidenceID, userID string) []AuditLog {
	bwc.auditMu.Lock()
	defer bwc.auditMu.Unlock()

	logs := make([]AuditLog, 0)

	for _, log := range bwc.auditLogs {
		match := true

		if evidenceID != "" && log.EvidenceID != evidenceID {
			match = false
		}
		if userID != "" && log.UserID != userID {
			match = false
		}

		if match {
			logs = append(logs, log)
		}
	}

	return logs
}

// logAudit logs system activity for audit trail
func (bwc *BWCSystem) logAudit(userID, action, evidenceID, details, ipAddress string) {
	bwc.auditMu.Lock()
	defer bwc.auditMu.Unlock()

	log := AuditLog{
		Timestamp:  time.Now(),
		UserID:     userID,
		Action:     action,
		EvidenceID: evidenceID,
		Details:    details,
		IPAddress:  ipAddress,
	}

	bwc.auditLogs = append(bwc.auditLogs, log)
}

// GenerateReport generates a comprehensive report for a case
func (bwc *BWCSystem) GenerateReport(caseNumber string) (string, error) {
	bwc.mu.RLock()
	defer bwc.mu.RUnlock()

	evidence := bwc.SearchEvidence(caseNumber, "", "")
	if len(evidence) == 0 {
		return "", errors.New("no evidence found for case")
	}

	report := fmt.Sprintf("FORENSIC BWC EVIDENCE REPORT\n")
	report += fmt.Sprintf("Case Number: %s\n", caseNumber)
	report += fmt.Sprintf("Report Generated: %s\n", time.Now().Format(time.RFC3339))
	report += fmt.Sprintf("Total Evidence Items: %d\n\n", len(evidence))

	for _, ev := range evidence {
		report += fmt.Sprintf("Evidence ID: %s\n", ev.ID)
		report += fmt.Sprintf("  Officer: %s (%s)\n", ev.OfficerName, ev.OfficerID)
		report += fmt.Sprintf("  Timestamp: %s\n", ev.Timestamp.Format(time.RFC3339))
		report += fmt.Sprintf("  Location: %s\n", ev.Location)
		report += fmt.Sprintf("  Status: %s\n", ev.Status)
		report += fmt.Sprintf("  File Hash: %s\n", ev.FileHash)
		report += fmt.Sprintf("  File Size: %d bytes\n", ev.FileSize)
		report += fmt.Sprintf("  Integrity Checks: %d\n", len(ev.IntegrityChecks))
		report += fmt.Sprintf("  Chain of Custody Entries: %d\n", len(ev.ChainOfCustody))
		report += fmt.Sprintf("\n")
	}

	return report, nil
}

// Utility functions

func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return destFile.Sync()
}

func generateEvidenceID(caseNumber, officerID string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("BWC-%s-%s-%d", caseNumber, officerID, timestamp)
}

// Main demonstration
func main() {
	// Initialize the BWC system
	system, err := NewBWCSystem("./bwc_storage")
	if err != nil {
		fmt.Printf("Error initializing system: %v\n", err)
		return
	}

	fmt.Println("Forensic Body-Worn Camera System Initialized")
	fmt.Println("============================================\n")

	// Example: Create a test video file
	testVideoPath := "./test_video.mp4"
	testFile, err := os.Create(testVideoPath)
	if err != nil {
		fmt.Printf("Error creating test file: %v\n", err)
		return
	}
	testFile.WriteString("This is test video content for demonstration")
	testFile.Close()

	// Ingest evidence
	fmt.Println("1. Ingesting Evidence...")
	evidence, err := system.IngestEvidence(
		testVideoPath,
		"CASE-2025-001",
		"OFF-12345",
		"Officer John Smith",
		"123 Main St, City",
		[]string{"traffic-stop", "incident"},
	)
	if err != nil {
		fmt.Printf("Error ingesting evidence: %v\n", err)
		return
	}
	fmt.Printf("   Evidence ID: %s\n", evidence.ID)
	fmt.Printf("   File Hash: %s\n", evidence.FileHash)
	fmt.Printf("   Status: %s\n\n", evidence.Status)

	// Verify integrity
	fmt.Println("2. Verifying Evidence Integrity...")
	isValid, err := system.VerifyIntegrity(evidence.ID, "OFF-12345")
	if err != nil {
		fmt.Printf("Error verifying integrity: %v\n", err)
		return
	}
	fmt.Printf("   Integrity Check: %v\n\n", isValid)

	// Transfer custody
	fmt.Println("3. Transferring Custody...")
	err = system.TransferCustody(evidence.ID, "OFF-12345", "DET-67890", "Evidence analysis")
	if err != nil {
		fmt.Printf("Error transferring custody: %v\n", err)
		return
	}
	fmt.Printf("   Custody transferred successfully\n\n")

	// Update status
	fmt.Println("4. Updating Evidence Status...")
	err = system.UpdateStatus(evidence.ID, "DET-67890", StatusAnalyzed, "Analysis completed")
	if err != nil {
		fmt.Printf("Error updating status: %v\n", err)
		return
	}
	fmt.Printf("   Status updated to: %s\n\n", StatusAnalyzed)

	// Get chain of custody
	fmt.Println("5. Chain of Custody:")
	custody, err := system.GetChainOfCustody(evidence.ID)
	if err != nil {
		fmt.Printf("Error getting chain of custody: %v\n", err)
		return
	}
	for i, entry := range custody {
		fmt.Printf("   [%d] %s: %s -> %s (%s)\n", i+1, entry.Timestamp.Format("2006-01-02 15:04:05"),
			entry.FromOfficer, entry.ToOfficer, entry.Action)
	}
	fmt.Println()

	// Generate report
	fmt.Println("6. Generating Case Report...")
	report, err := system.GenerateReport("CASE-2025-001")
	if err != nil {
		fmt.Printf("Error generating report: %v\n", err)
		return
	}
	fmt.Println(report)

	// Get audit logs
	fmt.Println("7. Audit Logs:")
	logs := system.GetAuditLogs(evidence.ID, "")
	for i, log := range logs {
		fmt.Printf("   [%d] %s: %s by %s - %s\n", i+1, log.Timestamp.Format("2006-01-02 15:04:05"),
			log.Action, log.UserID, log.Details)
	}

	// Export evidence record
	fmt.Println("\n8. Exporting Evidence Record...")
	err = system.ExportEvidence(evidence.ID, "./evidence_export.json")
	if err != nil {
		fmt.Printf("Error exporting evidence: %v\n", err)
		return
	}
	fmt.Printf("   Evidence exported to: ./evidence_export.json\n")

	fmt.Println("\nDemo completed successfully!")
}
