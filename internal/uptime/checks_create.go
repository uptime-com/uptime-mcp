package uptime

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
)

// Common fields shared across check types
type checkCommon struct {
	Name        string   `json:"name" jsonschema:"description=Check name"`
	Address     string   `json:"address" jsonschema:"description=URL or hostname to monitor"`
	Interval    int      `json:"interval,omitempty" jsonschema:"description=Check interval in seconds (default 60)"`
	Locations   []string `json:"locations,omitempty" jsonschema:"description=Monitoring locations (e.g. US-NY, EU-LON)"`
	Tags        []string `json:"tags,omitempty" jsonschema:"description=Tags to assign to the check"`
	Sensitivity int      `json:"sensitivity,omitempty" jsonschema:"description=Number of locations that must detect failure (default 2)"`
	Notes       string   `json:"notes,omitempty" jsonschema:"description=Notes about this check"`
}

// CreateHTTPCheckInput defines parameters for creating an HTTP check.
type CreateHTTPCheckInput struct {
	checkCommon
	Port         int    `json:"port,omitempty" jsonschema:"description=Port number (default 80 for HTTP, 443 for HTTPS)"`
	Username     string `json:"username,omitempty" jsonschema:"description=Basic auth username"`
	Password     string `json:"password,omitempty" jsonschema:"description=Basic auth password"`
	Headers      string `json:"headers,omitempty" jsonschema:"description=Custom HTTP headers (one per line)"`
	SendString   string `json:"send_string,omitempty" jsonschema:"description=String to send in request body"`
	ExpectString string `json:"expect_string,omitempty" jsonschema:"description=String that must appear in response"`
}

var createHTTPCheckTool = &mcp.Tool{
	Name:        "create_http_check",
	Description: "Create a new HTTP/HTTPS monitoring check",
}

func (p *Provider) handleCreateHTTPCheck(ctx context.Context, _ *mcp.ServerSession, req *mcp.CallToolParamsFor[CreateHTTPCheckInput]) (*mcp.CallToolResultFor[any], error) {
	client, err := getClient(ctx)
	if err != nil {
		return errorResult(err), nil
	}

	in := req.Arguments
	if in.Name == "" || in.Address == "" {
		return errorResult(fmt.Errorf("name and address are required")), nil
	}

	check := &api.Check{
		CheckType:    "HTTP",
		Name:         in.Name,
		Address:      in.Address,
		Port:         in.Port,
		Interval:     in.Interval,
		Locations:    in.Locations,
		Tags:         in.Tags,
		Sensitivity:  in.Sensitivity,
		Notes:        in.Notes,
		Username:     in.Username,
		Password:     in.Password,
		Headers:      in.Headers,
		SendString:   in.SendString,
		ExpectString: in.ExpectString,
	}

	created, _, err := client.Checks.Create(ctx, check)
	if err != nil {
		return errorResult(fmt.Errorf("failed to create HTTP check: %w", err)), nil
	}

	return textResult(fmt.Sprintf("Created HTTP check #%d: %s", created.PK, created.Name)), nil
}

// CreateDNSCheckInput defines parameters for creating a DNS check.
type CreateDNSCheckInput struct {
	checkCommon
	DNSServer     string `json:"dns_server,omitempty" jsonschema:"description=DNS server to query (default: authoritative)"`
	DNSRecordType string `json:"dns_record_type,omitempty" jsonschema:"description=DNS record type: A, AAAA, CNAME, MX, NS, TXT (default A)"`
	ExpectString  string `json:"expect_string,omitempty" jsonschema:"description=Expected value in DNS response"`
}

var createDNSCheckTool = &mcp.Tool{
	Name:        "create_dns_check",
	Description: "Create a new DNS monitoring check",
}

func (p *Provider) handleCreateDNSCheck(ctx context.Context, _ *mcp.ServerSession, req *mcp.CallToolParamsFor[CreateDNSCheckInput]) (*mcp.CallToolResultFor[any], error) {
	client, err := getClient(ctx)
	if err != nil {
		return errorResult(err), nil
	}

	in := req.Arguments
	if in.Name == "" || in.Address == "" {
		return errorResult(fmt.Errorf("name and address are required")), nil
	}

	check := &api.Check{
		CheckType:     "DNS",
		Name:          in.Name,
		Address:       in.Address,
		Interval:      in.Interval,
		Locations:     in.Locations,
		Tags:          in.Tags,
		Sensitivity:   in.Sensitivity,
		Notes:         in.Notes,
		DNSServer:     in.DNSServer,
		DNSRecordType: in.DNSRecordType,
		ExpectString:  in.ExpectString,
	}

	created, _, err := client.Checks.Create(ctx, check)
	if err != nil {
		return errorResult(fmt.Errorf("failed to create DNS check: %w", err)), nil
	}

	return textResult(fmt.Sprintf("Created DNS check #%d: %s", created.PK, created.Name)), nil
}

// CreateSSLCheckInput defines parameters for creating an SSL certificate check.
type CreateSSLCheckInput struct {
	checkCommon
	Port     int    `json:"port,omitempty" jsonschema:"description=Port number (default 443)"`
	Protocol string `json:"protocol,omitempty" jsonschema:"description=Protocol: HTTPS, SMTPS, IMAPS, POP3S (default HTTPS)"`
}

var createSSLCheckTool = &mcp.Tool{
	Name:        "create_ssl_check",
	Description: "Create a new SSL certificate monitoring check",
}

func (p *Provider) handleCreateSSLCheck(ctx context.Context, _ *mcp.ServerSession, req *mcp.CallToolParamsFor[CreateSSLCheckInput]) (*mcp.CallToolResultFor[any], error) {
	client, err := getClient(ctx)
	if err != nil {
		return errorResult(err), nil
	}

	in := req.Arguments
	if in.Name == "" || in.Address == "" {
		return errorResult(fmt.Errorf("name and address are required")), nil
	}

	check := &api.Check{
		CheckType:   "SSL",
		Name:        in.Name,
		Address:     in.Address,
		Port:        in.Port,
		Interval:    in.Interval,
		Locations:   in.Locations,
		Tags:        in.Tags,
		Sensitivity: in.Sensitivity,
		Notes:       in.Notes,
		Protocol:    in.Protocol,
	}

	created, _, err := client.Checks.Create(ctx, check)
	if err != nil {
		return errorResult(fmt.Errorf("failed to create SSL check: %w", err)), nil
	}

	return textResult(fmt.Sprintf("Created SSL check #%d: %s", created.PK, created.Name)), nil
}

// CreateTCPCheckInput defines parameters for creating a TCP port check.
type CreateTCPCheckInput struct {
	checkCommon
	Port         int    `json:"port" jsonschema:"description=Port number to check"`
	SendString   string `json:"send_string,omitempty" jsonschema:"description=String to send after connection"`
	ExpectString string `json:"expect_string,omitempty" jsonschema:"description=String expected in response"`
}

var createTCPCheckTool = &mcp.Tool{
	Name:        "create_tcp_check",
	Description: "Create a new TCP port connectivity check",
}

func (p *Provider) handleCreateTCPCheck(ctx context.Context, _ *mcp.ServerSession, req *mcp.CallToolParamsFor[CreateTCPCheckInput]) (*mcp.CallToolResultFor[any], error) {
	client, err := getClient(ctx)
	if err != nil {
		return errorResult(err), nil
	}

	in := req.Arguments
	if in.Name == "" || in.Address == "" {
		return errorResult(fmt.Errorf("name and address are required")), nil
	}
	if in.Port == 0 {
		return errorResult(fmt.Errorf("port is required for TCP check")), nil
	}

	check := &api.Check{
		CheckType:    "TCP",
		Name:         in.Name,
		Address:      in.Address,
		Port:         in.Port,
		Interval:     in.Interval,
		Locations:    in.Locations,
		Tags:         in.Tags,
		Sensitivity:  in.Sensitivity,
		Notes:        in.Notes,
		SendString:   in.SendString,
		ExpectString: in.ExpectString,
	}

	created, _, err := client.Checks.Create(ctx, check)
	if err != nil {
		return errorResult(fmt.Errorf("failed to create TCP check: %w", err)), nil
	}

	return textResult(fmt.Sprintf("Created TCP check #%d: %s", created.PK, created.Name)), nil
}

// CreateICMPCheckInput defines parameters for creating an ICMP/Ping check.
type CreateICMPCheckInput struct {
	checkCommon
}

var createICMPCheckTool = &mcp.Tool{
	Name:        "create_icmp_check",
	Description: "Create a new ICMP/Ping monitoring check",
}

func (p *Provider) handleCreateICMPCheck(ctx context.Context, _ *mcp.ServerSession, req *mcp.CallToolParamsFor[CreateICMPCheckInput]) (*mcp.CallToolResultFor[any], error) {
	client, err := getClient(ctx)
	if err != nil {
		return errorResult(err), nil
	}

	in := req.Arguments
	if in.Name == "" || in.Address == "" {
		return errorResult(fmt.Errorf("name and address are required")), nil
	}

	check := &api.Check{
		CheckType:   "ICMP",
		Name:        in.Name,
		Address:     in.Address,
		Interval:    in.Interval,
		Locations:   in.Locations,
		Tags:        in.Tags,
		Sensitivity: in.Sensitivity,
		Notes:       in.Notes,
	}

	created, _, err := client.Checks.Create(ctx, check)
	if err != nil {
		return errorResult(fmt.Errorf("failed to create ICMP check: %w", err)), nil
	}

	return textResult(fmt.Sprintf("Created ICMP check #%d: %s", created.PK, created.Name)), nil
}

// CreateSMTPCheckInput defines parameters for creating an SMTP check.
type CreateSMTPCheckInput struct {
	checkCommon
	Port       int    `json:"port,omitempty" jsonschema:"description=Port number (default 25, or 465/587 for encrypted)"`
	Encryption string `json:"encryption,omitempty" jsonschema:"description=Encryption: none, SSL, STARTTLS"`
	Username   string `json:"username,omitempty" jsonschema:"description=SMTP auth username"`
	Password   string `json:"password,omitempty" jsonschema:"description=SMTP auth password"`
}

var createSMTPCheckTool = &mcp.Tool{
	Name:        "create_smtp_check",
	Description: "Create a new SMTP email server monitoring check",
}

func (p *Provider) handleCreateSMTPCheck(ctx context.Context, _ *mcp.ServerSession, req *mcp.CallToolParamsFor[CreateSMTPCheckInput]) (*mcp.CallToolResultFor[any], error) {
	client, err := getClient(ctx)
	if err != nil {
		return errorResult(err), nil
	}

	in := req.Arguments
	if in.Name == "" || in.Address == "" {
		return errorResult(fmt.Errorf("name and address are required")), nil
	}

	check := &api.Check{
		CheckType:   "SMTP",
		Name:        in.Name,
		Address:     in.Address,
		Port:        in.Port,
		Interval:    in.Interval,
		Locations:   in.Locations,
		Tags:        in.Tags,
		Sensitivity: in.Sensitivity,
		Notes:       in.Notes,
		Encryption:  in.Encryption,
		Username:    in.Username,
		Password:    in.Password,
	}

	created, _, err := client.Checks.Create(ctx, check)
	if err != nil {
		return errorResult(fmt.Errorf("failed to create SMTP check: %w", err)), nil
	}

	return textResult(fmt.Sprintf("Created SMTP check #%d: %s", created.PK, created.Name)), nil
}

// CreateIMAPCheckInput defines parameters for creating an IMAP check.
type CreateIMAPCheckInput struct {
	checkCommon
	Port       int    `json:"port,omitempty" jsonschema:"description=Port number (default 143, or 993 for SSL)"`
	Encryption string `json:"encryption,omitempty" jsonschema:"description=Encryption: none, SSL, STARTTLS"`
	Username   string `json:"username,omitempty" jsonschema:"description=IMAP username"`
	Password   string `json:"password,omitempty" jsonschema:"description=IMAP password"`
}

var createIMAPCheckTool = &mcp.Tool{
	Name:        "create_imap_check",
	Description: "Create a new IMAP email server monitoring check",
}

func (p *Provider) handleCreateIMAPCheck(ctx context.Context, _ *mcp.ServerSession, req *mcp.CallToolParamsFor[CreateIMAPCheckInput]) (*mcp.CallToolResultFor[any], error) {
	client, err := getClient(ctx)
	if err != nil {
		return errorResult(err), nil
	}

	in := req.Arguments
	if in.Name == "" || in.Address == "" {
		return errorResult(fmt.Errorf("name and address are required")), nil
	}

	check := &api.Check{
		CheckType:   "IMAP",
		Name:        in.Name,
		Address:     in.Address,
		Port:        in.Port,
		Interval:    in.Interval,
		Locations:   in.Locations,
		Tags:        in.Tags,
		Sensitivity: in.Sensitivity,
		Notes:       in.Notes,
		Encryption:  in.Encryption,
		Username:    in.Username,
		Password:    in.Password,
	}

	created, _, err := client.Checks.Create(ctx, check)
	if err != nil {
		return errorResult(fmt.Errorf("failed to create IMAP check: %w", err)), nil
	}

	return textResult(fmt.Sprintf("Created IMAP check #%d: %s", created.PK, created.Name)), nil
}

// CreatePOPCheckInput defines parameters for creating a POP3 check.
type CreatePOPCheckInput struct {
	checkCommon
	Port       int    `json:"port,omitempty" jsonschema:"description=Port number (default 110, or 995 for SSL)"`
	Encryption string `json:"encryption,omitempty" jsonschema:"description=Encryption: none, SSL, STARTTLS"`
	Username   string `json:"username,omitempty" jsonschema:"description=POP3 username"`
	Password   string `json:"password,omitempty" jsonschema:"description=POP3 password"`
}

var createPOPCheckTool = &mcp.Tool{
	Name:        "create_pop_check",
	Description: "Create a new POP3 email server monitoring check",
}

func (p *Provider) handleCreatePOPCheck(ctx context.Context, _ *mcp.ServerSession, req *mcp.CallToolParamsFor[CreatePOPCheckInput]) (*mcp.CallToolResultFor[any], error) {
	client, err := getClient(ctx)
	if err != nil {
		return errorResult(err), nil
	}

	in := req.Arguments
	if in.Name == "" || in.Address == "" {
		return errorResult(fmt.Errorf("name and address are required")), nil
	}

	check := &api.Check{
		CheckType:   "POP",
		Name:        in.Name,
		Address:     in.Address,
		Port:        in.Port,
		Interval:    in.Interval,
		Locations:   in.Locations,
		Tags:        in.Tags,
		Sensitivity: in.Sensitivity,
		Notes:       in.Notes,
		Encryption:  in.Encryption,
		Username:    in.Username,
		Password:    in.Password,
	}

	created, _, err := client.Checks.Create(ctx, check)
	if err != nil {
		return errorResult(fmt.Errorf("failed to create POP check: %w", err)), nil
	}

	return textResult(fmt.Sprintf("Created POP check #%d: %s", created.PK, created.Name)), nil
}
