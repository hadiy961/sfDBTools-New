// Package connection menangani database connection dengan support untuk direct dan SSH tunnel.
//
// Package ini menyediakan abstraksi untuk koneksi database yang mendukung:
//   - Direct connection ke MySQL/MariaDB
//   - SSH tunnel (bastion host) untuk secured connection
//   - Connection testing dengan detailed diagnostics
//   - Pre-flight checks sebelum connection attempt
//
// # Main Functions
//
// ConnectWithProfile(cfg, profile, initialDB):
//   - Primary connection function
//   - Auto-detect direct atau SSH tunnel mode
//   - Setup SSH tunnel jika needed
//   - Return connected database client
//
// TestConnection(cfg, profile, initialDB):
//   - Test connection dengan step-by-step diagnostics
//   - Return detailed report (DNS, TCP, SSH, Auth, Version)
//   - Used untuk health checks dan validation
//
// ValidateConnectPreflight(profile):
//   - Pre-flight validation sebelum connect attempt
//   - Check required fields presence
//   - Validate SSH tunnel config consistency
//   - Fast fail untuk obvious errors
//
// # Connection Modes
//
// Direct Connection:
//   - Connect langsung ke database host:port
//   - Simple dan fast (no overhead)
//   - Require direct network access ke database
//
// SSH Tunnel Connection:
//   - Start SSH tunnel ke bastion host
//   - Forward remote database port ke local port
//   - Connect ke localhost:localPort
//   - Support password dan private key authentication
//   - Auto cleanup tunnel saat connection closed
//
// # SSH Tunnel Configuration
//
// Required fields (jika Enabled=true):
//   - Host: SSH bastion hostname/IP
//   - Port: SSH port (default 22)
//   - User: SSH username
//   - Auth: Password OR IdentityFile (tidak boleh keduanya kosong)
//
// Optional fields:
//   - LocalPort: Local port untuk forward (0 = auto assign)
//   - ServerAlive: Keep-alive interval (default 30s)
//   - ConnectTimeout: SSH connection timeout
//
// Tunnel lifecycle:
//   - Start: Saat ConnectWithProfile dipanggil
//   - Running: Background process dengan keep-alive
//   - Stop: Saat database client closed (via OnClose hook)
//
// # Connection Testing
//
// TestConnection menjalankan test step-by-step:
//
//  1. DNS Resolution: Resolve hostname ke IP address
//  2. TCP Connection: Test TCP socket ke host:port
//  3. SSH Tunnel: Start tunnel (jika enabled) dan test forwarding
//  4. Authentication: Login ke database dengan credentials
//  5. DB Version: Query version untuk verify connectivity
//
// Report format:
//
//	type ConnectionTestReport struct {
//		DNSResolution  StepResult  // Success/Failed dengan detail
//		TCPConnection  StepResult
//		SSHTunnel      StepResult
//		Authentication StepResult
//		DBVersion      string      // "MariaDB 10.11.8"
//		TotalLatency   time.Duration
//		Healthy        bool
//		Err            error       // Overall error jika failed
//	}
//
// # Error Diagnostics
//
// DescribeConnectError(cfg, err):
//   - Parse connection error
//   - Return user-friendly error info
//   - Provide actionable hints untuk troubleshooting
//
// Error categories:
//   - DNS errors: hostname not found, DNS timeout
//   - TCP errors: connection refused, timeout, network unreachable
//   - SSH errors: authentication failed, host key verification, tunnel failed
//   - Auth errors: access denied, invalid credentials
//   - DB errors: unknown database, query failed
//
// Error hints examples:
//   - "Check if database server is running"
//   - "Verify firewall rules allow connection to port 3306"
//   - "Confirm SSH credentials are correct"
//   - "Check if bastion host is accessible"
//
// # Pre-flight Validation
//
// ValidateConnectPreflight checks:
//   - Profile not nil
//   - Database host not empty
//   - Database port in valid range (1-65535)
//   - Database user not empty
//   - SSH tunnel config consistency (jika enabled)
//   - SSH auth method available (password atau key file)
//
// Fast fail untuk obvious errors (sebelum network calls):
//
//	if err := connection.ValidateConnectPreflight(profile); err != nil {
//		// Handle validation error (no network attempt made)
//	}
//
// # Performance Considerations
//
// Connection Timeout:
//   - Default: 10 seconds (dari config.ProfileConnectTimeout)
//   - Configurable via config
//   - Apply untuk DNS, TCP, dan authentication
//
// SSH Tunnel Overhead:
//   - Startup latency: ~200-500ms (SSH handshake)
//   - Query latency: +10-50ms (tunnel overhead)
//   - Keep-alive: 30 seconds (prevent idle timeout)
//
// Connection Pooling:
//   - Handled by database.Client
//   - Max connections: 10 (default)
//   - Max idle: 5
//   - Connection lifetime: unlimited (reuse)
//
// # Security
//
// Password handling:
//   - Never logged
//   - Encrypted in profile file
//   - Decrypted only saat connection time
//   - Not stored in memory after connection
//
// SSH private key:
//   - File path validated (existence check)
//   - Key passphrase support (via SSH agent)
//   - Secure key material handling (no logging)
//
// Known hosts:
//   - Optional verification (default disabled untuk automation)
//   - Configurable via SSH config
//   - Auto-accept mode untuk CI/CD
//
// # Usage Examples
//
// Direct connection:
//
//	profile := &domain.ProfileInfo{
//		DBInfo: domain.DBInfo{
//			Host: "10.0.0.5",
//			Port: 3306,
//			User: "admin",
//			Password: "secret",
//		},
//		SSHTunnel: domain.SSHTunnel{
//			Enabled: false,
//		},
//	}
//	client, err := connection.ConnectWithProfile(cfg, profile, "mydb")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
// SSH tunnel connection:
//
//	profile := &domain.ProfileInfo{
//		DBInfo: domain.DBInfo{
//			Host: "10.0.0.5",
//			Port: 3306,
//			User: "admin",
//			Password: "secret",
//		},
//		SSHTunnel: domain.SSHTunnel{
//			Enabled:      true,
//			Host:         "bastion.example.com",
//			Port:         22,
//			User:         "sshuser",
//			IdentityFile: "~/.ssh/id_rsa",
//			LocalPort:    0, // Auto assign
//		},
//	}
//	client, err := connection.ConnectWithProfile(cfg, profile, "mydb")
//	// Tunnel started automatically, cleanup on Close()
//
// Connection testing:
//
//	report := connection.TestConnection(cfg, profile, "information_schema")
//	if !report.Healthy {
//		fmt.Printf("Connection failed: %v\n", report.Err)
//		fmt.Printf("DNS: %s\n", report.DNSResolution.Display())
//		fmt.Printf("TCP: %s\n", report.TCPConnection.Display())
//	} else {
//		fmt.Printf("Connection healthy (latency: %v)\n", report.TotalLatency)
//	}
package connection
