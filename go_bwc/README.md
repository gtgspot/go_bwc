# Forensic Body-Worn Camera (BWC) Management System

## Overview

A comprehensive forensic evidence management system designed for law enforcement body-worn camera footage. This system ensures evidence integrity, maintains chain of custody, and provides complete audit trails for legal proceedings.

## Core Features

### 1. Evidence Integrity Management
- **SHA-256 Hash Verification**: Each file receives a cryptographic hash upon ingestion
- **Automated Integrity Checks**: Verify evidence hasn't been tampered with
- **Tamper Detection**: Alerts when file hash mismatches occur
- **Historical Tracking**: Complete record of all integrity verification attempts

### 2. Chain of Custody
- **Complete Custody Trail**: Every evidence transfer is documented
- **Verification at Transfer**: Integrity checked before custody changes
- **Purpose Documentation**: Reason for each transfer recorded
- **Timestamp Precision**: RFC3339 formatted timestamps for legal compliance

### 3. Evidence Lifecycle Management
- **Status Tracking**: COLLECTED → PROCESSING → ANALYZED → ARCHIVED → DELETED
- **Metadata Rich**: Case numbers, officer details, location, tags
- **Search Capabilities**: Multi-criteria evidence retrieval
- **Secure Storage**: Files copied to protected storage with restricted permissions

### 4. Comprehensive Audit Logging
- **Complete Activity Trail**: Every system action logged
- **User Attribution**: Links actions to specific officers/users
- **Evidence Tracking**: Associates logs with specific evidence
- **Compliance Ready**: Supports legal discovery and compliance audits

### 5. Reporting Capabilities
- **Case Reports**: Generate comprehensive reports by case number
- **JSON Export**: Export evidence records for external systems
- **Chain of Custody Reports**: Complete custody documentation
- **Audit Trail Export**: Full activity history retrieval

## System Architecture

```
BWCSystem
├── Evidence Database (Thread-safe map)
├── Audit Log System
├── Secure Storage
└── Integrity Verification Engine
```

### Key Components

**Evidence Structure**:
- Unique identification
- Case association
- Officer attribution
- File integrity (hash, size)
- Status tracking
- Chain of custody
- Integrity check history
- Temporal tracking

**Security Features**:
- Thread-safe operations (mutex locks)
- Secure file storage (0700 permissions)
- Cryptographic hashing (SHA-256)
- Immutable audit logs
- Access control ready

## Usage Examples

### Initialize System
```go
system, err := NewBWCSystem("./secure_storage")
if err != nil {
    log.Fatal(err)
}
```

### Ingest Evidence
```go
evidence, err := system.IngestEvidence(
    "/path/to/video.mp4",
    "CASE-2025-001",
    "OFF-12345",
    "Officer John Smith",
    "123 Main St, City",
    []string{"traffic-stop", "incident"},
)
```

### Verify Integrity
```go
isValid, err := system.VerifyIntegrity(evidenceID, "OFF-12345")
if !isValid {
    // Handle potential tampering
}
```

### Transfer Custody
```go
err := system.TransferCustody(
    evidenceID,
    "OFF-12345",
    "DET-67890",
    "Evidence analysis and processing",
)
```

### Update Status
```go
err := system.UpdateStatus(
    evidenceID,
    "DET-67890",
    StatusAnalyzed,
    "Analysis completed - no anomalies found",
)
```

### Search Evidence
```go
results := system.SearchEvidence(
    "CASE-2025-001",  // Case number
    "",               // Officer ID (empty = all)
    StatusCollected,  // Status filter
)
```

### Generate Report
```go
report, err := system.GenerateReport("CASE-2025-001")
fmt.Println(report)
```

### Get Audit Logs
```go
logs := system.GetAuditLogs(evidenceID, "")
for _, log := range logs {
    fmt.Printf("%s: %s by %s\n", log.Timestamp, log.Action, log.UserID)
}
```

## Evidence Status Flow

```
COLLECTED → PROCESSING → ANALYZED → ARCHIVED
                                   ↓
                                DELETED
```

## Chain of Custody Actions

- **INGESTED**: Initial evidence collection
- **TRANSFERRED**: Custody change between officers
- **VERIFIED**: Integrity check performed
- **ACCESSED**: Evidence file accessed
- **EXPORTED**: Evidence data exported

## Audit Actions

- `INGEST_EVIDENCE`: Evidence added to system
- `VERIFY_INTEGRITY`: Integrity check performed
- `TRANSFER_CUSTODY`: Custody transferred
- `UPDATE_STATUS`: Evidence status changed
- `ACCESS_EVIDENCE`: Evidence accessed
- `EXPORT_EVIDENCE`: Evidence exported

## Security Considerations

### File Integrity
- SHA-256 cryptographic hashing
- Hash verification before custody transfers
- Automated tamper detection
- Historical integrity tracking

### Access Control
- User/officer attribution on all actions
- Complete audit trail
- Secure file storage (0700 permissions)
- Thread-safe concurrent operations

### Chain of Custody
- Immutable custody records
- Verification at each transfer
- Purpose documentation
- Timestamp accuracy

### Data Protection
- Secure storage location
- File permission restrictions
- JSON export for backup
- Audit log preservation

## Running the Demo

```bash
go run forensic_bwc_system.go
```

The demo will:
1. Initialize the system
2. Ingest test evidence
3. Verify integrity
4. Transfer custody
5. Update status
6. Display chain of custody
7. Generate case report
8. Show audit logs
9. Export evidence record

## Production Deployment Considerations

### Database Integration
Replace in-memory maps with:
- PostgreSQL for evidence records
- Append-only audit log tables
- Indexed searches for performance

### Storage Backend
- Network-attached storage (NAS)
- Cloud storage (AWS S3, Azure Blob)
- Redundant storage (RAID)
- Automated backups

### Authentication & Authorization
- LDAP/Active Directory integration
- Role-based access control (RBAC)
- Multi-factor authentication (MFA)
- API key management

### Scalability
- Horizontal scaling with load balancers
- Microservices architecture
- Message queues for async processing
- Caching layer (Redis)

### Monitoring
- Evidence integrity monitoring
- System health checks
- Alert system for tampering
- Performance metrics

### Compliance
- CJIS compliance (Criminal Justice Information Services)
- GDPR data protection
- Evidence retention policies
- Legal hold capabilities

## API Extensions

Consider adding:
- RESTful API endpoints
- GraphQL interface
- Webhook notifications
- Real-time integrity monitoring
- Automated video transcription
- Facial recognition integration
- Geolocation tracking
- Mobile app support

## Testing

Recommended test coverage:
- Unit tests for all functions
- Integration tests for workflows
- Concurrency tests (race conditions)
- Integrity verification tests
- Chain of custody validation
- Audit log completeness

## License

This is a demonstration system for educational purposes. Consult legal and compliance teams before production deployment in law enforcement contexts.

## Future Enhancements

1. **Video Processing**
   - Automatic redaction
   - Video analytics
   - Thumbnail generation

2. **Advanced Search**
   - Full-text search
   - Geospatial queries
   - Time-range filtering

3. **Integration**
   - CAD system integration
   - Records management systems
   - Court filing systems

4. **Machine Learning**
   - Object detection
   - Activity recognition
   - Anomaly detection

5. **Blockchain**
   - Immutable chain of custody
   - Distributed ledger for integrity

## Support

For production implementation, consider:
- Legal compliance review
- Security audit
- Performance testing
- User training
- Documentation customization
