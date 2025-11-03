package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestSystem creates a temporary BWC system for testing
func setupTestSystem(t *testing.T) (*BWCSystem, string, func()) {
	tmpDir, err := os.MkdirTemp("", "bwc_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	system, err := NewBWCSystem(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create BWC system: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return system, tmpDir, cleanup
}

// createTestFile creates a temporary test video file
func createTestFile(t *testing.T, tmpDir string) string {
	testFile := filepath.Join(tmpDir, "test_video.mp4")
	content := []byte("This is test video content for BWC system testing")
	
	if err := os.WriteFile(testFile, content, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	return testFile
}

func TestNewBWCSystem(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bwc_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	system, err := NewBWCSystem(tmpDir)
	if err != nil {
		t.Errorf("NewBWCSystem failed: %v", err)
	}

	if system == nil {
		t.Error("NewBWCSystem returned nil system")
	}

	if system.storagePath != tmpDir {
		t.Errorf("Expected storage path %s, got %s", tmpDir, system.storagePath)
	}

	if system.evidenceDB == nil {
		t.Error("Evidence database not initialized")
	}

	if system.auditLogs == nil {
		t.Error("Audit logs not initialized")
	}
}

func TestIngestEvidence(t *testing.T) {
	system, tmpDir, cleanup := setupTestSystem(t)
	defer cleanup()

	testFile := createTestFile(t, tmpDir)

	evidence, err := system.IngestEvidence(
		testFile,
		"CASE-TEST-001",
		"OFF-123",
		"Officer Test",
		"Test Location",
		[]string{"test", "demo"},
	)

	if err != nil {
		t.Fatalf("IngestEvidence failed: %v", err)
	}

	if evidence == nil {
		t.Fatal("IngestEvidence returned nil evidence")
	}

	// Verify evidence fields
	if evidence.CaseNumber != "CASE-TEST-001" {
		t.Errorf("Expected case number CASE-TEST-001, got %s", evidence.CaseNumber)
	}

	if evidence.OfficerID != "OFF-123" {
		t.Errorf("Expected officer ID OFF-123, got %s", evidence.OfficerID)
	}

	if evidence.Status != StatusCollected {
		t.Errorf("Expected status %s, got %s", StatusCollected, evidence.Status)
	}

	if len(evidence.FileHash) != 64 { // SHA-256 produces 64 hex characters
		t.Errorf("Expected hash length 64, got %d", len(evidence.FileHash))
	}

	if len(evidence.ChainOfCustody) != 1 {
		t.Errorf("Expected 1 custody entry, got %d", len(evidence.ChainOfCustody))
	}

	if len(evidence.IntegrityChecks) != 1 {
		t.Errorf("Expected 1 integrity check, got %d", len(evidence.IntegrityChecks))
	}

	// Verify file was copied to secure storage
	if _, err := os.Stat(evidence.FilePath); os.IsNotExist(err) {
		t.Error("Evidence file not copied to secure storage")
	}
}

func TestVerifyIntegrity(t *testing.T) {
	system, tmpDir, cleanup := setupTestSystem(t)
	defer cleanup()

	testFile := createTestFile(t, tmpDir)

	evidence, err := system.IngestEvidence(
		testFile,
		"CASE-TEST-002",
		"OFF-123",
		"Officer Test",
		"Test Location",
		[]string{"test"},
	)

	if err != nil {
		t.Fatalf("IngestEvidence failed: %v", err)
	}

	// Test successful integrity check
	isValid, err := system.VerifyIntegrity(evidence.ID, "OFF-123")
	if err != nil {
		t.Fatalf("VerifyIntegrity failed: %v", err)
	}

	if !isValid {
		t.Error("Expected integrity check to pass")
	}

	// Verify integrity check was recorded
	updatedEvidence, _ := system.GetEvidence(evidence.ID)
	if len(updatedEvidence.IntegrityChecks) != 2 {
		t.Errorf("Expected 2 integrity checks, got %d", len(updatedEvidence.IntegrityChecks))
	}

	// Test integrity check after file modification
	file, err := os.OpenFile(evidence.FilePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatalf("Failed to open evidence file: %v", err)
	}
	file.WriteString("TAMPERED")
	file.Close()

	isValid, err = system.VerifyIntegrity(evidence.ID, "OFF-123")
	if err != nil {
		t.Fatalf("VerifyIntegrity failed: %v", err)
	}

	if isValid {
		t.Error("Expected integrity check to fail after tampering")
	}
}

func TestTransferCustody(t *testing.T) {
	system, tmpDir, cleanup := setupTestSystem(t)
	defer cleanup()

	testFile := createTestFile(t, tmpDir)

	evidence, err := system.IngestEvidence(
		testFile,
		"CASE-TEST-003",
		"OFF-123",
		"Officer Test",
		"Test Location",
		[]string{"test"},
	)

	if err != nil {
		t.Fatalf("IngestEvidence failed: %v", err)
	}

	// Test successful custody transfer
	err = system.TransferCustody(evidence.ID, "OFF-123", "DET-456", "Analysis")
	if err != nil {
		t.Fatalf("TransferCustody failed: %v", err)
	}

	// Verify custody entry was added
	updatedEvidence, _ := system.GetEvidence(evidence.ID)
	if len(updatedEvidence.ChainOfCustody) != 2 {
		t.Errorf("Expected 2 custody entries, got %d", len(updatedEvidence.ChainOfCustody))
	}

	lastEntry := updatedEvidence.ChainOfCustody[len(updatedEvidence.ChainOfCustody)-1]
	if lastEntry.FromOfficer != "OFF-123" {
		t.Errorf("Expected from officer OFF-123, got %s", lastEntry.FromOfficer)
	}
	if lastEntry.ToOfficer != "DET-456" {
		t.Errorf("Expected to officer DET-456, got %s", lastEntry.ToOfficer)
	}

	// Test custody transfer with tampered evidence
	file, err := os.OpenFile(evidence.FilePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatalf("Failed to open evidence file: %v", err)
	}
	file.WriteString("TAMPERED")
	file.Close()

	err = system.TransferCustody(evidence.ID, "DET-456", "INV-789", "Investigation")
	if err == nil {
		t.Error("Expected transfer to fail with tampered evidence")
	}
}

func TestUpdateStatus(t *testing.T) {
	system, tmpDir, cleanup := setupTestSystem(t)
	defer cleanup()

	testFile := createTestFile(t, tmpDir)

	evidence, err := system.IngestEvidence(
		testFile,
		"CASE-TEST-004",
		"OFF-123",
		"Officer Test",
		"Test Location",
		[]string{"test"},
	)

	if err != nil {
		t.Fatalf("IngestEvidence failed: %v", err)
	}

	// Test status update
	err = system.UpdateStatus(evidence.ID, "OFF-123", StatusAnalyzed, "Analysis complete")
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	// Verify status was updated
	updatedEvidence, _ := system.GetEvidence(evidence.ID)
	if updatedEvidence.Status != StatusAnalyzed {
		t.Errorf("Expected status %s, got %s", StatusAnalyzed, updatedEvidence.Status)
	}

	if updatedEvidence.Notes != "Analysis complete" {
		t.Errorf("Expected notes 'Analysis complete', got %s", updatedEvidence.Notes)
	}

	// Test updating non-existent evidence
	err = system.UpdateStatus("INVALID-ID", "OFF-123", StatusArchived, "Test")
	if err == nil {
		t.Error("Expected error when updating non-existent evidence")
	}
}

func TestSearchEvidence(t *testing.T) {
	system, tmpDir, cleanup := setupTestSystem(t)
	defer cleanup()

	testFile := createTestFile(t, tmpDir)

	// Ingest multiple evidence items
	evidence1, _ := system.IngestEvidence(testFile, "CASE-001", "OFF-123", "Officer A", "Location A", []string{"tag1"})
	evidence2, _ := system.IngestEvidence(testFile, "CASE-001", "OFF-456", "Officer B", "Location B", []string{"tag2"})
	evidence3, _ := system.IngestEvidence(testFile, "CASE-002", "OFF-123", "Officer A", "Location C", []string{"tag3"})

	// Update status for one evidence
	system.UpdateStatus(evidence2.ID, "OFF-456", StatusAnalyzed, "Done")

	// Test search by case number
	results := system.SearchEvidence("CASE-001", "", "")
	if len(results) != 2 {
		t.Errorf("Expected 2 results for CASE-001, got %d", len(results))
	}

	// Test search by officer ID
	results = system.SearchEvidence("", "OFF-123", "")
	if len(results) != 2 {
		t.Errorf("Expected 2 results for OFF-123, got %d", len(results))
	}

	// Test search by status
	results = system.SearchEvidence("", "", StatusAnalyzed)
	if len(results) != 1 {
		t.Errorf("Expected 1 result for status ANALYZED, got %d", len(results))
	}

	// Test combined search
	results = system.SearchEvidence("CASE-001", "OFF-123", StatusCollected)
	if len(results) != 1 {
		t.Errorf("Expected 1 result for combined search, got %d", len(results))
	}

	// Test search with no matches
	results = system.SearchEvidence("CASE-999", "", "")
	if len(results) != 0 {
		t.Errorf("Expected 0 results for non-existent case, got %d", len(results))
	}

	// Suppress unused variable warnings
	_ = evidence1
	_ = evidence3
}

func TestGetEvidence(t *testing.T) {
	system, tmpDir, cleanup := setupTestSystem(t)
	defer cleanup()

	testFile := createTestFile(t, tmpDir)

	evidence, err := system.IngestEvidence(
		testFile,
		"CASE-TEST-005",
		"OFF-123",
		"Officer Test",
		"Test Location",
		[]string{"test"},
	)

	if err != nil {
		t.Fatalf("IngestEvidence failed: %v", err)
	}

	// Test getting existing evidence
	retrieved, err := system.GetEvidence(evidence.ID)
	if err != nil {
		t.Fatalf("GetEvidence failed: %v", err)
	}

	if retrieved.ID != evidence.ID {
		t.Errorf("Expected ID %s, got %s", evidence.ID, retrieved.ID)
	}

	// Test getting non-existent evidence
	_, err = system.GetEvidence("INVALID-ID")
	if err == nil {
		t.Error("Expected error when getting non-existent evidence")
	}
}

func TestGetChainOfCustody(t *testing.T) {
	system, tmpDir, cleanup := setupTestSystem(t)
	defer cleanup()

	testFile := createTestFile(t, tmpDir)

	evidence, _ := system.IngestEvidence(
		testFile,
		"CASE-TEST-006",
		"OFF-123",
		"Officer Test",
		"Test Location",
		[]string{"test"},
	)

	system.TransferCustody(evidence.ID, "OFF-123", "DET-456", "Analysis")
	system.TransferCustody(evidence.ID, "DET-456", "INV-789", "Investigation")

	custody, err := system.GetChainOfCustody(evidence.ID)
	if err != nil {
		t.Fatalf("GetChainOfCustody failed: %v", err)
	}

	if len(custody) != 3 {
		t.Errorf("Expected 3 custody entries, got %d", len(custody))
	}

	// Verify custody chain integrity
	expectedOfficers := []string{"SYSTEM", "OFF-123", "DET-456"}
	for i, entry := range custody {
		if entry.FromOfficer != expectedOfficers[i] {
			t.Errorf("Entry %d: expected from %s, got %s", i, expectedOfficers[i], entry.FromOfficer)
		}
	}
}

func TestAuditLogs(t *testing.T) {
	system, tmpDir, cleanup := setupTestSystem(t)
	defer cleanup()

	testFile := createTestFile(t, tmpDir)

	evidence, _ := system.IngestEvidence(
		testFile,
		"CASE-TEST-007",
		"OFF-123",
		"Officer Test",
		"Test Location",
		[]string{"test"},
	)

	system.VerifyIntegrity(evidence.ID, "OFF-123")
	system.UpdateStatus(evidence.ID, "OFF-123", StatusAnalyzed, "Done")

	// Get all audit logs for this evidence
	logs := system.GetAuditLogs(evidence.ID, "")
	if len(logs) < 3 {
		t.Errorf("Expected at least 3 audit logs, got %d", len(logs))
	}

	// Verify audit log content
	foundIngest := false
	for _, log := range logs {
		if log.Action == "INGEST_EVIDENCE" {
			foundIngest = true
			if log.UserID != "OFF-123" {
				t.Errorf("Expected user ID OFF-123, got %s", log.UserID)
			}
		}
	}

	if !foundIngest {
		t.Error("INGEST_EVIDENCE action not found in audit logs")
	}

	// Get audit logs for specific user
	userLogs := system.GetAuditLogs("", "OFF-123")
	if len(userLogs) != len(logs) {
		t.Errorf("Expected %d user logs, got %d", len(logs), len(userLogs))
	}
}

func TestGenerateReport(t *testing.T) {
	system, tmpDir, cleanup := setupTestSystem(t)
	defer cleanup()

	testFile := createTestFile(t, tmpDir)

	// Ingest evidence for the same case
	system.IngestEvidence(testFile, "CASE-REPORT", "OFF-123", "Officer A", "Location A", []string{"tag1"})
	system.IngestEvidence(testFile, "CASE-REPORT", "OFF-456", "Officer B", "Location B", []string{"tag2"})

	// Generate report
	report, err := system.GenerateReport("CASE-REPORT")
	if err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	if len(report) == 0 {
		t.Error("Generated report is empty")
	}

	// Verify report contains expected information
	if !contains(report, "CASE-REPORT") {
		t.Error("Report doesn't contain case number")
	}

	if !contains(report, "Total Evidence Items: 2") {
		t.Error("Report doesn't contain correct evidence count")
	}

	// Test report for non-existent case
	_, err = system.GenerateReport("CASE-NONEXISTENT")
	if err == nil {
		t.Error("Expected error when generating report for non-existent case")
	}
}

func TestExportEvidence(t *testing.T) {
	system, tmpDir, cleanup := setupTestSystem(t)
	defer cleanup()

	testFile := createTestFile(t, tmpDir)

	evidence, _ := system.IngestEvidence(
		testFile,
		"CASE-TEST-008",
		"OFF-123",
		"Officer Test",
		"Test Location",
		[]string{"test"},
	)

	exportPath := filepath.Join(tmpDir, "export.json")
	err := system.ExportEvidence(evidence.ID, exportPath)
	if err != nil {
		t.Fatalf("ExportEvidence failed: %v", err)
	}

	// Verify export file exists
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Error("Export file was not created")
	}

	// Verify export file contains valid JSON
	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	if len(data) == 0 {
		t.Error("Export file is empty")
	}

	// Test exporting non-existent evidence
	err = system.ExportEvidence("INVALID-ID", filepath.Join(tmpDir, "invalid.json"))
	if err == nil {
		t.Error("Expected error when exporting non-existent evidence")
	}
}

func TestConcurrentOperations(t *testing.T) {
	system, tmpDir, cleanup := setupTestSystem(t)
	defer cleanup()

	testFile := createTestFile(t, tmpDir)

	// Ingest initial evidence
	evidence, _ := system.IngestEvidence(
		testFile,
		"CASE-CONCURRENT",
		"OFF-123",
		"Officer Test",
		"Test Location",
		[]string{"test"},
	)

	// Perform concurrent operations
	done := make(chan bool)
	iterations := 10

	// Concurrent integrity checks
	for i := 0; i < iterations; i++ {
		go func(id int) {
			system.VerifyIntegrity(evidence.ID, fmt.Sprintf("OFF-%d", id))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < iterations; i++ {
		<-done
	}

	// Verify all integrity checks were recorded
	updatedEvidence, _ := system.GetEvidence(evidence.ID)
	expectedChecks := 1 + iterations // Initial + concurrent checks
	if len(updatedEvidence.IntegrityChecks) != expectedChecks {
		t.Errorf("Expected %d integrity checks, got %d", expectedChecks, len(updatedEvidence.IntegrityChecks))
	}
}

func TestFileHashCalculation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bwc_hash_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "hash_test.txt")
	content := []byte("test content for hash calculation")
	
	if err := os.WriteFile(testFile, content, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hash1, err := calculateFileHash(testFile)
	if err != nil {
		t.Fatalf("calculateFileHash failed: %v", err)
	}

	if len(hash1) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash1))
	}

	// Verify hash consistency
	hash2, _ := calculateFileHash(testFile)
	if hash1 != hash2 {
		t.Error("Hash calculation is not consistent")
	}

	// Test with non-existent file
	_, err = calculateFileHash("/nonexistent/file.txt")
	if err == nil {
		t.Error("Expected error when calculating hash for non-existent file")
	}
}

func TestEvidenceIDGeneration(t *testing.T) {
	id1 := generateEvidenceID("CASE-001", "OFF-123")
	id2 := generateEvidenceID("CASE-001", "OFF-123")

	// IDs should be unique even for same inputs (due to timestamp)
	time.Sleep(time.Millisecond * 10)
	id3 := generateEvidenceID("CASE-001", "OFF-123")

	if id1 == id3 {
		t.Error("Evidence IDs should be unique")
	}

	// Verify ID format
	if !contains(id1, "BWC-") || !contains(id1, "CASE-001") || !contains(id1, "OFF-123") {
		t.Errorf("Evidence ID format incorrect: %s", id1)
	}

	// Suppress unused variable warning
	_ = id2
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestMain(m *testing.M) {
	// Setup
	fmt.Println("Running BWC System Tests...")
	
	// Run tests
	code := m.Run()
	
	// Cleanup
	fmt.Println("Tests completed.")
	
	os.Exit(code)
}
