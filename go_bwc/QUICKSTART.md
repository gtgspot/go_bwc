# Quick Start Guide - Forensic BWC System

## Prerequisites

- Go 1.21 or higher
- Basic understanding of Go programming
- Access to body-worn camera video files

## Installation

### Option 1: From Source

```bash
# Clone or download the repository
cd forensic_bwc_system

# Run the application
go run forensic_bwc_system.go
```

### Option 2: Build Binary

```bash
# Build the binary
make build

# Run the binary
./bwc-system
```

### Option 3: Docker

```bash
# Build Docker image
make docker-build

# Run container
make docker-run
```

## Quick Usage Examples

### 1. Initialize the System

```go
system, err := NewBWCSystem("./bwc_storage")
if err != nil {
    log.Fatal(err)
}
```

### 2. Ingest Evidence

```go
evidence, err := system.IngestEvidence(
    "/path/to/bodycam_video.mp4",
    "CASE-2025-001",           // Case number
    "OFF-12345",               // Officer ID
    "Officer John Smith",       // Officer name
    "123 Main St, City",       // Location
    []string{"traffic", "dui"}, // Tags
)
```

### 3. Verify Integrity

```go
isValid, err := system.VerifyIntegrity(evidence.ID, "OFF-12345")
if !isValid {
    log.Println("WARNING: Evidence integrity compromised!")
}
```

### 4. Transfer Custody

```go
err = system.TransferCustody(
    evidence.ID,
    "OFF-12345",              // From officer
    "DET-67890",              // To officer
    "Evidence analysis",      // Purpose
)
```

### 5. Update Status

```go
err = system.UpdateStatus(
    evidence.ID,
    "DET-67890",
    StatusAnalyzed,
    "Analysis completed - no anomalies found",
)
```

### 6. Search Evidence

```go
// Search by case number
results := system.SearchEvidence("CASE-2025-001", "", "")

// Search by officer
results = system.SearchEvidence("", "OFF-12345", "")

// Search by status
results = system.SearchEvidence("", "", StatusCollected)
```

### 7. Generate Report

```go
report, err := system.GenerateReport("CASE-2025-001")
if err != nil {
    log.Fatal(err)
}
fmt.Println(report)
```

### 8. Export Evidence

```go
err = system.ExportEvidence(evidence.ID, "./exports/evidence.json")
if err != nil {
    log.Fatal(err)
}
```

### 9. View Chain of Custody

```go
custody, err := system.GetChainOfCustody(evidence.ID)
if err != nil {
    log.Fatal(err)
}

for _, entry := range custody {
    fmt.Printf("%s: %s -> %s (%s)\n",
        entry.Timestamp,
        entry.FromOfficer,
        entry.ToOfficer,
        entry.Action)
}
```

### 10. Get Audit Logs

```go
// Get all logs for specific evidence
logs := system.GetAuditLogs(evidence.ID, "")

// Get all logs for specific user
logs = system.GetAuditLogs("", "OFF-12345")

for _, log := range logs {
    fmt.Printf("%s: %s by %s - %s\n",
        log.Timestamp,
        log.Action,
        log.UserID,
        log.Details)
}
```

## Common Workflows

### Complete Evidence Processing Workflow

```go
// 1. Ingest evidence
evidence, _ := system.IngestEvidence(
    videoPath,
    caseNum,
    officerID,
    officerName,
    location,
    tags,
)

// 2. Verify integrity immediately
isValid, _ := system.VerifyIntegrity(evidence.ID, officerID)

// 3. Transfer to detective for analysis
system.TransferCustody(
    evidence.ID,
    officerID,
    detectiveID,
    "Initial analysis",
)

// 4. Update status as processing begins
system.UpdateStatus(
    evidence.ID,
    detectiveID,
    StatusProcessing,
    "Beginning video analysis",
)

// 5. After analysis, update status
system.UpdateStatus(
    evidence.ID,
    detectiveID,
    StatusAnalyzed,
    "Analysis complete",
)

// 6. Generate case report
report, _ := system.GenerateReport(caseNum)

// 7. Export evidence record
system.ExportEvidence(evidence.ID, exportPath)
```

### Integrity Monitoring Workflow

```go
// Get all evidence
allEvidence := system.SearchEvidence("", "", "")

// Verify integrity of all items
for _, ev := range allEvidence {
    isValid, err := system.VerifyIntegrity(ev.ID, "SYSTEM")
    if err != nil {
        log.Printf("Error verifying %s: %v", ev.ID, err)
        continue
    }
    
    if !isValid {
        log.Printf("ALERT: Evidence %s integrity check failed!", ev.ID)
        // Trigger alert system
    }
}
```

## Testing

Run the test suite:

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run quick tests
make test-short
```

## Building for Production

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build for specific platforms
make build-linux
make build-windows
make build-mac
```

## Docker Deployment

```bash
# Build image
docker build -t bwc-system:1.0.0 .

# Run container with volume mounts
docker run -d \
    --name bwc-system \
    -v $(pwd)/bwc_storage:/app/bwc_storage \
    -v $(pwd)/evidence_data:/app/evidence_data:ro \
    bwc-system:1.0.0

# Using docker-compose
docker-compose up -d
```

## Configuration

Copy the example configuration:

```bash
cp config.example.json config.json
```

Edit `config.json` to customize:

- Storage paths
- Security settings
- Retention policies
- Compliance requirements
- Notification settings

## Troubleshooting

### Problem: File not found during ingestion

**Solution**: Verify the file path is correct and the file exists:
```bash
ls -la /path/to/video.mp4
```

### Problem: Permission denied on storage directory

**Solution**: Ensure the directory has proper permissions:
```bash
chmod 700 ./bwc_storage
```

### Problem: Integrity check fails unexpectedly

**Solution**: Check if the file has been modified or moved:
1. Verify file still exists at original location
2. Check file permissions
3. Review audit logs for any access

### Problem: Out of memory errors

**Solution**: For large video files:
1. Increase system memory allocation
2. Process files in batches
3. Enable compression in configuration

## Best Practices

1. **Always verify integrity** before and after custody transfers
2. **Document all actions** with detailed notes
3. **Regular integrity checks** on stored evidence
4. **Backup regularly** to secondary storage
5. **Review audit logs** periodically for suspicious activity
6. **Keep system updated** with latest security patches
7. **Limit access** to authorized personnel only
8. **Test disaster recovery** procedures regularly

## Security Reminders

- Use strong, unique passwords
- Enable two-factor authentication
- Regularly review access logs
- Keep systems patched and updated
- Use encrypted connections
- Follow data retention policies
- Implement least-privilege access
- Monitor for anomalies

## Getting Help

- Review the main README.md for detailed documentation
- Check the test files for usage examples
- Review the source code comments
- Contact your system administrator for support

## Next Steps

1. Run the demo application to understand the workflow
2. Review the test suite for comprehensive examples
3. Customize the configuration for your needs
4. Integrate with existing systems (CAD, RMS, etc.)
5. Train users on proper evidence handling procedures
6. Establish monitoring and alerting procedures
7. Document your organization's specific workflows

## Useful Make Commands

```bash
make help          # Show all available commands
make build         # Build the application
make test          # Run tests
make clean         # Clean build artifacts
make run           # Run the application
make fmt           # Format code
make vet           # Run go vet
make check         # Run all checks
```

## Sample Integration Code

```go
package main

import (
    "log"
    "time"
)

func main() {
    // Initialize system
    system, err := NewBWCSystem("./production_storage")
    if err != nil {
        log.Fatal(err)
    }

    // Set up scheduled integrity checks
    ticker := time.NewTicker(24 * time.Hour)
    go func() {
        for range ticker.C {
            verifyAllEvidence(system)
        }
    }()

    // Your application logic here
    log.Println("BWC System running...")
    select {} // Keep running
}

func verifyAllEvidence(system *BWCSystem) {
    evidence := system.SearchEvidence("", "", "")
    for _, ev := range evidence {
        isValid, _ := system.VerifyIntegrity(ev.ID, "SYSTEM")
        if !isValid {
            log.Printf("ALERT: Evidence %s failed integrity check", ev.ID)
            // Send alert notification
        }
    }
}
```

---

**For production deployment, always consult with legal and IT security teams to ensure compliance with regulations and organizational policies.**
